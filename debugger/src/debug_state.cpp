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

// program-specific includes
#include "debug_state.h"


// constructor
DebugState::DebugState(DebugClient& client) :
    client_{client}
{
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


