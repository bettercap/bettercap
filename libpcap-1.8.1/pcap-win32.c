/*
 * Copyright (c) 1999 - 2005 NetGroup, Politecnico di Torino (Italy)
 * Copyright (c) 2005 - 2010 CACE Technologies, Davis (California)
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 *
 * 1. Redistributions of source code must retain the above copyright
 * notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 * notice, this list of conditions and the following disclaimer in the
 * documentation and/or other materials provided with the distribution.
 * 3. Neither the name of the Politecnico di Torino, CACE Technologies
 * nor the names of its contributors may be used to endorse or promote
 * products derived from this software without specific prior written
 * permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
 * "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
 * LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
 * A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
 * OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
 * SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
 * LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
 * DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
 * THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

#define PCAP_DONT_INCLUDE_PCAP_BPF_H
#include <Packet32.h>
#include <pcap-int.h>
#include <pcap/dlt.h>
#ifdef __MINGW32__
#ifdef __MINGW64__
#include <ntddndis.h>
#else  /*__MINGW64__*/
#include <ddk/ntddndis.h>
#include <ddk/ndis.h>
#endif /*__MINGW64__*/
#else /*__MINGW32__*/
#include <ntddndis.h>
#endif /*__MINGW32__*/
#ifdef HAVE_DAG_API
#include <dagnew.h>
#include <dagapi.h>
#endif /* HAVE_DAG_API */
#ifdef __MINGW32__
int* _errno();
#define errno (*_errno())
#endif /* __MINGW32__ */
#ifdef HAVE_REMOTE
#include "pcap-rpcap.h"
#endif /* HAVE_REMOTE */

static int pcap_setfilter_win32_npf(pcap_t *, struct bpf_program *);
static int pcap_setfilter_win32_dag(pcap_t *, struct bpf_program *);
static int pcap_getnonblock_win32(pcap_t *, char *);
static int pcap_setnonblock_win32(pcap_t *, int, char *);

/*dimension of the buffer in the pcap_t structure*/
#define	WIN32_DEFAULT_USER_BUFFER_SIZE 256000

/*dimension of the buffer in the kernel driver NPF */
#define	WIN32_DEFAULT_KERNEL_BUFFER_SIZE 1000000

/* Equivalent to ntohs(), but a lot faster under Windows */
#define SWAPS(_X) ((_X & 0xff) << 8) | (_X >> 8)

/*
 * Private data for capturing on WinPcap devices.
 */
struct pcap_win {
	int nonblock;
	int rfmon_selfstart;		/* a flag tells whether the monitor mode is set by itself */
	int filtering_in_kernel;	/* using kernel filter */

#ifdef HAVE_DAG_API
	int	dag_fcs_bits;		/* Number of checksum bits from link layer */
#endif
};

BOOL WINAPI DllMain(
  HANDLE hinstDLL,
  DWORD dwReason,
  LPVOID lpvReserved
)
{
	return (TRUE);
}

/*
 * Define stub versions of the monitor-mode support routines if this
 * isn't Npcap. HAVE_NPCAP_PACKET_API is defined by Npcap but not
 * WinPcap.
 */
#ifndef HAVE_NPCAP_PACKET_API
static int
PacketIsMonitorModeSupported(PCHAR AdapterName _U_)
{
	/*
	 * We don't support monitor mode.
	 */
	return (0);
}

static int
PacketSetMonitorMode(PCHAR AdapterName _U_, int mode _U_)
{
	/*
	 * This should never be called, as PacketIsMonitorModeSupported()
	 * will return 0, meaning "we don't support monitor mode, so
	 * don't try to turn it on or off".
	 */
	return (0);
}

static int
PacketGetMonitorMode(PCHAR AdapterName _U_)
{
	/*
	 * This should fail, so that pcap_activate_win32() returns
	 * PCAP_ERROR_RFMON_NOTSUP if our caller requested monitor
	 * mode.
	 */
	return (-1);
}
#endif

/* Start winsock */
int
wsockinit(void)
{
	WORD wVersionRequested;
	WSADATA wsaData;
	static int err = -1;
	static int done = 0;

	if (done)
		return (err);

	wVersionRequested = MAKEWORD( 1, 1);
	err = WSAStartup( wVersionRequested, &wsaData );
	atexit ((void(*)(void))WSACleanup);
	done = 1;

	if ( err != 0 )
		err = -1;
	return (err);
}

int
pcap_wsockinit(void)
{
       return (wsockinit());
}

static int
pcap_stats_win32(pcap_t *p, struct pcap_stat *ps)
{
	struct bpf_stat bstats;
	char errbuf[PCAP_ERRBUF_SIZE+1];

	/*
	 * Try to get statistics.
	 *
	 * (Please note - "struct pcap_stat" is *not* the same as
	 * WinPcap's "struct bpf_stat". It might currently have the
	 * same layout, but let's not cheat.
	 *
	 * Note also that we don't fill in ps_capt, as we might have
	 * been called by code compiled against an earlier version of
	 * WinPcap that didn't have ps_capt, in which case filling it
	 * in would stomp on whatever comes after the structure passed
	 * to us.
	 */
	if (!PacketGetStats(p->adapter, &bstats)) {
		pcap_win32_err_to_str(GetLastError(), errbuf);
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "PacketGetStats error: %s", errbuf);
		return (-1);
	}
	ps->ps_recv = bstats.bs_recv;
	ps->ps_drop = bstats.bs_drop;

	/*
	 * XXX - PacketGetStats() doesn't fill this in, so we just
	 * return 0.
	 */
#if 0
	ps->ps_ifdrop = bstats.ps_ifdrop;
#else
	ps->ps_ifdrop = 0;
#endif

	return (0);
}

