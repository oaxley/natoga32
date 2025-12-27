cmake_minimum_required(VERSION 3.25)
project(natoga32 VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

# build options
option(VC_BUILD_DEBUGGER "Build debugger" ON)
option(VC_BUILD_TESTS "Build tests" ON)
option(VC_BUILD_HOST "Build host runtime" ON)
option(VC_BUILD_SHARED "Build shared library" OFF)

# Global compiler flags
if(MSVC)
    add_compile_options(/W4)
else()
    add_compile_options(-Wall -Wextra -pedantic)
endif()

# Add third-party libraries
add_subdirectory(third_party)

# Subdirectories
add_subdirectory(console)

if(VC_BUILD_DEBUGGER)
    add_subdirectory(debugger)
endif()

if(VC_BUILD_HOST)
    add_subdirectory(host)
endif()

if(VC_BUILD_TESTS)
    enable_testing()
endif()

