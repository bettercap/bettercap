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

/*
 * This doesn't actually test libpcap itself; it tests whether
 * valgrind properly handles the APIs libpcap uses.  If it doesn't,
 * we end up getting patches submitted to "fix" references that
 * valgrind claims are being made to uninitialized data, when, in
 * fact, the OS isn't making any such references - or we get
 * valgrind *not* detecting *actual* incorrect references.
 *
 * Both BPF and Linux socket filters aren't handled correctly
 * by some versions of valgrind.  See valgrind bug 318203 for
 * Linux:
 *
 *	https://bugs.kde.org/show_bug.cgi?id=318203
 *
 * and valgrind bug 312989 for OS X:
 *
 *	https://bugs.kde.org/show_bug.cgi?id=312989
 *
 * The fixes for both of those are checked into the official valgrind
 * repository.
 *
 * The unofficial FreeBSD port has similar issues to the official OS X
 * port, for similar reasons.
 */
#ifndef lint
static const char copyright[] _U_ =
    "@(#) Copyright (c) 1988, 1989, 1990, 1991, 1992, 1993, 1994, 1995, 1996, 1997, 2000\n\
The Regents of the University of California.  All rights reserved.\n";
#endif

#ifdef HAVE_CONFIG_H
#include "config.h"
#endif

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <unistd.h>
#include <fcntl.h>
#include <errno.h>
#include <arpa/inet.h>
#include <sys/types.h>
#include <sys/stat.h>

#if defined(__APPLE__) || defined(__FreeBSD__) || defined(__NetBSD__) || defined(__OpenBSD__) || defined(__DragonFly__)
/* BSD-flavored OS - use BPF */
#define USE_BPF
#elif defined(linux)
/* Linux - use socket filters */
#define USE_SOCKET_FILTERS
#else
#error "Unknown platform or platform that doesn't support Valgrind"
#endif

#if defined(USE_BPF)

#include <sys/ioctl.h>
#include <net/bpf.h>

/*
 * Make "pcap.h" not include "pcap/bpf.h"; we are going to include the
 * native OS version, as we're going to be doing our own ioctls to
 * make sure that, in the uninitialized-data tests, the filters aren't
 * checked by libpcap before being handed to BPF.
 */
#define PCAP_DONT_INCLUDE_PCAP_BPF_H

#elif defined(USE_SOCKET_FILTERS)

#include <sys/socket.h>
#include <linux/types.h>
#include <linux/filter.h>

#endif

#include <pcap.h>

static char *program_name;

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
static void PCAP_NORETURN usage(void);
static void PCAP_NORETURN error(const char *, ...) PCAP_PRINTFLIKE(1, 2);
static void warning(const char *, ...) PCAP_PRINTFLIKE(1, 2);

/*
 * On Windows, we need to open the file in binary mode, so that
 * we get all the bytes specified by the size we get from "fstat()".
 * On UNIX, that's not necessary.  O_BINARY is defined on Windows;
 * we define it as 0 if it's not defined, so it does nothing.
 */
#ifndef O_BINARY
#define O_BINARY	0
#endif

