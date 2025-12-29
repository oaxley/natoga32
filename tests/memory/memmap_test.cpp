#include <catch2/catch_all.hpp>

#include "core/memory_map.h"
#include "core/cpu.h"

using namespace vc;

// check that all the segments are valid
TEST_CASE("Memory Segments validity", "[memory]") {

    SECTION("32-bit aligned", "Ensure segments are aligned") {
        REQUIRE( (ROM_BOOT_LOADER % alignof(u32)) == 0 );
        REQUIRE( (INTERRUPT_VECTOR % alignof(u32)) == 0 );
        REQUIRE( (CSR_REGISTERS % alignof(u32)) == 0 );
        REQUIRE( (THREADS_REG_BASE % alignof(u32)) == 0 );
        REQUIRE( (THREADS_META_BASE % alignof(u32)) == 0 );
        REQUIRE( (CARTRIDGE_IO % alignof(u32)) == 0 );
        REQUIRE( (HOST_S_RAM % alignof(u32)) == 0 );
        REQUIRE( (AUDIO_RAM % alignof(u32)) == 0 );
        REQUIRE( (VIDEO_RAM % alignof(u32)) == 0 );
        REQUIRE( (HEAP % alignof(u32)) == 0 );
        REQUIRE( (STACK_TOP_BASE % alignof(u32)) == 0 );
        REQUIRE( (BSS % alignof(u32)) == 0 );
        REQUIRE( (DATA % alignof(u32)) == 0 );
        REQUIRE( (CODE % alignof(u32)) == 0 );
    }

    SECTION("Threads RegSize", "Thread Save Area can fit 8 threads") {
        REQUIRE( (TREG_SIZE % NUM_THREADS) == 0 );
        REQUIRE( (TREG_SIZE / (NUM_REGISTERS * sizeof(u32))) == NUM_THREADS );
    }

    SECTION("Threads StackSize", "Minimum stack size is 64K per thread") {
        u32 min_stack_size = 64 * 1024;
        REQUIRE( STACK_SIZE == (min_stack_size * NUM_THREADS) );
    }
}

// check that the size of all the segments is less than the total ram size
TEST_CASE("Sum of segments less or equal total RAM", "[memory]") {
    u32 sum = ROM_SIZE + IV_SIZE + CSR_SIZE + TREG_SIZE;
    sum += TMETA_SIZE + CARTIO_SIZE + SRAM_SIZE + ARAM_SIZE;
    sum += VRAM_SIZE + HEAP_SIZE + STACK_SIZE + BSS_SIZE;
    sum += DATA_SIZE + CODE_SIZE;

    REQUIRE( sum <= RAM_SIZE );
}
