/* -*- Mode: c; tab-width: 8; indent-tabs-mode: 1; c-basic-offset: 8; -*- */
/*
 * Copyright (c) 1994, 1995, 1996, 1997, 1998
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

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#ifdef _WIN32
#include <pcap-stdinc.h>
#else /* _WIN32 */

#include <sys/param.h>
#ifndef MSDOS
#include <sys/file.h>
#endif
#include <sys/ioctl.h>
#include <sys/socket.h>
#ifdef HAVE_SYS_SOCKIO_H
#include <sys/sockio.h>
#endif

struct mbuf;		/* Squelch compiler warnings on some platforms for */
struct rtentry;		/* declarations in <net/if.h> */
#include <net/if.h>
#include <netinet/in.h>
#endif /* _WIN32 */

#include <errno.h>
#include <memory.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#if !defined(_WIN32) && !defined(__BORLANDC__)
#include <unistd.h>
#endif /* !_WIN32 && !__BORLANDC__ */

#include "pcap-int.h"

#ifdef HAVE_OS_PROTO_H
#include "os-proto.h"
#endif

#if !defined(_WIN32) && !defined(MSDOS)

/*
 * Return the name of a network interface attached to the system, or NULL
 * if none can be found.  The interface must be configured up; the
 * lowest unit number is preferred; loopback is ignored.
 */
char *
pcap_lookupdev(errbuf)
	register char *errbuf;
{
	pcap_if_t *alldevs;
/* for old BSD systems, including bsdi3 */
#ifndef IF_NAMESIZE
#define IF_NAMESIZE IFNAMSIZ
#endif
	static char device[IF_NAMESIZE + 1];
	char *ret;

	if (pcap_findalldevs(&alldevs, errbuf) == -1)
		return (NULL);

	if (alldevs == NULL || (alldevs->flags & PCAP_IF_LOOPBACK)) {
		/*
		 * There are no devices on the list, or the first device
		 * on the list is a loopback device, which means there
		 * are no non-loopback devices on the list.  This means
		 * we can't return any device.
		 *
		 * XXX - why not return a loopback device?  If we can't
		 * capture on it, it won't be on the list, and if it's
		 * on the list, there aren't any non-loopback devices,
		 * so why not just supply it as the default device?
		 */
		(void)strlcpy(errbuf, "no suitable device found",
		    PCAP_ERRBUF_SIZE);
		ret = NULL;
	} else {
		/*
		 * Return the name of the first device on the list.
		 */
		(void)strlcpy(device, alldevs->name, sizeof(device));
		ret = device;
	}

	pcap_freealldevs(alldevs);
	return (ret);
}

int
pcap_lookupnet(device, netp, maskp, errbuf)
	register const char *device;
	register bpf_u_int32 *netp, *maskp;
	register char *errbuf;
{
	register int fd;
	register struct sockaddr_in *sin4;
	struct ifreq ifr;

	/*
	 * The pseudo-device "any" listens on all interfaces and therefore
	 * has the network address and -mask "0.0.0.0" therefore catching
	 * all traffic. Using NULL for the interface is the same as "any".
	 */
	if (!device || strcmp(device, "any") == 0
#ifdef HAVE_DAG_API
	    || strstr(device, "dag") != NULL
#endif
#ifdef HAVE_SEPTEL_API
	    || strstr(device, "septel") != NULL
#endif
#ifdef PCAP_SUPPORT_BT
	    || strstr(device, "bluetooth") != NULL
#endif
#ifdef PCAP_SUPPORT_USB
	    || strstr(device, "usbmon") != NULL
#endif
#ifdef HAVE_SNF_API
	    || strstr(device, "snf") != NULL
#endif
	    ) {
		*netp = *maskp = 0;
		return 0;
	}

	fd = socket(AF_INET, SOCK_DGRAM, 0);
	if (fd < 0) {
		(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "socket: %s",
		    pcap_strerror(errno));
		return (-1);
	}
	memset(&ifr, 0, sizeof(ifr));
#ifdef linux
	/* XXX Work around Linux kernel bug */
	ifr.ifr_addr.sa_family = AF_INET;
#endif
	(void)strlcpy(ifr.ifr_name, device, sizeof(ifr.ifr_name));
	if (ioctl(fd, SIOCGIFADDR, (char *)&ifr) < 0) {
		if (errno == EADDRNOTAVAIL) {
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			    "%s: no IPv4 address assigned", device);
		} else {
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			    "SIOCGIFADDR: %s: %s",
			    device, pcap_strerror(errno));
		}
		(void)close(fd);
		return (-1);
	}
	sin4 = (struct sockaddr_in *)&ifr.ifr_addr;
	*netp = sin4->sin_addr.s_addr;
	memset(&ifr, 0, sizeof(ifr));
#ifdef linux
	/* XXX Work around Linux kernel bug */
	ifr.ifr_addr.sa_family = AF_INET;
#endif
	(void)strlcpy(ifr.ifr_name, device, sizeof(ifr.ifr_name));
	if (ioctl(fd, SIOCGIFNETMASK, (char *)&ifr) < 0) {
		(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
		    "SIOCGIFNETMASK: %s: %s", device, pcap_strerror(errno));
		(void)close(fd);
		return (-1);
	}
	(void)close(fd);
	*maskp = sin4->sin_addr.s_addr;
	if (*maskp == 0) {
		if (IN_CLASSA(*netp))
			*maskp = IN_CLASSA_NET;
		else if (IN_CLASSB(*netp))
			*maskp = IN_CLASSB_NET;
		else if (IN_CLASSC(*netp))
			*maskp = IN_CLASSC_NET;
		else {
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
			    "inet class for 0x%x unknown", *netp);
			return (-1);
		}
	}
	*netp &= *maskp;
	return (0);
}

