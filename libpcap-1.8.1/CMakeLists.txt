cmake_minimum_required( VERSION 2.8.8 )

project( pcap )
#
# Call the library "wpcap" on Windows, for backwards compatibility.
#
if( WIN32 )
    set( LIBRARY_NAME wpcap )
else()
    set( LIBRARY_NAME pcap )
endif()

###################################################################
#   Parameters
###################################################################

option (INET6 "Enable IPv6" ON)
if( MSVC )
    option (USE_STATIC_RT "Use static Runtime" ON)
endif( MSVC )
option (BUILD_SHARED_LIBS "Build shared libraries" ON)
if( WIN32 )
    set(PACKET_DLL_DIR "" CACHE PATH "Path to directory with include and lib subdirectories for packet.dll")
endif( WIN32 )

#
# XXX - this should be an option, defaulting to "yes" for Windows and to
# "no", for now, on UN*X.
#
if( WIN32 )
    set( HAVE_REMOTE 1 )
endif( WIN32 )

######################################
# Project settings
######################################

add_definitions( -DHAVE_CONFIG_H )

include_directories(
    ${CMAKE_CURRENT_BINARY_DIR}
    ${pcap_SOURCE_DIR}
)

if( WIN32 )
    if( NOT "${PACKET_DLL_DIR}" STREQUAL "" )
        include_directories("${PACKET_DLL_DIR}/Include")
        if( CMAKE_CL_64 )
            link_directories("${PACKET_DLL_DIR}/Lib/x64")
        else( CMAKE_CL_64 )
            link_directories("${PACKET_DLL_DIR}/Lib")
        endif( CMAKE_CL_64 )
    endif()
    include_directories(
        ../Common/
        Win32/Include
    )
endif( WIN32)

add_definitions( -DBUILDING_PCAP )

if( MSVC )
    add_definitions( -D__STDC__ )
    add_definitions( -D_CRT_SECURE_NO_WARNINGS )
    add_definitions( "-D_U_=" )
elseif( CMAKE_COMPILER_IS_GNUCXX )
    add_definitions( "-D_U_=__attribute__((unused))" )
else(MSVC)
    add_definitions( "-D_U_=" )
endif( MSVC )

if( MSVC )
    if (USE_STATIC_RT)
        MESSAGE( STATUS "Use STATIC runtime" )
        set(NAME_RT MT)
        set (CMAKE_CXX_FLAGS_MINSIZEREL     "${CMAKE_CXX_FLAGS_MINSIZEREL} /MT")
        set (CMAKE_CXX_FLAGS_RELWITHDEBINFO "${CMAKE_CXX_FLAGS_RELWITHDEBINFO} /MT")
        set (CMAKE_CXX_FLAGS_RELEASE        "${CMAKE_CXX_FLAGS_RELEASE} /MT")
        set (CMAKE_CXX_FLAGS_DEBUG          "${CMAKE_CXX_FLAGS_DEBUG} /MTd")

        set (CMAKE_C_FLAGS_MINSIZEREL       "${CMAKE_C_FLAGS_MINSIZEREL} /MT")
        set (CMAKE_C_FLAGS_RELWITHDEBINFO   "${CMAKE_C_FLAGS_RELWITHDEBINFO} /MT")
        set (CMAKE_C_FLAGS_RELEASE          "${CMAKE_C_FLAGS_RELEASE} /MT")
        set (CMAKE_C_FLAGS_DEBUG            "${CMAKE_C_FLAGS_DEBUG} /MTd")
    else (USE_STATIC_RT)
        MESSAGE( STATUS "Use DYNAMIC runtime" )
        set(NAME_RT MD)
        set (CMAKE_CXX_FLAGS_MINSIZEREL     "${CMAKE_CXX_FLAGS_MINSIZEREL} /MD")
        set (CMAKE_CXX_FLAGS_RELWITHDEBINFO "${CMAKE_CXX_FLAGS_RELWITHDEBINFO} /MD")
        set (CMAKE_CXX_FLAGS_RELEASE        "${CMAKE_CXX_FLAGS_RELEASE} /MD")
        set (CMAKE_CXX_FLAGS_DEBUG          "${CMAKE_CXX_FLAGS_DEBUG} /MDd")

        set (CMAKE_C_FLAGS_MINSIZEREL       "${CMAKE_C_FLAGS_MINSIZEREL} /MD")
        set (CMAKE_C_FLAGS_RELWITHDEBINFO   "${CMAKE_C_FLAGS_RELWITHDEBINFO} /MD")
        set (CMAKE_C_FLAGS_RELEASE          "${CMAKE_C_FLAGS_RELEASE} /MD")
        set (CMAKE_C_FLAGS_DEBUG            "${CMAKE_C_FLAGS_DEBUG} /MDd")
   endif (USE_STATIC_RT)
