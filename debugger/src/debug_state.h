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
 * @brief	Debugger State tracker - interface
 */
#pragma once

// standard library headers
#include <string>
#include <array>
#include <stdint.h>

// program-specific includes
#include "debug_client.h"

// maximum number of windows in the debugger
constexpr uint8_t WindowCount = 10;

// class definition
class DebugState
{
public:
    DebugState(DebugClient& client);
    virtual ~DebugState();

    // visibility accessors for windows
    void set(uint8_t id, bool value) { visibility_[id] = value; }
    bool get(uint8_t id) const { return visibility_[id]; }
    bool isVisible(uint8_t id) const { return get(id); }
    void show(uint8_t id) { set(id, true); }
    void hide(uint8_t id) { set(id, false); }
    void toggleVisibility(uint8_t id) { visibility_[id] = !visibility_[id]; }

private:
    //----- members
    DebugClient& client_;               //< TCP/IP debug client

    // visibility array for debugger windows
    std::array<bool, WindowCount> visibility_{};
};
