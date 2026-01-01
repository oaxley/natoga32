#/* -*- coding: utf-8 -*-
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
 * @brief	Virtual Console - CPU Header
 */
#pragma once

// standard library headers
#include <array>
#include <memory>

// program-specific headers
#include "vc/types.h"

namespace vc {

// enum
enum class CPUState
{
    Running,        // CPU is actively executing instructions
    Idle,           // All threads sleeping on events/timers
    Halted          // Exception occured, stopped permanently
};

// forward declarations
class Memory;
struct ThreadContext;

// CPU class definition
class CPU
{
public:
    CPU(Memory& mem);

    void reset();
    int step();

    void wakeThreadOnEvent(u32 event);

private:
    //----- private members
    std::array<ThreadContext, THREADS_COUNT> threads_;
    int current_thread_ = 0;
    u64 total_cycles_ = 0;

    Memory& mem_;
    CPUState cpu_state_;


    //----- private methods
    // threads methods
    void initT(int tid, u32 entrypoint = 0);
    void setTPC(int tid, u32 entrypoint);
    void yieldT();
    void sleepT(u32 rs1, u32 rs2);
    void wakeT(u32 rs1);
    void endT();
    void newT(u32 rd, u32 rs1);

    // exception triggering
    void triggerException(u32 value);

};


} // namespace vc

