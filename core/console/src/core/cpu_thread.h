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
 * @brief	Virtual Console - CPU Thread Header
 */
#pragma once

// standard library headers
#include <span>

// program-specific headers
#include "vc/types.h"
#include "constants.h"

namespace vc
{
// forward declarations
class Memory;


//----- enums
enum class ThreadState : u8
{
    Free = 0,
    Ready = 1,
    Running = 2,
    Sleeping = 4,
    Dead = 8
};


//----- class
class CPUThread
{
public:
    CPUThread();

    void init(Memory& mem, int tid, u32 entrypoint = 0);

    bool checkStackCanary(Memory& mem) const;
    void sleep(u64 current_cycles, u32 waitkey_value, u32 sleep_cycles);
    void wake();
    void end();

    // Getters
    u32 getPC() const { return pc_; }
    u32 getSP() const { return sp_; }
    u32 getTP() const { return tp_; }
    ThreadState getState() const { return state_; }
    u32 getWaitKey() const { return waitkey_; }
    u64 getSleepUntil() const { return sleep_until_; }
    u8 getID() const { return id_; }
    u64 getTotalCycles() const { return total_cycles_; }
    std::span<u32>& getRegisters() { return registers_; }

    // Setters
    void setPC(u32 pc) { pc_ = pc; }
    void setSP(u32 sp) { sp_ = sp; }
    void setTP(u32 tp) { tp_ = tp; }
    void setState(ThreadState state) { state_ = state; }
    void incrementCycles() { total_cycles_++; }
    void incrementPC(u32 offset = 4) { pc_ += offset; }

private:
    u32 pc_ = 0;                                //< Program Counter
    u32 sp_ = 0;                                //< Stack Pointer
    u32 tp_ = 0;                                //< Thread Pointer
    ThreadState state_ = ThreadState::Free;     //< Thread State
    u32 sp_base_ = 0;                           //< Stack base address
    u32 canary_addr_ = 0;                       //< Stack Canary address
    u32 waitkey_ = 0;                           //< Waitkey value
    u64 sleep_until_ = 0;                       //< Time wait value
    std::span<u32> registers_ = {};             //< RISC-V registers
    u8 id_ = 0xFF;                              //< Thread ID
    u64 total_cycles_ = 0;                      //< Thread Total Cycles
};


} // namespace vc
