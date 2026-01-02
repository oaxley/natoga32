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
 * @brief	Virtual Console - main class header
 */
#pragma once

// standard library headers
#include <memory>
#include <tuple>

// program-specific headers
#include "vc/types.h"

namespace vc
{

// forward declarations
class CPU;
class Memory;

// console state
enum class ConsoleState
{
    Stopped,
    Running,
    Idle,
    Halted
};

// main console class
class Console
{
public:
    Console();
    virtual ~Console();

    // execution
    void reset();
    void tick();                    // execute one instruction
    void runCycles(u64 cycles);     // execute N cycles

    // state queries
    ConsoleState getState() const { return state_; }
    u64 getTotalCycles() const { return total_cycles_; }
    std::tuple<u32, u32> getLastException() const { }

    // subsystem access (for host)
    Memory* getMemoryPtr() { return memory_.get(); }
    CPU* getCPUPtr() { return cpu_.get(); }

private:
    std::unique_ptr<Memory> memory_;
    std::unique_ptr<CPU> cpu_;

    ConsoleState state_ = ConsoleState::Stopped;
    u64 total_cycles_ = 0;

    u32 exception_id_ = 0;
    u32 exception_pc_ = 0;
};


} // namespace vc