/*
 * Win32-only routine for getting statistics.
 *
 * This way is definitely safer than passing the pcap_stat * from the userland.
 * In fact, there could happen than the user allocates a variable which is not
 * big enough for the new structure, and the library will write in a zone
 * which is not allocated to this variable.
 *
 * In this way, we're pretty sure we are writing on memory allocated to this
 * variable.
 *
 * XXX - but this is the wrong way to handle statistics.  Instead, we should
 * have an API that returns data in a form like the Options section of a
 * pcapng Interface Statistics Block:
 *
 *    http://xml2rfc.tools.ietf.org/cgi-bin/xml2rfc.cgi?url=https://raw.githubusercontent.com/pcapng/pcapng/master/draft-tuexen-opsawg-pcapng.xml&modeAsFormat=html/ascii&type=ascii#rfc.section.4.6
 *
 * which would let us add new statistics straightforwardly and indicate which
 * statistics we are and are *not* providing, rather than having to provide
 * possibly-bogus values for statistics we can't provide.
 */
struct pcap_stat *
pcap_stats_ex_win32(pcap_t *p, int *pcap_stat_size)
{
	struct bpf_stat bstats;
	char errbuf[PCAP_ERRBUF_SIZE+1];

	*pcap_stat_size = sizeof (p->stat);

	/*
	 * Try to get statistics.
	 *
	 * (Please note - "struct pcap_stat" is *not* the same as
	 * WinPcap's "struct bpf_stat". It might currently have the
	 * same layout, but let's not cheat.)
	 */
	if (!PacketGetStatsEx(p->adapter, &bstats)) {
		pcap_win32_err_to_str(GetLastError(), errbuf);
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "PacketGetStatsEx error: %s", errbuf);
		return (NULL);
	}
	p->stat.ps_recv = bstats.bs_recv;
	p->stat.ps_drop = bstats.bs_drop;
	p->stat.ps_ifdrop = bstats.ps_ifdrop;
#ifdef HAVE_REMOTE
	p->stat.ps_capt = bstats.bs_capt;
#endif
	return (&p->stat);
}

/* Set the dimension of the kernel-level capture buffer */
static int
pcap_setbuff_win32(pcap_t *p, int dim)
{
	if(PacketSetBuff(p->adapter,dim)==FALSE)
	{
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "driver error: not enough memory to allocate the kernel buffer");
		return (-1);
	}
	return (0);
}

/* Set the driver working mode */
static int
pcap_setmode_win32(pcap_t *p, int mode)
{
	if(PacketSetMode(p->adapter,mode)==FALSE)
	{
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "driver error: working mode not recognized");
		return (-1);
	}

	return (0);
}

/*set the minimum amount of data that will release a read call*/
static int
pcap_setmintocopy_win32(pcap_t *p, int size)
{
	if(PacketSetMinToCopy(p->adapter, size)==FALSE)
	{
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "driver error: unable to set the requested mintocopy size");
		return (-1);
	}
	return (0);
}

static HANDLE
pcap_getevent_win32(pcap_t *p)
{
	return (PacketGetReadEvent(p->adapter));
}

static int
pcap_oid_get_request_win32(pcap_t *p, bpf_u_int32 oid, void *data, size_t *lenp)
{
	PACKET_OID_DATA *oid_data_arg;
	char errbuf[PCAP_ERRBUF_SIZE+1];

	/*
	 * Allocate a PACKET_OID_DATA structure to hand to PacketRequest().
	 * It should be big enough to hold "*lenp" bytes of data; it
	 * will actually be slightly larger, as PACKET_OID_DATA has a
	 * 1-byte data array at the end, standing in for the variable-length
	 * data that's actually there.
	 */
	oid_data_arg = malloc(sizeof (PACKET_OID_DATA) + *lenp);
	if (oid_data_arg == NULL) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Couldn't allocate argument buffer for PacketRequest");
		return (PCAP_ERROR);
	}

	/*
	 * No need to copy the data - we're doing a fetch.
	 */
	oid_data_arg->Oid = oid;
	oid_data_arg->Length = (ULONG)(*lenp);	/* XXX - check for ridiculously large value? */
	if (!PacketRequest(p->adapter, FALSE, oid_data_arg)) {
		pcap_win32_err_to_str(GetLastError(), errbuf);
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error calling PacketRequest: %s", errbuf);
		free(oid_data_arg);
		return (PCAP_ERROR);
	}

	/*
	 * Get the length actually supplied.
	 */
	*lenp = oid_data_arg->Length;

	/*
	 * Copy back the data we fetched.
	 */
	memcpy(data, oid_data_arg->Data, *lenp);
	free(oid_data_arg);
	return (0);
}

static int
pcap_oid_set_request_win32(pcap_t *p, bpf_u_int32 oid, const void *data,
    size_t *lenp)
{
	PACKET_OID_DATA *oid_data_arg;
	char errbuf[PCAP_ERRBUF_SIZE+1];

	/*
	 * Allocate a PACKET_OID_DATA structure to hand to PacketRequest().
	 * It should be big enough to hold "*lenp" bytes of data; it
	 * will actually be slightly larger, as PACKET_OID_DATA has a
	 * 1-byte data array at the end, standing in for the variable-length
	 * data that's actually there.
	 */
	oid_data_arg = malloc(sizeof (PACKET_OID_DATA) + *lenp);
	if (oid_data_arg == NULL) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Couldn't allocate argument buffer for PacketRequest");
		return (PCAP_ERROR);
	}

	oid_data_arg->Oid = oid;
	oid_data_arg->Length = (ULONG)(*lenp);	/* XXX - check for ridiculously large value? */
	memcpy(oid_data_arg->Data, data, *lenp);
	if (!PacketRequest(p->adapter, TRUE, oid_data_arg)) {
		pcap_win32_err_to_str(GetLastError(), errbuf);
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error calling PacketRequest: %s", errbuf);
		free(oid_data_arg);
		return (PCAP_ERROR);
	}

	/*
	 * Get the length actually copied.
	 */
	*lenp = oid_data_arg->Length;

	/*
	 * No need to copy the data - we're doing a set.
	 */
	free(oid_data_arg);
	return (0);
}

