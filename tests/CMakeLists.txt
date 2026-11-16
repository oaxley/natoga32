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
    libvc/test_console.cpp

    libvc/test_cpu/basic.cpp
    libvc/test_cpu/opcode-memory.cpp
    libvc/test_cpu/opcode-auipc-lui.cpp
    libvc/test_cpu/branches.cpp
    libvc/test_cpu/opcodes_19.cpp
    libvc/test_cpu/opcodes_51.cpp
    libvc/test_cpu/opcodes-csr.cpp
    libvc/test_cpu/idle.cpp
    libvc/test_cpu/opcodes-thread.cpp
    libvc/test_cpu/private-zone.cpp
    libvc/test_cpu/traps.cpp
)

# link targets
target_link_libraries(vctests PRIVATE vc Catch2::Catch2WithMain)

# include directories
target_include_directories(vctests
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/../core/console/src
        ${CMAKE_CURRENT_SOURCE_DIR}/../core/console/include
)

# Catch2 tests auto-discovery
include(CTest)
include(Catch)
catch_discover_tests(vctests)

# vchost command line: a missing cartridge must print the usage, not crash
if(TARGET vchost)
    add_test(NAME vchost_missing_file COMMAND vchost)
    set_tests_properties(vchost_missing_file PROPERTIES
        PASS_REGULAR_EXPRESSION "Usage: vchost"
        FAIL_REGULAR_EXPRESSION "terminate called")
    add_test(NAME vchost_with_file COMMAND vchost game.on32)
    set_tests_properties(vchost_with_file PROPERTIES
        PASS_REGULAR_EXPRESSION "File    : game.on32")
endif()
