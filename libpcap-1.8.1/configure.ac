dnl
dnl Copyright (c) 1994, 1995, 1996, 1997
dnl	The Regents of the University of California.  All rights reserved.
dnl
dnl Process this file with autoconf to produce a configure script.
dnl

#
# See
#
#	http://ftp.gnu.org/gnu/config/README
#
# for the URLs to use to fetch new versions of config.guess and
# config.sub.
#

AC_PREREQ(2.61)
AC_INIT(pcap.c)

AC_CANONICAL_SYSTEM

AC_LBL_C_INIT_BEFORE_CC(V_CCOPT, V_INCLS)
AC_PROG_CC
AC_LBL_C_INIT(V_CCOPT, V_INCLS)
AC_LBL_SHLIBS_INIT
AC_LBL_C_INLINE
AC_C___ATTRIBUTE__
if test "$ac_cv___attribute__" = "yes"; then
	AC_C___ATTRIBUTE___UNUSED
	AC_C___ATTRIBUTE___FORMAT
fi

AC_CHECK_HEADERS(sys/bitypes.h)

AC_CHECK_TYPE([int8_t], ,
	[AC_DEFINE([int8_t], [signed char],
	[Define to `signed char' if int8_t not defined.])],
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])
AC_CHECK_TYPE([u_int8_t], ,
	[AC_DEFINE([u_int8_t], [unsigned char],
	[Define to `unsigned char' if u_int8_t not defined.])],
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])
AC_CHECK_TYPE([int16_t], ,
	[AC_DEFINE([int16_t], [short],
	[Define to `short' if int16_t not defined.])]
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])
AC_CHECK_TYPE([u_int16_t], ,
	[AC_DEFINE([u_int16_t], [unsigned short],
	[Define to `unsigned short' if u_int16_t not defined.])],
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])
AC_CHECK_TYPE([int32_t], ,
	[AC_DEFINE([int32_t], [int],
	[Define to `int' if int32_t not defined.])],
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])
AC_CHECK_TYPE([u_int32_t], ,
	[AC_DEFINE([u_int32_t], [unsigned int],
	[Define to `unsigned int' if u_int32_t not defined.])],
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])
AC_CHECK_TYPE([int64_t], ,
	[AC_DEFINE([int64_t], [long long],
	[Define to `long long' if int64_t not defined.])],
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])
AC_CHECK_TYPE([u_int64_t], ,
	[AC_DEFINE([u_int64_t], [unsigned long long],
	[Define to `unsigned long long' if u_int64_t not defined.])],
	[AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif])

#
# Try to arrange for large file support.
#
AC_SYS_LARGEFILE
AC_FUNC_FSEEKO

dnl
dnl Even if <net/bpf.h> were, on all OSes that support BPF, fixed to
dnl include <sys/ioccom.h>, and we were to drop support for older
dnl releases without that fix, so that pcap-bpf.c doesn't need to
dnl include <sys/ioccom.h>, the test program in "AC_LBL_FIXINCLUDES"
dnl in "aclocal.m4" uses it, so we would still have to test for it
dnl and set "HAVE_SYS_IOCCOM_H" if we have it, otherwise
dnl "AC_LBL_FIXINCLUDES" wouldn't work on some platforms such as Solaris.
dnl
AC_CHECK_HEADERS(sys/ioccom.h sys/select.h sys/sockio.h limits.h)
AC_CHECK_HEADERS(linux/types.h)
AC_CHECK_HEADERS(linux/if_packet.h netpacket/packet.h netpacket/if_packet.h)
AC_CHECK_HEADERS(net/pfvar.h, , , [#include <sys/types.h>
#include <sys/socket.h>
#include <net/if.h>])
if test "$ac_cv_header_net_pfvar_h" = yes; then
	#
	# Check for various PF actions.
	#
	AC_MSG_CHECKING(whether net/pfvar.h defines PF_NAT through PF_NORDR)
	AC_TRY_COMPILE(
	    [#include <sys/types.h>
	    #include <sys/socket.h>
	    #include <net/if.h>
	    #include <net/pfvar.h>],
	    [return PF_NAT+PF_NONAT+PF_BINAT+PF_NOBINAT+PF_RDR+PF_NORDR;],
	    [
		AC_MSG_RESULT(yes)
		AC_DEFINE(HAVE_PF_NAT_THROUGH_PF_NORDR, 1,
		    [define if net/pfvar.h defines PF_NAT through PF_NORDR])
	    ],
	    AC_MSG_RESULT(no))
fi
AC_CHECK_HEADERS(netinet/if_ether.h, , , [#include <sys/types.h>
#include <sys/socket.h>])
if test "$ac_cv_header_netinet_if_ether_h" != yes; then
	#
	# The simple test didn't work.
	# Do we need to include <net/if.h> first?
	# Unset ac_cv_header_netinet_if_ether_h so we don't
	# treat the previous failure as a cached value and
	# suppress the next test.
	#
	AC_MSG_NOTICE([Rechecking with some additional includes])
	unset ac_cv_header_netinet_if_ether_h
	AC_CHECK_HEADERS(netinet/if_ether.h, , , [#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
struct mbuf;
struct rtentry;
#include <net/if.h>])
fi

case "$host_os" in
linux*|uclinux*)
	AC_CHECK_HEADERS(linux/sockios.h linux/if_bonding.h,,,
	[
#include <sys/socket.h>
#include <linux/if.h>
	])
	;;
esac

AC_LBL_FIXINCLUDES

AC_CHECK_FUNCS(strerror strlcpy)

needsnprintf=no
AC_CHECK_FUNCS(vsnprintf snprintf,,
	[needsnprintf=yes])
if test $needsnprintf = yes; then
	AC_LIBOBJ([snprintf])
fi

needstrtok_r=no
AC_CHECK_FUNCS(strtok_r,,
	[needstrtok_r=yes])
if test $needstrtok_r = yes; then
	AC_LIBOBJ([strtok_r])
fi

#
# Do this before checking for ether_hostton(), as it's a
# "gethostbyname() -ish function".
#
AC_LBL_LIBRARY_NET

#
# You are in a twisty little maze of UN*Xes, all different.
# Some might not have ether_hostton().
# Some might have it, but not declare it in any header file.
# Some might have it, but declare it in <netinet/if_ether.h>.
# Some might have it, but declare it in <netinet/ether.h>
# (And some might have it but document it as something declared in
# <netinet/ethernet.h>, although <netinet/if_ether.h> appears to work.)
#
# Before you is a C compiler.
#
AC_CHECK_FUNCS(ether_hostton)
if test "$ac_cv_func_ether_hostton" = yes; then
	#
	# OK, we have ether_hostton().  Do we have <netinet/if_ether.h>?
	#
	if test "$ac_cv_header_netinet_if_ether_h" = yes; then
		#
		# Yes.  Does it declare ether_hostton()?
		#
		AC_CHECK_DECL(ether_hostton,
		    [
			AC_DEFINE(NETINET_IF_ETHER_H_DECLARES_ETHER_HOSTTON,,
			    [Define to 1 if netinet/if_ether.h declares `ether_hostton'])
		    ],,
		    [
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
struct mbuf;
struct rtentry;
#include <net/if.h>
#include <netinet/if_ether.h>
		    ])
	fi
	#
	# Did that succeed?
	#
	if test "$ac_cv_have_decl_ether_hostton" != yes; then
		#
		# No, how about <netinet/ether.h>, as on Linux?
		#
		AC_CHECK_HEADERS(netinet/ether.h)
		if test "$ac_cv_header_netinet_ether_h" = yes; then
			#
			# We have it - does it declare ether_hostton()?
			# Unset ac_cv_have_decl_ether_hostton so we don't
			# treat the previous failure as a cached value and
			# suppress the next test.
			#
			unset ac_cv_have_decl_ether_hostton
			AC_CHECK_DECL(ether_hostton,
			    [
				AC_DEFINE(NETINET_ETHER_H_DECLARES_ETHER_HOSTTON,,
				    [Define to 1 if netinet/ether.h declares `ether_hostton'])
			    ],,
			    [
#include <netinet/ether.h>
			    ])
		fi
	fi
	#
	# Is ether_hostton() declared?
	#
	if test "$ac_cv_have_decl_ether_hostton" != yes; then
		#
		# No, we'll have to declare it ourselves.
		# Do we have "struct ether_addr"?
		#
		AC_CHECK_TYPES(struct ether_addr,,,
		    [
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
struct mbuf;
struct rtentry;
#include <net/if.h>
#include <netinet/if_ether.h>
		    ])
		AC_DEFINE(HAVE_DECL_ETHER_HOSTTON, 0,
		    [Define to 1 if you have the declaration of `ether_hostton', and to 0 if you
don't.])
	else
		AC_DEFINE(HAVE_DECL_ETHER_HOSTTON, 1,
		    [Define to 1 if you have the declaration of `ether_hostton', and to 0 if you
don't.])
	fi
fi

dnl to pacify those who hate protochain insn
AC_MSG_CHECKING(if --disable-protochain option is specified)
AC_ARG_ENABLE(protochain,
AC_HELP_STRING([--disable-protochain],[disable \"protochain\" insn]))
case "x$enable_protochain" in
xyes)	enable_protochain=enabled	;;
xno)	enable_protochain=disabled	;;
x)	enable_protochain=enabled	;;
esac

if test "$enable_protochain" = "disabled"; then
	AC_DEFINE(NO_PROTOCHAIN,1,[do not use protochain])
fi
AC_MSG_RESULT(${enable_protochain})

#
# valgrindtest directly uses the native capture mechanism, but
# only tests with BPF and PF_PACKET sockets; only enable it if
# we have BPF or PF_PACKET sockets.
#
VALGRINDTEST=

#
# SITA support is mutually exclusive with native capture support;
# "--with-sita" selects SITA support.
#
AC_ARG_WITH(sita,
AC_HELP_STRING([--with-sita],[include SITA support]),
[
	if test ! "x$withval" = "xno" ; then
		AC_DEFINE(SITA,1,[include ACN support])
		AC_MSG_NOTICE(Enabling SITA ACN support)
		V_PCAP=sita
	fi
],
[
dnl
dnl Not all versions of test support -c (character special) but it's a
dnl better way of testing since the device might be protected. So we
dnl check in our normal order using -r and then check the for the /dev
dnl guys again using -c.
dnl
dnl XXX This could be done for cross-compiling, but for now it's not.
dnl
if test -z "$with_pcap" && test "$cross_compiling" = yes; then
	AC_MSG_ERROR(pcap type not determined when cross-compiling; use --with-pcap=...)
fi
AC_ARG_WITH(pcap,
AC_HELP_STRING([--with-pcap=TYPE],[use packet capture TYPE]))
AC_MSG_CHECKING(packet capture type)
if test ! -z "$with_pcap" ; then
	V_PCAP="$withval"
elif test -r /dev/bpf -o -h /dev/bpf ; then
	#
	# Cloning BPF device.
	#
	V_PCAP=bpf
	AC_DEFINE(HAVE_CLONING_BPF,1,[define if you have a cloning BPF device])

	#
	# We have BPF, so build valgrindtest with "make test".
	#
	VALGRINDTEST=valgrindtest
elif test -r /dev/bpf0 ; then
	V_PCAP=bpf

	#
	# We have BPF, so build valgrindtest with "make test".
	#
	VALGRINDTEST=valgrindtest
elif test -r /usr/include/net/pfilt.h ; then
	V_PCAP=pf
elif test -r /dev/enet ; then
	V_PCAP=enet
elif test -r /dev/nit ; then
	V_PCAP=snit
elif test -r /usr/include/sys/net/nit.h ; then
	V_PCAP=nit
elif test -r /usr/include/linux/socket.h ; then
	V_PCAP=linux

	#
	# XXX - this won't work with older kernels that have SOCK_PACKET
	# sockets but not PF_PACKET sockets.
	#
	VALGRINDTEST=valgrindtest
elif test -r /usr/include/net/raw.h ; then
	V_PCAP=snoop
elif test -r /usr/include/odmi.h ; then
	#
	# On AIX, the BPF devices might not yet be present - they're
	# created the first time libpcap runs after booting.
	# We check for odmi.h instead.
	#
	V_PCAP=bpf
elif test -c /dev/bpf0 ; then		# check again in case not readable
	V_PCAP=bpf

	#
	# We have BPF, so build valgrindtest with "make test".
	#
	VALGRINDTEST=valgrindtest
elif test -r /usr/include/sys/dlpi.h ; then
	V_PCAP=dlpi
elif test -c /dev/enet ; then		# check again in case not readable
	V_PCAP=enet
elif test -c /dev/nit ; then		# check again in case not readable
	V_PCAP=snit
else
	V_PCAP=null
fi
AC_MSG_RESULT($V_PCAP)
AC_SUBST(VALGRINDTEST)

#
# Do capture-mechanism-dependent tests.
#
case "$V_PCAP" in
dlpi)
	#
	# Needed for common functions used by pcap-[dlpi,libdlpi].c
	#
 	SSRC="dlpisubs.c"

	#
	# Checks for some header files.
	#
	AC_CHECK_HEADERS(sys/bufmod.h sys/dlpi_ext.h)

	#
	# Checks to see if Solaris has the public libdlpi(3LIB) library.
	# Note: The existence of /usr/include/libdlpi.h does not mean it is the
	# public libdlpi(3LIB) version. Before libdlpi was made public, a
	# private version also existed, which did not have the same APIs.
	# Due to a gcc bug, the default search path for 32-bit libraries does
	# not include /lib, we add it explicitly here.
	# [http://bugs.opensolaris.org/view_bug.do?bug_id=6619485].
	# Also, due to the bug above applications that link to libpcap with
	# libdlpi will have to add "-L/lib" option to "configure".
	#
	saved_ldflags=$LDFLAGS
	LDFLAGS="$LIBS -L/lib"
	AC_CHECK_LIB(dlpi, dlpi_walk,
		[
			LIBS="-ldlpi $LIBS"
			V_PCAP=libdlpi
			AC_DEFINE(HAVE_LIBDLPI,1,[if libdlpi exists])
		],
		V_PCAP=dlpi)
	LDFLAGS=$saved_ldflags

	#
	# Checks whether <sys/dlpi.h> is usable, to catch weird SCO
	# versions of DLPI.
	#
	AC_MSG_CHECKING(whether <sys/dlpi.h> is usable)
	AC_CACHE_VAL(ac_cv_sys_dlpi_usable,
		AC_TRY_COMPILE(
		    [
			#include <sys/types.h>
			#include <sys/time.h>
			#include <sys/dlpi.h>
		    ],
		    [int i = DL_PROMISC_PHYS;],
		    ac_cv_sys_dlpi_usable=yes,
		    ac_cv_sys_dlpi_usable=no))
	AC_MSG_RESULT($ac_cv_sys_dlpi_usable)
	if test $ac_cv_sys_dlpi_usable = no ; then
		AC_MSG_ERROR(<sys/dlpi.h> is not usable on this system; it probably has a non-standard DLPI)
	fi

	#
	# Check whether we have a /dev/dlpi device or have multiple devices.
	#
	AC_MSG_CHECKING(for /dev/dlpi device)
	if test -c /dev/dlpi ; then
		AC_MSG_RESULT(yes)
		AC_DEFINE(HAVE_DEV_DLPI, 1, [define if you have a /dev/dlpi])
	else
		AC_MSG_RESULT(no)
		dir="/dev/dlpi"
		AC_MSG_CHECKING(for $dir directory)
		if test -d $dir ; then
			AC_MSG_RESULT(yes)
			AC_DEFINE_UNQUOTED(PCAP_DEV_PREFIX, "$dir", [/dev/dlpi directory])
		else
			AC_MSG_RESULT(no)
		fi
	fi

	#
	# This check is for Solaris with DLPI support for passive modes.
	# See dlpi(7P) for more details.
	#
	AC_LBL_DL_PASSIVE_REQ_T
	;;

linux)
	#
	# Do we have the wireless extensions?
	#
	AC_CHECK_HEADERS(linux/wireless.h, [], [],
	[
#include <sys/socket.h>
#include <linux/if.h>
#include <linux/types.h>
	])

	#
	# Do we have libnl?
	#
	AC_ARG_WITH(libnl,
	AC_HELP_STRING([--without-libnl],[disable libnl support @<:@default=yes, on Linux, if present@:>@]),
		with_libnl=$withval,,)

	if test x$with_libnl != xno ; then
		have_any_nl="no"

                incdir=-I/usr/include/libnl3
                libnldir=
                if test x$withval != x ; then
                  libnldir=-L${withval}/lib/.libs
                  incdir=-I${withval}/include
                fi

		#
		# Try libnl 3.x first.
		#
		AC_CHECK_LIB(nl-3, nl_socket_alloc,
		[
			#
			# Yes, we have libnl 3.x.
			#
			LIBS="${libnldir} -lnl-genl-3 -lnl-3 $LIBS"
			AC_DEFINE(HAVE_LIBNL,1,[if libnl exists])
			AC_DEFINE(HAVE_LIBNL_3_x,1,[if libnl exists and is version 3.x])
			AC_DEFINE(HAVE_LIBNL_NLE,1,[libnl has NLE_FAILURE])
			AC_DEFINE(HAVE_LIBNL_SOCKETS,1,[libnl has new-style socket api])
			V_INCLS="$V_INCLS ${incdir}"
			have_any_nl="yes"
		],[], ${incdir} ${libnldir} -lnl-genl-3 -lnl-3 )

		if test x$have_any_nl = xno ; then
			#
			# Try libnl 2.x
			#
			AC_CHECK_LIB(nl, nl_socket_alloc,
			[
				#
				# Yes, we have libnl 2.x.
				#
				LIBS="${libnldir} -lnl-genl -lnl $LIBS"
				AC_DEFINE(HAVE_LIBNL,1,[if libnl exists])
				AC_DEFINE(HAVE_LIBNL_2_x,1,[if libnl exists and is version 2.x])
				AC_DEFINE(HAVE_LIBNL_NLE,1,[libnl has NLE_FAILURE])
				AC_DEFINE(HAVE_LIBNL_SOCKETS,1,[libnl has new-style socket api])
				have_any_nl="yes"
			])
		fi

		if test x$have_any_nl = xno ; then
			#
			# No, we don't; do we have libnl 1.x?
			#
			AC_CHECK_LIB(nl, nl_handle_alloc,
			[
				#
				# Yes.
				#
				LIBS="${libnldir} -lnl $LIBS"
				AC_DEFINE(HAVE_LIBNL,1,[if libnl exists])
				have_any_nl="yes"
			])
		fi

		if test x$have_any_nl = xno ; then
			#
			# No, we don't have libnl at all.
			#
			if test x$with_libnl = xyes ; then
				AC_MSG_ERROR([libnl support requested but libnl not found])
			fi
		fi
	fi

	AC_CHECK_HEADERS(linux/ethtool.h,,,
	    [
AC_INCLUDES_DEFAULT
#include <linux/types.h>
	    ])
	AC_LBL_TPACKET_STATS
	AC_LBL_LINUX_TPACKET_AUXDATA_TP_VLAN_TCI
	;;

bpf)
	#
	# Check whether we have the *BSD-style ioctls.
	#
	AC_CHECK_HEADERS(net/if_media.h)

	AC_MSG_CHECKING(whether the system supports zerocopy BPF)
	AC_TRY_COMPILE(
	    [#include <sys/socket.h>
	    #include <sys/ioctl.h>
	    #include <net/if.h>
	    #include <net/bpf.h>],
	    [return (BIOCROTZBUF + BPF_BUFMODE_ZBUF);],
	    [
		AC_MSG_RESULT(yes)
		AC_DEFINE(HAVE_ZEROCOPY_BPF, 1,
		[define if the system supports zerocopy BPF])
	    ],
	    AC_MSG_RESULT(no))

	    #
	    # Check whether we have struct BPF_TIMEVAL.
	    #
	    AC_CHECK_TYPES(struct BPF_TIMEVAL,,,
		[
#include <sys/types.h>
#include <sys/ioctl.h>
#ifdef HAVE_SYS_IOCCOM_H
#include <sys/ioccom.h>
#endif
#include <net/bpf.h>
		    ])
	;;

dag)
	#
	# --with-pcap=dag is the only way to get here, and it means
	# "DAG support but nothing else"
	#
	V_DEFS="$V_DEFS -DDAG_ONLY"
	xxx_only=yes
	;;

septel)
	#
	# --with-pcap=septel is the only way to get here, and it means
	# "Septel support but nothing else"
	#
	V_DEFS="$V_DEFS -DSEPTEL_ONLY"
	xxx_only=yes
	;;

snf)
	#
	# --with-pcap=snf is the only way to get here, and it means
	# "SNF support but nothing else"
	#
	V_DEFS="$V_DEFS -DSNF_ONLY"
	xxx_only=yes
	;;

null)
	AC_MSG_WARN(cannot determine packet capture interface)
	AC_MSG_WARN((see the INSTALL doc for more info))
	;;
esac

dnl
dnl Now figure out how we get a list of interfaces and addresses,
dnl if we support capturing.  Don't bother if we don't support
dnl capturing.
dnl
if test "$V_PCAP" != null
then
	AC_CHECK_FUNC(getifaddrs,[
		#
		# We have "getifaddrs()"; make sure we have <ifaddrs.h>
		# as well, just in case some platform is really weird.
		#
		AC_CHECK_HEADER(ifaddrs.h,[
		    #
		    # We have the header, so we use "getifaddrs()" to
		    # get the list of interfaces.
		    #
		    V_FINDALLDEVS=fad-getad.c
		],[
		    #
		    # We don't have the header - give up.
		    # XXX - we could also fall back on some other
		    # mechanism, but, for now, this'll catch this
		    # problem so that we can at least try to figure
		    # out something to do on systems with "getifaddrs()"
		    # but without "ifaddrs.h", if there is something
		    # we can do on those systems.
		    #
		    AC_MSG_ERROR([Your system has getifaddrs() but doesn't have a usable <ifaddrs.h>.])
		])
	],[
		#
		# Well, we don't have "getifaddrs()", so we have to use
		# some other mechanism; determine what that mechanism is.
		#
		# The first thing we use is the type of capture mechanism,
		# which is somewhat of a proxy for the OS we're using.
		#
		case "$V_PCAP" in

		dlpi|libdlpi)
			#
			# This might be Solaris 8 or later, with
			# SIOCGLIFCONF, or it might be some other OS
			# or some older version of Solaris, with
			# just SIOCGIFCONF.
			#
			AC_MSG_CHECKING(whether we have SIOCGLIFCONF)
			AC_CACHE_VAL(ac_cv_lbl_have_siocglifconf,
			    AC_TRY_COMPILE(
				[#include <sys/param.h>
				#include <sys/file.h>
				#include <sys/ioctl.h>
				#include <sys/socket.h>
				#include <sys/sockio.h>],
				[ioctl(0, SIOCGLIFCONF, (char *)0);],
				ac_cv_lbl_have_siocglifconf=yes,
				ac_cv_lbl_have_siocglifconf=no))
			AC_MSG_RESULT($ac_cv_lbl_have_siocglifconf)
			if test $ac_cv_lbl_have_siocglifconf = yes ; then
				V_FINDALLDEVS=fad-glifc.c
			else
				V_FINDALLDEVS=fad-gifc.c
			fi
			;;

		*)
			#
			# Assume we just have SIOCGIFCONF.
			# (XXX - on at least later Linux kernels, there's
			# another mechanism, and we should be using that
			# instead.)
			#
			V_FINDALLDEVS=fad-gifc.c
			;;
		esac])
fi
])

AC_MSG_CHECKING(for socklen_t)
AC_TRY_COMPILE([
	#include <sys/types.h>
	#include <sys/socket.h>
	],
	[ socklen_t x; ],
	have_socklen_t=yes,
	have_socklen_t=no)
if test "x$have_socklen_t" = "xyes"; then
	AC_DEFINE(HAVE_SOCKLEN_T, 1, [define if socklen_t is defined])
fi
AC_MSG_RESULT($have_socklen_t)

AC_ARG_ENABLE(ipv6,
AC_HELP_STRING([--enable-ipv6],[build IPv6-capable version @<:@default=yes, if getaddrinfo available@:>@]),
    [],
    [enable_ipv6=ifavailable])
if test "$enable_ipv6" != "no"; then
	AC_CHECK_FUNC(getaddrinfo,
	    [
		AC_DEFINE(INET6,1,[IPv6])
	    ],
	    [
		if test "$enable_ipv6" != "ifavailable"; then
			AC_MSG_FAILURE([--enable-ipv6 was given, but getaddrinfo isn't available])
		fi
	    ])
fi

AC_MSG_CHECKING(whether to build optimizer debugging code)
AC_ARG_ENABLE(optimizer-dbg,
AC_HELP_STRING([--enable-optimizer-dbg],[build optimizer debugging code]))
if test "$enable_optimizer_dbg" = "yes"; then
	AC_DEFINE(BDEBUG,1,[Enable optimizer debugging])
fi
AC_MSG_RESULT(${enable_optimizer_dbg-no})

AC_MSG_CHECKING(whether to build parser debugging code)
AC_ARG_ENABLE(yydebug,
AC_HELP_STRING([--enable-yydebug],[build parser debugging code]))
if test "$enable_yydebug" = "yes"; then
	AC_DEFINE(YYDEBUG,1,[Enable parser debugging])
fi
AC_MSG_RESULT(${enable_yydebug-no})

# Check for Endace DAG card support.
AC_ARG_WITH([dag],
AC_HELP_STRING([--with-dag@<:@=DIR@:>@],[include Endace DAG support (located in directory DIR, if supplied).  @<:@default=yes, if present@:>@]),
[
	if test "$withval" = no
	then
		# User doesn't want DAG support.
		want_dag=no
	elif test "$withval" = yes
	then
		# User wants DAG support but hasn't specified a directory.
		want_dag=yes
	else
		# User wants DAG support and has specified a directory, so use the provided value.
		want_dag=yes
		dag_root=$withval
	fi
],[
	if test "$V_PCAP" = dag; then
		# User requested DAG-only libpcap, so we'd better have
		# the DAG API.
		want_dag=yes
	elif test "xxx_only" = yes; then
		# User requested something-else-only pcap, so they don't
		# want DAG support.
		want_dag=no
	else
		#
		# Use DAG API if present, otherwise don't
		#
		want_dag=ifpresent
	fi
])

AC_ARG_WITH([dag-includes],
AC_HELP_STRING([--with-dag-includes=IDIR],[Endace DAG include directory, if not DIR/include]),
[
	# User wants DAG support and has specified a header directory, so use the provided value.
	want_dag=yes
	dag_include_dir=$withval
],[])

AC_ARG_WITH([dag-libraries],
AC_HELP_STRING([--with-dag-libraries=LDIR],[Endace DAG library directory, if not DIR/lib]),
[
	# User wants DAG support and has specified a library directory, so use the provided value.
	want_dag=yes
	dag_lib_dir=$withval
],[])

ac_cv_lbl_dag_api=no
if test "$want_dag" != no; then

	AC_MSG_CHECKING([whether we have DAG API headers])

	# If necessary, set default paths for DAG API headers and libraries.
	if test -z "$dag_root"; then
		dag_root=/usr/local
	fi

	if test -z "$dag_include_dir"; then
		dag_include_dir="$dag_root/include"
	fi

	if test -z "$dag_lib_dir"; then
		dag_lib_dir="$dag_root/lib"
	fi

	if test -z "$dag_tools_dir"; then
		dag_tools_dir="$dag_root/tools"
	fi

	if test -r $dag_include_dir/dagapi.h; then
		ac_cv_lbl_dag_api=yes
	fi

	if test "$ac_cv_lbl_dag_api" = yes; then
		AC_MSG_RESULT([yes ($dag_include_dir)])

		V_INCLS="$V_INCLS -I$dag_include_dir"

		if test $V_PCAP != dag ; then
			 SSRC="$SSRC pcap-dag.c"
		fi

		# See if we can find a general version string.
		# Don't need to save and restore LIBS to prevent -ldag being
		# included if there's a found-action (arg 3).
		saved_ldflags=$LDFLAGS
		LDFLAGS="-L$dag_lib_dir"
		AC_CHECK_LIB([dag], [dag_attach_stream], [dag_streams="1"], [dag_streams="0"])
		AC_CHECK_LIB([dag],[dag_get_erf_types], [
			AC_DEFINE(HAVE_DAG_GET_ERF_TYPES, 1, [define if you have dag_get_erf_types()])])
		AC_CHECK_LIB([dag],[dag_get_stream_erf_types], [
			AC_DEFINE(HAVE_DAG_GET_STREAM_ERF_TYPES, 1, [define if you have dag_get_stream_erf_types()])])

		LDFLAGS=$saved_ldflags

		if test "$dag_streams" = 1; then
			AC_DEFINE(HAVE_DAG_STREAMS_API, 1, [define if you have streams capable DAG API])
			LIBS="$LIBS -ldag"
			LDFLAGS="$LDFLAGS -L$dag_lib_dir"

			AC_CHECK_LIB([vdag],[vdag_set_device_info], [ac_dag_have_vdag="1"], [ac_dag_have_vdag="0"])
			if test "$ac_dag_have_vdag" = 1; then
				AC_DEFINE(HAVE_DAG_VDAG, 1, [define if you have vdag_set_device_info()])
				LIBS="$LIBS -lpthread"
			fi
		fi

		AC_DEFINE(HAVE_DAG_API, 1, [define if you have the DAG API])
	else
		AC_MSG_RESULT(no)

		if test "$V_PCAP" = dag; then
			# User requested "dag" capture type but we couldn't
			# find the DAG API support.
			AC_MSG_ERROR([DAG support requested with --with-pcap=dag, but the DAG headers weren't found at $dag_include_dir: make sure the DAG support is installed, specify a different path or paths if necessary, or don't request DAG support])
		fi

		if test "$want_dag" = yes; then
	        	# User wanted DAG support but we couldn't find it.
			AC_MSG_ERROR([DAG support requested with --with-dag, but the DAG headers weren't found at $dag_include_dir: make sure the DAG support is installed, specify a different path or paths if necessary, or don't request DAG support])
		fi
	fi
fi

AC_ARG_WITH(septel,
AC_HELP_STRING([--with-septel@<:@=DIR@:>@],[include Septel support (located in directory DIR, if supplied).  @<:@default=yes, if present@:>@]),
[
	if test "$withval" = no
	then
		want_septel=no
	elif test "$withval" = yes
	then
		want_septel=yes
		septel_root=
	else
		want_septel=yes
		septel_root=$withval
	fi
],[
	if test "$V_PCAP" = septel; then
		# User requested Septel-only libpcap, so we'd better have
		# the Septel API.
		want_septel=yes
	elif test "xxx_only" = yes; then
		# User requested something-else-only pcap, so they don't
		# want Septel support.
		want_septel=no
	else
		#
		# Use Septel API if present, otherwise don't
		#
		want_septel=ifpresent
	fi
])

ac_cv_lbl_septel_api=no
if test "$with_septel" != no; then

	AC_MSG_CHECKING([whether we have Septel API headers])

	# If necessary, set default paths for Septel API headers and libraries.
	if test -z "$septel_root"; then
		septel_root=$srcdir/../septel
	fi

	septel_tools_dir="$septel_root"
	septel_include_dir="$septel_root/INC"

	if test -r "$septel_include_dir/msg.h"; then
		ac_cv_lbl_septel_api=yes
	fi

	if test "$ac_cv_lbl_septel_api" = yes; then
		AC_MSG_RESULT([yes ($septel_include_dir)])

		V_INCLS="$V_INCLS -I$septel_include_dir"
		ADDLOBJS="$ADDLOBJS $septel_tools_dir/asciibin.o $septel_tools_dir/bit2byte.o $septel_tools_dir/confirm.o $septel_tools_dir/fmtmsg.o $septel_tools_dir/gct_unix.o $septel_tools_dir/hqueue.o $septel_tools_dir/ident.o $septel_tools_dir/mem.o $septel_tools_dir/pack.o $septel_tools_dir/parse.o $septel_tools_dir/pool.o $septel_tools_dir/sdlsig.o $septel_tools_dir/strtonum.o $septel_tools_dir/timer.o $septel_tools_dir/trace.o"
		ADDLARCHIVEOBJS="$ADDLARCHIVEOBJS $septel_tools_dir/asciibin.o $septel_tools_dir/bit2byte.o $septel_tools_dir/confirm.o $septel_tools_dir/fmtmsg.o $septel_tools_dir/gct_unix.o $septel_tools_dir/hqueue.o $septel_tools_dir/ident.o $septel_tools_dir/mem.o $septel_tools_dir/pack.o $septel_tools_dir/parse.o $septel_tools_dir/pool.o $septel_tools_dir/sdlsig.o $septel_tools_dir/strtonum.o $septel_tools_dir/timer.o $septel_tools_dir/trace.o"

		if test "$V_PCAP" != septel ; then
			 SSRC="$SSRC pcap-septel.c"
		fi

		AC_DEFINE(HAVE_SEPTEL_API, 1, [define if you have the Septel API])
	else
		AC_MSG_RESULT(no)

		if test "$V_PCAP" = septel; then
			# User requested "septel" capture type but
			# we couldn't find the Septel API support.
			AC_MSG_ERROR([Septel support requested with --with-pcap=septel, but the Septel headers weren't found at $septel_include_dir: make sure the Septel support is installed, specify a different path or paths if necessary, or don't request Septel support])
		fi

		if test "$want_septel" = yes; then
	        	# User wanted Septel support but we couldn't find it.
			AC_MSG_ERROR([Septel support requested with --with-septel, but the Septel headers weren't found at $septel_include_dir: make sure the Septel support is installed, specify a different path or paths if necessary, or don't request Septel support])
		fi
	fi
fi

# Check for Myricom SNF support.
AC_ARG_WITH([snf],
AC_HELP_STRING([--with-snf@<:@=DIR@:>@],[include Myricom SNF support (located in directory DIR, if supplied).  @<:@default=yes, if present@:>@]),
[
	if test "$withval" = no
	then
		# User explicitly doesn't want SNF
		want_snf=no
	elif test "$withval" = yes
	then
		# User wants SNF support but hasn't specified a directory.
		want_snf=yes
	else
		# User wants SNF support with a specified directory.
		want_snf=yes
		snf_root=$withval
	fi
],[
	if test "$V_PCAP" = snf; then
		# User requested Sniffer-only libpcap, so we'd better have
		# the Sniffer API.
		want_snf=yes
	elif test "xxx_only" = yes; then
		# User requested something-else-only pcap, so they don't
		# want SNF support.
		want_snf=no
	else
		#
		# Use Sniffer API if present, otherwise don't
		#
		want_snf=ifpresent
	fi
])

AC_ARG_WITH([snf-includes],
AC_HELP_STRING([--with-snf-includes=IDIR],[Myricom SNF include directory, if not DIR/include]),
[
	# User wants SNF with specific header directory
	want_snf=yes
	snf_include_dir=$withval
],[])

AC_ARG_WITH([snf-libraries],
AC_HELP_STRING([--with-snf-libraries=LDIR],[Myricom SNF library directory, if not DIR/lib]),
[
	# User wants SNF with specific lib directory
	want_snf=yes
	snf_lib_dir=$withval
],[])

ac_cv_lbl_snf_api=no
if test "$with_snf" != no; then

	AC_MSG_CHECKING(whether we have Myricom Sniffer API)

	# If necessary, set default paths for Sniffer headers and libraries.
	if test -z "$snf_root"; then
		snf_root=/opt/snf
	fi

	if test -z "$snf_include_dir"; then
		snf_include_dir="$snf_root/include"
	fi

	if test -z "$snf_lib_dir"; then
		snf_lib_dir="$snf_root/lib"
	fi

	if test -f "$snf_include_dir/snf.h"; then
		# We found a header; make sure we can link with the library
		saved_ldflags=$LDFLAGS
		LDFLAGS="$LDFLAGS -L$snf_lib_dir"
		AC_CHECK_LIB([snf], [snf_init], [ac_cv_lbl_snf_api="yes"])
		LDFLAGS="$saved_ldflags"
		if test "$ac_cv_lbl_snf_api" = no; then
			AC_MSG_ERROR(SNF API cannot correctly be linked; check config.log)
		fi
	fi

	if test "$ac_cv_lbl_snf_api" = yes; then
		AC_MSG_RESULT([yes ($snf_root)])

		V_INCLS="$V_INCLS -I$snf_include_dir"
		LIBS="$LIBS -lsnf"
		LDFLAGS="$LDFLAGS -L$snf_lib_dir"

		if test "$V_PCAP" != snf ; then
			SSRC="$SSRC pcap-snf.c"
		fi

		AC_DEFINE(HAVE_SNF_API, 1, [define if you have the Myricom SNF API])
	else
		AC_MSG_RESULT(no)

		if test "$want_snf" = yes; then
			# User requested "snf" capture type but
			# we couldn't find the Sniffer API support.
			AC_MSG_ERROR([Myricom Sniffer support requested with --with-pcap=snf, but the Sniffer headers weren't found at $snf_include_dir: make sure the Sniffer support is installed, specify a different path or paths if necessary, or don't request Sniffer support])
		fi

		if test "$want_snf" = yes; then
			AC_MSG_ERROR([Myricom Sniffer support requested with --with-snf, but the Sniffer headers weren't found at $snf_include_dir: make sure the Sniffer support is installed, specify a different path or paths if necessary, or don't request Sniffer support])
		fi
	fi
fi

# Check for Riverbed TurboCap support.
AC_ARG_WITH([turbocap],
AC_HELP_STRING([--with-turbocap@<:@=DIR@:>@],[include Riverbed TurboCap support (located in directory DIR, if supplied).  @<:@default=yes, if present@:>@]),
[
	if test "$withval" = no
	then
		# User explicitly doesn't want TurboCap
		want_turbocap=no
	elif test "$withval" = yes
	then
		# User wants TurboCap support but hasn't specified a directory.
		want_turbocap=yes
	else
		# User wants TurboCap support with a specified directory.
		want_turbocap=yes
		turbocap_root=$withval
	fi
],[
	if test "xxx_only" = yes; then
		# User requested something-else-only pcap, so they don't
		# want TurboCap support.
		want_turbocap=no
	else
		#
		# Use TurboCap API if present, otherwise don't
		#
		want_turbocap=ifpresent
	fi
])

ac_cv_lbl_turbocap_api=no
if test "$want_turbocap" != no; then

	AC_MSG_CHECKING(whether TurboCap is supported)

	save_CFLAGS="$CFLAGS"
	save_LIBS="$LIBS"
	if test ! -z "$turbocap_root"; then
		TURBOCAP_CFLAGS="-I$turbocap_root/include"
		TURBOCAP_LIBS="-L$turbocap_root/lib"
		CFLAGS="$CFLAGS $TURBOCAP_CFLAGS"
	fi

	AC_TRY_COMPILE(
	[
	    #include <TcApi.h>
	],
	[
	    TC_INSTANCE a; TC_PORT b; TC_BOARD c;
	    TC_INSTANCE i;
	    (void)TcInstanceCreateByName("foo", &i);
	],
	ac_cv_lbl_turbocap_api=yes)

	CFLAGS="$save_CFLAGS"
	if test $ac_cv_lbl_turbocap_api = yes; then
		AC_MSG_RESULT(yes)

		SSRC="$SSRC pcap-tc.c"
		V_INCLS="$V_INCLS $TURBOCAP_CFLAGS"
		LIBS="$LIBS $TURBOCAP_LIBS -lTcApi -lpthread -lstdc++"

		AC_DEFINE(HAVE_TC_API, 1, [define if you have the TurboCap API])
	else
		AC_MSG_RESULT(no)

		if test "$want_turbocap" = yes; then
	        	# User wanted Turbo support but we couldn't find it.
			AC_MSG_ERROR([TurboCap support requested with --with-turbocap, but the TurboCap headers weren't found: make sure the TurboCap support is installed or don't request TurboCap support])
		fi
	fi
fi

#
# Look for {f}lex.
#
AC_PROG_LEX
if test "$LEX" = ":"; then
	AC_MSG_ERROR([Neither flex nor lex was found.])
fi

#
# Make sure {f}lex supports the -P, --header-file, and --nounput flags
# and supports processing our scanner.l.
#
AC_CACHE_CHECK([for capable lex], tcpdump_cv_capable_lex,
	if $LEX -P pcap_ --header-file=/dev/null --nounput -t $srcdir/scanner.l > /dev/null 2>&1; then
	    tcpdump_cv_capable_lex=yes
	else
	    tcpdump_cv_capable_lex=insufficient
	fi)
if test $tcpdump_cv_capable_lex = insufficient ; then
	AC_MSG_ERROR([$LEX is insufficient to compile libpcap.
 libpcap requires Flex 2.5.31 or later, or a compatible version of lex.])
fi

#
# Look for yacc/bison/byacc.
#
AC_PROG_YACC

#
# Make sure it supports the -p flag and supports processing our
# grammar.y.
#
AC_CACHE_CHECK([for capable yacc/bison], tcpdump_cv_capable_yacc,
	if $YACC -p pcap_ -o /dev/null $srcdir/grammar.y >/dev/null 2>&1; then
	    tcpdump_cv_capable_yacc=yes
	else
	    tcpdump_cv_capable_yacc=insufficient
	fi)
if test $tcpdump_cv_capable_yacc = insufficient ; then
	AC_MSG_ERROR([$YACC is insufficient to compile libpcap.
 libpcap requires Bison, Berkeley YACC, or another YACC compatible with them.])
fi

#
# Assume, by default, no support for shared libraries and V7/BSD convention
# for man pages (file formats in section 5, miscellaneous info in section 7).
# Individual cases can override this.
#
DYEXT="none"
MAN_FILE_FORMATS=5
MAN_MISC_INFO=7
case "$host_os" in

aix*)
	dnl Workaround to enable certain features
	AC_DEFINE(_SUN,1,[define on AIX to get certain functions])

	#
	# AIX makes it fun to build shared and static libraries,
	# because they're *both* ".a" archive libraries.  We
	# build the static library for the benefit of the traditional
	# scheme of building libpcap and tcpdump in subdirectories of
	# the same directory, with tcpdump statically linked with the
	# libpcap in question, but we also build a shared library as
	# "libpcap.shareda" and install *it*, rather than the static
	# library, as "libpcap.a".
	#
	DYEXT="shareda"

	case "$V_PCAP" in

	dlpi)
		#
		# If we're using DLPI, applications will need to
		# use /lib/pse.exp if present, as we use the
		# STREAMS routines.
		#
		pseexe="/lib/pse.exp"
		AC_MSG_CHECKING(for $pseexe)
		if test -f $pseexe ; then
			AC_MSG_RESULT(yes)
			LIBS="-I:$pseexe"
		fi
		;;

	bpf)
		#
		# If we're using BPF, we need "-lodm" and "-lcfg", as
		# we use them to load the BPF module.
		#
		LIBS="-lodm -lcfg"
		;;
	esac
	;;

darwin*)
	DYEXT="dylib"
	V_CCOPT="$V_CCOPT -fno-common"
	AC_ARG_ENABLE(universal,
	AC_HELP_STRING([--disable-universal],[don't build universal on OS X]))
	if test "$enable_universal" != "no"; then
		case "$host_os" in

		darwin[0-7].*)
			#
			# Pre-Tiger.  Build only for 32-bit PowerPC; no
			# need for any special compiler or linker flags.
			#
			;;

		darwin8.[0123]*)
			#
			# Tiger, prior to Intel support.  Build for 32-bit
			# PowerPC and 64-bit PowerPC, with 32-bit PowerPC
			# first.  (I'm guessing that's what Apple does.)
			#
			V_CCOPT="$V_CCOPT -arch ppc -arch ppc64"
			LDFLAGS="$LDFLAGS -arch ppc -arch ppc64"
			;;

		darwin8.[456]*)
			#
			# Tiger, subsequent to Intel support but prior to
			# x86-64 support.  Build for 32-bit PowerPC, 64-bit
			# PowerPC, and x86, with 32-bit PowerPC first.
			# (I'm guessing that's what Apple does.)
			#
			V_CCOPT="$V_CCOPT -arch ppc -arch ppc64 -arch i386"
			LDFLAGS="$LDFLAGS -arch ppc -arch ppc64 -arch i386"
			;;

		darwin8.*)
			#
			# All other Tiger, so subsequent to x86-64
			# support. Build for 32-bit PowerPC, 64-bit
			# PowerPC, x86, and x86-64, and with 32-bit PowerPC
			# first.  (I'm guessing that's what Apple does.)
			#
			V_CCOPT="$V_CCOPT -arch ppc -arch ppc64 -arch i386 -arch x86_64"
			LDFLAGS="$LDFLAGS -arch ppc -arch ppc64 -arch i386 -arch x86_64"
			;;

		darwin9.*)
			#
			# Leopard.  Build for 32-bit PowerPC, 64-bit
			# PowerPC, x86, and x86-64, with 32-bit PowerPC
			# first.  (That's what Apple does.)
			#
			V_CCOPT="$V_CCOPT -arch ppc -arch ppc64 -arch i386 -arch x86_64"
			LDFLAGS="$LDFLAGS -arch ppc -arch ppc64 -arch i386 -arch x86_64"
			;;

		darwin10.*)
			#
			# Snow Leopard.  Build for x86-64, x86, and
			# 32-bit PowerPC, with x86-64 first.  (That's
			# what Apple does, even though Snow Leopard
			# doesn't run on PPC, so PPC libpcap runs under
			# Rosetta, and Rosetta doesn't support BPF
			# ioctls, so PPC programs can't do live
			# captures.)
			#
			V_CCOPT="$V_CCOPT -arch x86_64 -arch i386 -arch ppc"
			LDFLAGS="$LDFLAGS -arch x86_64 -arch i386 -arch ppc"
			;;

		darwin*)
			#
			# Post-Snow Leopard.  Build for x86-64 and
			# x86, with x86-64 first.  (That's probably what
			# Apple does, given that Rosetta is gone.)
			# XXX - update if and when Apple drops support
			# for 32-bit x86 code.
			#
			V_CCOPT="$V_CCOPT -arch x86_64 -arch i386"
			LDFLAGS="$LDFLAGS -arch x86_64 -arch i386"
			;;
		esac
	fi
	;;

hpux9*)
	AC_DEFINE(HAVE_HPUX9,1,[on HP-UX 9.x])

	#
	# Use System V conventions for man pages.
	#
	MAN_FILE_FORMATS=4
	MAN_MISC_INFO=5
	;;

hpux10.0*)

	#
	# Use System V conventions for man pages.
	#
	MAN_FILE_FORMATS=4
	MAN_MISC_INFO=5
	;;

hpux10.1*)

	#
	# Use System V conventions for man pages.
	#
	MAN_FILE_FORMATS=4
	MAN_MISC_INFO=5
	;;

hpux*)
	dnl HPUX 10.20 and above is similar to HPUX 9, but
	dnl not the same....
	dnl
	dnl XXX - DYEXT should be set to "sl" if this is building
	dnl for 32-bit PA-RISC, but should be left as "so" for
	dnl 64-bit PA-RISC or, I suspect, IA-64.
	AC_DEFINE(HAVE_HPUX10_20_OR_LATER,1,[on HP-UX 10.20 or later])
	if test "`uname -m`" = "ia64"; then
		DYEXT="so"
	else
		DYEXT="sl"
	fi

	#
	# "-b" builds a shared library; "+h" sets the soname.
	#
	SHLIB_OPT="-b"
	SONAME_OPT="+h"

	#
	# Use System V conventions for man pages.
	#
	MAN_FILE_FORMATS=4
	MAN_MISC_INFO=5
	;;

irix*)
	#
	# Use System V conventions for man pages.
	#
	MAN_FILE_FORMATS=4
	MAN_MISC_INFO=5
	;;

linux*|freebsd*|netbsd*|openbsd*|dragonfly*|kfreebsd*|gnu*)
	DYEXT="so"

	#
	# Compiler assumed to be GCC; run-time linker may require a -R
	# flag.
	#
	if test "$libdir" != "/usr/lib"; then
		V_RFLAGS=-Wl,-R$libdir
	fi
	;;

osf*)
	DYEXT="so"

	#
	# Use System V conventions for man pages.
	#
	MAN_FILE_FORMATS=4
	MAN_MISC_INFO=5
	;;

sinix*)
	AC_MSG_CHECKING(if SINIX compiler defines sinix)
	AC_CACHE_VAL(ac_cv_cc_sinix_defined,
		AC_TRY_COMPILE(
		    [],
		    [int i = sinix;],
		    ac_cv_cc_sinix_defined=yes,
		    ac_cv_cc_sinix_defined=no))
	    AC_MSG_RESULT($ac_cv_cc_sinix_defined)
	    if test $ac_cv_cc_sinix_defined = no ; then
		    AC_DEFINE(sinix,1,[on sinix])
	    fi
	;;