static u_int
pcap_sendqueue_transmit_win32(pcap_t *p, pcap_send_queue *queue, int sync)
{
	u_int res;
	char errbuf[PCAP_ERRBUF_SIZE+1];

	if (p->adapter==NULL) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Cannot transmit a queue to an offline capture or to a TurboCap port");
		return (0);
	}

	res = PacketSendPackets(p->adapter,
		queue->buffer,
		queue->len,
		(BOOLEAN)sync);

	if(res != queue->len){
		pcap_win32_err_to_str(GetLastError(), errbuf);
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error opening adapter: %s", errbuf);
	}

	return (res);
}

static int
pcap_setuserbuffer_win32(pcap_t *p, int size)
{
	unsigned char *new_buff;

	if (size<=0) {
		/* Bogus parameter */
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error: invalid size %d",size);
		return (-1);
	}

	/* Allocate the buffer */
	new_buff=(unsigned char*)malloc(sizeof(char)*size);

	if (!new_buff) {
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error: not enough memory");
		return (-1);
	}

	free(p->buffer);

	p->buffer=new_buff;
	p->bufsize=size;

	return (0);
}

static int
pcap_live_dump_win32(pcap_t *p, char *filename, int maxsize, int maxpacks)
{
	BOOLEAN res;

	/* Set the packet driver in dump mode */
	res = PacketSetMode(p->adapter, PACKET_MODE_DUMP);
	if(res == FALSE){
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error setting dump mode");
		return (-1);
	}

	/* Set the name of the dump file */
	res = PacketSetDumpName(p->adapter, filename, (int)strlen(filename));
	if(res == FALSE){
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error setting kernel dump file name");
		return (-1);
	}

	/* Set the limits of the dump file */
	res = PacketSetDumpLimits(p->adapter, maxsize, maxpacks);

	return (0);
}

static int
pcap_live_dump_ended_win32(pcap_t *p, int sync)
{
	return (PacketIsDumpEnded(p->adapter, (BOOLEAN)sync));
}

static PAirpcapHandle
pcap_get_airpcap_handle_win32(pcap_t *p)
{
#ifdef HAVE_AIRPCAP_API
	return (PacketGetAirPcapHandle(p->adapter));
#else
	return (NULL);
#endif /* HAVE_AIRPCAP_API */
}

static int
pcap_read_win32_npf(pcap_t *p, int cnt, pcap_handler callback, u_char *user)
{
	PACKET Packet;
	int cc;
	int n = 0;
	register u_char *bp, *ep;
	u_char *datap;
	struct pcap_win *pw = p->priv;

	cc = p->cc;
	if (p->cc == 0) {
		/*
		 * Has "pcap_breakloop()" been called?
		 */
		if (p->break_loop) {
			/*
			 * Yes - clear the flag that indicates that it
			 * has, and return PCAP_ERROR_BREAK to indicate
			 * that we were told to break out of the loop.
			 */
			p->break_loop = 0;
			return (PCAP_ERROR_BREAK);
		}

		/*
		 * Capture the packets.
		 *
		 * The PACKET structure had a bunch of extra stuff for
		 * Windows 9x/Me, but the only interesting data in it
		 * in the versions of Windows that we support is just
		 * a copy of p->buffer, a copy of p->buflen, and the
		 * actual number of bytes read returned from
		 * PacketReceivePacket(), none of which has to be
		 * retained from call to call, so we just keep one on
		 * the stack.
		 */
		PacketInitPacket(&Packet, (BYTE *)p->buffer, p->bufsize);
		if (!PacketReceivePacket(p->adapter, &Packet, TRUE)) {
			pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "read error: PacketReceivePacket failed");
			return (PCAP_ERROR);
		}

		cc = Packet.ulBytesReceived;

		bp = p->buffer;
	}
	else
		bp = p->bp;

	/*
	 * Loop through each packet.
	 */
#define bhp ((struct bpf_hdr *)bp)
	ep = bp + cc;
	while (1) {
		register int caplen, hdrlen;

		/*
		 * Has "pcap_breakloop()" been called?
		 * If so, return immediately - if we haven't read any
		 * packets, clear the flag and return PCAP_ERROR_BREAK
		 * to indicate that we were told to break out of the loop,
		 * otherwise leave the flag set, so that the *next* call
		 * will break out of the loop without having read any
		 * packets, and return the number of packets we've
		 * processed so far.
		 */
		if (p->break_loop) {
			if (n == 0) {
				p->break_loop = 0;
				return (PCAP_ERROR_BREAK);
			} else {
				p->bp = bp;
				p->cc = (int) (ep - bp);
				return (n);
			}
		}
		if (bp >= ep)
			break;

		caplen = bhp->bh_caplen;
		hdrlen = bhp->bh_hdrlen;
		datap = bp + hdrlen;

		/*
		 * Short-circuit evaluation: if using BPF filter
		 * in kernel, no need to do it now - we already know
		 * the packet passed the filter.
		 *
		 * XXX - bpf_filter() should always return TRUE if
		 * handed a null pointer for the program, but it might
		 * just try to "run" the filter, so we check here.
		 */
		if (pw->filtering_in_kernel ||
		    p->fcode.bf_insns == NULL ||
		    bpf_filter(p->fcode.bf_insns, datap, bhp->bh_datalen, caplen)) {
			/*
			 * XXX A bpf_hdr matches a pcap_pkthdr.
			 */
			(*callback)(user, (struct pcap_pkthdr*)bp, datap);
			bp += Packet_WORDALIGN(caplen + hdrlen);
			if (++n >= cnt && !PACKET_COUNT_IS_UNLIMITED(cnt)) {
				p->bp = bp;
				p->cc = (int) (ep - bp);
				return (n);
			}
		} else {
			/*
			 * Skip this packet.
			 */
			bp += Packet_WORDALIGN(caplen + hdrlen);
		}
	}
