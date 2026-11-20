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
 * @brief	Virtual Console - Cartridge checker: loads a cartridge in the emulator and prints its status
 */

// standard library headers
#include <cstdio>

// program-specific headers
#include "vc/vc.h"
#include "core/cartridge.h"


int main(int argc, char* argv[])
{
    if (argc != 2) {
        std::fprintf(stderr, "usage: vc_cartcheck FILE\n");
        return 2;
    }

    VC_Console_t* console = vc_create();
    if (!console) {
        std::fprintf(stderr, "vc_cartcheck: cannot create the console\n");
        return 2;
    }
    if (!vc_initialize(console)) {
        std::fprintf(stderr, "vc_cartcheck: %s\n", vc_getLastError(console));
        vc_destroy(console);
        return 2;
    }
    if (!vc_loadCartridgeFile(console, argv[1])) {
        std::fprintf(stderr, "vc_cartcheck: %s\n", vc_getLastError(console));
        vc_destroy(console);
        return 2;
    }

    // the cartridge device checks the file at reset
    vc_reset(console);
    const uint32_t status = vc_memRead32(console, vc::CART_STATUS);
    const uint32_t count = vc_memRead32(console, vc::CART_INDX_COUNT);
    vc_destroy(console);

    if ((status & vc::CART_STATUS_VALID) != 0) {
        std::printf("cartridge valid, %u chunks indexed\n", static_cast<unsigned>(count));
        return 0;
    }
    if ((status & vc::CART_STATUS_PRESENT) != 0) {
        std::printf("cartridge invalid: error %u\n", static_cast<unsigned>((status >> 8) & 0xFF));
        return 1;
    }
    std::printf("no cartridge\n");
    return 1;
}
