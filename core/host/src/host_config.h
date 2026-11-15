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
    bool debug;
    int debug_port;
    std::string rombios;
    std::string cartridge;
};
