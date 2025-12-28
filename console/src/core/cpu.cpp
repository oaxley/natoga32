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

// program-specific headers
#include "cpu.h"
#include "memory.h"

namespace vc {

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

// extended opcodes (csr, ecal, mret, ...)
constexpr u32 OP_EXT    = 0b111'0011;       // 0x73

// custom opcodes for thread management
constexpr u32 OP_CUSTOM = 0b000'1011;       // 0x0B


CPU::CPU() {
    reset();
}

void CPU::reset() {
    for (auto& thread : threads_) {
        thread.regs.fill(0);
        thread.pc = 0;
        thread.state = ThreadState::Unused;
        thread.sleep_until = 0;
    }

    // Thread 0 is the main thread
    current_thread_ = 0;
    threads_[0].state = ThreadState::Running;
    interrupt_pending_ = false;
}

int CPU::step(Memory& mem, u64 current_cycle) {
    // check for sleeping threads that should wake up
    for (size_t i = 0; i < threads_.size(); i++) {
        if (threads_[i].state == ThreadState::Sleeping && threads_[i].sleep_until <= current_cycle) {
            threads_[i].state = ThreadState::Running;
        }
    }

    // get the current thread
    auto& thread = threads_[current_thread_];

    // if the current thread is not running, switch to the next one
    if (thread.state != ThreadState::Running) {
        switch_thread();
        thread = threads_[current_thread_];

        // no thread running at the moment
        if (thread.state != ThreadState::Running) {
            return 1;       // return only 1 cycle burned
        }
    }

    // fetch a new instruction
    u32 instruction = mem.read_u32(thread.pc);

    // decode / execute
    execute_instruction(instruction, mem);

    // ensure x0 remains at 0
    thread.regs[0] = 0;

    // 1 instruction = 1 cycle
    return 1;
}

void CPU::execute_instruction(u32 instr, Memory& mem) {

}

void CPU::switch_thread() {
    // round robin scheduling
    int start_thread = current_thread_;

    do {
        // next thread
        current_thread_ = (current_thread_ + 1) % NUM_THREADS;

        // is this thread able to run?
        if (threads_[current_thread_].state == ThreadState::Running) {
            return;
        }

        // did we wrap around?
        if (current_thread_ == start_thread) {
            break;
        }
    } while (true);
}

void CPU::wake_thread(int thread_id) {
    if (thread_id >= 0 && thread_id < NUM_THREADS) {
        threads_[thread_id].state = ThreadState::Running;
    }
}

void CPU::sleep_thread(int thread_id, u64 cycles) {
    if (thread_id >= 0 && thread_id < NUM_THREADS) {
        threads_[thread_id].state = ThreadState::Sleeping;
        threads_[thread_id].sleep_until = cycles;
    }
}

i32 CPU::decode_i_immediate(u32 instr) const {
    return static_cast<i32>((instr & 0xFFF0'0000) >> 20);
}

i32 CPU::decode_s_immediate(u32 instr) const {
    u32 imm = ((instr >> 7) & 0x1F) | ((instr >> 25) << 5);
    return static_cast<i32>((imm << 20) >> 20);     // sign extended
}

i32 CPU::decode_b_immediate(u32 instr) const {
    u32 imm = ((instr >>  8) & 0x0F) <<  1 |
              ((instr >> 25) & 0x3F) <<  5 |
              ((instr >>  7) & 0x01) << 11 |
              ((instr >> 31) & 0x01) << 12;
    return static_cast<i32>((imm << 19) >> 19);     // sign extended
}

i32 CPU::decode_u_immediate(u32 instr) const {
    return static_cast<i32>(instr & 0xFFFFF000);    // no shift, value is imm[32:12]
}

i32 CPU::decode_j_immediate(u32 instr) const {
    u32 imm = ((instr >> 21) & 0x3FF) <<  1 |
              ((instr >> 20) &   0x1) << 11 |
              ((instr >> 12) &  0xFF) << 12 |
              ((instr >> 31) &   0x1) << 20;
    return static_cast<i32>((imm << 11) >> 11);     // sign extendee
}

} // namespace vc


