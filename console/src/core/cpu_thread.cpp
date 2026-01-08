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
 * @brief	Virtual Console - CPU Thread Implementation
 */

// program-specific headers
#include "cpu_thread.h"
#include "memory.h"


namespace vc
{

//----- constants
// Specific Registers
constexpr int SP_REG = 2;
constexpr int TP_REG = 4;


//----- CPUThread class definition

CPUThread::CPUThread()
{
}

void CPUThread::init(Memory& mem, int tid, u32 entrypoint)
{
    // thread id / cycles counter
    id_ = tid;
    total_cycles_ = 0;

    // stack
    sp_base_ = MMAP_STACK_BASE + ((tid + 1) * MMAP_STACK_SIZE) - 1;
    sp_ = sp_base_;

    // set the canary for the stack
    canary_addr_ = MMAP_STACK_BASE + tid * MMAP_STACK_SIZE;
    mem.write32(canary_addr_, STACK_CANARY_VALUE);

    // thread local storage
    tp_ = MMAP_THREADS_TLS_BASE + tid * MMAP_TTLS_SIZE;

    // thread initial state
    state_ = ThreadState::Free;

    // events reset
    waitkey_ = 0;
    sleep_until_ = 0;

    // set registers
    int regsize = THREADS_REGISTERS * sizeof(u32);
    registers_ = mem.memview<u32>(MMAP_THREADS_REG_BASE + tid * regsize, THREADS_REGISTERS);
    for (auto& reg : registers_) {
        reg = 0;
    }

    // program counter
    pc_ = entrypoint;

    // ensure sp and tp are mirror properly in the registers
    registers_[SP_REG] = sp_;
    registers_[TP_REG] = tp_;
}

bool CPUThread::checkStackCanary(Memory& mem) const
{
    u32 value = mem.read32(canary_addr_);
    return value == STACK_CANARY_VALUE;
}

void CPUThread::sleep(u64 current_cycles, u32 waitkey_value, u32 sleep_cycles)
{
    if (waitkey_value > 0) {
        waitkey_ = waitkey_value;
    }

    if (sleep_cycles > 0) {
        sleep_until_ = current_cycles + static_cast<u64>(sleep_cycles);
    }

    state_ = ThreadState::Sleeping;
}

void CPUThread::wake()
{
    state_ = ThreadState::Ready;
    waitkey_ = 0;
    sleep_until_ = 0;
}

void CPUThread::end()
{
    state_ = ThreadState::Dead;
}


} // namespace vc