#undef bhp
	p->cc = 0;
	return (n);
}

#ifdef HAVE_DAG_API
static int
pcap_read_win32_dag(pcap_t *p, int cnt, pcap_handler callback, u_char *user)
{
	struct pcap_win *pw = p->priv;
	PACKET Packet;
	u_char *dp = NULL;
	int	packet_len = 0, caplen = 0;
	struct pcap_pkthdr	pcap_header;
	u_char *endofbuf;
	int n = 0;
	dag_record_t *header;
	unsigned erf_record_len;
	ULONGLONG ts;
	int cc;
	unsigned swt;
	unsigned dfp = p->adapter->DagFastProcess;

	cc = p->cc;
	if (cc == 0) /* Get new packets only if we have processed all the ones of the previous read */
	{
		/*
		 * Get new packets from the network.
		 *
		 * The PACKET structure had a bunch of extra stuff for
		 * Windows 9x/Me, but the only interesting data in it
		 * in the versions of Windows that we support is just
		 * a copy of p->buffer, a copy of p->buflen, and the
		 * actual number of bytes read returned from
		 * PacketReceivePacket(), none of which has to be
		 * retained from call to call, so we just keep one on
		 * the stack.
		 */
		PacketInitPacket(&Packet, (BYTE *)p->buffer, p->bufsize);
		if (!PacketReceivePacket(p->adapter, &Packet, TRUE)) {
			pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "read error: PacketReceivePacket failed");
			return (-1);
		}

		cc = Packet.ulBytesReceived;
		if(cc == 0)
			/* The timeout has expired but we no packets arrived */
			return (0);
		header = (dag_record_t*)p->adapter->DagBuffer;
	}
	else
		header = (dag_record_t*)p->bp;

	endofbuf = (char*)header + cc;

	/*
	 * Cycle through the packets
	 */
	do
	{
		erf_record_len = SWAPS(header->rlen);
		if((char*)header + erf_record_len > endofbuf)
			break;

		/* Increase the number of captured packets */
		p->stat.ps_recv++;

		/* Find the beginning of the packet */
		dp = ((u_char *)header) + dag_record_size;

		/* Determine actual packet len */
		switch(header->type)
		{
		case TYPE_ATM:
			packet_len = ATM_SNAPLEN;
			caplen = ATM_SNAPLEN;
			dp += 4;

			break;

		case TYPE_ETH:
			swt = SWAPS(header->wlen);
			packet_len = swt - (pw->dag_fcs_bits);
			caplen = erf_record_len - dag_record_size - 2;
			if (caplen > packet_len)
			{
				caplen = packet_len;
			}
			dp += 2;

			break;

		case TYPE_HDLC_POS:
			swt = SWAPS(header->wlen);
			packet_len = swt - (pw->dag_fcs_bits);
			caplen = erf_record_len - dag_record_size;
			if (caplen > packet_len)
			{
				caplen = packet_len;
			}

			break;
		}

		if(caplen > p->snapshot)
			caplen = p->snapshot;

		/*
		 * Has "pcap_breakloop()" been called?
		 * If so, return immediately - if we haven't read any
		 * packets, clear the flag and return -2 to indicate
		 * that we were told to break out of the loop, otherwise
		 * leave the flag set, so that the *next* call will break
		 * out of the loop without having read any packets, and
		 * return the number of packets we've processed so far.
		 */
		if (p->break_loop)
		{
			if (n == 0)
			{
				p->break_loop = 0;
				return (-2);
			}
			else
			{
				p->bp = (char*)header;
				p->cc = endofbuf - (char*)header;
				return (n);
			}
		}

		if(!dfp)
		{
			/* convert between timestamp formats */
			ts = header->ts;
			pcap_header.ts.tv_sec = (int)(ts >> 32);
			ts = (ts & 0xffffffffi64) * 1000000;
			ts += 0x80000000; /* rounding */
			pcap_header.ts.tv_usec = (int)(ts >> 32);
			if (pcap_header.ts.tv_usec >= 1000000) {
				pcap_header.ts.tv_usec -= 1000000;
				pcap_header.ts.tv_sec++;
			}
		}

		/* No underlaying filtering system. We need to filter on our own */
		if (p->fcode.bf_insns)
		{
			if (bpf_filter(p->fcode.bf_insns, dp, packet_len, caplen) == 0)
			{
				/* Move to next packet */
				header = (dag_record_t*)((char*)header + erf_record_len);
				continue;
			}
		}

		/* Fill the header for the user suppplied callback function */
		pcap_header.caplen = caplen;
		pcap_header.len = packet_len;

		/* Call the callback function */
		(*callback)(user, &pcap_header, dp);

		/* Move to next packet */
		header = (dag_record_t*)((char*)header + erf_record_len);

		/* Stop if the number of packets requested by user has been reached*/
		if (++n >= cnt && !PACKET_COUNT_IS_UNLIMITED(cnt))
		{
			p->bp = (char*)header;
			p->cc = endofbuf - (char*)header;
			return (n);
		}
	}
	while((u_char*)header < endofbuf);

	return (1);
}
#endif /* HAVE_DAG_API */

