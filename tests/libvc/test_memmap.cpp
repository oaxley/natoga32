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
        REQUIRE( (MMAP_CART_CONTROL % alignof(u32)) == 0 );
        REQUIRE( (MMAP_CART_WINDOW % alignof(u32)) == 0 );
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

// check that the segments match the reference memory map (workspace/memory_map.md)
TEST_CASE("Memory Segments match the memory map", "[memory]") {

    SECTION("Stack segment: 0x0078_0000 - 0x007F_FFFF") {
        REQUIRE( MMAP_STACK_BASE == 0x0078'0000 );
        REQUIRE( MMAP_STACK_BASE + MMAP_STACK_SIZE == 0x0080'0000 );
    }

    SECTION("Heap segment: 0x0080_0000 - 0x00FF_FFFF") {
        REQUIRE( MMAP_HEAP == 0x0080'0000 );
        REQUIRE( MMAP_HEAP + MMAP_HEAP_SIZE == MMAP_BSS );
    }

    SECTION("Per-thread stack is 64KiB") {
        REQUIRE( MMAP_THREAD_STACK_SIZE == 64 * 1024 );
        REQUIRE( MMAP_THREAD_STACK_SIZE * THREADS_COUNT == MMAP_STACK_SIZE );
    }

    SECTION("Per-thread TLS is 256B and fits in the Threads Metadata block") {
        REQUIRE( MMAP_THREAD_TLS_SIZE == 256 );
        REQUIRE( MMAP_THREAD_TLS_SIZE * THREADS_COUNT == MMAP_TTLS_SIZE );
    }

    SECTION("Thread Control Blocks fit between Registers Save and TLS") {
        REQUIRE( MMAP_THREAD_TCB_SIZE == 32 );
        REQUIRE( MMAP_THREADS_TCB_BASE == MMAP_THREADS_REG_BASE + MMAP_TREG_SIZE );
        REQUIRE( MMAP_THREADS_TCB_BASE + THREADS_COUNT * MMAP_THREAD_TCB_SIZE <= MMAP_THREADS_TLS_BASE );
    }

    SECTION("Private thread zone covers Registers Save up to the TLS") {
        REQUIRE( MMAP_THREADS_PRIVATE_BASE == MMAP_THREADS_REG_BASE );
        REQUIRE( MMAP_THREADS_PRIVATE_BASE + MMAP_THREADS_PRIVATE_SIZE == MMAP_THREADS_TLS_BASE );
    }

    SECTION("Cartridge window fills the gap before Video RAM") {
        REQUIRE( MMAP_AUDIO_RAM + MMAP_ARAM_SIZE == MMAP_CART_WINDOW );
        REQUIRE( MMAP_CART_WINDOW + MMAP_CART_WINDOW_SIZE == MMAP_VIDEO_RAM );
        REQUIRE( MMAP_CART_CONTROL + MMAP_CART_CONTROL_SIZE == MMAP_HOST_S_RAM );
    }
}

// check that the size of all the segments is less than the total ram size
TEST_CASE("Sum of segments less or equal total RAM", "[memory]") {
    std::vector<u32> segments = {
        MMAP_ROM_SIZE, MMAP_IV_SIZE, MMAP_CSR_SIZE,
        MMAP_TREG_SIZE, MMAP_TTLS_SIZE, MMAP_CART_CONTROL_SIZE, MMAP_CART_WINDOW_SIZE,
        MMAP_SRAM_SIZE, MMAP_ARAM_SIZE, MMAP_VRAM_SIZE,
        MMAP_HEAP_SIZE, MMAP_STACK_SIZE, MMAP_BSS_SIZE,
        MMAP_RODATA_SIZE, MMAP_DATA_SIZE, MMAP_CODE_SIZE,
        MMAP_UNUSED_SIZE
    };

    u32 sum = 0;
    for (auto value : segments) {
        sum += value;
    }

    REQUIRE( sum == MMAP_RAM_SIZE );
}

// check that .bss, .rodata, .data and .text follow each other up to the end of RAM
TEST_CASE("Program sections are contiguous", "[memory]") {
    REQUIRE( MMAP_BSS + MMAP_BSS_SIZE == MMAP_RODATA );
    REQUIRE( MMAP_RODATA + MMAP_RODATA_SIZE == MMAP_DATA );
    REQUIRE( MMAP_DATA + MMAP_DATA_SIZE == MMAP_CODE );
    REQUIRE( MMAP_CODE + MMAP_CODE_SIZE == MMAP_RAM_SIZE );
}
