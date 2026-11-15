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
    // thread id
    id_ = tid;

    // thread control block: reset all the private state
    tcb_ = mem.memview<ThreadControlBlock>(MMAP_THREADS_TCB_BASE + tid * MMAP_THREAD_TCB_SIZE, 1).data();
    *tcb_ = ThreadControlBlock{};
    tcb_->state = static_cast<u8>(ThreadState::Free);
    tcb_->pc = entrypoint;

    // set the canary at the bottom of the thread stack
    mem.write32(getCanaryAddr(), STACK_CANARY_VALUE);

    // set registers
    int regsize = THREADS_REGISTERS * sizeof(u32);
    registers_ = mem.memview<u32>(MMAP_THREADS_REG_BASE + tid * regsize, THREADS_REGISTERS);
    for (auto& reg : registers_) {
        reg = 0;
    }

    // SP starts at the top of the thread stack (16-byte aligned, grows downward), TP on the TLS
    registers_[SP_REG] = getSP();
    registers_[TP_REG] = getTP();
}

bool CPUThread::checkStackCanary(Memory& mem) const
{
    u32 value = mem.read32(getCanaryAddr());
    return value == STACK_CANARY_VALUE;
}

void CPUThread::sleep(u64 current_cycles, u32 waitkey_value, u32 sleep_cycles)
{
    if (waitkey_value > 0) {
        tcb_->waitkey = waitkey_value;
    }

    if (sleep_cycles > 0) {
        tcb_->sleep_until = current_cycles + static_cast<u64>(sleep_cycles);
    }

    setState(ThreadState::Sleeping);
}

void CPUThread::wake()
{
    setState(ThreadState::Ready);
    tcb_->waitkey = 0;
    tcb_->sleep_until = 0;
}

void CPUThread::end()
{
    setState(ThreadState::Dead);
}


} // namespace vc