/* Send a packet to the network */
static int
pcap_inject_win32(pcap_t *p, const void *buf, size_t size){
	LPPACKET PacketToSend;

	PacketToSend=PacketAllocatePacket();

	if (PacketToSend == NULL)
	{
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "send error: PacketAllocatePacket failed");
		return (-1);
	}

	PacketInitPacket(PacketToSend, (PVOID)buf, (UINT)size);
	if(PacketSendPacket(p->adapter,PacketToSend,TRUE) == FALSE){
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "send error: PacketSendPacket failed");
		PacketFreePacket(PacketToSend);
		return (-1);
	}

	PacketFreePacket(PacketToSend);

	/*
	 * We assume it all got sent if "PacketSendPacket()" succeeded.
	 * "pcap_inject()" is expected to return the number of bytes
	 * sent.
	 */
	return ((int)size);
}

static void
pcap_cleanup_win32(pcap_t *p)
{
	struct pcap_win *pw = p->priv;
	if (p->adapter != NULL) {
		PacketCloseAdapter(p->adapter);
		p->adapter = NULL;
	}
	if (pw->rfmon_selfstart)
	{
		PacketSetMonitorMode(p->opt.device, 0);
	}
	pcap_cleanup_live_common(p);
}

static int
pcap_activate_win32(pcap_t *p)
{
	struct pcap_win *pw = p->priv;
	NetType type;
	int res;
	char errbuf[PCAP_ERRBUF_SIZE+1];

#ifdef HAVE_REMOTE
	char host[PCAP_BUF_SIZE + 1];
	char port[PCAP_BUF_SIZE + 1];
	char name[PCAP_BUF_SIZE + 1];
	int srctype;
	int opensource_remote_result;

	struct pcap_md *md;				/* structure used when doing a remote live capture */
	md = (struct pcap_md *) ((u_char*)p->priv + sizeof(struct pcap_win));

	/*
	Retrofit; we have to make older applications compatible with the remote capture
	So, we're calling the pcap_open_remote() from here, that is a very dirty thing.
	Obviously, we cannot exploit all the new features; for instance, we cannot
	send authentication, we cannot use a UDP data connection, and so on.
	*/
	if (pcap_parsesrcstr(p->opt.device, &srctype, host, port, name, p->errbuf))
		return PCAP_ERROR;

	if (srctype == PCAP_SRC_IFREMOTE)
	{
		opensource_remote_result = pcap_opensource_remote(p, NULL);

		if (opensource_remote_result != 0)
			return opensource_remote_result;

		md->rmt_flags = (p->opt.promisc) ? PCAP_OPENFLAG_PROMISCUOUS : 0;

		return 0;
	}

	if (srctype == PCAP_SRC_IFLOCAL)
	{
		/*
		* If it starts with rpcap://, cut down the string
		*/
		if (strncmp(p->opt.device, PCAP_SRC_IF_STRING, strlen(PCAP_SRC_IF_STRING)) == 0)
		{
			size_t len = strlen(p->opt.device) - strlen(PCAP_SRC_IF_STRING) + 1;
			char *new_string;
			/*
			* allocate a new string and free the old one
			*/
			if (len > 0)
			{
				new_string = (char*)malloc(len);
				if (new_string != NULL)
				{
					char *tmp;
					strcpy_s(new_string, len, p->opt.device + strlen(PCAP_SRC_IF_STRING));
					tmp = p->opt.device;
					p->opt.device = new_string;
					free(tmp);
				}
			}
		}
	}

#endif	/* HAVE_REMOTE */

	if (p->opt.rfmon) {
		/*
		 * Monitor mode is supported on Windows Vista and later.
		 */
		if (PacketGetMonitorMode(p->opt.device) == 1)
		{
			pw->rfmon_selfstart = 0;
		}
		else
		{
			if ((res = PacketSetMonitorMode(p->opt.device, 1)) != 1)
			{
				pw->rfmon_selfstart = 0;
				// Monitor mode is not supported.
				if (res == 0)
				{
					return PCAP_ERROR_RFMON_NOTSUP;
				}
				else
				{
					return PCAP_ERROR;
				}
			}
			else
			{
				pw->rfmon_selfstart = 1;
			}
		}
	}

	/* Init WinSock */
	wsockinit();

	p->adapter = PacketOpenAdapter(p->opt.device);

	if (p->adapter == NULL)
	{
		/* Adapter detected but we are not able to open it. Return failure. */
		pcap_win32_err_to_str(GetLastError(), errbuf);
		if (pw->rfmon_selfstart)
		{
			PacketSetMonitorMode(p->opt.device, 0);
		}
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Error opening adapter: %s", errbuf);
		return (PCAP_ERROR);
	}

	/*get network type*/
	if(PacketGetNetType (p->adapter,&type) == FALSE)
	{
		pcap_win32_err_to_str(GetLastError(), errbuf);
		pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
		    "Cannot determine the network type: %s", errbuf);
		goto bad;
	}

	/*Set the linktype*/
	switch (type.LinkType)
	{
	case NdisMediumWan:
		p->linktype = DLT_EN10MB;
		break;

	case NdisMedium802_3:
		p->linktype = DLT_EN10MB;
		/*
		 * This is (presumably) a real Ethernet capture; give it a
		 * link-layer-type list with DLT_EN10MB and DLT_DOCSIS, so
		 * that an application can let you choose it, in case you're
		 * capturing DOCSIS traffic that a Cisco Cable Modem
		 * Termination System is putting out onto an Ethernet (it
		 * doesn't put an Ethernet header onto the wire, it puts raw
		 * DOCSIS frames out on the wire inside the low-level
		 * Ethernet framing).
		 */
		p->dlt_list = (u_int *) malloc(sizeof(u_int) * 2);
		/*
		 * If that fails, just leave the list empty.
		 */
		if (p->dlt_list != NULL) {
			p->dlt_list[0] = DLT_EN10MB;
			p->dlt_list[1] = DLT_DOCSIS;
			p->dlt_count = 2;
		}
		break;

	case NdisMediumFddi:
		p->linktype = DLT_FDDI;
		break;

	case NdisMedium802_5:
		p->linktype = DLT_IEEE802;
		break;

	case NdisMediumArcnetRaw:
		p->linktype = DLT_ARCNET;
		break;

	case NdisMediumArcnet878_2:
		p->linktype = DLT_ARCNET;
		break;

	case NdisMediumAtm:
		p->linktype = DLT_ATM_RFC1483;
		break;

	case NdisMediumCHDLC:
		p->linktype = DLT_CHDLC;
		break;

	case NdisMediumPPPSerial:
		p->linktype = DLT_PPP_SERIAL;
		break;

	case NdisMediumNull:
		p->linktype = DLT_NULL;
		break;

	case NdisMediumBare80211:
		p->linktype = DLT_IEEE802_11;
		break;

	case NdisMediumRadio80211:
		p->linktype = DLT_IEEE802_11_RADIO;
		break;

	case NdisMediumPpi:
		p->linktype = DLT_PPI;
		break;

	default:
		p->linktype = DLT_EN10MB;			/*an unknown adapter is assumed to be ethernet*/
		break;
	}

	/* Set promiscuous mode */
	if (p->opt.promisc)
	{

		if (PacketSetHwFilter(p->adapter,NDIS_PACKET_TYPE_PROMISCUOUS) == FALSE)
		{
			pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "failed to set hardware filter to promiscuous mode");
			goto bad;
		}
	}
	else
	{
		if (PacketSetHwFilter(p->adapter,NDIS_PACKET_TYPE_ALL_LOCAL) == FALSE)
		{
			pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "failed to set hardware filter to non-promiscuous mode");
			goto bad;
		}
	}

	/* Set the buffer size */
	p->bufsize = WIN32_DEFAULT_USER_BUFFER_SIZE;

	if(!(p->adapter->Flags & INFO_FLAG_DAG_CARD))
	{
	/*
	 * Traditional Adapter
	 */
		/*
		 * If the buffer size wasn't explicitly set, default to
		 * WIN32_DEFAULT_KERNEL_BUFFER_SIZE.
		 */
	 	if (p->opt.buffer_size == 0)
	 		p->opt.buffer_size = WIN32_DEFAULT_KERNEL_BUFFER_SIZE;

		if(PacketSetBuff(p->adapter,p->opt.buffer_size)==FALSE)
		{
			pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "driver error: not enough memory to allocate the kernel buffer");
			goto bad;
		}

		p->buffer = malloc(p->bufsize);
		if (p->buffer == NULL)
		{
			pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE, "malloc: %s", pcap_strerror(errno));
			goto bad;
		}

		if (p->opt.immediate)
		{
			/* tell the driver to copy the buffer as soon as data arrives */
			if(PacketSetMinToCopy(p->adapter,0)==FALSE)
			{
				pcap_win32_err_to_str(GetLastError(), errbuf);
				pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
				    "Error calling PacketSetMinToCopy: %s",
				    errbuf);
				goto bad;
			}
		}
		else
		{
			/* tell the driver to copy the buffer only if it contains at least 16K */
			if(PacketSetMinToCopy(p->adapter,16000)==FALSE)
			{
				pcap_win32_err_to_str(GetLastError(), errbuf);
				pcap_snprintf(p->errbuf, PCAP_ERRBUF_SIZE,
				    "Error calling PacketSetMinToCopy: %s",
				    errbuf);
				goto bad;
			}
		}
	}
	else