endif( MSVC )

###################################################################
#   Detect available platform features
###################################################################

include(CheckIncludeFile)
include(CheckFunctionExists)
include(CheckStructHasMember)
include(CheckTypeSize)

#
# Header files.
#
check_include_file( inttypes.h HAVE_INTTYPES_H )
check_include_file( stdint.h HAVE_STDINT_H )
check_include_file( unistd.h HAVE_UNISTD_H )
if( NOT HAVE_UNISTD_H )
    add_definitions( -DYY_NO_UNISTD_H )
endif( NOT HAVE_UNISTD_H )
check_include_file( bitypes.h HAVE_SYS_BITYPES_H )
check_include_file( limits.h HAVE_LIMITS_H )

#
# Functions.
#
check_function_exists( strerror HAVE_STRERROR )
check_function_exists( strlcpy HAVE_STRLCPY )
check_function_exists( snprintf HAVE_SNPRINTF )
check_function_exists( vsnprintf HAVE_VSNPRINTF )
check_function_exists( strtok_r HAVE_STRTOK_R )

if (WIN32)
    #
    # Check for Windows-only functions, such as packet.dll functions.
    #
    check_function_exists( PacketIsLoopbackAdapter HAVE_PACKET_IS_LOOPBACK_ADAPTER )
endif()

#
# Data types.
#
# XXX - there's no check_struct() macro that's like check_struct_has_member()
# except that it only checks for the existence of the structure type,
# so we use check_struct_has_member() and look for ss_family.
#
check_struct_has_member("struct sockaddr_storage" ss_family sys/socket.h  HAVE_SOCKADDR_STORAGE)
set(CMAKE_EXTRA_INCLUDE_FILES unistd.h sys/socket.h)
check_type_size("socklen_t" SOCKLEN_T)
set(CMAKE_EXTRA_INCLUDE_FILES unistd.h)

#
# Structure fields.
#
check_struct_has_member("struct sockaddr" sa_len sys/socket.h HAVE_SOCKADDR_SA_LEN )

if( INET6 )
    MESSAGE( STATUS "Use IPv6" )
endif( INET6 )

if( WIN32 )
    add_definitions( -DHAVE_ADDRINFO )
endif( WIN32 )

######################################
# External dependencies
######################################

######################################
# Input files
######################################

set(PROJECT_SOURCE_LIST_C
    bpf_dump.c
    bpf_image.c
    etherent.c
    fad-helpers.c
    gencode.c
    inet.c
    nametoaddr.c
    optimize.c
    pcap-common.c
    pcap.c
    savefile.c
    sf-pcap-ng.c
    sf-pcap.c
    bpf/net/bpf_filter.c
)

if( WIN32 )
    set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} missing/win_snprintf.c )
else()
    if( NOT HAVE_SNPRINTF )
        set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} missing/snprintf.c )
    endif( NOT HAVE_SNPRINTF )
    if( NOT HAVE_STRTOK_R )
        set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} missing/strtok_r.c )
    endif( NOT HAVE_STRTOK_R )
endif( WIN32 )

if( HAVE_REMOTE )
    set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C}
        pcap-new.c pcap-rpcap.c sockutils.c)
endif( HAVE_REMOTE )

#
# Determine the main pcap-XXX.c file to use.
#
if( WIN32 )
    #
    # WinPcap.
    #
    set( PCAP_TYPE win32 )