solaris*)
	AC_DEFINE(HAVE_SOLARIS,1,[On solaris])

	DYEXT="so"
	#
	# Use System V conventions for man pages.
	#
	MAN_FILE_FORMATS=4
	MAN_MISC_INFO=5
	;;
esac

AC_ARG_ENABLE(shared,
AC_HELP_STRING([--enable-shared],[build shared libraries @<:@default=yes, if support available@:>@]))
test "x$enable_shared" = "xno" && DYEXT="none"

AC_PROG_RANLIB
AC_CHECK_TOOL([AR], [ar])

AC_PROG_LN_S
AC_SUBST(LN_S)

AC_LBL_DEVEL(V_CCOPT)

AC_LBL_SOCKADDR_SA_LEN

AC_LBL_SOCKADDR_STORAGE

AC_LBL_HP_PPA_INFO_T_DL_MODULE_ID_1

AC_LBL_UNALIGNED_ACCESS

rm -f net
ln -s ${srcdir}/bpf/net net

AC_SUBST(V_CCOPT)
AC_SUBST(V_DEFS)
AC_SUBST(V_FINDALLDEVS)
AC_SUBST(V_INCLS)
AC_SUBST(V_LEX)
AC_SUBST(V_PCAP)
AC_SUBST(V_SHLIB_CMD)
AC_SUBST(V_SHLIB_OPT)
AC_SUBST(V_SONAME_OPT)
AC_SUBST(V_RPATH_OPT)
AC_SUBST(V_YACC)
AC_SUBST(ADDLOBJS)
AC_SUBST(ADDLARCHIVEOBJS)
AC_SUBST(SSRC)
AC_SUBST(DYEXT)
AC_SUBST(MAN_FILE_FORMATS)
AC_SUBST(MAN_MISC_INFO)

