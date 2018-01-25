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

#include <pcap.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>

/* Forwards */
static void error(const char *, ...);

int
main(void)
{
	char ebuf[PCAP_ERRBUF_SIZE];
	pcap_t *pd;
	int status = 0;

	pd = pcap_open_live("lo0", 65535, 0, 1000, ebuf);
	if (pd == NULL) {
		pd = pcap_open_live("lo", 65535, 0, 1000, ebuf);
		if (pd == NULL) {
			error("Neither lo0 nor lo could be opened: %s",
			    ebuf);
			return 2;
		}
	}
	status = pcap_activate(pd);
	if (status != PCAP_ERROR_ACTIVATED) {
		if (status == 0)
			error("pcap_activate() of opened pcap_t succeeded");
		else if (status == PCAP_ERROR)
			error("pcap_activate() of opened pcap_t failed with %s, not PCAP_ERROR_ACTIVATED",
			    pcap_geterr(pd));
		else
			error("pcap_activate() of opened pcap_t failed with %s, not PCAP_ERROR_ACTIVATED",
			    pcap_statustostr(status));
	}
	return 0;
}

/* VARARGS */
static void
error(const char *fmt, ...)
{
	va_list ap;

	(void)fprintf(stderr, "reactivatetest: ");
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
