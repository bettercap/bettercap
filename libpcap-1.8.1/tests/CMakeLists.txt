cmake_minimum_required( VERSION 2.8.8 )

project( pcap_tests )

######################################
# Register targets
######################################

if( MSVC )
    file(GLOB PROJECT_SOURCE_LIST_WIN32_C
        ${pcap_SOURCE_DIR}/missing/getopt.c
    )
endif( MSVC )

ADD_EXECUTABLE (can_set_rfmon_test can_set_rfmon_test.c ${PROJECT_SOURCE_LIST_WIN32_C} )
target_link_libraries ( can_set_rfmon_test pcap )

ADD_EXECUTABLE (capturetest capturetest.c ${PROJECT_SOURCE_LIST_WIN32_C} )
target_link_libraries ( capturetest pcap )

ADD_EXECUTABLE (filtertest filtertest.c ${PROJECT_SOURCE_LIST_WIN32_C} )
target_link_libraries ( filtertest pcap )

ADD_EXECUTABLE (indalldevstest findalldevstest.c ${PROJECT_SOURCE_LIST_WIN32_C} )
target_link_libraries ( indalldevstest pcap )

ADD_EXECUTABLE (opentest opentest.c ${PROJECT_SOURCE_LIST_WIN32_C} )
target_link_libraries ( opentest pcap )

#ADD_EXECUTABLE (pcap_compile_test pcap_compile_test.c ${PROJECT_SOURCE_LIST_WIN32_C} )
#target_link_libraries ( pcap_compile_test pcap )

ADD_EXECUTABLE (reactivatetest reactivatetest.c ${PROJECT_SOURCE_LIST_WIN32_C} )
target_link_libraries ( reactivatetest pcap )

if( NOT WIN32 )
    ADD_EXECUTABLE (selpolltest selpolltest.c ${PROJECT_SOURCE_LIST_WIN32_C} )
    target_link_libraries ( selpolltest pcap )
endif()

ADD_EXECUTABLE (valgrindtest valgrindtest.c ${PROJECT_SOURCE_LIST_WIN32_C} )
target_link_libraries ( valgrindtest pcap )
