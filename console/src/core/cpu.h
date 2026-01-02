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
#include <tuple>

// program-specific headers
#include "vc/types.h"

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

enum class ThreadState : u8
{
    Free = 0,
    Ready = 1,
    Running = 2,
    Sleeping = 4,
    Dead = 8
};


//----- structs
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


//----- class
class CPU
{
public:
    CPU(Memory& mem);

    void reset();
    int tick();

    void wakeThreadOnEvent(u32 event);

    // accessors
    CPUState getState() const;
    int getThreadId() const;
    std::tuple<u32, u32> getLastException() const;

private:
    //----- private members
    std::array<ThreadContext, THREADS_COUNT> threads_;
    int current_thread_ = 0;
    u64 total_cycles_ = 0;

    Memory& mem_;
    CPUState cpu_state_;

    // error management
    u32 exception_id_ = 0;
    u32 exception_pc_ = 0;

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

    // instruction executor
    void execute_instruction();
};


} // namespace vc