#ifdef HAVE_DAG_API
	{
	/*
	 * Dag Card
	 */
		LONG	status;
		HKEY	dagkey;
		DWORD	lptype;
		DWORD	lpcbdata;
		int		postype = 0;
		char	keyname[512];

		pcap_snprintf(keyname, sizeof(keyname), "%s\\CardParams\\%s",
			"SYSTEM\\CurrentControlSet\\Services\\DAG",
			strstr(_strlwr(p->opt.device), "dag"));
		do
		{
			status = RegOpenKeyEx(HKEY_LOCAL_MACHINE, keyname, 0, KEY_READ, &dagkey);
			if(status != ERROR_SUCCESS)
				break;

			status = RegQueryValueEx(dagkey,
				"PosType",
				NULL,
				&lptype,
				(char*)&postype,
				&lpcbdata);

			if(status != ERROR_SUCCESS)
			{
				postype = 0;
			}

			RegCloseKey(dagkey);
		}
		while(FALSE);


		p->snapshot = PacketSetSnapLen(p->adapter, snaplen);

		/* Set the length of the FCS associated to any packet. This value
		 * will be subtracted to the packet length */
		pw->dag_fcs_bits = p->adapter->DagFcsLen;
	}
#else
	goto bad;
#endif /* HAVE_DAG_API */

	PacketSetReadTimeout(p->adapter, p->opt.timeout);

#ifdef HAVE_DAG_API
	if(p->adapter->Flags & INFO_FLAG_DAG_CARD)
	{
		/* install dag specific handlers for read and setfilter */
		p->read_op = pcap_read_win32_dag;
		p->setfilter_op = pcap_setfilter_win32_dag;
	}
	else
	{
#endif /* HAVE_DAG_API */
		/* install traditional npf handlers for read and setfilter */
		p->read_op = pcap_read_win32_npf;
		p->setfilter_op = pcap_setfilter_win32_npf;
#ifdef HAVE_DAG_API
	}