else()
    #
    # UN*X - figure out what type of packet capture mechanism we
    # have.
    #
    if( EXISTS /dev/bpf )
	#
	# Cloning BPF device.
	#
	set( PCAP_TYPE bpf )
	AC_DEFINE(HAVE_CLONING_BPF,1,[define if you have a cloning BPF device])
    elseif( EXISTS /dev/bpf0 )
	set( PCAP_TYPE bpf )

        #
        # XXX - many more BPF checks.
        #
    elseif( EXISTS /usr/include/net/pfilt.h )
        #
        # DEC OSF/1, Digital UNIX, Tru64 UNIX
        #
	set( PCAP_TYPE pf )
    elseif( EXISTS /dev/enet )
        set( PCAP_TYPE enet )
    elseif( EXISTS /dev/nit )
        set( PCAP_TYPE snit )
    elseif( EXISTS /usr/include/sys/net/nit.h )
        set( PCAP_TYPE nit )
    elseif( EXISTS /usr/include/linux/socket.h )
        set( PCAP_TYPE linux )

	#
	# Do we have the wireless extensions?
	#
        check_include_file( linux/wireless.h HAVE_LINUX_WIRELESS_H )

        #
        # XXX - many more Linux checks.
	#
    elseif( EXISTS /usr/include/net/raw.h )
        set( PCAP_TYPE snoop )
    elseif( EXISTS /usr/include/odmi.h )
        #
        # On AIX, the BPF devices might not yet be present - they're
        # created the first time libpcap runs after booting.
        # We check for odmi.h instead.
        #
        set( PCAP_TYPE bpf )
    elseif( /usr/include/sys/dlpi.h )
        set( PCAP_TYPE dlpi )

        #
        # XXX - many more DLPI checks.
        #
    else()
	set( PCAP_TYPE null )
    endif()
endif( WIN32 )
message(STATUS "Packet capture mechanism type: ${PCAP_TYPE}")
set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} pcap-${PCAP_TYPE}.c)

#
# Now figure out how we get a list of interfaces and addresses,
# if we support capturing.  Don't bother if we don't support
# capturing.
#
if( NOT WIN32 )
    #
    # UN*X - figure out what type of interface list mechanism we
    # have.
    #
    if( ${PCAP_TYPE} STREQUAL "null" )
        #
        # We can't capture, so we can't open any capture
        # devices, so we won't return any interfaces.
        #
        set( FINDALLDEVS_TYPE null )
    else()
        check_function_exists( getifaddrs HAVE_GETIFADDRS )
        if( ${HAVE_GETIFADDRS} )
            #
            # We have "getifaddrs()"; make sure we have <ifaddrs.h>
            # as well, just in case some platform is really weird.
            #
            check_include_file( ifaddrs.h HAVE_IFADDRS_H )
            if( ${HAVE_IFADDRS_H} )
                #
                # We have the header, so we use "getifaddrs()" to
                # get the list of interfaces.
                #
                set( FINDALLDEVS_TYPE getad )
            else()
                #
                # We don't have the header - give up.
                # XXX - we could also fall back on some other
                # mechanism, but, for now, this'll catch this
                # problem so that we can at least try to figure
                # out something to do on systems with "getifaddrs()"
                # but without "ifaddrs.h", if there is something
                # we can do on those systems.
                #
                message(FATAL_ERROR "Your system has getifaddrs() but doesn't have a usable <ifaddrs.h>." )
            endif()
        else()
            #
            # Well, we don't have "getifaddrs()", so we have to use
            # some other mechanism; determine what that mechanism is.
            #
            # The first thing we use is the type of capture mechanism,
            # which is somewhat of a proxy for the OS we're using.
            #
            if( ${PCAP_TYPE} STREQUAL "dlpi" OR ${PCAP_TYPE} STREQUAL "libdlpi" )
                #
                # This might be Solaris 8 or later, with
                # SIOCGLIFCONF, or it might be some other OS
                # or some older version of Solaris, with
                # just SIOCGIFCONF.
                #
                try_compile( HAVE_SIOCGLIFCONF ${CMAKE_CURRENT_BINARY_DIR} "${pcap_SOURCE_DIR}/config/have_siocglifconf.c"  )
                message( STATUS "HAVE_SIOCGLIFCONF = ${HAVE_SIOCGLIFCONF}" )
                if( HAVE_SIOCGLIFCONF )
                    set( FINDALLDEVS_TYPE glifc )
                else()
                    set( FINDALLDEVS_TYPE gifc )
                endif()
            else()
                #
                # Assume we just have SIOCGIFCONF.
                # (XXX - on at least later Linux kernels, there's
                # another mechanism, and we should be using that
                # instead.)
                #
                set( FINDALLDEVS_TYPE gifc )
            endif()
        endif()
    endif()
    message(STATUS "Find-interfaces mechanism type: ${FINDALLDEVS_TYPE}")
    set( PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} fad-${FINDALLDEVS_TYPE}.c )
