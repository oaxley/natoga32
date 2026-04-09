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
 * @brief	Debugger State tracker - implementation
 */

// standard library headers
#include <random>

// program-specific includes
#include "debug_state.h"


// constructor
DebugState::DebugState(DebugClient& client) :
    client_{client}, ram_(ram_size)
{
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 255);

    for (size_t x = 0; x < ram_size; x++) {
        ram_[x] = dis(gen);
    }

    for (int i = 0; i < ThreadCount; i++)
    {
        threads_[i].regs[i+1] = 0xDEADC0DE;
    }

    threads_[5].state = ThreadState::Running;
    threads_[5].cycles = 1 << 13;
    threads_[5].waitkey = 0xCAFEBABE;
    threads_[5].sleep_until = 1 << 18;
}

/*virtual*/ DebugState::~DebugState()
{ }


