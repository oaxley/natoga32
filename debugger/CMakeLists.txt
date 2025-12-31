cmake_minimum_required(VERSION 3.25)

# SDL2 required for imgui
find_package(SDL2 REQUIRED)

# CMake fetch argparse library
include(FetchContent)
FetchContent_Declare(
    argparse
    GIT_REPOSITORY https://github.com/p-ranav/argparse.git
)
FetchContent_MakeAvailable(argparse)

# create the Virtual Console debugger binary
add_executable(vcdebug
    src/main.cpp
)

# include directory
target_include_directories(vcdebug
    PUBLIC
        ${SDL2_INCLUDE_DIRS}
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# link targets
target_link_libraries(vcdebug
    PUBLIC
        ${SDL2_LIBRARIES}
    PRIVATE
        argparse
        imgui
        vc
)