AC_ARG_ENABLE([usb],
[AC_HELP_STRING([--enable-usb],[enable nusb support @<:@default=yes, if support available@:>@])],
    [],
    [enable_usb=yes])

if test "xxx_only" = yes; then
	# User requested something-else-only pcap, so they don't
	# want USB support.
	enable_usb=no
fi

if test "x$enable_usb" != "xno" ; then
   dnl check for USB sniffing support
   AC_MSG_CHECKING(for USB sniffing support)
   case "$host_os" in
   linux*)
	AC_DEFINE(PCAP_SUPPORT_USB, 1, [target host supports USB sniffing])
	USB_SRC=pcap-usb-linux.c
	AC_MSG_RESULT(yes)
	ac_usb_dev_name=`udevinfo -q name -p /sys/class/usb_device/usbmon 2>/dev/null`
	if test $? -ne 0 ; then
	  ac_usb_dev_name="usbmon"
	fi
	AC_DEFINE_UNQUOTED(LINUX_USB_MON_DEV, "/dev/$ac_usb_dev_name", [path for device for USB sniffing])
	AC_MSG_NOTICE(Device for USB sniffing is /dev/$ac_usb_dev_name)
	#
	# Do we have a version of <linux/compiler.h> available?
	# If so, we might need it for <linux/usbdevice_fs.h>.
	#
	AC_CHECK_HEADERS(linux/compiler.h)
	if test "$ac_cv_header_linux_compiler_h" = yes; then
	  #
	  # Yes - include it when testing for <linux/usbdevice_fs.h>.
	  #
	  AC_CHECK_HEADERS(linux/usbdevice_fs.h,,,[#include <linux/compiler.h>])
	else
	  AC_CHECK_HEADERS(linux/usbdevice_fs.h)
	fi
	if test "$ac_cv_header_linux_usbdevice_fs_h" = yes; then
	  #
	  # OK, does it define bRequestType?  Older versions of the kernel
	  # define fields with names like "requesttype, "request", and
	  # "value", rather than "bRequestType", "bRequest", and
	  # "wValue".
	  #
	  AC_MSG_CHECKING(if usbdevfs_ctrltransfer struct has bRequestType member)
	  AC_CACHE_VAL(ac_cv_usbdevfs_ctrltransfer_has_bRequestType,
	    AC_TRY_COMPILE([
AC_INCLUDES_DEFAULT
#ifdef HAVE_SYS_BITYPES_H
#include <sys/bitypes.h>
#endif
#ifdef HAVE_LINUX_COMPILER_H
#include <linux/compiler.h>
#endif
#include <linux/usbdevice_fs.h>],
	      [u_int i = sizeof(((struct usbdevfs_ctrltransfer *)0)->bRequestType)],
	      ac_cv_usbdevfs_ctrltransfer_has_bRequestType=yes,
	      ac_cv_usbdevfs_ctrltransfer_has_bRequestType=no))
	    AC_MSG_RESULT($ac_cv_usbdevfs_ctrltransfer_has_bRequestType)
	    if test $ac_cv_usbdevfs_ctrltransfer_has_bRequestType = yes ; then
	      AC_DEFINE(HAVE_USBDEVFS_CTRLTRANSFER_BREQUESTTYPE,1,
	        [if struct usbdevfs_ctrltransfer has bRequestType])
	    fi
	fi
	;;
    freebsd*)
	#
	# This just uses BPF in FreeBSD 8.4 and later; we don't need
	# to check for anything special for capturing.
	#
	AC_MSG_RESULT([yes, in FreeBSD 8.4 and later])
	;;

    *)
	AC_MSG_RESULT(no)
	;;
esac
fi
AC_SUBST(PCAP_SUPPORT_USB)
AC_SUBST(USB_SRC)

dnl check for netfilter sniffing support
if test "xxx_only" != yes; then
	AC_MSG_CHECKING(whether the platform could support netfilter sniffing)
	case "$host_os" in
	linux*)
		AC_MSG_RESULT(yes)
		#
		# Life's too short to deal with trying to get this to compile
		# if you don't get the right types defined with
		# __KERNEL_STRICT_NAMES getting defined by some other include.
		#
		# Check whether the includes Just Work.  If not, don't turn on
		# netfilter support.
		#
		AC_MSG_CHECKING(whether we can compile the netfilter support)
		AC_CACHE_VAL(ac_cv_netfilter_can_compile,
		  AC_TRY_COMPILE([
AC_INCLUDES_DEFAULT
#include <sys/socket.h>
#include <netinet/in.h>
#include <linux/types.h>

#include <linux/netlink.h>
#include <linux/netfilter.h>
#include <linux/netfilter/nfnetlink.h>
#include <linux/netfilter/nfnetlink_log.h>
#include <linux/netfilter/nfnetlink_queue.h>],
	    [],
	    ac_cv_netfilter_can_compile=yes,
	    ac_cv_netfilter_can_compile=no))
		AC_MSG_RESULT($ac_cv_netfilter_can_compile)
		if test $ac_cv_netfilter_can_compile = yes ; then
		  AC_DEFINE(PCAP_SUPPORT_NETFILTER, 1,
		    [target host supports netfilter sniffing])
		  NETFILTER_SRC=pcap-netfilter-linux.c
		fi
		;;
	*)
		AC_MSG_RESULT(no)
		;;
	esac
