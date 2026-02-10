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
#include <span>
#include <cstdint>
#include <cassert>

// program-specific includes
#include "debug_client.h"
#include "constants.h"

#include "console/cpu_threads.h"


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



    int getActiveThreadID() const { return active_thread_id_; }
    void setActiveThreadID(int tid) { active_thread_id_ = tid; }

    // ThreadInfo accessors
    std::span<uint32_t> getThreadRegisters(int tid) {
        assert(tid >= 0 && tid < ThreadCount);
        return std::span(threads_[tid].regs);
    }

    std::span<const uint32_t> getThreadRegisters(int tid) const {
        assert(tid >= 0 && tid < ThreadCount);
        return std::span(threads_[tid].regs);
    }

    ThreadInfo& getThreadInfo(int tid) {
        assert(tid >= 0 && tid < ThreadCount);
        return threads_[tid];
    }

    const ThreadInfo& getThreadInfo(int tid) const {
        assert(tid >= 0 && tid < ThreadCount);
        return threads_[tid];
    }

private:
    //----- members
    DebugClient& client_;               //< TCP/IP debug client

    // visibility array for debugger windows
    std::array<bool, WindowCount> visibility_{};

    int active_thread_id_ = 0;
    std::array<struct ThreadInfo, ThreadCount> threads_ = {};
};
