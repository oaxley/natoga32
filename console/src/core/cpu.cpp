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

int CPU::tick()
{
    if (cpu_state_ != CPUState::Running) {
        return;
    }

    // execute one instruction
    execute_instruction();

    // increment the global instructions/thread counters
    total_cycles_++;
    threads_[current_thread_].total_cycles++;
}

// wake any threads waiting for this event
void CPU::wakeThreadOnEvent(u32 event)
{
    wakeT(event);
}

// retrieve the current CPU state
CPUState CPU::getState() const
{
    return cpu_state_;
}

// retrieve the current thread ID
int CPU::getThreadId() const
{
    return current_thread_;
}

std::tuple<u32, u32> CPU::getLastException() const
{
    return std::make_tuple(exception_id_, exception_pc_);
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
    // CPU is halted
    cpu_state_ = CPUState::Halted;

    // record the exception and the current PC
    exception_id_ = value;
    exception_pc_ = threads_[current_thread_].pc;
}

void CPU::execute_instruction()
{
    // thread accessor
    auto& thread  = threads_[current_thread_];

    // fetch the instruction
    u32 instruction = mem_.read32(thread.pc);
    u32 opcode = decoder::opcode(instruction);

    switch(opcode)
    {
        case OP_LUI:        // load upper immediate
        {
            u32 rd = decoder::rd(instruction);
            i32 imm = decoder::immTypeU(instruction);
            thread.registers[rd] = imm;
            thread.pc += 4;
            break;
        }

        case OP_AUIPC:      // add upper immediate to PC
        {
            u32 rd = decoder::rd(instruction);
            i32 imm = decoder::immTypeU(instruction);
            thread.registers[rd] = thread.pc + imm;
            thread.pc += 4;
            break;
        }

        case OP_JAL:        // jump and link
        {
            u32 rd = decoder::rd(instruction);
            i32 imm = decoder::immTypeJ(instruction);
            thread.registers[rd] = thread.pc + 4;
            thread.pc += imm;
            break;
        }

        case OP_JALR:       // jump and link register
        {
            u32 rd = decoder::rd(instruction);
            u32 rs1 = decoder::rs1(instruction);
            i32 imm = decoder::immTypeI(instruction);
            u32 target = (thread.registers[rs1] + imm) & ~1;
            thread.registers[rd] = thread.pc + 4;
            thread.pc = target;
            break;
        }

        case OP_BRANCH:     // branches
        {
            u32 rs1 = decoder::rs1(instruction);
            u32 rs2 = decoder::rs2(instruction);
            i32 imm = decoder::immTypeB(instruction);
            u32 funct3 = decoder::funct3(instruction);

            bool take_branch = false;
            i32 val1 = thread.registers[rs1];
            i32 val2 = thread.registers[rs2];

            switch(funct3)
            {
                case 0b000: // BEQ
                    take_branch = (val1 == val2);
                    break;
                case 0b001: // BNE
                    take_branch = (val1 != val2);
                    break;
                case 0b100: // BLT
                    take_branch = (val1 < val2);
                    break;
                case 0b101: // BGE
                    take_branch = (val1 >= val2);
                    break;
                case 0b110: // BLTU
                    take_branch = ((u32)val1 < (u32)val2);
                    break;
                case 0b111: // BGEU
                    take_branch = ((u32)val1 >= (u32)val2);
                    break;
            }

            if (take_branch) {
                thread.pc += imm;
            } else {
                thread.pc += 4;
            }
            break;
        }

        case OP_LOAD:       // load
        {
            u32 rd = decoder::rd(instruction);
            u32 rs1 = decoder::rs1(instruction);
            i32 imm = decoder::immTypeI(instruction);
            u32 funct3 = decoder::funct3(instruction);

            u32 addr = thread.registers[rs1] + imm;

            switch(funct3)
            {
                case 0b000:     // LB
                    thread.registers[rd] = static_cast<i32>(static_cast<i8>(mem_.read8(addr)));
                    break;
                case 0b001:     // LH
                    thread.registers[rd] = static_cast<i32>(static_cast<i16>(mem_.read16(addr)));
                    break;
                case 0b010:     // LW
                    thread.registers[rd] = mem_.read32(addr);
                    break;
                case 0b100:     // LBU
                    thread.registers[rd] = mem_.read8(addr);
                    break;
                case 0b101:     // LWU
                    thread.registers[rd] = mem_.read16(addr);
                    break;
            }
            thread.pc += 4;
            break;
        }

        case OP_STORE:      // store
        {
            u32 rs1 = decoder::rs1(instruction);
            u32 rs2 = decoder::rs2(instruction);
            i32 imm = decoder::immTypeS(instruction);
            u32 funct3 = decoder::funct3(instruction);

            u32 addr = thread.registers[rs1] + imm;

            switch(funct3)
            {
                case 0b000:     // SB
                    mem_.write8(addr, thread.registers[rs2]);
                    break;
                case 0b001:     // SH
                    mem_.write16(addr, thread.registers[rs2]);
                    break;
                case 0b010:     // SW
                    mem_.write32(addr, thread.registers[rs2]);
                    break;
            }
            thread.pc += 4;
            break;
        }

        case OP_OP_IMM:     // Integer operations with Immediate
        {
            u32 rd = decoder::rd(instruction);
            u32 rs1 = decoder::rs1(instruction);
            i32 imm = decoder::immTypeI(instruction);

            u32 funct3 = decoder::funct3(instruction);
            u32 funct7 = decoder::funct7(instruction);
            u32 shamt = decoder::rs2(instruction);

            switch(funct3)
            {
                case 0b000:     // ADDI
                    thread.registers[rd] = thread.registers[rs1] + imm;
                    break;
                case 0b001:     // SLLI or Zbb instruction
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLLI
                            thread.registers[rd] = thread.registers[rs1] << shamt;
                            break;
                        case 0b001'0100:    // BSETI
                            thread.registers[rd] = thread.registers[rs1] | (1u << shamt);
                            break;
                        case 0b010'0100:    // BCLRI
                            thread.registers[rd] = thread.registers[rs1] & ~(1u << shamt);
                            break;
                        case 0b011'0000:    // SEXT.[H,B] / CLZ / CTZ / CPOP
                        {
                            switch (shamt)
                            {
                                case 0b00000:       // CLZ
                                    thread.registers[rd] = std::countl_zero(thread.registers[rs1]);
                                    break;
                                case 0b00001:       // CTZ
                                    thread.registers[rd] = std::countr_zero(thread.registers[rs1]);
                                    break;
                                case 0b00010:       // CPOP
                                    thread.registers[rd] = std::popcount(thread.registers[rs1]);
                                    break;
                                case 0b00100:       // SEXT.B
                                    thread.registers[rd] = static_cast<u32>(static_cast<i8>(thread.registers[rs1]));
                                    break;
                                case 0b00101:       // SEXT.H
                                    thread.registers[rd] = static_cast<u32>(static_cast<i16>(thread.registers[rs1]));
                                    break;
                            }
                            break;
                        }
                        case 0b011'0100:    // BINVI
                                thread.registers[rd] = thread.registers[rs1] ^ (1u << shamt);
                            break;

                    }
                    break;
                }
                case 0b010:     // SLTI
                    thread.registers[rd] = ((i32)thread.registers[rs1] < imm) ? 1 : 0;
                    break;
                case 0b011:     // SLTIU
                    thread.registers[rd] = (thread.registers[rs1] < (u32)imm) ? 1 : 0;
                    break;
                case 0b100:     // XORI
                    thread.registers[rd] = thread.registers[rs1] ^ imm;
                    break;
                case 0b101:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SRLI
                            thread.registers[rd] = thread.registers[rs1] >> shamt;
                            break;
                        case 0b010'0000:    // SRAI
                            thread.registers[rd] = static_cast<i32>(thread.registers[rs1]) >> shamt;
                            break;
                        case 0b010'0100:    // BEXTI
                            thread.registers[rd] = (thread.registers[rs1] >> shamt) & 1u;
                            break;
                        case 0b011'0000:    // RORI
                            thread.registers[rd] = std::rotr(thread.registers[rs1], shamt);
                            break;
                        case 0b011'0100:    // REV8
                        {
                            u32 value = thread.registers[rs1];
                            thread.registers[rd] = ((value >> 24) & 0xFF)
                                                 | ((value >> 8 ) & 0xFF00)
                                                 | ((value << 8) & 0xFF0000)
                                                 | ((value << 24) & 0xFF000000);
                            break;
                        }
                    }
                    break;
                }
                case 0b110:     // ORI
                    thread.registers[rd] = thread.registers[rs1] | imm;
                    break;
                case 0b111:     // ANDI
                    thread.registers[rd] = thread.registers[rs1] & imm;
                    break;
            }
            thread.pc += 4;
            break;
        }

        case OP_OP:         // register-register operations
        {
            u32 rd = decoder::rd(instruction);
            u32 rs1 = decoder::rs1(instruction);
            u32 rs2 = decoder::rs2(instruction);

            u32 funct3 = decoder::funct3(instruction);
            u32 funct7 = decoder::funct7(instruction);

            switch(funct3)
            {
                case 0b000:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // ADD
                            thread.registers[rd] = thread.registers[rs1] + thread.registers[rs2];
                            break;
                        case 0b010'0000:    // SUB
                            thread.registers[rd] = thread.registers[rs1] - thread.registers[rs2];
                            break;
                    }
                    break;
                }
                case 0b001:     // SLL
                    thread.registers[rd] = thread.registers[rs1] << (thread.registers[rs2] & 0x1F);
                    break;
                case 0b010:     // SLT
                    thread.registers[rd] = ((i32)thread.registers[rs1] < (i32)thread.registers[rs2]) ? 1 : 0;
                    break;
                case 0b011:     // SLTU
                    thread.registers[rd] = (thread.registers[rs1] < thread.registers[rs2]) ? 1 : 0;
                    break;
                case 0b100:     // XOR, MIN, ZEXT.H
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // XOR
                            thread.registers[rd] = thread.registers[rs1] ^ thread.registers[rs2];
                            break;
                        case 0b000'0100:    // ZEXT.H
                            thread.registers[rd] = thread.registers[rs1] & 0xFFFF;
                            break;
                        case 0b000'0101:    // MIN
                        {
                            i32 a = static_cast<i32>(thread.registers[rs1]);
                            i32 b = static_cast<i32>(thread.registers[rs2]);
                            thread.registers[rd] = static_cast<u32>(std::min(a, b));
                            break;
                        }
                    }
                    break;
                }
                case 0b101:     // SRL, SRA, MINU
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SRL
                            thread.registers[rd] = thread.registers[rs1] >> (thread.registers[rs2] & 0x1F);
                            break;
                        case 0b010'0000:    // SRA
                            thread.registers[rd] = (i32)thread.registers[rs1] >> (thread.registers[rs2] & 0x1F);
                            break;
                        case 0b000'0101:    // MINU
                            thread.registers[rd] = std::min(thread.registers[rs1], thread.registers[rs2]);
                            break;
                    }
                    break;
                }
                case 0b110:     // OR, MAX
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // OR
                            thread.registers[rd] = thread.registers[rs1] | thread.registers[rs2];
                            break;
                        case 0b000'0101:    // MAX
                        {
                            i32 a = static_cast<i32>(thread.registers[rs1]);
                            i32 b = static_cast<i32>(thread.registers[rs2]);
                            thread.registers[rd] = static_cast<u32>(std::max(a, b));
                            break;
                        }
                    }
                    break;
                }
                case 0b111:     // AND, MAXU
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // AND
                            thread.registers[rd] = thread.registers[rs1] & thread.registers[rs2];
                            break;
                        case 0b000'0101:    // MAXU
                            thread.registers[rd] = std::max(thread.registers[rs1], thread.registers[rs2]);
                            break;
                    }
                    break;
                }
            }
            thread.pc += 4;
            break;
        }

        case OP_CUSTOM:     // custom instruction
        {
            u32 funct3 = decoder::funct3(instruction);

            switch (funct3)
            {
                case 0b000:     // new.t rd, rs1
                {
                    u32 rd = decoder::rd(instruction);
                    u32 rs1 = decoder::rs1(instruction);
                    newT(rd, rs1);
                    break;
                }
                case 0b001:     // yield.t
                {
                    yieldT();
                    break;
                }
                case 0b010:     // id.t rd
                {
                    u32 rd = decoder::rd(instruction);
                    thread.registers[rd] = static_cast<u32>(getThreadId());
                    break;
                }
                case 0b100:     // sleep.t rs1, rs2
                {
                    u32 rs1 = decoder::rs1(instruction);
                    u32 rs2 = decoder::rs2(instruction);
                    sleepT(rs1, rs2);
                    break;
                }
                case 0b101:     // wake.t rs1
                {
                    u32 rs1 = decoder::rs1(instruction);
                    wakeT(rs1);
                    break;
                }
                case 0b111:     // end.t
                {
                    endT();
                    break;
                }
            }
            thread.pc += 4;
            break;
        }

        default:            // unknown instruction
            triggerException(CPU_MAIN_ILLEGAL_INSTRUCTION);
            break;
    }
}

} // namespace vc