fi
AC_SUBST(PCAP_SUPPORT_NETFILTER)
AC_SUBST(NETFILTER_SRC)

AC_ARG_ENABLE([bluetooth],
[AC_HELP_STRING([--enable-bluetooth],[enable Bluetooth support @<:@default=yes, if support available@:>@])],
    [],
    [enable_bluetooth=ifsupportavailable])

if test "xxx_only" = yes; then
	# User requested something-else-only pcap, so they don't
	# want Bluetooth support.
	enable_bluetooth=no
fi

if test "x$enable_bluetooth" != "xno" ; then
	dnl check for Bluetooth sniffing support
	case "$host_os" in
	linux*)
		AC_CHECK_HEADER(bluetooth/bluetooth.h,
		[
		  AC_DEFINE(PCAP_SUPPORT_BT, 1, [target host supports Bluetooth sniffing])
		  BT_SRC=pcap-bt-linux.c
		  AC_MSG_NOTICE(Bluetooth sniffing is supported)

		  #
		  # OK, does struct sockaddr_hci have an hci_channel
		  # member?
		  #
		  AC_MSG_CHECKING(if struct sockaddr_hci has hci_channel member)
		  AC_CACHE_VAL(ac_cv_lbl_sockaddr_hci_has_hci_channel,
		    AC_TRY_COMPILE(
[
#include <bluetooth/bluetooth.h>
#include <bluetooth/hci.h>
],
		      [u_int i = sizeof(((struct sockaddr_hci *)0)->hci_channel)],
		      ac_cv_lbl_sockaddr_hci_has_hci_channel=yes,
		      ac_cv_lbl_sockaddr_hci_has_hci_channel=no))
		    AC_MSG_RESULT($ac_cv_lbl_sockaddr_hci_has_hci_channel)
		    if test $ac_cv_lbl_sockaddr_hci_has_hci_channel = yes ; then
		      AC_DEFINE(SOCKADDR_HCI_HAS_HCI_CHANNEL,,
			[if struct sockaddr_hci has hci_channel member])

		      #
		      # OK, is HCI_CHANNEL_MONITOR defined?
		      #
		      AC_MSG_CHECKING(if HCI_CHANNEL_MONITOR is defined)
		      AC_CACHE_VAL(ac_cv_lbl_hci_channel_monitor_is_defined,
		      AC_TRY_COMPILE(
[
#include <bluetooth/bluetooth.h>
#include <bluetooth/hci.h>
],
			[u_int i = HCI_CHANNEL_MONITOR],
			ac_cv_lbl_hci_channel_monitor_is_defined=yes,
			ac_cv_lbl_hci_channel_monitor_is_defined=no))
		      AC_MSG_RESULT($ac_cv_lbl_hci_channel_monitor_is_defined)
		      if test $ac_cv_lbl_hci_channel_monitor_is_defined = yes ; then
			AC_DEFINE(PCAP_SUPPORT_BT_MONITOR,,
			  [target host supports Bluetooth Monitor])
			BT_MONITOR_SRC=pcap-bt-monitor-linux.c
		      fi
		    fi
		    ac_lbl_bluetooth_available=yes
		],
		ac_lbl_bluetooth_available=no
		)
		if test "x$ac_lbl_bluetooth_available" == "xno" ; then
		    if test "x$enable_bluetooth" = "xyes" ; then
			AC_MSG_ERROR(Bluetooth sniffing is not supported; install bluez-lib devel to enable it)
		    else
			AC_MSG_NOTICE(Bluetooth sniffing is not supported; install bluez-lib devel to enable it)
		    fi
		fi
		;;
	*)
		if test "x$enable_bluetooth" = "xyes" ; then
		    AC_MSG_ERROR(no Bluetooth sniffing support implemented for $host_os)
		else
		    AC_MSG_NOTICE(no Bluetooth sniffing support implemented for $host_os)
		fi
		;;
	esac
	AC_SUBST(PCAP_SUPPORT_BT)
	AC_SUBST(BT_SRC)
	AC_SUBST(BT_MONITOR_SRC)
