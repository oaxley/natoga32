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
 * @brief	Host main file
 */
// standard library headers
#include <iostream>
#include <vector>
#include <chrono>

// program-specific includes
#include <argparse/argparse.hpp>
#include "vc/vc.h"
#include "debug_server.h"

//----- functions

std::vector<uint8_t> createTestROM()
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

    // LUI x12, 0x12345000    - Load upper immediate
    rom.push_back(0x37); rom.push_back(0x56); rom.push_back(0x34); rom.push_back(0x12);

    // ADDI x16 x0, 0x678    - Add immediate
    rom.push_back(0x13); rom.push_back(0x08); rom.push_back(0x80); rom.push_back(0x67);

    // ADD x20, x12, x16      - Add registers
    rom.push_back(0x33); rom.push_back(0x0a); rom.push_back(0x06); rom.push_back(0x01);

    // JAL x0, -12           - Jump back (infinite loop)
    rom.push_back(0x6F); rom.push_back(0xf0); rom.push_back(0x5f); rom.push_back(0xFF);

    return rom;
}



//----- main entry
int main(int argc, char* argv[])
{
    //----- Read the command line arguments

    // argument parser
    argparse::ArgumentParser program("vchost", "0.1.0");

    // debug flag
    program.add_argument("--debug")
        .flag()
        .default_value(false)
        .help("Enable debug server");

    // debug port
    program.add_argument("--debug-port")
        .scan<'i', int>()
        .default_value(2600)
        .help("Default debug listening port");

    // read the command line arguments
    try {
        program.parse_args(argc, argv);
    } catch (const std::exception& err) {
        std::cerr << err.what() << "\n";
        std::cerr << program;
        return EXIT_FAILURE;
    }

    //----- Create the console

    // create the console
    VC_Console_t* console = vc_create();
    if (!console) {
        std::cerr << "Error: failed to create the console\n";
        return EXIT_FAILURE;
    }

    // initialize the console
    if (!vc_initialize(console)) {
        std::cerr << "Error: failed to initialize the console\n";
        std::cerr << "       " << vc_getLastError(console) << "\n";
        return EXIT_FAILURE;
    }

    // load the ROM
    auto rom = createTestROM();
    if (!vc_loadRomData(console, rom.data(), rom.size())) {
        std::cerr << "Error: failed to load the ROM\n";
        std::cerr << "       " << vc_getLastError(console) << "\n";
        return EXIT_FAILURE;
    }

    //------ Lookup for a potential debug mode

    // retrieve the debug values
    auto has_debug = program.get<bool>("--debug");
    auto debug_port = program.get<int>("--debug-port");

    if (has_debug) {
        // create the debug server
        auto debug_server = DebugServer(console, debug_port);
        debug_server.start();

        // main loop waiting for debug thread to finish
        while (true) {
            std::this_thread::sleep_for(std::chrono::milliseconds(100));
        }

        return EXIT_SUCCESS;
    } else {
        // start the console without debug
        vc_run(console, 100);
    }

    // cleanup
    vc_destroy(console);

    return EXIT_SUCCESS;
}
