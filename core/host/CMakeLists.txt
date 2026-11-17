cmake_minimum_required(VERSION 3.25)

# SDL3 required for the host
find_package(SDL3 REQUIRED)

# CMake fetch argparse library
include(FetchContent)
FetchContent_Declare(
    argparse
    GIT_REPOSITORY https://github.com/p-ranav/argparse.git
)
FetchContent_MakeAvailable(argparse)

# create the Virtual Console host binary
add_executable(vchost
    src/main.cpp
    src/host_config.h
    src/console_runner.h
    src/console_runner.cpp
    src/frame_pacer.h
)

# include directories
target_include_directories(vchost
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# link targets
target_link_libraries(vchost
    PRIVATE
        SDL3::SDL3
        argparse
        vc
)

# the host loads the boot ROM built next to it
if(TARGET rombios)
    add_dependencies(vchost rombios)
endif()