#endif /* HAVE_DAG_API */
	p->setdirection_op = NULL;	/* Not implemented. */
	    /* XXX - can this be implemented on some versions of Windows? */
	p->inject_op = pcap_inject_win32;
	p->set_datalink_op = NULL;	/* can't change data link type */
	p->getnonblock_op = pcap_getnonblock_win32;
	p->setnonblock_op = pcap_setnonblock_win32;
	p->stats_op = pcap_stats_win32;
	p->stats_ex_op = pcap_stats_ex_win32;
	p->setbuff_op = pcap_setbuff_win32;
	p->setmode_op = pcap_setmode_win32;
	p->setmintocopy_op = pcap_setmintocopy_win32;
	p->getevent_op = pcap_getevent_win32;
	p->oid_get_request_op = pcap_oid_get_request_win32;
	p->oid_set_request_op = pcap_oid_set_request_win32;
	p->sendqueue_transmit_op = pcap_sendqueue_transmit_win32;
	p->setuserbuffer_op = pcap_setuserbuffer_win32;
	p->live_dump_op = pcap_live_dump_win32;
	p->live_dump_ended_op = pcap_live_dump_ended_win32;
	p->get_airpcap_handle_op = pcap_get_airpcap_handle_win32;
	p->cleanup_op = pcap_cleanup_win32;

	return (0);
bad:
	pcap_cleanup_win32(p);
	return (PCAP_ERROR);
}

/*
* Check if rfmon mode is supported on the pcap_t for Windows systems.
*/
static int
pcap_can_set_rfmon_win32(pcap_t *p)
{
	return (PacketIsMonitorModeSupported(p->opt.device) == 1);
}

pcap_t *
pcap_create_interface(const char *device _U_, char *ebuf)
{
	pcap_t *p;

#ifdef HAVE_REMOTE
	p = pcap_create_common(ebuf, sizeof(struct pcap_win) + sizeof(struct pcap_md));
#else
	p = pcap_create_common(ebuf, sizeof(struct pcap_win));
#endif /* HAVE_REMOTE */
	if (p == NULL)
		return (NULL);

	p->activate_op = pcap_activate_win32;
	p->can_set_rfmon_op = pcap_can_set_rfmon_win32;
	return (p);
}

static int
pcap_setfilter_win32_npf(pcap_t *p, struct bpf_program *fp)
{
	struct pcap_win *pw = p->priv;

	if(PacketSetBpf(p->adapter,fp)==FALSE){
		/*
		 * Kernel filter not installed.
		 *
		 * XXX - we don't know whether this failed because:
		 *
		 *  the kernel rejected the filter program as invalid,
		 *  in which case we should fall back on userland
		 *  filtering;
		 *
		 *  the kernel rejected the filter program as too big,
		 *  in which case we should again fall back on
		 *  userland filtering;
		 *
		 *  there was some other problem, in which case we
		 *  should probably report an error.
		 *
		 * For NPF devices, the Win32 status will be
		 * STATUS_INVALID_DEVICE_REQUEST for invalid
		 * filters, but I don't know what it'd be for
		 * other problems, and for some other devices
		 * it might not be set at all.
		 *
		 * So we just fall back on userland filtering in
		 * all cases.
		 */

		/*
		 * install_bpf_program() validates the program.
		 *
		 * XXX - what if we already have a filter in the kernel?
		 */
		if (install_bpf_program(p, fp) < 0)
			return (-1);
		pw->filtering_in_kernel = 0;	/* filtering in userland */
		return (0);
	}

	/*
	 * It worked.
	 */
	pw->filtering_in_kernel = 1;	/* filtering in the kernel */

	/*
	 * Discard any previously-received packets, as they might have
	 * passed whatever filter was formerly in effect, but might
	 * not pass this filter (BIOCSETF discards packets buffered
	 * in the kernel, so you can lose packets in any case).
	 */
	p->cc = 0;
	return (0);
}

/*
 * We filter at user level, since the kernel driver does't process the packets
 */
static int
pcap_setfilter_win32_dag(pcap_t *p, struct bpf_program *fp) {

	if(!fp)
	{
		strlcpy(p->errbuf, "setfilter: No filter specified", sizeof(p->errbuf));
		return (-1);
	}

	/* Install a user level filter */
	if (install_bpf_program(p, fp) < 0)
	{
		pcap_snprintf(p->errbuf, sizeof(p->errbuf),
			"setfilter, unable to install the filter: %s", pcap_strerror(errno));
		return (-1);
	}

	return (0);
}

static int
pcap_getnonblock_win32(pcap_t *p, char *errbuf)
{
	struct pcap_win *pw = p->priv;

	/*
	 * XXX - if there were a PacketGetReadTimeout() call, we
	 * would use it, and return 1 if the timeout is -1
	 * and 0 otherwise.
	 */
	return (pw->nonblock);
}

static int
pcap_setnonblock_win32(pcap_t *p, int nonblock, char *errbuf)
{
	struct pcap_win *pw = p->priv;
	int newtimeout;
	char win_errbuf[PCAP_ERRBUF_SIZE+1];

	if (nonblock) {
		/*
		 * Set the read timeout to -1 for non-blocking mode.
		 */
		newtimeout = -1;
	} else {
		/*
		 * Restore the timeout set when the device was opened.
		 * (Note that this may be -1, in which case we're not
		 * really leaving non-blocking mode.  However, although
		 * the timeout argument to pcap_set_timeout() and
		 * pcap_open_live() is an int, you're not supposed to
		 * supply a negative value, so that "shouldn't happen".)
		 */
		newtimeout = p->opt.timeout;
	}
	if (!PacketSetReadTimeout(p->adapter, newtimeout)) {
		pcap_win32_err_to_str(GetLastError(), win_errbuf);
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
		    "PacketSetReadTimeout: %s", win_errbuf);
		return (-1);
	}
	pw->nonblock = (newtimeout == -1);
	return (0);
}

