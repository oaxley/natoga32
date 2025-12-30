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
#include "constants.h"
#include "memory.h"
#include "decoders.h"
#include "cpu.h"


namespace vc {

//----- constants

// RISC-V Opcodes
constexpr u32 OP_LOAD   = 0b000'0011;       // 0x03
constexpr u32 OP_STORE  = 0b010'0011;       // 0x23
constexpr u32 OP_BRANCH = 0b110'0011;       // 0x63

constexpr u32 OP_AUIPC  = 0b001'0111;       // 0x17
constexpr u32 OP_LUI    = 0b011'0111;       // 0x37
constexpr u32 OP_JALR   = 0b110'0111;       // 0x67
constexpr u32 OP_JAL    = 0b110'1111;       // 0x6F

constexpr u32 OP_OP_IMM = 0b001'0011;       // 0x13
constexpr u32 OP_OP     = 0b011'0011;       // 0x33

constexpr u32 OP_EXT    = 0b111'0011;       // 0x73
constexpr u32 OP_CUSTOM = 0b000'1011;       // 0x0B

// Specific Registers
constexpr int SP_REG = 2;
constexpr int TP_REG = 4;


//----- Enums
enum class ThreadState : u8 {
    Free = 0,
    Ready = 1,
    Running = 2,
    Sleeping = 4,
    Dead = 8
};

struct ThreadContext {
    u32 pc = 0;                                 //< Program Counter
    u32 sp = 0;                                 //< Stack Pointer
    u32 tp = 0;                                 //< Thread Pointer
    ThreadState state = ThreadState::Free;      //< Thread State
    u32 sp_base = 0;                            //< Stack base address
    u32 canary_addr = 0;                        //< Stack Canary address
    u32 waitkey = 0;                            //< Waitkey value
    u64 sleep_until = 0;                        //< Time wait value
    std::span<u32> registers = {};              //< RISC-V registers
};

//----- Opaque Data definition
struct CPU::OpaqueData {
    Memory& mem;
    std::array<ThreadContext, THREADS_COUNT> threads_;
    int current_thread;

    OpaqueData(Memory& memref);
    void reset();
    void init_thread(int tid);
};

CPU::OpaqueData::OpaqueData(Memory& memref) :
    mem(memref)
{
    reset();
}

void CPU::OpaqueData::reset() {
    // init threads
    for (int tid = 0; tid < THREADS_COUNT; tid++) {
        init_thread(tid);
    }

    // Thread 0 is the main thread
    current_thread = 0;
    threads_[0].state = ThreadState::Running;
    threads_[0].pc = 0;
}

// initialize a thread
void CPU::OpaqueData::init_thread(int tid) {
    if ((tid < 0) || (tid >= THREADS_COUNT))
        return;

    // stack
    threads_[tid].sp_base = MMAP_STACK_BASE + ((tid + 1) * MMAP_STACK_SIZE) - 1;
    threads_[tid].sp = threads_[tid].sp_base;

    // set the canary for the stack
    threads_[tid].canary_addr = MMAP_STACK_BASE + tid * MMAP_STACK_SIZE;
    mem.write_u32(threads_[tid].canary_addr, STACK_CANARY_VALUE);

    // thread local storage
    threads_[tid].tp = MMAP_THREADS_TLS_BASE + tid * MMAP_TTLS_SIZE;

    // thread initial state
    threads_[tid].state = ThreadState::Free;

    // events reset
    threads_[tid].waitkey = 0;
    threads_[tid].sleep_until = 0;

    // set registers
    int regsize = THREADS_REGISTERS * sizeof(u32);
    threads_[tid].registers = mem.memview<u32>(MMAP_THREADS_REG_BASE + tid * regsize, THREADS_REGISTERS);
    for (auto& reg : threads_[tid].registers) {
        reg = 0;
    }

    // program counter
    threads_[tid].pc = 0;

    // ensure sp and tp are mirror properly in the registers
    threads_[tid].registers[SP_REG] = threads_[tid].sp;
    threads_[tid].registers[TP_REG] = threads_[tid].tp;
}

//----- CPU class definition

CPU::CPU(Memory& mem) :
    data_(new (std::nothrow) OpaqueData(mem))
{ }

void CPU::reset() {
    data_.reset();
}

int CPU::step() {
    // TODO
}

} // namespace vc


