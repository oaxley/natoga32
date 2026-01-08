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
#include <span>
#include <memory>
#include <tuple>

// program-specific headers
#include "vc/types.h"
#include "constants.h"
#include "cpu_thread.h"

namespace vc
{
// forward declarations
class Memory;


//----- enums
enum class CPUState
{
    Running,        // CPU is actively executing instructions
    Idle,           // All threads sleeping on events/timers
    Halted          // Exception occured, stopped permanently
};


//----- class
class CPU
{
public:
    CPU(Memory& mem);

    void reset();
    void tick();

    void wakeThreadOnEvent(u32 event);

    // debug access
    CPUThread& getThread(int id);
    int getCurrentThreadID() const;
    CPUState getState() const;

private:
    //----- private members
    std::array<CPUThread, THREADS_COUNT> threads_;
    int current_thread_ = 0;
    u64 total_cycles_ = 0;

    Memory& mem_;
    CPUState cpu_state_;

    //----- private methods
    // threads methods
    void yieldT();
    void sleepT(u8 rs1, u8 rs2);
    void wakeT(u8 rs1);
    void endT();
    void newT(u8 rd, u8 rs1);

    // interruption / exception trigger
    void triggerTrap(bool is_interrupt, u32 cause, u32 trap_value = 0);

    // instruction executor
    void execute_instruction();

    // CSR read/write
    u32 readCSR(u16 csr);
    void writeCSR(u16 csr, u32 value);
};


} // namespace vc

