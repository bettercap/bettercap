/*
 * Copyright (c) 1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997, 1998
 *	The Regents of the University of California.  All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that: (1) source code distributions
 * retain the above copyright notice and this paragraph in its entirety, (2)
 * distributions including binary code include the above copyright notice and
 * this paragraph in its entirety in the documentation or other materials
 * provided with the distribution, and (3) all advertising materials mentioning
 * features or use of this software display the following acknowledgement:
 * ``This product includes software developed by the University of California,
 * Lawrence Berkeley Laboratory and its contributors.'' Neither the name of
 * the University nor the names of its contributors may be used to endorse
 * or promote products derived from this software without specific prior
 * written permission.
 * THIS SOFTWARE IS PROVIDED ``AS IS'' AND WITHOUT ANY EXPRESS OR IMPLIED
 * WARRANTIES, INCLUDING, WITHOUT LIMITATION, THE IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE.
 *
 * Name to id translation routines used by the scanner.
 * These functions are not time critical.
 */

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#ifdef DECNETLIB
#include <sys/types.h>
#include <netdnet/dnetdb.h>
#endif

#ifdef _WIN32
#include <pcap-stdinc.h>

#ifdef INET6
/*
 * To quote the MSDN page for getaddrinfo() at
 *
 *    https://msdn.microsoft.com/en-us/library/windows/desktop/ms738520(v=vs.85).aspx
 *
 * "Support for getaddrinfo on Windows 2000 and older versions
 * The getaddrinfo function was added to the Ws2_32.dll on Windows XP and
 * later. To execute an application that uses this function on earlier
 * versions of Windows, then you need to include the Ws2tcpip.h and
 * Wspiapi.h files. When the Wspiapi.h include file is added, the
 * getaddrinfo function is defined to the WspiapiGetAddrInfo inline
 * function in the Wspiapi.h file. At runtime, the WspiapiGetAddrInfo
 * function is implemented in such a way that if the Ws2_32.dll or the
 * Wship6.dll (the file containing getaddrinfo in the IPv6 Technology
 * Preview for Windows 2000) does not include getaddrinfo, then a
 * version of getaddrinfo is implemented inline based on code in the
 * Wspiapi.h header file. This inline code will be used on older Windows
 * platforms that do not natively support the getaddrinfo function."
 *
 * We use getaddrinfo(), so we include Wspiapi.h here.  pcap-stdinc.h
 * includes Ws2tcpip.h, so we don't need to include it ourselves.
 */
#include <Wspiapi.h>
#endif

#else /* _WIN32 */

#include <sys/param.h>
#include <sys/types.h>				/* concession to AIX */
#include <sys/socket.h>
#include <sys/time.h>

#include <netinet/in.h>
#endif /* _WIN32 */

#ifndef _WIN32
#ifdef HAVE_ETHER_HOSTTON
/*
 * XXX - do we need any of this if <netinet/if_ether.h> doesn't declare
 * ether_hostton()?
 */
#ifdef HAVE_NETINET_IF_ETHER_H
struct mbuf;		/* Squelch compiler warnings on some platforms for */
struct rtentry;		/* declarations in <net/if.h> */
#include <net/if.h>	/* for "struct ifnet" in "struct arpcom" on Solaris */
#include <netinet/if_ether.h>
#endif /* HAVE_NETINET_IF_ETHER_H */
#ifdef NETINET_ETHER_H_DECLARES_ETHER_HOSTTON
#include <netinet/ether.h>
#endif /* NETINET_ETHER_H_DECLARES_ETHER_HOSTTON */
#endif /* HAVE_ETHER_HOSTTON */
#include <arpa/inet.h>
#include <netdb.h>
#endif /* _WIN32 */

#include <ctype.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#include "pcap-int.h"

#include "gencode.h"
#include <pcap/namedb.h>
#include "nametoaddr.h"

#ifdef HAVE_OS_PROTO_H
#include "os-proto.h"
#endif

