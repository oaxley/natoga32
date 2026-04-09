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
 * @brief	Debugger constants
 */
#pragma once

// standard library headers
#include <string>
#include <stdint.h>

// program-specific includes


// semantic versionning
constexpr std::string VERSION = "0.1.0";

// registers + threads
constexpr uint8_t RegisterCount = 32;
constexpr uint8_t ThreadCount = 8;


// windows ID
// maximum number of windows in the debugger
constexpr uint8_t WindowCount = 10;

constexpr uint8_t QUIT_HWND = 0x00;

constexpr uint8_t ABOUT_HWND = 0x01;
constexpr uint8_t THREADINFO_HWND = 0x02;
constexpr uint8_t MEMORYVIEW_HWND = 0x03;
