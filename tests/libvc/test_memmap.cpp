#include <catch2/catch_all.hpp>

#include "core/constants.h"
#include "core/cpu.h"

using namespace vc;

// check that all the segments are valid
TEST_CASE("Memory Segments validity", "[memory]") {

    SECTION("32-bit aligned", "Ensure segments are aligned") {
        REQUIRE( (MMAP_ROM_BOOT_LOADER % alignof(u32)) == 0 );
        REQUIRE( (MMAP_INTERRUPT_VECTOR % alignof(u32)) == 0 );
        REQUIRE( (MMAP_CSR_REGISTERS % alignof(u32)) == 0 );
        REQUIRE( (MMAP_THREADS_REG_BASE % alignof(u32)) == 0 );
        REQUIRE( (MMAP_THREADS_TLS_BASE % alignof(u32)) == 0 );
        REQUIRE( (MMAP_CARTRIDGE_IO % alignof(u32)) == 0 );
        REQUIRE( (MMAP_HOST_S_RAM % alignof(u32)) == 0 );
        REQUIRE( (MMAP_AUDIO_RAM % alignof(u32)) == 0 );
        REQUIRE( (MMAP_VIDEO_RAM % alignof(u32)) == 0 );
        REQUIRE( (MMAP_HEAP % alignof(u32)) == 0 );
        REQUIRE( (MMAP_STACK_BASE % alignof(u32)) == 0 );
        REQUIRE( (MMAP_BSS % alignof(u32)) == 0 );
        REQUIRE( (MMAP_DATA % alignof(u32)) == 0 );
        REQUIRE( (MMAP_CODE % alignof(u32)) == 0 );
    }

    SECTION("Threads RegSize", "Thread Register Area can fit 8 threads") {
        REQUIRE( (MMAP_TREG_SIZE % THREADS_COUNT) == 0 );
        REQUIRE( (MMAP_TREG_SIZE / (THREADS_REGISTERS * sizeof(u32))) == THREADS_COUNT );
    }

    SECTION("Threads StackSize", "Minimum stack size is 64K per thread") {
        u32 min_stack_size = 64 * 1024;
        REQUIRE( MMAP_STACK_SIZE == (min_stack_size * THREADS_COUNT) );
    }
}

// check that the size of all the segments is less than the total ram size
TEST_CASE("Sum of segments less or equal total RAM", "[memory]") {
    std::vector<u32> segments = {
        MMAP_ROM_SIZE, MMAP_IV_SIZE, MMAP_CSR_SIZE,
        MMAP_TREG_SIZE, MMAP_TTLS_SIZE, MMAP_CARTIO_SIZE,
        MMAP_SRAM_SIZE, MMAP_ARAM_SIZE, MMAP_VRAM_SIZE,
        MMAP_HEAP_SIZE, MMAP_STACK_SIZE, MMAP_BSS_SIZE,
        MMAP_DATA_SIZE, MMAP_CODE_SIZE, MMAP_UNUSED_SIZE
    };

    u32 sum = 0;
    for (auto value : segments) {
        sum += value;
    }

    REQUIRE( sum == MMAP_RAM_SIZE );
}
