/*
 * Copyright (c) 1994, 1995, 1996
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

#ifndef portability_h
#define	portability_h

/*
 * Helpers for portability between Windows and UN*X and between different
 * flavors of UN*X.
 */

#ifdef __cplusplus
extern "C" {
#endif

#ifndef HAVE_STRLCPY
 /*
  * Macro that does the same thing as strlcpy().
  */
 #ifdef _MSC_VER
  /*
   * strncpy_s() is supported at least back to Visual
   * Studio 2005.
   */
  #define strlcpy(x, y, z) \
	strncpy_s((x), (z), (y), _TRUNCATE)

 #else
  #define strlcpy(x, y, z) \
	(strncpy((x), (y), (z)), \
	 ((z) <= 0 ? 0 : ((x)[(z) - 1] = '\0')), \
	 (void) strlen((y)))
 #endif
#endif

/*
 * For flagging arguments as format strings in MSVC.
 */
#if _MSC_VER >= 1400
 #include <sal.h>
 #if _MSC_VER > 1400
  #define FORMAT_STRING(p) _Printf_format_string_ p
 #else
  #define FORMAT_STRING(p) __format_string p
 #endif
#else
 #define FORMAT_STRING(p) p
#endif

#ifdef _MSC_VER
  #define strdup	_strdup
  #define sscanf	sscanf_s
  #define setbuf(x, y) \
	setvbuf((x), (y), _IONBF, 0)
  #define fopen(x, y) \
	fopen_safe((x), (y))
  FILE *fopen_safe(const char *filename, const char* mode);
#endif

#if defined(_MSC_VER) || defined(__MINGW32__)
  #define strlcat(x, y, z) \
	strncat_s((x), (z), (y), _TRUNCATE)
#endif

#ifdef _MSC_VER
  /*
   * MSVC.
   */
  #if _MSC_VER >= 1900
    /*
     * VS 2015 or newer; we have snprintf() function.
     */
    #define HAVE_SNPRINTF
  #endif
#endif

/*
 * On Windows, snprintf(), with that name and with C99 behavior - i.e.,
 * guaranteeing that the formatted string is null-terminated - didn't
 * appear until Visual Studio 2015.  Prior to that, the C runtime had
 * only _snprintf(), which *doesn't* guarantee that the string is
 * null-terminated if it is truncated due to the buffer being too
 * small.  We therefore can't just define snprintf to be _snprintf
 * and define vsnprintf to be _vsnprintf, as we're relying on null-
 * termination of strings in all cases.
 *
 * We also want to allow this to be built with versions of Visual Studio
 * prior to VS 2015, so we can't rely on snprintf() being present.
 *
 * And we want to make sure that, if we support plugins in the future,
 * a routine with C99 snprintf() behavior will be available to them.
 * We also don't want it to collide with the C library snprintf() if
 * there is one.
 *
 * So we make pcap_snprintf() and pcap_vsnprintf() available, either by
 * #defining them to be snprintf or vsnprintf, respectively, or by
 * defining our own versions and exporting them.
 */
#ifdef HAVE_SNPRINTF
#define pcap_snprintf snprintf
#else
extern int pcap_snprintf(char *, size_t, FORMAT_STRING(const char *), ...)
#ifdef __ATTRIBUTE___FORMAT_OK
    __attribute__((format (printf, 3, 4)))
#endif /* __ATTRIBUTE___FORMAT_OK */
    ;
#endif

#ifdef HAVE_VSNPRINTF
#define pcap_vsnprintf vsnprintf
#else
extern int pcap_vsnprintf(char *, size_t, const char *, va_list ap);
#endif

#ifdef HAVE_STRTOK_R
  #define pcap_strtok_r	strtok_r
#else
  #ifdef _MSC_VER
    /*
     * Microsoft gives it a different name.
     */
    #define pcap_strtok_r	strtok_s
  #else
    /*
     * Define it ourselves.
     */
    #define NEED_STRTOK_R
    extern int pcap_strtok_r(char *, const char *, char **);
  #endif
#endif /* HAVE_STRTOK_R */

#ifdef _WIN32
  /*
   * These may be defined by <inttypes.h>.
   *
   * XXX - for MSVC, we always want the _MSC_EXTENSIONS versions.
   * What about other compilers?  If, as the MinGW Web site says MinGW
   * does, the other compilers just use Microsoft's run-time library,
   * then they should probably use the _MSC_EXTENSIONS even if the
   * compiler doesn't define _MSC_EXTENSIONS.
   *
   * XXX - we currently aren't using any of these, but this allows
   * their use in the future.
   */
  #ifndef PRId64
    #ifdef _MSC_EXTENSIONS
      #define PRId64	"I64d"
    #else
      #define PRId64	"lld"
    #endif
  #endif /* PRId64 */

  #ifndef PRIo64
    #ifdef _MSC_EXTENSIONS
      #define PRIo64	"I64o"
    #else
      #define PRIo64	"llo"
    #endif
  #endif /* PRIo64 */

  #ifndef PRIx64
    #ifdef _MSC_EXTENSIONS
      #define PRIx64	"I64x"
    #else
      #define PRIx64	"llx"
    #endif
  #endif

  #ifndef PRIu64
    #ifdef _MSC_EXTENSIONS
      #define PRIu64	"I64u"
    #else
      #define PRIu64	"llu"
    #endif
  #endif

  #if !defined(__cplusplus)
    #define inline __inline
  #endif
#endif /* _WIN32 */

#ifdef __cplusplus
}
#endif

#endif
