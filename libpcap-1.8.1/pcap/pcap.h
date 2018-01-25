/* -*- Mode: c; tab-width: 8; indent-tabs-mode: 1; c-basic-offset: 8; -*- */
/*
 * Copyright (c) 1993, 1994, 1995, 1996, 1997
 *	The Regents of the University of California.  All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions
 * are met:
 * 1. Redistributions of source code must retain the above copyright
 *    notice, this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright
 *    notice, this list of conditions and the following disclaimer in the
 *    documentation and/or other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software
 *    must display the following acknowledgement:
 *	This product includes software developed by the Computer Systems
 *	Engineering Group at Lawrence Berkeley Laboratory.
 * 4. Neither the name of the University nor of the Laboratory may be used
 *    to endorse or promote products derived from this software without
 *    specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE REGENTS AND CONTRIBUTORS ``AS IS'' AND
 * ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED.  IN NO EVENT SHALL THE REGENTS OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS
 * OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
 * HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT
 * LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY
 * OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF
 * SUCH DAMAGE.
 */

#ifndef lib_pcap_pcap_h
#define lib_pcap_pcap_h

#include <pcap/export-defs.h>

#if defined(_WIN32)
  #include <pcap-stdinc.h>
#elif defined(MSDOS)
  #include <sys/types.h>
  #include <sys/socket.h>  /* u_int, u_char etc. */
#else /* UN*X */
  #include <sys/types.h>
  #include <sys/time.h>
#endif /* _WIN32/MSDOS/UN*X */

#ifndef PCAP_DONT_INCLUDE_PCAP_BPF_H
#include <pcap/bpf.h>
#endif

#include <stdio.h>

