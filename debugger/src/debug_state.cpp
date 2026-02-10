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
{ }

/*virtual*/ DebugState::~DebugState()
{ }