#elif defined(_WIN32)

/*
 * Return the name of a network interface attached to the system, or NULL
 * if none can be found.  The interface must be configured up; the
 * lowest unit number is preferred; loopback is ignored.
 *
 * In the best of all possible worlds, this would be the same as on
 * UN*X, but there may be software that expects this to return a
 * full list of devices after the first device.
 */
#define ADAPTERSNAME_LEN	8192
char *
pcap_lookupdev(errbuf)
	register char *errbuf;
{
	DWORD dwVersion;
	DWORD dwWindowsMajorVersion;
	char our_errbuf[PCAP_ERRBUF_SIZE+1];

#pragma warning (push)
#pragma warning (disable: 4996) /* disable MSVC's GetVersion() deprecated warning here */
	dwVersion = GetVersion();	/* get the OS version */
#pragma warning (pop)
	dwWindowsMajorVersion = (DWORD)(LOBYTE(LOWORD(dwVersion)));

	if (dwVersion >= 0x80000000 && dwWindowsMajorVersion >= 4) {
		/*
		 * Windows 95, 98, ME.
		 */
		ULONG NameLength = ADAPTERSNAME_LEN;
		static char AdaptersName[ADAPTERSNAME_LEN];

		if (PacketGetAdapterNames(AdaptersName,&NameLength) )
			return (AdaptersName);
		else
			return NULL;
	} else {
		/*
		 * Windows NT (NT 4.0 and later).
		 * Convert the names to Unicode for backward compatibility.
		 */
		ULONG NameLength = ADAPTERSNAME_LEN;
		static WCHAR AdaptersName[ADAPTERSNAME_LEN];
		size_t BufferSpaceLeft;
		char *tAstr;
		WCHAR *Unameptr;
		char *Adescptr;
		size_t namelen, i;
		WCHAR *TAdaptersName = (WCHAR*)malloc(ADAPTERSNAME_LEN * sizeof(WCHAR));
		int NAdapts = 0;

		if(TAdaptersName == NULL)
		{
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE, "memory allocation failure");
			return NULL;
		}

		if ( !PacketGetAdapterNames((PTSTR)TAdaptersName,&NameLength) )
		{
			pcap_win32_err_to_str(GetLastError(), our_errbuf);
			(void)pcap_snprintf(errbuf, PCAP_ERRBUF_SIZE,
				"PacketGetAdapterNames: %s", our_errbuf);
			free(TAdaptersName);
			return NULL;
		}


		BufferSpaceLeft = ADAPTERSNAME_LEN * sizeof(WCHAR);
		tAstr = (char*)TAdaptersName;
		Unameptr = AdaptersName;

		/*
		 * Convert the device names to Unicode into AdapterName.
		 */
		do {
			/*
			 * Length of the name, including the terminating
			 * NUL.
			 */
			namelen = strlen(tAstr) + 1;

			/*
			 * Do we have room for the name in the Unicode
			 * buffer?
			 */
			if (BufferSpaceLeft < namelen * sizeof(WCHAR)) {
				/*
				 * No.
				 */
				goto quit;
			}
			BufferSpaceLeft -= namelen * sizeof(WCHAR);

			/*
			 * Copy the name, converting ASCII to Unicode.
			 * namelen includes the NUL, so we copy it as
			 * well.
			 */
			for (i = 0; i < namelen; i++)
				*Unameptr++ = *tAstr++;

			/*
			 * Count this adapter.
			 */
			NAdapts++;
		} while (namelen != 1);

		/*
		 * Copy the descriptions, but don't convert them from
		 * ASCII to Unicode.
		 */
		Adescptr = (char *)Unameptr;
		while(NAdapts--)
		{
			size_t desclen;

			desclen = strlen(tAstr) + 1;

			/*
			 * Do we have room for the name in the Unicode
			 * buffer?
			 */
			if (BufferSpaceLeft < desclen) {
				/*
				 * No.
				 */
				goto quit;
			}

			/*
			 * Just copy the ASCII string.
			 * namelen includes the NUL, so we copy it as
			 * well.
			 */
			memcpy(Adescptr, tAstr, desclen);
			Adescptr += desclen;
			tAstr += desclen;
			BufferSpaceLeft -= desclen;
		}

	quit:
		free(TAdaptersName);
		return (char *)(AdaptersName);
	}
}


int
pcap_lookupnet(device, netp, maskp, errbuf)
	register const char *device;
	register bpf_u_int32 *netp, *maskp;
	register char *errbuf;
{
	/*
	 * We need only the first IPv4 address, so we must scan the array returned by PacketGetNetInfo()
	 * in order to skip non IPv4 (i.e. IPv6 addresses)
	 */
	npf_if_addr if_addrs[MAX_NETWORK_ADDRESSES];
	LONG if_addr_size = 1;
	struct sockaddr_in *t_addr;
	unsigned int i;

	if (!PacketGetNetInfoEx((void *)device, if_addrs, &if_addr_size)) {
		*netp = *maskp = 0;
		return (0);
	}

	for(i=0; i<MAX_NETWORK_ADDRESSES; i++)
	{
		if(if_addrs[i].IPAddress.ss_family == AF_INET)
		{
			t_addr = (struct sockaddr_in *) &(if_addrs[i].IPAddress);
			*netp = t_addr->sin_addr.S_un.S_addr;
			t_addr = (struct sockaddr_in *) &(if_addrs[i].SubnetMask);
			*maskp = t_addr->sin_addr.S_un.S_addr;

			*netp &= *maskp;
			return (0);
		}

	}

	*netp = *maskp = 0;
	return (0);
}

#endif /* !_WIN32 && !MSDOS */
