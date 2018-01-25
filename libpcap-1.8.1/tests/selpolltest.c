/*
 * Copyright (c) 1988, 1989, 1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997, 2000
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
 */

#ifndef lint
static const char copyright[] _U_ =
    "@(#) Copyright (c) 1988, 1989, 1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997, 2000\n\
The Regents of the University of California.  All rights reserved.\n";
#endif

/*
 * Tests how select() and poll() behave on the selectable file descriptor
 * for a pcap_t.
 *
 * This would be significantly different on Windows, as it'd test
 * how WaitForMultipleObjects() would work on the event handle for a
 * pcap_t.
 */
#include <pcap.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <unistd.h>
#include <errno.h>
#include <sys/types.h>
#ifdef HAVE_SYS_SELECT_H
#include <sys/select.h>
#else
#include <sys/time.h>	/* older UN*Xes */
#endif
#include <poll.h>

char *program_name;

/*
 * This was introduced by Clang:
 *
 *     http://clang.llvm.org/docs/LanguageExtensions.html#has-attribute
 *
 * in some version (which version?); it has been picked up by GCC 5.0.
 */
#ifndef __has_attribute
  /*
   * It's a macro, so you can check whether it's defined to check
   * whether it's supported.
   *
   * If it's not, define it to always return 0, so that we move on to
   * the fallback checks.
   */
  #define __has_attribute(x) 0
#endif

#if __has_attribute(noreturn) \
    || (defined(__GNUC__) && ((__GNUC__ * 100 + __GNUC_MINOR__) >= 205)) \
    || (defined(__SUNPRO_C) && (__SUNPRO_C >= 0x590)) \
    || (defined(__xlC__) && __xlC__ >= 0x0A01) \
    || (defined(__HP_aCC) && __HP_aCC >= 61000)
  /*
   * Compiler with support for it, or GCC 2.5 and later, or Solaris Studio 12
   * (Sun C 5.9) and later, or IBM XL C 10.1 and later (do any earlier
   * versions of XL C support this?), or HP aCC A.06.10 and later.
   */
  #define PCAP_NORETURN __attribute((noreturn))
#elif defined( _MSC_VER )
  #define PCAP_NORETURN __declspec(noreturn)
#else
  #define PCAP_NORETURN
#endif

#if __has_attribute(__format__) \
    || (defined(__GNUC__) && ((__GNUC__ * 100 + __GNUC_MINOR__) >= 203)) \
    || (defined(__xlC__) && __xlC__ >= 0x0A01) \
    || (defined(__HP_aCC) && __HP_aCC >= 61000)
  /*
   * Compiler with support for it, or GCC 2.3 and later, or IBM XL C 10.1
   * and later (do any earlier versions of XL C support this?),
   * or HP aCC A.06.10 and later.
   */
  #define PCAP_PRINTFLIKE(x,y) __attribute__((__format__(__printf__,x,y)))
#else
  #define PCAP_PRINTFLIKE(x,y)
#endif

/* Forwards */
static void countme(u_char *, const struct pcap_pkthdr *, const u_char *);
static void PCAP_NORETURN usage(void);
static void PCAP_NORETURN error(const char *, ...) PCAP_PRINTFLIKE(1, 2);
static void warning(const char *, ...) PCAP_PRINTFLIKE(1, 2);
static char *copy_argv(char **);

static pcap_t *pd;

