/* -*- coding: utf-8 -*-
 * vim: filetype=cpp
 *
 * This source file is subject to the MIT License
 * that is bundled with this package in the file LICENSE.txt.
 * It is also available through the Internet at this address:
 * https://opensource.org/license/mit
 *
 * @author	Sebastien LEGRAND
 * @license	MIT License
 *
 * @brief	Virtual Console - Basic Host Program
 */

#include <iostream>
#include <iomanip>
#include <cstdint>
#include <vector>

// Use the C API
#include "vc/vc.h"


//----- Helper functions

void print_separator()
{
    std::cout << "----------------------------------------\n";
}

void print_console_state(vc_console_t* console)
{
    std::cout << "Console State: ";
    auto state = vc_console_get_state(console);
    switch (state) {
        case VC_STATE_UNINITIALIZED:
            std::cout << "UNINITIALIZED\n";
            break;
        case VC_STATE_READY:
            std::cout << "READY\n";
            break;
        case VC_STATE_RUNNING:
            std::cout << "RUNNING\n";
            break;
        case VC_STATE_PAUSED:
            std::cout << "PAUSED\n";
            break;
        case VC_STATE_ERROR:
            std::cout << "ERROR\n";
            break;
    }

    std::cout << "CPU State: ";
    auto cpu_state = vc_console_get_cpu_state(console);
    switch (cpu_state) {
        case VC_CPU_RUNNING:
            std::cout << "RUNNING\n";
            break;
        case VC_CPU_IDLE:
            std::cout << "IDLE\n";
            break;
        case VC_CPU_HALTED:
            std::cout << "HALTED\n";
            break;
    }

    std::cout << "Current Thread: " << vc_console_get_current_thread_id(console) << "\n";
    std::cout << "Total Cycles: " << vc_console_get_total_cycles(console) << "\n";
}

void print_thread_info(vc_console_t* console, int thread_id)
{
    std::cout << "Thread " << thread_id << ": ";

    auto state = vc_console_get_thread_state(console, thread_id);
    switch (state) {
        case VC_THREAD_FREE:
            std::cout << "FREE";
            break;
        case VC_THREAD_READY:
            std::cout << "READY";
            break;
        case VC_THREAD_RUNNING:
            std::cout << "RUNNING";
            break;
        case VC_THREAD_SLEEPING:
            std::cout << "SLEEPING";
            break;
        case VC_THREAD_DEAD:
            std::cout << "DEAD";
            break;
    }

    if (state != VC_THREAD_FREE) {
        uint32_t pc = vc_console_get_thread_pc(console, thread_id);
        std::cout << " | PC: 0x" << std::hex << std::setfill('0') << std::setw(8) << pc << std::dec;
    }

    std::cout << "\n";
}

void print_all_threads(vc_console_t* console)
{
    std::cout << "\nThread Status:\n";
    for (int i = 0; i < 8; i++) {
        print_thread_info(console, i);
    }
}

void print_registers(vc_console_t* console, int thread_id)
{
    std::cout << "\nThread " << thread_id << " Registers:\n";
    for (int i = 0; i < 32; i++) {
        if (i % 4 == 0 && i > 0) {
            std::cout << "\n";
        }
        uint32_t value = vc_console_get_thread_register(console, thread_id, i);
        std::cout << "x" << std::setw(2) << i << ": 0x"
                  << std::hex << std::setfill('0') << std::setw(8) << value
                  << std::dec << "  ";
    }
    std::cout << "\n";
}


//----- Test ROM creation

std::vector<uint8_t> create_test_rom()
{
    std::vector<uint8_t> rom;

    // CPU starts at address 0x100, so we need to pad the ROM
    // Fill first 0x100 bytes with zeros (or could be boot loader code)
    for (int i = 0; i < 0x100; i++) {
        rom.push_back(0x00);
    }

    // Now add our test program at offset 0x100
    // Simple test program that will:
    // 1. Set some registers
    // 2. Add values
    // 3. Loop forever

    // LUI x1, 0x12345000    - Load upper immediate
    rom.push_back(0x37); rom.push_back(0x51); rom.push_back(0x34); rom.push_back(0x12);

    // ADDI x2, x0, 100      - Add immediate
    rom.push_back(0x13); rom.push_back(0x01); rom.push_back(0x40); rom.push_back(0x06);

    // ADD x3, x1, x2        - Add registers
    rom.push_back(0xB3); rom.push_back(0x01); rom.push_back(0x20); rom.push_back(0x00);

    // JAL x0, -12           - Jump back (infinite loop)
    rom.push_back(0x6F); rom.push_back(0x00); rom.push_back(0x40); rom.push_back(0xFF);

    return rom;
}


