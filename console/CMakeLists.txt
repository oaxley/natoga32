cmake_minimum_required(VERSION 3.25)

# library name
set(VC_NAME "vc")

# source files
set(VC_SOURCES
    src/core/cpu.cpp
    src/core/memory.cpp
    src/vc.cpp
)

set(VC_HEADERS
    src/core/constants.h
    src/core/csr.h
    src/core/cpu.h
    src/core/decoders.h
    src/core/helpers.h
    src/core/memory.h
    src/vc.h
)

# library target
if(VC_BUILD_SHARED)
    add_library(${VC_NAME} SHARED ${VC_SOURCES} ${VC_HEADERS})
else()
    add_library(${VC_NAME} STATIC ${VC_SOURCES} ${VC_HEADERS})
endif()


target_include_directories(${VC_NAME}
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# enable position-independent code for shared library
set_target_properties(${VC_NAME} PROPERTIES
    POSITION_INDEPENDENT_CODE ON
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
)

# Installation
install(TARGETS ${VC_NAME}
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(DIRECTORY include/${VC_NAME}
    DESTINATION include
)