int
main(int argc, char **argv)
{
	register int op;
	bpf_u_int32 localnet, netmask;
	register char *cp, *cmdbuf, *device;
	int doselect, dopoll, dotimeout, dononblock;
	struct bpf_program fcode;
	char ebuf[PCAP_ERRBUF_SIZE];
	int selectable_fd;
	int status;
	int packet_count;

	device = NULL;
	doselect = 0;
	dopoll = 0;
	dotimeout = 0;
	dononblock = 0;
	if ((cp = strrchr(argv[0], '/')) != NULL)
		program_name = cp + 1;
	else
		program_name = argv[0];

	opterr = 0;
	while ((op = getopt(argc, argv, "i:sptn")) != -1) {
		switch (op) {

		case 'i':
			device = optarg;
			break;

		case 's':
			doselect = 1;
			break;

		case 'p':
			dopoll = 1;
			break;

		case 't':
			dotimeout = 1;
			break;

		case 'n':
			dononblock = 1;
			break;

		default:
			usage();
			/* NOTREACHED */
		}
	}

	if (doselect && dopoll) {
		fprintf(stderr, "selpolltest: choose select (-s) or poll (-p), but not both\n");
		return 1;
	}
	if (dotimeout && !doselect && !dopoll) {
		fprintf(stderr, "selpolltest: timeout (-t) requires select (-s) or poll (-p)\n");
		return 1;
	}
	if (device == NULL) {
		device = pcap_lookupdev(ebuf);
		if (device == NULL)
			error("%s", ebuf);
	}
	*ebuf = '\0';
	pd = pcap_open_live(device, 65535, 0, 1000, ebuf);
	if (pd == NULL)
		error("%s", ebuf);
	else if (*ebuf)
		warning("%s", ebuf);
	if (pcap_lookupnet(device, &localnet, &netmask, ebuf) < 0) {
		localnet = 0;
		netmask = 0;
		warning("%s", ebuf);
	}
	cmdbuf = copy_argv(&argv[optind]);

	if (pcap_compile(pd, &fcode, cmdbuf, 1, netmask) < 0)
		error("%s", pcap_geterr(pd));

	if (pcap_setfilter(pd, &fcode) < 0)
		error("%s", pcap_geterr(pd));
	if (pcap_get_selectable_fd(pd) == -1)
		error("pcap_get_selectable_fd() fails");
	if (dononblock) {
		if (pcap_setnonblock(pd, 1, ebuf) == -1)
			error("pcap_setnonblock failed: %s", ebuf);
	}
	selectable_fd = pcap_get_selectable_fd(pd);
	printf("Listening on %s\n", device);
	if (doselect) {
		for (;;) {
			fd_set setread, setexcept;
			struct timeval seltimeout;

			FD_ZERO(&setread);
			FD_SET(selectable_fd, &setread);
			FD_ZERO(&setexcept);
			FD_SET(selectable_fd, &setexcept);
			if (dotimeout) {
				seltimeout.tv_sec = 0;
				seltimeout.tv_usec = 1000;
				status = select(selectable_fd + 1, &setread,
				    NULL, &setexcept, &seltimeout);
			} else {
				status = select(selectable_fd + 1, &setread,
				    NULL, &setexcept, NULL);
			}
			if (status == -1) {
				printf("Select returns error (%s)\n",
				    strerror(errno));
			} else {
				if (status == 0)
					printf("Select timed out: ");
				else
					printf("Select returned a descriptor: ");
				if (FD_ISSET(selectable_fd, &setread))
					printf("readable, ");
				else
					printf("not readable, ");
				if (FD_ISSET(selectable_fd, &setexcept))
					printf("exceptional condition\n");
				else
					printf("no exceptional condition\n");
				packet_count = 0;
				status = pcap_dispatch(pd, -1, countme,
				    (u_char *)&packet_count);
				if (status < 0)
					break;
				printf("%d packets seen, %d packets counted after select returns\n",
				    status, packet_count);
			}
		}
	} else if (dopoll) {
		for (;;) {
			struct pollfd fd;
			int polltimeout;

			fd.fd = selectable_fd;
			fd.events = POLLIN;
			if (dotimeout)
				polltimeout = 1;
			else
				polltimeout = -1;
			status = poll(&fd, 1, polltimeout);
			if (status == -1) {
				printf("Poll returns error (%s)\n",
				    strerror(errno));
			} else {
				if (status == 0)
					printf("Poll timed out\n");
				else {
					printf("Poll returned a descriptor: ");
					if (fd.revents & POLLIN)
						printf("readable, ");
					else
						printf("not readable, ");
					if (fd.revents & POLLERR)
						printf("exceptional condition, ");
					else
						printf("no exceptional condition, ");
					if (fd.revents & POLLHUP)
						printf("disconnect, ");
					else
						printf("no disconnect, ");
					if (fd.revents & POLLNVAL)
						printf("invalid\n");
					else
						printf("not invalid\n");
				}
				packet_count = 0;
				status = pcap_dispatch(pd, -1, countme,
				    (u_char *)&packet_count);
				if (status < 0)
					break;
				printf("%d packets seen, %d packets counted after poll returns\n",
				    status, packet_count);
			}
		}
	} else {
		for (;;) {
			packet_count = 0;
			status = pcap_dispatch(pd, -1, countme,
			    (u_char *)&packet_count);
			if (status < 0)
				break;
			printf("%d packets seen, %d packets counted after pcap_dispatch returns\n",
			    status, packet_count);
		}
	}
	if (status == -2) {
		/*
		 * We got interrupted, so perhaps we didn't
		 * manage to finish a line we were printing.
		 * Print an extra newline, just in case.
		 */
		putchar('\n');
	}
	(void)fflush(stdout);
	if (status == -1) {
		/*
		 * Error.  Report it.
		 */
		(void)fprintf(stderr, "%s: pcap_loop: %s\n",
		    program_name, pcap_geterr(pd));
	}
	pcap_close(pd);
	exit(status == -1 ? 1 : 0);
}

static void
countme(u_char *user, const struct pcap_pkthdr *h, const u_char *sp)
{
	int *counterp = (int *)user;

	(*counterp)++;
}

static void
usage(void)
{
	(void)fprintf(stderr, "Usage: %s [ -sptn ] [ -i interface ] [expression]\n",
	    program_name);
	exit(1);
}

/* VARARGS */
static void
error(const char *fmt, ...)
{
	va_list ap;

	(void)fprintf(stderr, "%s: ", program_name);
	va_start(ap, fmt);
	(void)vfprintf(stderr, fmt, ap);
	va_end(ap);
	if (*fmt) {
		fmt += strlen(fmt);
		if (fmt[-1] != '\n')
			(void)fputc('\n', stderr);
	}
	exit(1);
	/* NOTREACHED */
}

/* VARARGS */
static void
warning(const char *fmt, ...)
{
	va_list ap;

	(void)fprintf(stderr, "%s: WARNING: ", program_name);
	va_start(ap, fmt);
	(void)vfprintf(stderr, fmt, ap);
	va_end(ap);
	if (*fmt) {
		fmt += strlen(fmt);
		if (fmt[-1] != '\n')
			(void)fputc('\n', stderr);
	}
}

/*
 * Copy arg vector into a new buffer, concatenating arguments with spaces.
 */
static char *
copy_argv(register char **argv)
{
	register char **p;
	register u_int len = 0;
	char *buf;
	char *src, *dst;

	p = argv;
	if (*p == 0)
		return 0;

	while (*p)
		len += strlen(*p++) + 1;

	buf = (char *)malloc(len);
	if (buf == NULL)
		error("copy_argv: malloc");

	p = argv;
	dst = buf;
	while ((src = *p++) != NULL) {
		while ((*dst++ = *src++) != '\0')
			;
		dst[-1] = ' ';
	}
	dst[-1] = '\0';

	return buf;
}
