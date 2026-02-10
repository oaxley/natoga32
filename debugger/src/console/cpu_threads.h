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
 * @brief	Debugger thread info structure
 */
#pragma once

// standard library headers
#include <cstdint>
#include <array>

// program-specific includes
#include "../constants.h"


enum class ThreadState : uint8_t
{
    Free = 0,
    Ready = 1,
    Running = 2,
    Sleeping = 4,
    Dead = 8
};

struct ThreadInfo
{
    uint32_t pc;
    ThreadState state = ThreadState::Free;

    uint32_t waitkey = {0};
    uint64_t sleep_until = {0};
    uint64_t cycles = {0};

    std::array<uint32_t, RegisterCount> regs = {0};
};
