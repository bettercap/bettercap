if( NOT LIBPCAP_PRECONFIGURED )
    set( LIBPCAP_PRECONFIGURED TRUE )

    ###################################################################
    #   Parameters
    ###################################################################

    option (USE_STATIC_RT "Use static Runtime" ON)

    ######################################
    # Project setings
    ######################################

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

    if (USE_STATIC_RT)
        MESSAGE( STATUS "Use STATIC runtime" )

        if( MSVC )
            set (CMAKE_CXX_FLAGS_MINSIZEREL     "${CMAKE_CXX_FLAGS_MINSIZEREL} /MT")
            set (CMAKE_CXX_FLAGS_RELWITHDEBINFO "${CMAKE_CXX_FLAGS_RELWITHDEBINFO} /MT")
            set (CMAKE_CXX_FLAGS_RELEASE        "${CMAKE_CXX_FLAGS_RELEASE} /MT")
            set (CMAKE_CXX_FLAGS_DEBUG          "${CMAKE_CXX_FLAGS_DEBUG} /MTd")

            set (CMAKE_C_FLAGS_MINSIZEREL       "${CMAKE_C_FLAGS_MINSIZEREL} /MT")
            set (CMAKE_C_FLAGS_RELWITHDEBINFO   "${CMAKE_C_FLAGS_RELWITHDEBINFO} /MT")
            set (CMAKE_C_FLAGS_RELEASE          "${CMAKE_C_FLAGS_RELEASE} /MT")
            set (CMAKE_C_FLAGS_DEBUG            "${CMAKE_C_FLAGS_DEBUG} /MTd")
        endif( MSVC )
    else (USE_STATIC_RT)
        MESSAGE( STATUS "Use DYNAMIC runtime" )

        if( MSVC )
            set (CMAKE_CXX_FLAGS_MINSIZEREL     "${CMAKE_CXX_FLAGS_MINSIZEREL} /MD")
            set (CMAKE_CXX_FLAGS_RELWITHDEBINFO "${CMAKE_CXX_FLAGS_RELWITHDEBINFO} /MD")
            set (CMAKE_CXX_FLAGS_RELEASE        "${CMAKE_CXX_FLAGS_RELEASE} /MD")
            set (CMAKE_CXX_FLAGS_DEBUG          "${CMAKE_CXX_FLAGS_DEBUG} /MDd")

            set (CMAKE_C_FLAGS_MINSIZEREL       "${CMAKE_C_FLAGS_MINSIZEREL} /MD")
            set (CMAKE_C_FLAGS_RELWITHDEBINFO   "${CMAKE_C_FLAGS_RELWITHDEBINFO} /MD")
            set (CMAKE_C_FLAGS_RELEASE          "${CMAKE_C_FLAGS_RELEASE} /MD")
            set (CMAKE_C_FLAGS_DEBUG            "${CMAKE_C_FLAGS_DEBUG} /MDd")
        endif( MSVC )
    endif (USE_STATIC_RT)
endif( NOT LIBPCAP_PRECONFIGURED )
