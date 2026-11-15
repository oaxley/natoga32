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
#include <cstddef>
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


//----- structures
// Thread Control Block: private state of a thread, stored in console memory
// at MMAP_THREADS_TCB_BASE + tid * MMAP_THREAD_TCB_SIZE
struct ThreadControlBlock
{
    u8  state;              //< ThreadState
    u8  reserved0[3];
    u32 pc;                 //< Program Counter
    u32 waitkey;            //< Waitkey value
    u32 reserved1;
    u64 sleep_until;        //< Wake-up cycle (0 = none)
    u64 total_cycles;       //< Thread total cycles
};

static_assert(sizeof(ThreadControlBlock) == MMAP_THREAD_TCB_SIZE);
static_assert(offsetof(ThreadControlBlock, state) == 0x00);
static_assert(offsetof(ThreadControlBlock, pc) == 0x04);
static_assert(offsetof(ThreadControlBlock, waitkey) == 0x08);
static_assert(offsetof(ThreadControlBlock, sleep_until) == 0x10);
static_assert(offsetof(ThreadControlBlock, total_cycles) == 0x18);


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
    u32 getPC() const { return tcb_->pc; }
    u32 getSP() const { return MMAP_STACK_BASE + (id_ + 1) * MMAP_THREAD_STACK_SIZE; }
    u32 getTP() const { return MMAP_THREADS_TLS_BASE + id_ * MMAP_THREAD_TLS_SIZE; }
    ThreadState getState() const { return static_cast<ThreadState>(tcb_->state); }
    u32 getWaitKey() const { return tcb_->waitkey; }
    u64 getSleepUntil() const { return tcb_->sleep_until; }
    u8 getID() const { return id_; }
    u64 getTotalCycles() const { return tcb_->total_cycles; }
    std::span<u32>& getRegisters() { return registers_; }

    // Setters
    void setPC(u32 pc) { tcb_->pc = pc; }
    void setState(ThreadState state) { tcb_->state = static_cast<u8>(state); }
    void incrementCycles() { tcb_->total_cycles++; }
    void incrementPC(u32 offset = 4) { tcb_->pc += offset; }

private:
    u32 getCanaryAddr() const { return MMAP_STACK_BASE + id_ * MMAP_THREAD_STACK_SIZE; }

    // the thread state lives in console memory, the class only holds views on it
    ThreadControlBlock* tcb_ = nullptr;         //< Thread Control Block (console memory)
    std::span<u32> registers_ = {};             //< RISC-V registers (console memory)
    u8 id_ = 0xFF;                              //< Thread ID
};


} // namespace vc
