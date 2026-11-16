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
 * @brief	ConsoleRunner configuration object
 */
#pragma once

// standard library headers
#include <string>

// program-specific includes
#include <vc/vc.h>

//----- HostConfig

struct HostConfig
{
    bool debug = false;
    int debug_port = 2600;
    std::string rombios;
    std::string cartridge;
    bool superscalar = false;   // host-only: window scaled x2
    int frames = 0;             // 0 = windowed, unlimited; N > 0 = headless, N frames
};
