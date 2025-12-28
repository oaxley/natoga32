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

// program-specific headers
#include "vc/types.h"

namespace vc {

// constants
constexpr int NUM_REGISTERS = 32;       // 32 registers (x0 to x31)
constexpr int NUM_THREADS = 8;          // 8 virtual threads

// threads state
enum class ThreadState : u8 {
    Unused,
    Ready,
    Running,
    Sleeping,
    Dead
};

struct ThreadContext {
    std::array<u32, NUM_REGISTERS> regs{};
    u32 pc = 0;
    ThreadState state = ThreadState::Unused;
    u64 sleep_until = 0;
};

class Memory;       // forward declaration

class CPU
{
public:
    CPU();

    void reset();
    int step(Memory &mem, u64 current_cycle);
    void trigger_interrupt(u32 irq_num);

    // thread management
    void wake_thread(int thread_id);
    void sleep_thread(int thread_id, u64 cycles);

    // debug access
    const ThreadContext& get_thread(int id) const { return threads_[id]; }
    int current_thread() const { return current_thread_; }


private:
    std::array<ThreadContext, NUM_THREADS> threads_;
    int current_thread_ = 0;
    bool interrupt_pending_ = false;
    u32 interrupt_vector_ = 0;

    // instruction execution
    void execute_instruction(u32 instruction, Memory& mem);
    void switch_thread();

    // instruction decoder helpers
    u32 decode_opcode(u32 instr) const { return instr & 0x7F; }
    u32 decode_rd(u32 instr) const { return (instr >> 7) & 0x1F; }
    u32 decode_rs1(u32 instr) const { return (instr >> 15) & 0x1F; }
    u32 decode_rs2(u32 instr) const { return (instr >> 20) & 0x1F; }
    u32 decode_funct3(u32 instr) const { return (instr >> 12) & 0x7; }
    u32 decode_funct7(u32 instr) const { return (instr >> 25) & 0x7F; }

    // immediate decoding
    i32 decode_i_immediate(u32 instr) const;
    i32 decode_s_immediate(u32 instr) const;
    i32 decode_b_immediate(u32 instr) const;
    i32 decode_u_immediate(u32 instr) const;
    i32 decode_j_immediate(u32 instr) const;
};


} // namespace vc

