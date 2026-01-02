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
 * @brief	Virtual Console - Exceptions values
 */

#pragma once

// program-specific includes
#include "vc/types.h"


namespace vc
{

// CPU Exceptions
constexpr u32 CPU_THREAD_MAIN_EXIT_ERROR            = 0x00010101;
constexpr u32 CPU_THREAD_SPAWN_ERROR                = 0x00010102;
constexpr u32 CPU_THREAD_STACK_OVERFLOW_ERROR       = 0x00010103;
constexpr u32 CPU_THREAD_DEADLOCK                   = 0x00010104;

constexpr u32 CPU_MAIN_ILLEGAL_INSTRUCTION          = 0x00010201;

} // namespace vc