#ifndef NTOHL
#define NTOHL(x) (x) = ntohl(x)
#define NTOHS(x) (x) = ntohs(x)
#endif

static inline int xdtoi(int);

/*
 *  Convert host name to internet address.
 *  Return 0 upon failure.
 */
bpf_u_int32 **
pcap_nametoaddr(const char *name)
{
#ifndef h_addr
	static bpf_u_int32 *hlist[2];
#endif
	bpf_u_int32 **p;
	struct hostent *hp;

	if ((hp = gethostbyname(name)) != NULL) {
#ifndef h_addr
		hlist[0] = (bpf_u_int32 *)hp->h_addr;
		NTOHL(hp->h_addr);
		return hlist;
#else
		for (p = (bpf_u_int32 **)hp->h_addr_list; *p; ++p)
			NTOHL(**p);
		return (bpf_u_int32 **)hp->h_addr_list;
#endif
	}
	else
		return 0;
}

#ifdef INET6
struct addrinfo *
pcap_nametoaddrinfo(const char *name)
{
	struct addrinfo hints, *res;
	int error;

	memset(&hints, 0, sizeof(hints));
	hints.ai_family = PF_UNSPEC;
	hints.ai_socktype = SOCK_STREAM;	/*not really*/
	hints.ai_protocol = IPPROTO_TCP;	/*not really*/
	error = getaddrinfo(name, NULL, &hints, &res);
	if (error)
		return NULL;
	else
		return res;
}
#endif /*INET6*/

/*
 *  Convert net name to internet address.
 *  Return 0 upon failure.
 */
bpf_u_int32
pcap_nametonetaddr(const char *name)
{
#ifndef _WIN32
	struct netent *np;

	if ((np = getnetbyname(name)) != NULL)
		return np->n_net;
	else
		return 0;
#else
	/*
	 * There's no "getnetbyname()" on Windows.
	 *
	 * XXX - I guess we could use the BSD code to read
	 * C:\Windows\System32\drivers\etc/networks, assuming
	 * that's its home on all the versions of Windows
	 * we use, but that file probably just has the loopback
	 * network on 127/24 on 99 44/100% of Windows machines.
	 *
	 * (Heck, these days it probably just has that on 99 44/100%
	 * of *UN*X* machines.)
	 */
	return 0;
#endif
}

/*
 * Convert a port name to its port and protocol numbers.
 * We assume only TCP or UDP.
 * Return 0 upon failure.
 */
int
pcap_nametoport(const char *name, int *port, int *proto)
{
	struct servent *sp;
	int tcp_port = -1;
	int udp_port = -1;

	/*
	 * We need to check /etc/services for ambiguous entries.
	 * If we find the ambiguous entry, and it has the
	 * same port number, change the proto to PROTO_UNDEF
	 * so both TCP and UDP will be checked.
	 */
	sp = getservbyname(name, "tcp");
	if (sp != NULL) tcp_port = ntohs(sp->s_port);
	sp = getservbyname(name, "udp");
	if (sp != NULL) udp_port = ntohs(sp->s_port);
	if (tcp_port >= 0) {
		*port = tcp_port;
		*proto = IPPROTO_TCP;
		if (udp_port >= 0) {
			if (udp_port == tcp_port)
				*proto = PROTO_UNDEF;
#ifdef notdef
			else
				/* Can't handle ambiguous names that refer
				   to different port numbers. */
				warning("ambiguous port %s in /etc/services",
					name);
#endif
		}
		return 1;
	}
	if (udp_port >= 0) {
		*port = udp_port;
		*proto = IPPROTO_UDP;
		return 1;
	}
#if defined(ultrix) || defined(__osf__)
	/* Special hack in case NFS isn't in /etc/services */
	if (strcmp(name, "nfs") == 0) {
		*port = 2049;
		*proto = PROTO_UNDEF;
		return 1;
	}
#endif
	return 0;
}