static int
pcap_add_if_win32(pcap_if_t **devlist, char *name, bpf_u_int32 flags,
    const char *description, char *errbuf)
{
	pcap_if_t *curdev;
	npf_if_addr if_addrs[MAX_NETWORK_ADDRESSES];
	LONG if_addr_size;
	int res = 0;

	if_addr_size = MAX_NETWORK_ADDRESSES;

	/*
	 * Add an entry for this interface, with no addresses.
	 */
	if (add_or_find_if(&curdev, devlist, name, flags, description,
	    errbuf) == -1) {
		/*
		 * Failure.
		 */
		return (-1);
	}

	/*
	 * Get the list of addresses for the interface.
	 */
	if (!PacketGetNetInfoEx((void *)name, if_addrs, &if_addr_size)) {
		/*
		 * Failure.
		 *
		 * We don't return an error, because this can happen with
		 * NdisWan interfaces, and we want to supply them even
		 * if we can't supply their addresses.
		 *
		 * We return an entry with an empty address list.
		 */
		return (0);
	}

	/*
	 * Now add the addresses.
	 */
	while (if_addr_size-- > 0) {
		/*
		 * "curdev" is an entry for this interface; add an entry for
		 * this address to its list of addresses.
		 */
		if(curdev == NULL)
			break;
		res = add_addr_to_dev(curdev,
		    (struct sockaddr *)&if_addrs[if_addr_size].IPAddress,
		    sizeof (struct sockaddr_storage),
		    (struct sockaddr *)&if_addrs[if_addr_size].SubnetMask,
		    sizeof (struct sockaddr_storage),
		    (struct sockaddr *)&if_addrs[if_addr_size].Broadcast,
		    sizeof (struct sockaddr_storage),
		    NULL,
		    0,
		    errbuf);
		if (res == -1) {
			/*
			 * Failure.
			 */
			break;
		}
	}

	return (res);
}

int
pcap_platform_finddevs(pcap_if_t **alldevsp, char *errbuf)
{
	pcap_if_t *devlist = NULL;
	int ret = 0;
	const char *desc;
	char *AdaptersName;
	ULONG NameLength;
	char *name;
	char our_errbuf[PCAP_ERRBUF_SIZE+1];

	/*
	 * Find out how big a buffer we need.
	 *
	 * This call should always return FALSE; if the error is
	 * ERROR_INSUFFICIENT_BUFFER, NameLength will be set to
	 * the size of the buffer we need, otherwise there's a
	 * problem, and NameLength should be set to 0.
	 *
	 * It shouldn't require NameLength to be set, but,
	 * at least as of WinPcap 4.1.3, it checks whether
	 * NameLength is big enough before it checks for a
	 * NULL buffer argument, so, while it'll still do
	 * the right thing if NameLength is uninitialized and
	 * whatever junk happens to be there is big enough
	 * (because the pointer argument will be null), it's
	 * still reading an uninitialized variable.
	 */
	NameLength = 0;
	if (!PacketGetAdapterNames(NULL, &NameLength))
	{
		DWORD last_error = GetLastError();

		if (last_error != ERROR_INSUFFICIENT_BUFFER)
		{
			pcap_win32_err_to_str(last_error, our_errbuf);
			pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			    "PacketGetAdapterNames: %s", our_errbuf);
			return (-1);
		}
	}

	if (NameLength > 0)
		AdaptersName = (char*) malloc(NameLength);
	else
	{
		*alldevsp = NULL;
		return 0;
	}
	if (AdaptersName == NULL)
	{
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "Cannot allocate enough memory to list the adapters.");
		return (-1);
	}

	if (!PacketGetAdapterNames(AdaptersName, &NameLength)) {
		pcap_win32_err_to_str(GetLastError(), our_errbuf);
		pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "PacketGetAdapterNames: %s",
		    our_errbuf);
		free(AdaptersName);
		return (-1);
	}

	/*
	 * "PacketGetAdapterNames()" returned a list of
	 * null-terminated ASCII interface name strings,
	 * terminated by a null string, followed by a list
	 * of null-terminated ASCII interface description
	 * strings, terminated by a null string.
	 * This means there are two ASCII nulls at the end
	 * of the first list.
	 *
	 * Find the end of the first list; that's the
	 * beginning of the second list.
	 */
	desc = &AdaptersName[0];
	while (*desc != '\0' || *(desc + 1) != '\0')
		desc++;

	/*
 	 * Found it - "desc" points to the first of the two
	 * nulls at the end of the list of names, so the
	 * first byte of the list of descriptions is two bytes
	 * after it.
	 */
	desc += 2;

	/*
	 * Loop over the elements in the first list.
	 */
	name = &AdaptersName[0];
	while (*name != '\0') {
		bpf_u_int32 flags = 0;
#ifdef HAVE_PACKET_IS_LOOPBACK_ADAPTER
		/*
		 * Is this a loopback interface?
		 */
		if (PacketIsLoopbackAdapter(name)) {
			/* Yes */
			flags |= PCAP_IF_LOOPBACK;
		}
#endif

		/*
		 * Add an entry for this interface.
		 */
		if (pcap_add_if_win32(&devlist, name, flags, desc,
		    errbuf) == -1) {
			/*
			 * Failure.
			 */
			ret = -1;
			break;
		}
		name += strlen(name) + 1;
		desc += strlen(desc) + 1;
	}

	if (ret == -1) {
		/*
		 * We had an error; free the list we've been constructing.
		 */
		if (devlist != NULL) {
			pcap_freealldevs(devlist);
			devlist = NULL;
		}
	}

	*alldevsp = devlist;
	free(AdaptersName);
	return (ret);
}