fi

AC_ARG_ENABLE([dbus],
[AC_HELP_STRING([--enable-dbus],[enable D-Bus capture support @<:@default=yes, if support available@:>@])],
    [],
    [enable_dbus=ifavailable])

if test "xxx_only" = yes; then
	# User requested something-else-only pcap, so they don't
	# want D-Bus support.
	enable_dbus=no
fi

if test "x$enable_dbus" != "xno"; then
	if test "x$enable_dbus" = "xyes"; then
		case "$host_os" in

		darwin*)
			#
			# We don't support D-Bus sniffing on OS X; see
			#
			# https://bugs.freedesktop.org/show_bug.cgi?id=74029
			#
			# The user requested it, so fail.
			#
			AC_MSG_ERROR([Due to freedesktop.org bug 74029, D-Bus capture support is not available on OS X])
		esac
	else
		case "$host_os" in

		darwin*)
			#
			# We don't support D-Bus sniffing on OS X; see
			#
			# https://bugs.freedesktop.org/show_bug.cgi?id=74029
			#
			# The user dind't explicitly request it, so just
			# silently refuse to enable it.
			#
			enable_dbus="no"
			;;
		esac
	fi
fi

if test "x$enable_dbus" != "xno"; then
	AC_CHECK_PROG([PKGCONFIG], [pkg-config], [pkg-config], [no])
	if test "x$PKGCONFIG" != "xno"; then
		AC_MSG_CHECKING([for D-Bus])
		if "$PKGCONFIG" dbus-1; then
			AC_MSG_RESULT([yes])
			DBUS_CFLAGS=`"$PKGCONFIG" --cflags dbus-1`
			DBUS_LIBS=`"$PKGCONFIG" --libs dbus-1`
			save_CFLAGS="$CFLAGS"
			save_LIBS="$LIBS"
			CFLAGS="$CFLAGS $DBUS_CFLAGS"
			LIBS="$LIBS $DBUS_LIBS"
			AC_MSG_CHECKING(whether the D-Bus library defines dbus_connection_read_write)
			AC_TRY_LINK(
			    [#include <string.h>

			     #include <time.h>
			     #include <sys/time.h>

			     #include <dbus/dbus.h>],
			    [return dbus_connection_read_write(NULL, 0);],
			    [
				AC_MSG_RESULT([yes])
				AC_DEFINE(PCAP_SUPPORT_DBUS, 1, [support D-Bus sniffing])
				DBUS_SRC=pcap-dbus.c
				V_INCLS="$V_INCLS $DBUS_CFLAGS"
			    ],
			    [
				AC_MSG_RESULT([no])
				if test "x$enable_dbus" = "xyes"; then
				    AC_MSG_ERROR([--enable-dbus was given, but the D-Bus library doesn't define dbus_connection_read_write()])
				fi
				LIBS="$save_LIBS"
			     ])
			CFLAGS="$save_CFLAGS"
		else
			AC_MSG_RESULT([no])
			if test "x$enable_dbus" = "xyes"; then
				AC_MSG_ERROR([--enable-dbus was given, but the dbus-1 package is not installed])
			fi
		fi
	fi
	AC_SUBST(PCAP_SUPPORT_DBUS)
	AC_SUBST(DBUS_SRC)
fi

dnl check for hardware timestamp support
case "$host_os" in
linux*)
	AC_CHECK_HEADERS([linux/net_tstamp.h])
	;;
*)
	AC_MSG_NOTICE(no hardware timestamp support implemented for $host_os)
	;;
esac

dnl The packet ring capture facility of Linux, described in
dnl Documentation/networking/packet_mmap.txt, is not 32/64-bit compatible before
dnl version 2.6.27. A 32-bit kernel requires a 32-bit userland, and likewise for
dnl 64-bit. The effect of this is that a 32-bit libpcap binary will not run
dnl correctly on a 64-bit kernel (the binary will use the wrong offsets into a
dnl kernel struct). This problem was solved in Linux 2.6.27. Use
dnl --disable-packet-ring whenever a 32-bit application must run on a 64-bit
dnl target host, and either the build host or the target host run Linux 2.6.26
dnl or earlier.
AC_ARG_ENABLE([packet-ring],
[AC_HELP_STRING([--enable-packet-ring],[enable Linux packet ring support @<:@default=yes@:>@])],
,enable_packet_ring=yes)

if test "x$enable_packet_ring" != "xno" ; then
	AC_DEFINE(PCAP_SUPPORT_PACKET_RING, 1, [use Linux packet ring capture if available])
	AC_SUBST(PCAP_SUPPORT_PACKET_RING)
fi

AC_PROG_INSTALL

AC_CONFIG_HEADER(config.h)

AC_OUTPUT_COMMANDS([if test -f .devel; then
	echo timestamp > stamp-h
	cat Makefile-devel-adds >> Makefile
	make depend
fi])
AC_OUTPUT(Makefile pcap-filter.manmisc pcap-linktype.manmisc
	pcap-tstamp.manmisc pcap-savefile.manfile pcap.3pcap
	pcap_compile.3pcap pcap_datalink.3pcap pcap_dump_open.3pcap
	pcap_get_tstamp_precision.3pcap pcap_list_datalinks.3pcap
	pcap_list_tstamp_types.3pcap pcap_open_dead.3pcap
	pcap_open_offline.3pcap pcap_set_tstamp_precision.3pcap
	pcap_set_tstamp_type.3pcap)
exit 0