#ifdef __cplusplus
extern "C" {
#endif

/*
 * Version number of the current version of the pcap file format.
 *
 * NOTE: this is *NOT* the version number of the libpcap library.
 * To fetch the version information for the version of libpcap
 * you're using, use pcap_lib_version().
 */
#define PCAP_VERSION_MAJOR 2
#define PCAP_VERSION_MINOR 4

#define PCAP_ERRBUF_SIZE 256

/*
 * Compatibility for systems that have a bpf.h that
 * predates the bpf typedefs for 64-bit support.
 */
#if BPF_RELEASE - 0 < 199406
typedef	int bpf_int32;
typedef	u_int bpf_u_int32;
#endif

typedef struct pcap pcap_t;
typedef struct pcap_dumper pcap_dumper_t;
typedef struct pcap_if pcap_if_t;
typedef struct pcap_addr pcap_addr_t;

/*
 * The first record in the file contains saved values for some
 * of the flags used in the printout phases of tcpdump.
 * Many fields here are 32 bit ints so compilers won't insert unwanted
 * padding; these files need to be interchangeable across architectures.
 *
 * Do not change the layout of this structure, in any way (this includes
 * changes that only affect the length of fields in this structure).
 *
 * Also, do not change the interpretation of any of the members of this
 * structure, in any way (this includes using values other than
 * LINKTYPE_ values, as defined in "savefile.c", in the "linktype"
 * field).
 *
 * Instead:
 *
 *	introduce a new structure for the new format, if the layout
 *	of the structure changed;
 *
 *	send mail to "tcpdump-workers@lists.tcpdump.org", requesting
 *	a new magic number for your new capture file format, and, when
 *	you get the new magic number, put it in "savefile.c";
 *
 *	use that magic number for save files with the changed file
 *	header;
 *
 *	make the code in "savefile.c" capable of reading files with
 *	the old file header as well as files with the new file header
 *	(using the magic number to determine the header format).
 *
 * Then supply the changes by forking the branch at
 *
 *	https://github.com/the-tcpdump-group/libpcap/issues
 *
 * and issuing a pull request, so that future versions of libpcap and
 * programs that use it (such as tcpdump) will be able to read your new
 * capture file format.
 */
struct pcap_file_header {
	bpf_u_int32 magic;
	u_short version_major;
	u_short version_minor;
	bpf_int32 thiszone;	/* gmt to local correction */
	bpf_u_int32 sigfigs;	/* accuracy of timestamps */
	bpf_u_int32 snaplen;	/* max length saved portion of each pkt */
	bpf_u_int32 linktype;	/* data link type (LINKTYPE_*) */
};

/*
 * Macros for the value returned by pcap_datalink_ext().
 *
 * If LT_FCS_LENGTH_PRESENT(x) is true, the LT_FCS_LENGTH(x) macro
 * gives the FCS length of packets in the capture.
 */
#define LT_FCS_LENGTH_PRESENT(x)	((x) & 0x04000000)
#define LT_FCS_LENGTH(x)		(((x) & 0xF0000000) >> 28)
#define LT_FCS_DATALINK_EXT(x)		((((x) & 0xF) << 28) | 0x04000000)

typedef enum {
       PCAP_D_INOUT = 0,
       PCAP_D_IN,
       PCAP_D_OUT
} pcap_direction_t;

/*
 * Generic per-packet information, as supplied by libpcap.
 *
 * The time stamp can and should be a "struct timeval", regardless of
 * whether your system supports 32-bit tv_sec in "struct timeval",
 * 64-bit tv_sec in "struct timeval", or both if it supports both 32-bit
 * and 64-bit applications.  The on-disk format of savefiles uses 32-bit
 * tv_sec (and tv_usec); this structure is irrelevant to that.  32-bit
 * and 64-bit versions of libpcap, even if they're on the same platform,
 * should supply the appropriate version of "struct timeval", even if
 * that's not what the underlying packet capture mechanism supplies.
 */
struct pcap_pkthdr {
	struct timeval ts;	/* time stamp */
	bpf_u_int32 caplen;	/* length of portion present */
	bpf_u_int32 len;	/* length this packet (off wire) */
};

/*
 * As returned by the pcap_stats()
 */
struct pcap_stat {
	u_int ps_recv;		/* number of packets received */
	u_int ps_drop;		/* number of packets dropped */
	u_int ps_ifdrop;	/* drops by interface -- only supported on some platforms */
#if defined(_WIN32) && defined(HAVE_REMOTE)
	u_int ps_capt;		/* number of packets that reach the application */
	u_int ps_sent;		/* number of packets sent by the server on the network */
	u_int ps_netdrop;	/* number of packets lost on the network */
#endif /* _WIN32 && HAVE_REMOTE */
};

#ifdef MSDOS
/*
 * As returned by the pcap_stats_ex()
 */
struct pcap_stat_ex {
       u_long  rx_packets;        /* total packets received       */
       u_long  tx_packets;        /* total packets transmitted    */
       u_long  rx_bytes;          /* total bytes received         */
       u_long  tx_bytes;          /* total bytes transmitted      */
       u_long  rx_errors;         /* bad packets received         */
       u_long  tx_errors;         /* packet transmit problems     */
       u_long  rx_dropped;        /* no space in Rx buffers       */
       u_long  tx_dropped;        /* no space available for Tx    */
       u_long  multicast;         /* multicast packets received   */
       u_long  collisions;

       /* detailed rx_errors: */
       u_long  rx_length_errors;
       u_long  rx_over_errors;    /* receiver ring buff overflow  */
       u_long  rx_crc_errors;     /* recv'd pkt with crc error    */
       u_long  rx_frame_errors;   /* recv'd frame alignment error */
       u_long  rx_fifo_errors;    /* recv'r fifo overrun          */
       u_long  rx_missed_errors;  /* recv'r missed packet         */

       /* detailed tx_errors */
       u_long  tx_aborted_errors;
       u_long  tx_carrier_errors;
       u_long  tx_fifo_errors;
       u_long  tx_heartbeat_errors;
       u_long  tx_window_errors;
     };
#endif

/*
 * Item in a list of interfaces.
 */
struct pcap_if {
	struct pcap_if *next;
	char *name;		/* name to hand to "pcap_open_live()" */
	char *description;	/* textual description of interface, or NULL */
	struct pcap_addr *addresses;
	bpf_u_int32 flags;	/* PCAP_IF_ interface flags */
};

#define PCAP_IF_LOOPBACK	0x00000001	/* interface is loopback */
#define PCAP_IF_UP		0x00000002	/* interface is up */
#define PCAP_IF_RUNNING		0x00000004	/* interface is running */

/*
 * Representation of an interface address.
 */
struct pcap_addr {
	struct pcap_addr *next;
	struct sockaddr *addr;		/* address */
	struct sockaddr *netmask;	/* netmask for that address */
	struct sockaddr *broadaddr;	/* broadcast address for that address */
	struct sockaddr *dstaddr;	/* P2P destination address for that address */
};

typedef void (*pcap_handler)(u_char *, const struct pcap_pkthdr *,
			     const u_char *);

/*
 * Error codes for the pcap API.
 * These will all be negative, so you can check for the success or
 * failure of a call that returns these codes by checking for a
 * negative value.
 */
#define PCAP_ERROR			-1	/* generic error code */
#define PCAP_ERROR_BREAK		-2	/* loop terminated by pcap_breakloop */
#define PCAP_ERROR_NOT_ACTIVATED	-3	/* the capture needs to be activated */
#define PCAP_ERROR_ACTIVATED		-4	/* the operation can't be performed on already activated captures */
#define PCAP_ERROR_NO_SUCH_DEVICE	-5	/* no such device exists */
#define PCAP_ERROR_RFMON_NOTSUP		-6	/* this device doesn't support rfmon (monitor) mode */
#define PCAP_ERROR_NOT_RFMON		-7	/* operation supported only in monitor mode */
#define PCAP_ERROR_PERM_DENIED		-8	/* no permission to open the device */
#define PCAP_ERROR_IFACE_NOT_UP		-9	/* interface isn't up */
#define PCAP_ERROR_CANTSET_TSTAMP_TYPE	-10	/* this device doesn't support setting the time stamp type */
#define PCAP_ERROR_PROMISC_PERM_DENIED	-11	/* you don't have permission to capture in promiscuous mode */
#define PCAP_ERROR_TSTAMP_PRECISION_NOTSUP -12  /* the requested time stamp precision is not supported */

/*
 * Warning codes for the pcap API.
 * These will all be positive and non-zero, so they won't look like
 * errors.
 */
#define PCAP_WARNING			1	/* generic warning code */
#define PCAP_WARNING_PROMISC_NOTSUP	2	/* this device doesn't support promiscuous mode */
#define PCAP_WARNING_TSTAMP_TYPE_NOTSUP	3	/* the requested time stamp type is not supported */

/*
 * Value to pass to pcap_compile() as the netmask if you don't know what
 * the netmask is.
 */
#define PCAP_NETMASK_UNKNOWN	0xffffffff

PCAP_API char	*pcap_lookupdev(char *);
PCAP_API int	pcap_lookupnet(const char *, bpf_u_int32 *, bpf_u_int32 *, char *);

PCAP_API pcap_t	*pcap_create(const char *, char *);
PCAP_API int	pcap_set_snaplen(pcap_t *, int);
PCAP_API int	pcap_set_promisc(pcap_t *, int);
PCAP_API int	pcap_can_set_rfmon(pcap_t *);
PCAP_API int	pcap_set_rfmon(pcap_t *, int);
PCAP_API int	pcap_set_timeout(pcap_t *, int);
PCAP_API int	pcap_set_tstamp_type(pcap_t *, int);
PCAP_API int	pcap_set_immediate_mode(pcap_t *, int);
PCAP_API int	pcap_set_buffer_size(pcap_t *, int);
PCAP_API int	pcap_set_tstamp_precision(pcap_t *, int);
PCAP_API int	pcap_get_tstamp_precision(pcap_t *);
PCAP_API int	pcap_activate(pcap_t *);

PCAP_API int	pcap_list_tstamp_types(pcap_t *, int **);
PCAP_API void	pcap_free_tstamp_types(int *);
PCAP_API int	pcap_tstamp_type_name_to_val(const char *);
PCAP_API const char *pcap_tstamp_type_val_to_name(int);
PCAP_API const char *pcap_tstamp_type_val_to_description(int);

/*
 * Time stamp types.
 * Not all systems and interfaces will necessarily support all of these.
 *
 * A system that supports PCAP_TSTAMP_HOST is offering time stamps
 * provided by the host machine, rather than by the capture device,
 * but not committing to any characteristics of the time stamp;
 * it will not offer any of the PCAP_TSTAMP_HOST_ subtypes.
 *
 * PCAP_TSTAMP_HOST_LOWPREC is a time stamp, provided by the host machine,
 * that's low-precision but relatively cheap to fetch; it's normally done
 * using the system clock, so it's normally synchronized with times you'd
 * fetch from system calls.
 *
 * PCAP_TSTAMP_HOST_HIPREC is a time stamp, provided by the host machine,
 * that's high-precision; it might be more expensive to fetch.  It might
 * or might not be synchronized with the system clock, and might have
 * problems with time stamps for packets received on different CPUs,
 * depending on the platform.
 *
 * PCAP_TSTAMP_ADAPTER is a high-precision time stamp supplied by the
 * capture device; it's synchronized with the system clock.
 *
 * PCAP_TSTAMP_ADAPTER_UNSYNCED is a high-precision time stamp supplied by
 * the capture device; it's not synchronized with the system clock.
 *
 * Note that time stamps synchronized with the system clock can go
 * backwards, as the system clock can go backwards.  If a clock is
 * not in sync with the system clock, that could be because the
 * system clock isn't keeping accurate time, because the other
 * clock isn't keeping accurate time, or both.
 *
 * Note that host-provided time stamps generally correspond to the
 * time when the time-stamping code sees the packet; this could
 * be some unknown amount of time after the first or last bit of
 * the packet is received by the network adapter, due to batching
 * of interrupts for packet arrival, queueing delays, etc..
 */
#define PCAP_TSTAMP_HOST		0	/* host-provided, unknown characteristics */
#define PCAP_TSTAMP_HOST_LOWPREC	1	/* host-provided, low precision */
#define PCAP_TSTAMP_HOST_HIPREC		2	/* host-provided, high precision */
#define PCAP_TSTAMP_ADAPTER		3	/* device-provided, synced with the system clock */
#define PCAP_TSTAMP_ADAPTER_UNSYNCED	4	/* device-provided, not synced with the system clock */

/*
 * Time stamp resolution types.
 * Not all systems and interfaces will necessarily support all of these
 * resolutions when doing live captures; all of them can be requested
 * when reading a savefile.
 */
#define PCAP_TSTAMP_PRECISION_MICRO	0	/* use timestamps with microsecond precision, default */
#define PCAP_TSTAMP_PRECISION_NANO	1	/* use timestamps with nanosecond precision */

PCAP_API pcap_t	*pcap_open_live(const char *, int, int, int, char *);
PCAP_API pcap_t	*pcap_open_dead(int, int);
PCAP_API pcap_t	*pcap_open_dead_with_tstamp_precision(int, int, u_int);
PCAP_API pcap_t	*pcap_open_offline_with_tstamp_precision(const char *, u_int, char *);
PCAP_API pcap_t	*pcap_open_offline(const char *, char *);
#ifdef _WIN32
  PCAP_API pcap_t  *pcap_hopen_offline_with_tstamp_precision(intptr_t, u_int, char *);
  PCAP_API pcap_t  *pcap_hopen_offline(intptr_t, char *);
  /*
   * If we're building libpcap, these are internal routines in savefile.c,
   * so we mustn't define them as macros.
   */
  #ifndef BUILDING_PCAP
    #define pcap_fopen_offline_with_tstamp_precision(f,p,b) \
	pcap_hopen_offline_with_tstamp_precision(_get_osfhandle(_fileno(f)), p, b)
    #define pcap_fopen_offline(f,b) \
	pcap_hopen_offline(_get_osfhandle(_fileno(f)), b)
  #endif
#else /*_WIN32*/
  PCAP_API pcap_t	*pcap_fopen_offline_with_tstamp_precision(FILE *, u_int, char *);
  PCAP_API pcap_t	*pcap_fopen_offline(FILE *, char *);
#endif /*_WIN32*/

PCAP_API void	pcap_close(pcap_t *);
PCAP_API int	pcap_loop(pcap_t *, int, pcap_handler, u_char *);
PCAP_API int	pcap_dispatch(pcap_t *, int, pcap_handler, u_char *);
PCAP_API const u_char *pcap_next(pcap_t *, struct pcap_pkthdr *);
PCAP_API int 	pcap_next_ex(pcap_t *, struct pcap_pkthdr **, const u_char **);
PCAP_API void	pcap_breakloop(pcap_t *);
PCAP_API int	pcap_stats(pcap_t *, struct pcap_stat *);
PCAP_API int	pcap_setfilter(pcap_t *, struct bpf_program *);
PCAP_API int 	pcap_setdirection(pcap_t *, pcap_direction_t);
PCAP_API int	pcap_getnonblock(pcap_t *, char *);
PCAP_API int	pcap_setnonblock(pcap_t *, int, char *);
PCAP_API int	pcap_inject(pcap_t *, const void *, size_t);
PCAP_API int	pcap_sendpacket(pcap_t *, const u_char *, int);
PCAP_API const char *pcap_statustostr(int);
PCAP_API const char *pcap_strerror(int);
PCAP_API char	*pcap_geterr(pcap_t *);
PCAP_API void	pcap_perror(pcap_t *, const char *);
PCAP_API int	pcap_compile(pcap_t *, struct bpf_program *, const char *, int,
	    bpf_u_int32);
PCAP_API int	pcap_compile_nopcap(int, int, struct bpf_program *,
	    const char *, int, bpf_u_int32);
PCAP_API void	pcap_freecode(struct bpf_program *);
PCAP_API int	pcap_offline_filter(const struct bpf_program *,
	    const struct pcap_pkthdr *, const u_char *);
PCAP_API int	pcap_datalink(pcap_t *);
PCAP_API int	pcap_datalink_ext(pcap_t *);
PCAP_API int	pcap_list_datalinks(pcap_t *, int **);
PCAP_API int	pcap_set_datalink(pcap_t *, int);
PCAP_API void	pcap_free_datalinks(int *);
PCAP_API int	pcap_datalink_name_to_val(const char *);
PCAP_API const char *pcap_datalink_val_to_name(int);
PCAP_API const char *pcap_datalink_val_to_description(int);
PCAP_API int	pcap_snapshot(pcap_t *);
PCAP_API int	pcap_is_swapped(pcap_t *);
PCAP_API int	pcap_major_version(pcap_t *);
PCAP_API int	pcap_minor_version(pcap_t *);

/* XXX */
PCAP_API FILE	*pcap_file(pcap_t *);
PCAP_API int	pcap_fileno(pcap_t *);

#ifdef _WIN32
  PCAP_API int	pcap_wsockinit(void);
#endif

PCAP_API pcap_dumper_t *pcap_dump_open(pcap_t *, const char *);
PCAP_API pcap_dumper_t *pcap_dump_fopen(pcap_t *, FILE *fp);
PCAP_API pcap_dumper_t *pcap_dump_open_append(pcap_t *, const char *);
PCAP_API FILE	*pcap_dump_file(pcap_dumper_t *);
PCAP_API long	pcap_dump_ftell(pcap_dumper_t *);
PCAP_API int	pcap_dump_flush(pcap_dumper_t *);
PCAP_API void	pcap_dump_close(pcap_dumper_t *);
PCAP_API void	pcap_dump(u_char *, const struct pcap_pkthdr *, const u_char *);

PCAP_API int	pcap_findalldevs(pcap_if_t **, char *);
PCAP_API void	pcap_freealldevs(pcap_if_t *);

PCAP_API const char *pcap_lib_version(void);

/*
 * On at least some versions of NetBSD and QNX, we don't want to declare
 * bpf_filter() here, as it's also be declared in <net/bpf.h>, with a
 * different signature, but, on other BSD-flavored UN*Xes, it's not
 * declared in <net/bpf.h>, so we *do* want to declare it here, so it's
 * declared when we build pcap-bpf.c.
 */
#if !defined(__NetBSD__) && !defined(__QNX__)
  PCAP_API u_int	bpf_filter(const struct bpf_insn *, const u_char *, u_int, u_int);
#endif
PCAP_API int	bpf_validate(const struct bpf_insn *f, int len);
PCAP_API char	*bpf_image(const struct bpf_insn *, int);
PCAP_API void	bpf_dump(const struct bpf_program *, int);

#if defined(_WIN32)

  /*
   * Win32 definitions
   */

  /*!
    \brief A queue of raw packets that will be sent to the network with pcap_sendqueue_transmit().
  */
  struct pcap_send_queue
  {
	u_int maxlen;	/* Maximum size of the the queue, in bytes. This
			   variable contains the size of the buffer field. */
	u_int len;	/* Current size of the queue, in bytes. */
	char *buffer;	/* Buffer containing the packets to be sent. */
  };

  typedef struct pcap_send_queue pcap_send_queue;

  /*!
    \brief This typedef is a support for the pcap_get_airpcap_handle() function
  */
  #if !defined(AIRPCAP_HANDLE__EAE405F5_0171_9592_B3C2_C19EC426AD34__DEFINED_)
    #define AIRPCAP_HANDLE__EAE405F5_0171_9592_B3C2_C19EC426AD34__DEFINED_
    typedef struct _AirpcapHandle *PAirpcapHandle;
  #endif

  PCAP_API int pcap_setbuff(pcap_t *p, int dim);
  PCAP_API int pcap_setmode(pcap_t *p, int mode);
  PCAP_API int pcap_setmintocopy(pcap_t *p, int size);

  PCAP_API HANDLE pcap_getevent(pcap_t *p);

  PCAP_API int pcap_oid_get_request(pcap_t *, bpf_u_int32, void *, size_t *);
  PCAP_API int pcap_oid_set_request(pcap_t *, bpf_u_int32, const void *, size_t *);

  PCAP_API pcap_send_queue* pcap_sendqueue_alloc(u_int memsize);

  PCAP_API void pcap_sendqueue_destroy(pcap_send_queue* queue);

  PCAP_API int pcap_sendqueue_queue(pcap_send_queue* queue, const struct pcap_pkthdr *pkt_header, const u_char *pkt_data);

  PCAP_API u_int pcap_sendqueue_transmit(pcap_t *p, pcap_send_queue* queue, int sync);

  PCAP_API struct pcap_stat *pcap_stats_ex(pcap_t *p, int *pcap_stat_size);

  PCAP_API int pcap_setuserbuffer(pcap_t *p, int size);

  PCAP_API int pcap_live_dump(pcap_t *p, char *filename, int maxsize, int maxpacks);

  PCAP_API int pcap_live_dump_ended(pcap_t *p, int sync);

  PCAP_API int pcap_start_oem(char* err_str, int flags);

  PCAP_API PAirpcapHandle pcap_get_airpcap_handle(pcap_t *p);

  #define MODE_CAPT 0
  #define MODE_STAT 1
  #define MODE_MON 2

#elif defined(MSDOS)

  /*
   * MS-DOS definitions
   */

  PCAP_API int  pcap_stats_ex (pcap_t *, struct pcap_stat_ex *);
  PCAP_API void pcap_set_wait (pcap_t *p, void (*yield)(void), int wait);
  PCAP_API u_long pcap_mac_packets (void);

#else /* UN*X */

  /*
   * UN*X definitions
   */

  PCAP_API int	pcap_get_selectable_fd(pcap_t *);

#endif /* _WIN32/MSDOS/UN*X */

#ifdef HAVE_REMOTE
  /* Includes most of the public stuff that is needed for the remote capture */
  #include <remote-ext.h>
#endif	 /* HAVE_REMOTE */

#ifdef __cplusplus
}
#endif

#endif /* lib_pcap_pcap_h */