/*
 * Convert a string in the form PPP-PPP, where correspond to ports, to
 * a starting and ending port in a port range.
 * Return 0 on failure.
 */
int
pcap_nametoportrange(const char *name, int *port1, int *port2, int *proto)
{
	u_int p1, p2;
	char *off, *cpy;
	int save_proto;

	if (sscanf(name, "%d-%d", &p1, &p2) != 2) {
		if ((cpy = strdup(name)) == NULL)
			return 0;

		if ((off = strchr(cpy, '-')) == NULL) {
			free(cpy);
			return 0;
		}

		*off = '\0';

		if (pcap_nametoport(cpy, port1, proto) == 0) {
			free(cpy);
			return 0;
		}
		save_proto = *proto;

		if (pcap_nametoport(off + 1, port2, proto) == 0) {
			free(cpy);
			return 0;
		}
		free(cpy);

		if (*proto != save_proto)
			*proto = PROTO_UNDEF;
	} else {
		*port1 = p1;
		*port2 = p2;
		*proto = PROTO_UNDEF;
	}

	return 1;
}

int
pcap_nametoproto(const char *str)
{
	struct protoent *p;

	p = getprotobyname(str);
	if (p != 0)
		return p->p_proto;
	else
		return PROTO_UNDEF;
}

#include "ethertype.h"

struct eproto {
	const char *s;
	u_short p;
};

/*
 * Static data base of ether protocol types.
 * tcpdump used to import this, and it's declared as an export on
 * Debian, at least, so make it a public symbol, even though we
 * don't officially export it by declaring it in a header file.
 * (Programs *should* do this themselves, as tcpdump now does.)
 */
PCAP_API_DEF struct eproto eproto_db[] = {
	{ "pup", ETHERTYPE_PUP },
	{ "xns", ETHERTYPE_NS },
	{ "ip", ETHERTYPE_IP },
#ifdef INET6
	{ "ip6", ETHERTYPE_IPV6 },
#endif
	{ "arp", ETHERTYPE_ARP },
	{ "rarp", ETHERTYPE_REVARP },
	{ "sprite", ETHERTYPE_SPRITE },
	{ "mopdl", ETHERTYPE_MOPDL },
	{ "moprc", ETHERTYPE_MOPRC },
	{ "decnet", ETHERTYPE_DN },
	{ "lat", ETHERTYPE_LAT },
	{ "sca", ETHERTYPE_SCA },
	{ "lanbridge", ETHERTYPE_LANBRIDGE },
	{ "vexp", ETHERTYPE_VEXP },
	{ "vprod", ETHERTYPE_VPROD },
	{ "atalk", ETHERTYPE_ATALK },
	{ "atalkarp", ETHERTYPE_AARP },
	{ "loopback", ETHERTYPE_LOOPBACK },
	{ "decdts", ETHERTYPE_DECDTS },
	{ "decdns", ETHERTYPE_DECDNS },
	{ (char *)0, 0 }
};

int
pcap_nametoeproto(const char *s)
{
	struct eproto *p = eproto_db;

	while (p->s != 0) {
		if (strcmp(p->s, s) == 0)
			return p->p;
		p += 1;
	}
	return PROTO_UNDEF;
}

#include "llc.h"

/* Static data base of LLC values. */
static struct eproto llc_db[] = {
	{ "iso", LLCSAP_ISONS },
	{ "stp", LLCSAP_8021D },
	{ "ipx", LLCSAP_IPX },
	{ "netbeui", LLCSAP_NETBEUI },
	{ (char *)0, 0 }
};

int
pcap_nametollc(const char *s)
{
	struct eproto *p = llc_db;

	while (p->s != 0) {
		if (strcmp(p->s, s) == 0)
			return p->p;
		p += 1;
	}
	return PROTO_UNDEF;
}

/* Hex digit to integer. */
static inline int
xdtoi(c)
	register int c;
{
	if (isdigit(c))
		return c - '0';
	else if (islower(c))
		return c - 'a' + 10;
	else
		return c - 'A' + 10;
}

