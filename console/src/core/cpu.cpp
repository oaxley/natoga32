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
 * @brief	Virtual Console - CPU Implementation
 */

// standard library headers
#include <cstring>
#include <array>
#include <span>

// program-specific headers
#include "vc/exceptions.h"
#include "vc/events.h"
#include "constants.h"
#include "memory.h"
#include "decoders.h"
#include "helpers.h"
#include "cpu.h"


namespace vc
{

//----- constants
// Specific Registers
constexpr int SP_REG = 2;
constexpr int TP_REG = 4;


//----- Enums & Structs
enum class ThreadState : u8
{
    Free = 0,
    Ready = 1,
    Running = 2,
    Sleeping = 4,
    Dead = 8
};

struct ThreadContext
{
    u32 pc = 0;                                 //< Program Counter
    u32 sp = 0;                                 //< Stack Pointer
    u32 tp = 0;                                 //< Thread Pointer
    ThreadState state = ThreadState::Free;      //< Thread State
    u32 sp_base = 0;                            //< Stack base address
    u32 canary_addr = 0;                        //< Stack Canary address
    u32 waitkey = 0;                            //< Waitkey value
    u64 sleep_until = 0;                        //< Time wait value
    std::span<u32> registers = {};              //< RISC-V registers
    u8 id = 0xFF;                               //< Thread ID
    u64 total_cycles = 0;                       //< Thread Total Cycles
};


//----- CPU class definition

CPU::CPU(Memory& mem) :
    mem_{mem}, cpu_state_{CPUState::Idle}
{
    reset();
}

void CPU::reset()
{
    // reset all the threads
    for (int tid = 0; tid < THREADS_COUNT; tid++) {
        initT(tid);
    }

    // reset the global cycle counter
    total_cycles_ = 0;

    // Thread 0 is the main thread
    current_thread_ = 0;
    threads_[current_thread_].state = ThreadState::Running;
    threads_[current_thread_].pc = RESET_DEFAULT_ADDR;

    // CPU is running
    cpu_state_ = CPUState::Running;
}

int CPU::step()
{


}

// wake any threads waiting for this event
void CPU::wakeThreadOnEvent(u32 event)
{
    wakeT(event);
}


//----- private methods
// initialize a Thread
void CPU::initT(int tid, u32 entrypoint = 0)
{
    if ((tid < 0) || (tid >= THREADS_COUNT)) {
        return;
    }

    // thread id / cycles counter
    threads_[tid].id = tid;
    threads_[tid].total_cycles = 0;

    // stack
    threads_[tid].sp_base = MMAP_STACK_BASE + ((tid + 1) * MMAP_STACK_SIZE) - 1;
    threads_[tid].sp = threads_[tid].sp_base;

    // set the canary for the stack
    threads_[tid].canary_addr = MMAP_STACK_BASE + tid * MMAP_STACK_SIZE;
    mem_.write32(threads_[tid].canary_addr, STACK_CANARY_VALUE);

    // thread local storage
    threads_[tid].tp = MMAP_THREADS_TLS_BASE + tid * MMAP_TTLS_SIZE;

    // thread initial state
    threads_[tid].state = ThreadState::Free;

    // events reset
    threads_[tid].waitkey = 0;
    threads_[tid].sleep_until = 0;

    // set registers
    int regsize = THREADS_REGISTERS * sizeof(u32);
    threads_[tid].registers = mem_.memview<u32>(MMAP_THREADS_REG_BASE + tid * regsize, THREADS_REGISTERS);
    for (auto& reg : threads_[tid].registers) {
        reg = 0;
    }

    // program counter
    threads_[tid].pc = entrypoint;

    // ensure sp and tp are mirror properly in the registers
    threads_[tid].registers[SP_REG] = threads_[tid].sp;
    threads_[tid].registers[TP_REG] = threads_[tid].tp;
}

// set the Program Counter for a thread
void CPU::setTPC(int tid, u32 entrypoint)
{
    if ((tid < 0) || (tid >= THREADS_COUNT)) {
        return;
    }

    threads_[tid].pc = entrypoint;
}

// yield the current thread
void CPU::yieldT()
{
    ThreadContext& current = threads_[current_thread_];

    // step 1 : check the canary for the current thread
    if (current.state == ThreadState::Running) {
        u32 value = mem_.read32(current.canary_addr);
        if (value != STACK_CANARY_VALUE) {
            // stack overflow detected
            triggerException(CPU_THREAD_STACK_OVERFLOW_ERROR);
            return;
        }
        current.state = ThreadState::Ready;
    }

    // step 2 : update cycle-based sleepers
    for (auto& t : threads_) {
        if (t.state == ThreadState::Sleeping && t.sleep_until > 0) {
            if (total_cycles_ >= t.sleep_until) {
                t.state = ThreadState::Ready;
                t.sleep_until = 0;
                t.waitkey = 0;
            }
        }
    }

    // step 3 : find the next READY thread (round robin search)
    int start = current_thread_;
    do {
        current_thread_ = (current_thread_ + 1) % THREADS_COUNT;
        if (threads_[current_thread_].state == ThreadState::Ready) {
            threads_[current_thread_].state = ThreadState::Running;
            cpu_state_ = CPUState::Running;
            return;
        }
    } while (current_thread_ != start);

    // step 4 : no READY threads - check for deadlocks
    bool has_hardware_waiters = false;
    bool has_cycle_waiters = false;

    for (const auto& t : threads_) {
        if (t.state == ThreadState::Sleeping) {
            if (isHardwareEvent(t.waitkey)) {
                has_hardware_waiters = true;
            }
            if (t.sleep_until > 0) {
                has_cycle_waiters = true;
            }
        }
    }

    if (has_hardware_waiters || has_cycle_waiters) {
        // CPU idle - waiting for external events
        cpu_state_ = CPUState::Idle;
    } else {
        // deadlock situation
        triggerException(CPU_THREAD_DEADLOCK);
    }
}

// move the current thread to Sleeping status
void CPU::sleepT(u32 rs1, u32 rs2)
{
    ThreadContext& t = threads_[current_thread_];

    // waitkey
    if (rs1 > 0) {
        t.waitkey = rs1;
    }

    // sleep until
    if (rs2 > 0) {
        t.sleep_until = total_cycles_ + static_cast<u64>(rs2);
    }

    // switch to next thread
    if (rs1 > 0 || rs2 > 0) {
        t.state = ThreadState::Sleeping;
        yieldT();
    }
}

// wake all the threads that match the waitkey value in RS1
void CPU::wakeT(u32 rs1)
{
    for (auto& t : threads_) {
        if (t.state == ThreadState::Sleeping && t.waitkey == rs1) {
            t.state = ThreadState::Ready;
            t.waitkey = 0;
        }
    }
}

// end the current thread
void CPU::endT()
{
    // Thread 0 is immortal
    if (current_thread_ == 0) {
        triggerException(CPU_THREAD_MAIN_EXIT_ERROR);
        return;
    }

    // mark this thread as dead, and yield
    threads_[current_thread_].state = ThreadState::Dead;
    yieldT();
}

// create a new thread
void CPU::newT(u32 rd, u32 rs1)
{
    // find a free slot
    for (int tid = 0; tid < THREADS_COUNT; tid++) {
        if (threads_[tid].state == ThreadState::Free) {
            // initialize this thread with the user parameters
            initT(tid, rs1);
            threads_[tid].state = ThreadState::Ready;

            // return its id
            threads_[current_thread_].registers[rd] = tid;
            return;
        }
    }

    // no available slot found!
    triggerException(CPU_THREAD_SPAWN_ERROR);
}

void CPU::triggerException(u32 value)
{

}


} // namespace vc
