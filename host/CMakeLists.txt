cmake_minimum_required(VERSION 3.25)

# SDL2 required for the host
find_package(SDL2 REQUIRED)

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
)

# include directories
target_include_directories(vchost
    PUBLIC
        ${SDL2_INCLUDE_DIRS}
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# link targets
target_link_libraries(vchost
    PUBLIC
        ${SDL2_LIBRARIES}
    PRIVATE
        argparse
        vc
)