static char *
read_infile(char *fname)
{
	register int i, fd, cc;
	register char *cp;
	struct stat buf;

	fd = open(fname, O_RDONLY|O_BINARY);
	if (fd < 0)
		error("can't open %s: %s", fname, pcap_strerror(errno));

	if (fstat(fd, &buf) < 0)
		error("can't stat %s: %s", fname, pcap_strerror(errno));

	cp = malloc((u_int)buf.st_size + 1);
	if (cp == NULL)
		error("malloc(%d) for %s: %s", (u_int)buf.st_size + 1,
			fname, pcap_strerror(errno));
	cc = read(fd, cp, (u_int)buf.st_size);
	if (cc < 0)
		error("read %s: %s", fname, pcap_strerror(errno));
	if (cc != buf.st_size)
		error("short read %s (%d != %d)", fname, cc, (int)buf.st_size);

	close(fd);
	/* replace "# comment" with spaces */
	for (i = 0; i < cc; i++) {
		if (cp[i] == '#')
			while (i < cc && cp[i] != '\n')
				cp[i++] = ' ';
	}
	cp[cc] = '\0';
	return (cp);
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

#define INSN_COUNT	17

int
main(int argc, char **argv)
{
	char *cp, *device;
	int op;
	int dorfmon, useactivate;
	char ebuf[PCAP_ERRBUF_SIZE];
	char *infile;
	char *cmdbuf;
	pcap_t *pd;
	int status = 0;
	int pcap_fd;
#if defined(USE_BPF)
	struct bpf_program bad_fcode;
	struct bpf_insn uninitialized[INSN_COUNT];
#elif defined(USE_SOCKET_FILTERS)
	struct sock_fprog bad_fcode;
	struct sock_filter uninitialized[INSN_COUNT];
#endif
	struct bpf_program fcode;

	device = NULL;
	dorfmon = 0;
	useactivate = 0;
	infile = NULL;

	if ((cp = strrchr(argv[0], '/')) != NULL)
		program_name = cp + 1;
	else
		program_name = argv[0];

	opterr = 0;
	while ((op = getopt(argc, argv, "aF:i:I")) != -1) {
		switch (op) {

		case 'a':
			useactivate = 1;
			break;

		case 'F':
			infile = optarg;
			break;

		case 'i':
			device = optarg;
			break;

		case 'I':
			dorfmon = 1;
			useactivate = 1;	/* required for rfmon */
			break;

		default:
			usage();
			/* NOTREACHED */
		}
	}

	if (device == NULL) {
		/*
		 * No interface specified; get whatever pcap_lookupdev()
		 * finds.
		 */
		device = pcap_lookupdev(ebuf);
		if (device == NULL) {
			error("couldn't find interface to use: %s",
			    ebuf);
		}
	}

	if (infile != NULL) {
		/*
		 * Filter specified with "-F" and a file containing
		 * a filter.
		 */
		cmdbuf = read_infile(infile);
	} else {
		if (optind < argc) {
			/*
			 * Filter specified with arguments on the
			 * command line.
			 */
			cmdbuf = copy_argv(&argv[optind+1]);
		} else {
			/*
			 * No filter specified; use an empty string, which
			 * compiles to an "accept all" filter.
			 */
			cmdbuf = "";
		}
	}

	if (useactivate) {
		pd = pcap_create(device, ebuf);
		if (pd == NULL)
			error("%s: pcap_create() failed: %s", device, ebuf);
		status = pcap_set_snaplen(pd, 65535);
		if (status != 0)
			error("%s: pcap_set_snaplen failed: %s",
			    device, pcap_statustostr(status));
		status = pcap_set_promisc(pd, 1);
		if (status != 0)
			error("%s: pcap_set_promisc failed: %s",
			    device, pcap_statustostr(status));
		if (dorfmon) {
			status = pcap_set_rfmon(pd, 1);
			if (status != 0)
				error("%s: pcap_set_rfmon failed: %s",
				    device, pcap_statustostr(status));
		}
		status = pcap_set_timeout(pd, 1000);
		if (status != 0)
			error("%s: pcap_set_timeout failed: %s",
			    device, pcap_statustostr(status));
		status = pcap_activate(pd);
		if (status < 0) {
			/*
			 * pcap_activate() failed.
			 */
			error("%s: %s\n(%s)", device,
			    pcap_statustostr(status), pcap_geterr(pd));
		} else if (status > 0) {
			/*
			 * pcap_activate() succeeded, but it's warning us
			 * of a problem it had.
			 */
			warning("%s: %s\n(%s)", device,
			    pcap_statustostr(status), pcap_geterr(pd));
		}
	} else {
		*ebuf = '\0';
		pd = pcap_open_live(device, 65535, 1, 1000, ebuf);
		if (pd == NULL)
			error("%s", ebuf);
		else if (*ebuf)
			warning("%s", ebuf);
	}

	pcap_fd = pcap_fileno(pd);

	/*
	 * Try setting a filter with an uninitialized bpf_program
	 * structure.  This should cause valgrind to report a
	 * problem.
	 *
	 * We don't check for errors, because it could get an
	 * error due to a bad pointer or count.
	 */
#if defined(USE_BPF)
	ioctl(pcap_fd, BIOCSETF, &bad_fcode);
#elif defined(USE_SOCKET_FILTERS)
	setsockopt(pcap_fd, SOL_SOCKET, SO_ATTACH_FILTER, &bad_fcode,
	    sizeof(bad_fcode));
#endif

	/*
	 * Try setting a filter with an initialized bpf_program
	 * structure that points to an uninitialized program.
	 * That should also cause valgrind to report a problem.
	 *
	 * We don't check for errors, because it could get an
	 * error due to a bad pointer or count.
	 */
#if defined(USE_BPF)
	bad_fcode.bf_len = INSN_COUNT;
	bad_fcode.bf_insns = uninitialized;
	ioctl(pcap_fd, BIOCSETF, &bad_fcode);
#elif defined(USE_SOCKET_FILTERS)
	bad_fcode.len = INSN_COUNT;
	bad_fcode.filter = uninitialized;
	setsockopt(pcap_fd, SOL_SOCKET, SO_ATTACH_FILTER, &bad_fcode,
	    sizeof(bad_fcode));
#endif

	/*
	 * Now compile a filter and set the filter with that.
	 * That should *not* cause valgrind to report a
	 * problem.
	 */
	if (pcap_compile(pd, &fcode, cmdbuf, 1, 0) < 0)
		error("can't compile filter: %s", pcap_geterr(pd));
	if (pcap_setfilter(pd, &fcode) < 0)
		error("can't set filter: %s", pcap_geterr(pd));

	pcap_close(pd);
	exit(status < 0 ? 1 : 0);
}

static void
usage(void)
{
	(void)fprintf(stderr, "%s, with %s\n", program_name,
	    pcap_lib_version());
	(void)fprintf(stderr,
	    "Usage: %s [-aI] [ -F file ] [ -I interface ] [ expression ]\n",
	    program_name);
	exit(1);
}
