cmake_minimum_required(VERSION 3.25)
project(natoga32 VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
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

# Add third-party libraries (only when the directory provides a CMakeLists.txt)
if(EXISTS ${CMAKE_CURRENT_SOURCE_DIR}/third_party/CMakeLists.txt)
    add_subdirectory(third_party)
endif()

# Subdirectories
add_subdirectory(core/console)

# boot ROM (needs Go and Python+Pillow, must come before host and tests)
add_subdirectory(rom)

if(VC_BUILD_DEBUGGER)
    # the debugger needs the imgui target, provided by third_party/
    if(TARGET imgui)
        add_subdirectory(core/debugger)
    else()
        message(WARNING "imgui target not found (third_party/ is missing it): the debugger will not be built")
    endif()
endif()

if(VC_BUILD_HOST)
    add_subdirectory(core/host)
endif()

if(VC_BUILD_TESTS)
    enable_testing()
    add_subdirectory(tests)
endif()