int
__pcap_atoin(const char *s, bpf_u_int32 *addr)
{
	u_int n;
	int len;

	*addr = 0;
	len = 0;
	while (1) {
		n = 0;
		while (*s && *s != '.')
			n = n * 10 + *s++ - '0';
		*addr <<= 8;
		*addr |= n & 0xff;
		len += 8;
		if (*s == '\0')
			return len;
		++s;
	}
	/* NOTREACHED */
}

int
__pcap_atodn(const char *s, bpf_u_int32 *addr)
{
#define AREASHIFT 10
#define AREAMASK 0176000
#define NODEMASK 01777

	u_int node, area;

	if (sscanf(s, "%d.%d", &area, &node) != 2)
		return(0);

	*addr = (area << AREASHIFT) & AREAMASK;
	*addr |= (node & NODEMASK);

	return(32);
}

/*
 * Convert 's', which can have the one of the forms:
 *
 *	"xx:xx:xx:xx:xx:xx"
 *	"xx.xx.xx.xx.xx.xx"
 *	"xx-xx-xx-xx-xx-xx"
 *	"xxxx.xxxx.xxxx"
 *	"xxxxxxxxxxxx"
 *
 * (or various mixes of ':', '.', and '-') into a new
 * ethernet address.  Assumes 's' is well formed.
 */
u_char *
pcap_ether_aton(const char *s)
{
	register u_char *ep, *e;
	register u_int d;

	e = ep = (u_char *)malloc(6);
	if (e == NULL)
		return (NULL);

	while (*s) {
		if (*s == ':' || *s == '.' || *s == '-')
			s += 1;
		d = xdtoi(*s++);
		if (isxdigit((unsigned char)*s)) {
			d <<= 4;
			d |= xdtoi(*s++);
		}
		*ep++ = d;
	}

	return (e);
}

#ifndef HAVE_ETHER_HOSTTON
/* Roll our own */
u_char *
pcap_ether_hostton(const char *name)
{
	register struct pcap_etherent *ep;
	register u_char *ap;
	static FILE *fp = NULL;
	static int init = 0;

	if (!init) {
		fp = fopen(PCAP_ETHERS_FILE, "r");
		++init;
		if (fp == NULL)
			return (NULL);
	} else if (fp == NULL)
		return (NULL);
	else
		rewind(fp);

	while ((ep = pcap_next_etherent(fp)) != NULL) {
		if (strcmp(ep->name, name) == 0) {
			ap = (u_char *)malloc(6);
			if (ap != NULL) {
				memcpy(ap, ep->addr, 6);
				return (ap);
			}
			break;
		}
	}
	return (NULL);
}
#else

#if !defined(HAVE_DECL_ETHER_HOSTTON) || !HAVE_DECL_ETHER_HOSTTON
#ifndef HAVE_STRUCT_ETHER_ADDR
struct ether_addr {
	unsigned char ether_addr_octet[6];
};
#endif
extern int ether_hostton(const char *, struct ether_addr *);
#endif

/* Use the os supplied routines */
u_char *
pcap_ether_hostton(const char *name)
{
	register u_char *ap;
	u_char a[6];

	ap = NULL;
	if (ether_hostton(name, (struct ether_addr *)a) == 0) {
		ap = (u_char *)malloc(6);
		if (ap != NULL)
			memcpy((char *)ap, (char *)a, 6);
	}
	return (ap);
}
#endif

int
__pcap_nametodnaddr(const char *name, u_short *res)
{
#ifdef	DECNETLIB
	struct nodeent *getnodebyname();
	struct nodeent *nep;

	nep = getnodebyname(name);
	if (nep == ((struct nodeent *)0))
		return(0);

	memcpy((char *)res, (char *)nep->n_addr, sizeof(unsigned short));
	return(1);
#else
	return(0);
#endif
}