endif()

file(GLOB PROJECT_SOURCE_LIST_CORE_H
    *.h
    pcap/*.h
)
set( PROJECT_SOURCE_LIST_H ${PROJECT_SOURCE_LIST_H} ${PROJECT_SOURCE_LIST_CORE_H} )

if( WIN32 )
    file(GLOB PROJECT_SOURCE_LIST_WIN32_H
        Win32/Include/*.h
    )
    set( PROJECT_SOURCE_LIST_H ${PROJECT_SOURCE_LIST_H} ${PROJECT_SOURCE_LIST_WIN32_H} )
endif( WIN32 )

#
# {Flex} and YACC/Berkeley YACC/Bison.
# From a mail message to the CMake mailing list by Andy Cedilnik of
# Kitware.
#

#
# Try to find Flex, a Windows version of Flex, or Lex.
#
find_program(LEX_EXECUTABLE NAMES flex win_flex lex)
if( ${LEX_EXECUTABLE} STREQUAL "LEX_EXECUTABLE-NOTFOUND" )
    message(FATAL_ERROR "Neither flex nor win_flex nor lex was found." )
endif()
message(STATUS "Lexical analyzer generator: ${LEX_EXECUTABLE}")

add_custom_command(
    OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/scanner.c ${CMAKE_CURRENT_BINARY_DIR}/scanner.h
    SOURCE ${pcap_SOURCE_DIR}/scanner.l
    COMMAND ${LEX_EXECUTABLE} -P pcap_ --header-file=scanner.h --nounput -o${CMAKE_CURRENT_BINARY_DIR}/scanner.c ${pcap_SOURCE_DIR}/scanner.l
    DEPENDS ${pcap_SOURCE_DIR}/scanner.l
)

#
# Since scanner.c does not exist yet when cmake is run, mark
# it as generated.
#
# Since scanner.c includes grammar.h, mark that as a dependency.
#
set_source_files_properties(${CMAKE_CURRENT_BINARY_DIR}/scanner.c PROPERTIES
    GENERATED TRUE
    OBJECT_DEPENDS ${CMAKE_CURRENT_BINARY_DIR}/scanner.h
)

#
# Add scanner.c to the list of sources.
#
#set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} ${CMAKE_CURRENT_BINARY_DIR}/scanner.c)

#
# Try to find YACC or Bison.
#
find_program(YACC_EXECUTABLE NAMES bison win_bison byacc yacc)
if( ${YACC_EXECUTABLE} STREQUAL "YACC_EXECUTABLE-NOTFOUND" )
    message(FATAL_ERROR "Neither bison nor win_bison nor byacc nor yacc was found." )
endif()
message(STATUS "Parser generator: ${YACC_EXECUTABLE}")

#
# Create custom command for the scanner.
# Find out whether it's Bison or notby looking at the last component
# of the path (without a .exe extension, if this is Windows).
#
get_filename_component(YACC_NAME ${YACC_EXECUTABLE} NAME_WE)
if( "${YACC_NAME}" STREQUAL "bison" OR "${YACC_NAME}" STREQUAL "win_bison" )
    set( YACC_COMPATIBILITY_FLAG "-y" )
endif()
add_custom_command(
    OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/grammar.c ${CMAKE_CURRENT_BINARY_DIR}/grammar.h
    SOURCE ${pcap_SOURCE_DIR}/grammar.y
    COMMAND ${YACC_EXECUTABLE} ${YACC_COMPATIBILITY_FLAG} -p pcap_ -o ${CMAKE_CURRENT_BINARY_DIR}/grammar.c -d ${pcap_SOURCE_DIR}/grammar.y
    DEPENDS ${pcap_SOURCE_DIR}/grammar.y
)

#
# Since grammar.c does not exists yet when cmake is run, mark
# it as generated.
#
set_source_files_properties(${CMAKE_CURRENT_BINARY_DIR}/grammar.c PROPERTIES
    GENERATED TRUE
)

#
# Add grammar.c to the list of sources.
#
#set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} ${CMAKE_CURRENT_BINARY_DIR}/grammar.c)

if( WIN32 )
    #
    # CMake does not love Windows.
    #
    file(TO_NATIVE_PATH "${pcap_SOURCE_DIR}/GenVersion.bat" GenVersion_path)
    file(TO_NATIVE_PATH "${pcap_SOURCE_DIR}/VERSION" VERSION_path)
    file(TO_NATIVE_PATH "${pcap_SOURCE_DIR}/pcap_version.h.in" version_h_in_path)
    file(TO_NATIVE_PATH "${CMAKE_CURRENT_BINARY_DIR}/pcap_version.h" version_h_path)
    add_custom_command(
        OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/pcap_version.h
        SOURCE ${pcap_SOURCE_DIR}/VERSION ${pcap_SOURCE_DIR}/pcap_version.h.in
        COMMAND ${GenVersion_path} ${VERSION_path} ${version_h_in_path} ${version_h_path}
        DEPENDS ${pcap_SOURCE_DIR}/VERSION ${pcap_SOURCE_DIR}/pcap_version.h.in
    )
else( WIN32 )
    add_custom_command(
        OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/version.c
        SOURCE ${pcap_SOURCE_DIR}/VERSION
        COMMAND ${pcap_SOURCE_DIR}/gen_version_c.sh ${pcap_SOURCE_DIR}/VERSION ${CMAKE_CURRENT_BINARY_DIR}/version.c
        DEPENDS ${pcap_SOURCE_DIR}/VERSION
    )
    add_custom_command(
        OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/pcap_version.h
        SOURCE ${pcap_SOURCE_DIR}/VERSION
        COMMAND ${pcap_SOURCE_DIR}/gen_version_header.sh ${pcap_SOURCE_DIR}/VERSION ${pcap_SOURCE_DIR}/pcap_version.h.in ${CMAKE_CURRENT_BINARY_DIR}/pcap_version.h
        DEPENDS ${pcap_SOURCE_DIR}/VERSION
    )

    #
    # Since version.c does not exists yet when cmake is run, mark
    # it as generated.
    #
    set_source_files_properties(${CMAKE_CURRENT_BINARY_DIR}/version.c PROPERTIES
        GENERATED TRUE
    )

    #
    # Add version.c to the list of sources.
    #
    set(PROJECT_SOURCE_LIST_C ${PROJECT_SOURCE_LIST_C} ${CMAKE_CURRENT_BINARY_DIR}/version.c)
endif( WIN32 )

#
# Since pcap_version.h does not exists yet when cmake is run, mark
# it as generated.
#
set_source_files_properties(${CMAKE_CURRENT_BINARY_DIR}/pcap_version.h PROPERTIES
    GENERATED TRUE
)

#
# Add pcap_version.h to the list of headers.
#
set(PROJECT_SOURCE_LIST_H ${PROJECT_SOURCE_LIST_H} ${CMAKE_CURRENT_BINARY_DIR}/pcap_version.h)

source_group("Source Files" FILES ${PROJECT_SOURCE_LIST_C})
source_group("Header Files" FILES ${PROJECT_SOURCE_LIST_H})

######################################
# Register targets
######################################

add_library(${LIBRARY_NAME}
    ${PROJECT_SOURCE_LIST_C}
    ${CMAKE_CURRENT_BINARY_DIR}/grammar.c
    ${CMAKE_CURRENT_BINARY_DIR}/scanner.c
    ${PROJECT_SOURCE_LIST_H}
)

if( WIN32 )
    target_link_libraries ( ${LIBRARY_NAME}
        packet
        ws2_32
    )
endif( WIN32 )

######################################
# Write out the config.h file
######################################

configure_file(${CMAKE_CURRENT_SOURCE_DIR}/cmakeconfig.h.in ${CMAKE_CURRENT_BINARY_DIR}/config.h)
