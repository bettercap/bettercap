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

#ifndef lib_pcap_export_defs_h
#define lib_pcap_export_defs_h

/*
 * PCAP_API_DEF must be used when defining *data* exported from
 * libpcap.  It can be used when defining *functions* exported
 * from libpcap, but it doesn't have to be used there.  It
 * should not be used in declarations in headers.
 *
 * PCAP_API must be used when *declaring* data or functions
 * exported from libpcap; PCAP_API_DEF won't work on all platforms.
 */

/*
 * Check whether this is GCC major.minor or a later release, or some
 * compiler that claims to be "just like GCC" of that version or a
 * later release.
 */
#define IS_AT_LEAST_GNUC_VERSION(major, minor) \
	(defined(__GNUC__) && \
	    (__GNUC__ > (major) || \
	     (__GNUC__ == (major) && __GNUC_MINOR__ >= (minor))))

#if defined(_WIN32)
  #ifdef BUILDING_PCAP
    /*
     * We're compiling libpcap, so we should export functions in our
     * API.
     */
    #define PCAP_API_DEF	__declspec(dllexport)
  #else
    #define PCAP_API_DEF	__declspec(dllimport)
  #endif
#elif defined(MSDOS)
  /* XXX - does this need special treatment? */
  #define PCAP_API_DEF
#else /* UN*X */
  #ifdef BUILDING_PCAP
    /*
     * We're compiling libpcap, so we should export functions in our API.
     * The compiler might be configured not to export functions from a
     * shared library by default, so we might have to explicitly mark
     * functions as exported.
     */
    #if IS_AT_LEAST_GNUC_VERSION(3, 4)
      /*
       * GCC 3.4 or later, or some compiler asserting compatibility with
       * GCC 3.4 or later, so we have __attribute__((visibility()).
       */
      #define PCAP_API_DEF	__attribute__((visibility("default")))
    #elif defined(__SUNPRO_C) && (__SUNPRO_C >= 0x550)
      /*
       * Sun C 5.5 or later, so we have __global.
       * (Sun C 5.9 and later also have __attribute__((visibility()),
       * but there's no reason to prefer it with Sun C.)
       */
      #define PCAP_API_DEF	__global
    #else
      /*
       * We don't have anything to say.
       */
      #define PCAP_API_DEF
    #endif
  #else
    /*
     * We're not building libpcap.
     */
    #define PCAP_API_DEF
  #endif
#endif /* _WIN32/MSDOS/UN*X */

#define PCAP_API	PCAP_API_DEF extern

#endif /* lib_pcap_export_defs_h */