//----- Main program

int main(int argc, char* argv[])
{
    std::cout << "Virtual Console - Host Integration Test\n";
    print_separator();

    // Step 1: Create console
    std::cout << "\n1. Creating console...\n";
    vc_console_t* console = vc_console_create();
    if (!console) {
        std::cerr << "ERROR: Failed to create console\n";
        return 1;
    }
    std::cout << "   ✓ Console created\n";

    // Step 2: Initialize console
    std::cout << "\n2. Initializing console...\n";
    if (!vc_console_initialize(console)) {
        std::cerr << "ERROR: Failed to initialize console\n";
        std::cerr << "       " << vc_console_get_last_error(console) << "\n";
        vc_console_destroy(console);
        return 1;
    }
    std::cout << "   ✓ Console initialized\n";
    print_console_state(console);

    // Step 3: Create and load test ROM
    std::cout << "\n3. Loading test ROM...\n";
    auto rom = create_test_rom();
    if (!vc_console_load_rom_data(console, rom.data(), rom.size())) {
        std::cerr << "ERROR: Failed to load ROM\n";
        std::cerr << "       " << vc_console_get_last_error(console) << "\n";
        vc_console_destroy(console);
        return 1;
    }
    std::cout << "   ✓ ROM loaded (" << rom.size() << " bytes)\n";

    // Step 4: Reset console to start execution
    std::cout << "\n4. Resetting console...\n";
    vc_console_reset(console);
    std::cout << "   ✓ Console reset\n";
    print_console_state(console);
    print_all_threads(console);

    // Step 5: Change state to running
    std::cout << "\n5. Starting execution...\n";
    vc_console_resume(console);
    std::cout << "   ✓ Console resumed\n";

    // Step 6: Run a few cycles
    std::cout << "\n6. Executing 10 cycles...\n";
    vc_console_run(console, 10);
    print_console_state(console);
    print_all_threads(console);
    print_registers(console, 0);

    // Step 7: Memory read test
    std::cout << "\n7. Memory read test...\n";
    std::cout << "   Reading first 16 bytes at 0x00000000 (should be zeros):\n   ";
    for (int i = 0; i < 16; i++) {
        uint8_t byte = vc_console_mem_read8(console, 0x00000000 + i);
        std::cout << std::hex << std::setfill('0') << std::setw(2) << (int)byte << " ";
    }
    std::cout << "\n";

    std::cout << "   Reading 16 bytes at 0x00000100 (program code):\n   ";
    for (int i = 0; i < 16; i++) {
        uint8_t byte = vc_console_mem_read8(console, 0x00000100 + i);
        std::cout << std::hex << std::setfill('0') << std::setw(2) << (int)byte << " ";
    }
    std::cout << std::dec << "\n";

    // Step 8: Pause and resume test
    std::cout << "\n8. Pause/Resume test...\n";
    vc_console_pause(console);
    std::cout << "   Paused: " << (vc_console_is_paused(console) ? "YES" : "NO") << "\n";
    vc_console_resume(console);
    std::cout << "   Running: " << (vc_console_is_running(console) ? "YES" : "NO") << "\n";

    // Step 9: Run more cycles
    std::cout << "\n9. Running 1000 more cycles...\n";
    vc_console_run(console, 1000);
    print_console_state(console);
    print_registers(console, 0);

    // Step 10: Cleanup
    std::cout << "\n10. Cleaning up...\n";
    vc_console_destroy(console);
    std::cout << "    ✓ Console destroyed\n";

    print_separator();
    std::cout << "Test completed successfully!\n";

    return 0;
}
