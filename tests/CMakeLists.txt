cmake_minimum_required(VERSION 3.25)

# CMake fetch Catch2 library
include(FetchContent)
FetchContent_Declare(
    Catch2
    GIT_REPOSITORY https://github.com/catchorg/Catch2.git
    GIT_TAG        v3.12.0
)
FetchContent_MakeAvailable(Catch2)

# create the Virtual Console tests binary
add_executable(vctests
    libvc/test_memmap.cpp
    libvc/test_memory.cpp
    libvc/test_decoder.cpp

    libvc/test_cpu/basic.cpp
    libvc/test_cpu/opcode-memory.cpp
    libvc/test_cpu/opcode-auipc-lui.cpp
    libvc/test_cpu/branches.cpp
)

# link targets
target_link_libraries(vctests PRIVATE vc Catch2::Catch2WithMain)

# include directories
target_include_directories(vctests
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/../console/src
        ${CMAKE_CURRENT_SOURCE_DIR}/../console/include
)

# Catch2 tests auto-discovery
include(CTest)
include(Catch)
catch_discover_tests(vctests)
