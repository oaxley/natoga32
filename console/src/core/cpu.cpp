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
#include <bit>

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

void CPU::tick()
{
    if (cpu_state_ != CPUState::Running) {
        return;
    }

    // execute one instruction
    execute_instruction();

    // increment the global instructions cycle counter
    total_cycles_++;
}

// wake any threads waiting for this event
void CPU::wakeThreadOnEvent(u32 event)
{
    for (auto& t : threads_) {
        if (t.state == ThreadState::Sleeping && t.waitkey == event) {
            t.state = ThreadState::Ready;
            t.waitkey = 0;
            t.sleep_until = 0;
        }
    }
}

std::tuple<u32, u32> CPU::getLastException() const
{
    return std::make_tuple(exception_id_, exception_pc_);
}

// debug access
ThreadContext& CPU::getThreadContext(int id)
{
    return threads_[id];
}

int CPU::getCurrentThreadID() const
{
    return current_thread_;
}

CPUState CPU::getState() const
{
    return cpu_state_;
}


//----- private methods
// initialize a Thread
void CPU::initT(int tid, u32 entrypoint)
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
void CPU::sleepT(u8 rs1, u8 rs2)
{
    ThreadContext& t = threads_[current_thread_];
    bool yield = false;

    // waitkey
    if (t.registers[rs1] > 0) {
        t.waitkey = t.registers[rs1];
        yield = true;
    }

    // sleep until
    if (t.registers[rs2] > 0) {
        t.sleep_until = total_cycles_ + static_cast<u64>(t.registers[rs2]);
        yield = true;
    }

    // switch to next thread
    if (yield) {
        t.state = ThreadState::Sleeping;
        yieldT();
    }
}

// wake all the threads that match the waitkey value in RS1
void CPU::wakeT(u8 rs1)
{
    u32 waitkey = threads_[current_thread_].registers[rs1];

    for (auto& t : threads_) {
        if (t.state == ThreadState::Sleeping && t.waitkey == waitkey) {
            t.state = ThreadState::Ready;
            t.waitkey = 0;
            t.sleep_until = 0;
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
void CPU::newT(u8 rd, u8 rs1)
{
    auto& thread = threads_[current_thread_];

    // find a free slot
    for (int tid = 0; tid < THREADS_COUNT; tid++) {
        auto& t = threads_[tid];

        if (t.state == ThreadState::Free) {
            // initialize this thread with the user parameters
            initT(tid, thread.registers[rs1]);
            t.state = ThreadState::Ready;

            // return its id
            thread.registers[rd] = tid;
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

    // decode all the parts
    u8 opcode = decoder::opcode(instruction);
    u8 rd = decoder::rd(instruction);
    u8 rs1 = decoder::rs1(instruction);
    u8 rs2 = decoder::rs2(instruction);
    u8 funct3 = decoder::funct3(instruction);
    u8 funct7 = decoder::funct7(instruction);

    // retrieve the value for RS1 and RS2
    u32 urs1 = thread.registers[rs1];
    u32 urs2 = thread.registers[rs2];

    // Very BIG switch/case
    switch(opcode)
    {
        case OP_LUI:        // load upper immediate
        {
            thread.registers[rd] = decoder::immTypeU(instruction);
            thread.pc += 4;
            break;
        }

        case OP_AUIPC:      // add upper immediate to PC
        {
            thread.registers[rd] = thread.pc + decoder::immTypeU(instruction);
            thread.pc += 4;
            break;
        }

        case OP_JAL:        // jump and link
        {
            thread.registers[rd] = thread.pc + 4;
            thread.pc += decoder::immTypeJ(instruction);;
            break;
        }

        case OP_JALR:       // jump and link register
        {
            i32 imm = decoder::immTypeI(instruction);
            u32 target = (urs1 + imm) & ~1;
            thread.registers[rd] = thread.pc + 4;
            thread.pc = target;
            break;
        }

        case OP_BRANCH:     // branches
        {
            bool take_branch = false;
            i32 irs1 = static_cast<i32>(urs1);
            i32 irs2 = static_cast<i32>(urs2);
            i32 imm = decoder::immTypeB(instruction);

            switch(funct3)
            {
                case 0b000: // BEQ
                    take_branch = (irs1 == irs2);
                    break;
                case 0b001: // BNE
                    take_branch = (irs1 != irs2);
                    break;
                case 0b100: // BLT
                    take_branch = (irs1 < irs2);
                    break;
                case 0b101: // BGE
                    take_branch = (irs1 >= irs2);
                    break;
                case 0b110: // BLTU
                    take_branch = (urs1 < urs2);
                    break;
                case 0b111: // BGEU
                    take_branch = (urs1 >= urs2);
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
            i32 imm = decoder::immTypeI(instruction);
            u32 addr = urs1 + imm;

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
            i32 imm = decoder::immTypeS(instruction);
            u32 addr = urs1 + imm;

            switch(funct3)
            {
                case 0b000:     // SB
                    mem_.write8(addr, urs2);
                    break;
                case 0b001:     // SH
                    mem_.write16(addr, urs2);
                    break;
                case 0b010:     // SW
                    mem_.write32(addr, urs2);
                    break;
            }
            thread.pc += 4;
            break;
        }

        case OP_OP_IMM:     // Integer operations with Immediate
        {
            i32 imm = decoder::immTypeI(instruction);
            u32 shamt = decoder::rs2(instruction) & 0x1F;
            i32 irs1 = static_cast<i32>(urs1);

            switch(funct3)
            {
                case 0b000:     // ADDI
                    thread.registers[rd] = urs1 + imm;
                    break;
                case 0b001:     // SLLI or Zbb instruction
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLLI
                            thread.registers[rd] = urs1 << shamt;
                            break;
                        case 0b001'0100:    // BSETI
                            thread.registers[rd] = urs1 | (1u << shamt);
                            break;
                        case 0b010'0100:    // BCLRI
                            thread.registers[rd] = urs1 & ~(1u << shamt);
                            break;
                        case 0b011'0000:    // SEXT.[H,B] / CLZ / CTZ / CPOP
                        {
                            switch (shamt)
                            {
                                case 0b00000:       // CLZ
                                    thread.registers[rd] = std::countl_zero(urs1);
                                    break;
                                case 0b00001:       // CTZ
                                    thread.registers[rd] = std::countr_zero(urs1);
                                    break;
                                case 0b00010:       // CPOP
                                    thread.registers[rd] = std::popcount(urs1);
                                    break;
                                case 0b00100:       // SEXT.B
                                    thread.registers[rd] = static_cast<u32>(static_cast<i8>(urs1));
                                    break;
                                case 0b00101:       // SEXT.H
                                    thread.registers[rd] = static_cast<u32>(static_cast<i16>(urs1));
                                    break;
                            }
                            break;
                        }
                        case 0b011'0100:    // BINVI
                                thread.registers[rd] = urs1 ^ (1u << shamt);
                            break;

                    }
                    break;
                }
                case 0b010:     // SLTI
                    thread.registers[rd] = (irs1 < imm) ? 1 : 0;
                    break;
                case 0b011:     // SLTIU
                    thread.registers[rd] = (urs1 < static_cast<u32>(imm)) ? 1 : 0;
                    break;
                case 0b100:     // XORI
                    thread.registers[rd] = urs1 ^ imm;
                    break;
                case 0b101:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SRLI
                            thread.registers[rd] = urs1 >> shamt;
                            break;
                        case 0b010'0000:    // SRAI
                            thread.registers[rd] = irs1 >> shamt;
                            break;
                        case 0b010'0100:    // BEXTI
                            thread.registers[rd] = (urs1 >> shamt) & 1u;
                            break;
                        case 0b011'0000:    // RORI
                            thread.registers[rd] = std::rotr(urs1, shamt);
                            break;
                        case 0b011'0100:    // REV8
                        {
                            thread.registers[rd] = ((urs1 >> 24) & 0xFF)
                                                 | ((urs1 >> 8 ) & 0xFF00)
                                                 | ((urs1 << 8) & 0xFF0000)
                                                 | ((urs1 << 24) & 0xFF000000);
                            break;
                        }
                    }
                    break;
                }
                case 0b110:     // ORI
                    thread.registers[rd] = urs1 | imm;
                    break;
                case 0b111:     // ANDI
                    thread.registers[rd] = urs1 & imm;
                    break;
            }
            thread.pc += 4;
            break;
        }

        case OP_OP:         // register-register operations
        {
            u32 shamt = urs2 & 0x1F;
            i32 irs1 = static_cast<i32>(urs1);
            i32 irs2 = static_cast<i32>(urs2);

            switch(funct3)
            {
                case 0b000:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // ADD
                            thread.registers[rd] = urs1 + urs2;
                            break;
                        case 0b000'0001:    // MUL
                            thread.registers[rd] = urs1 * urs2;
                            break;
                        case 0b010'0000:    // SUB
                            thread.registers[rd] = urs1 - urs2;
                            break;
                    }
                    break;
                }
                case 0b001:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLL
                            thread.registers[rd] = urs1 << shamt;
                            break;
                        case 0b000'0001:    // MULH
                        {
                            i64 a = static_cast<i64>(irs1);
                            i64 b = static_cast<i64>(irs2);
                            i64 result = a * b;
                            thread.registers[rd] = static_cast<u32>(result >> 32);
                            break;
                        }
                        case 0b001'0100:    // BSET
                            thread.registers[rd] = urs1 | (1u << shamt);
                            break;
                        case 0b010'0100:    // BCLR
                            thread.registers[rd] = urs1 & ~(1u << shamt);
                            break;
                        case 0b011'0000:    // ROL
                            thread.registers[rd] = std::rotl(urs1, shamt);
                            break;
                        case 0b011'0100:    // BINV
                            thread.registers[rd] = urs1 ^ (1u << shamt);
                            break;
                    }
                    break;
                }
                case 0b010:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLT
                            thread.registers[rd] = (irs1 < irs2) ? 1 : 0;
                            break;
                        case 0b000'0001:    // MULHSU
                        {
                            i64 a = static_cast<i64>(irs1);
                            i64 b = static_cast<i64>(rs2);
                            i64 result = a * b;
                            thread.registers[rd] = static_cast<u32>(result >> 32);
                            break;
                        }
                    }
                    break;
                }
                case 0b011:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLTU
                            thread.registers[rd] = (urs1 < urs2) ? 1 : 0;
                            break;
                        case 0b000'0001:    // MULHU
                        {
                            u64 a = static_cast<u64>(urs1);
                            u64 b = static_cast<u64>(urs2);
                            u64 result = a * b;
                            thread.registers[rd] = static_cast<u32>(result >> 32);
                            break;
                        }
                    }
                    break;
                }
                case 0b100:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // XOR
                            thread.registers[rd] = urs1 ^ urs2;
                            break;
                        case 0b000'0001:     // DIV
                        {
                            i32 dividend = irs1;
                            i32 divisor = irs2;

                            if (divisor == 0) {
                                thread.registers[rd] = 0xFFFF'FFFF;     // -1
                            } else if (dividend == INT32_MIN && divisor == -1) {
                                thread.registers[rd] = static_cast<u32>(INT32_MIN);     // overflow
                            } else {
                                thread.registers[rd] = static_cast<u32>(dividend / divisor);
                            }
                            break;
                        }
                        case 0b000'0100:    // ZEXT.H & PACK
                        {
                            if (rs2 == 0) {         //< we need the register to be 0 (not its value)
                                // ZEXT.H
                                thread.registers[rd] = urs1 & 0xFFFF;
                            } else {
                                // PACK
                                thread.registers[rd] = (urs1 & 0xFFFF) | (urs2 << 16);
                            }
                            break;
                        }
                        case 0b000'0101:    // MIN
                            thread.registers[rd] = static_cast<u32>(std::min(irs1, irs2));
                            break;
                    }
                    break;
                }
                case 0b101:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SRL
                            thread.registers[rd] = urs1 >> shamt;
                            break;
                        case 0b000'0001:    // DIVU
                        {
                            if (urs2 == 0) {
                                thread.registers[rd] = 0xFFFF'FFFF; // max unsigned
                            } else {
                                thread.registers[rd] = urs1 / urs2;
                            }
                            break;
                        }
                        case 0b000'0101:    // MINU
                            thread.registers[rd] = std::min(urs1, urs2);
                            break;
                        case 0b000'0111:    // CZERO.EQZ
                            thread.registers[rd] = (urs2 == 0) ? 0 : urs1;
                            break;
                        case 0b010'0000:    // SRA
                            thread.registers[rd] = static_cast<u32>(irs1 >> shamt);
                            break;
                        case 0b010'0100:    // BEXT
                            thread.registers[rd] = (urs1 >> shamt) & 1u;
                            break;
                        case 0b011'0000:    // ROR
                            thread.registers[rd] = std::rotr(urs1, shamt);
                            break;
                    }
                    break;
                }
                case 0b110:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // OR
                            thread.registers[rd] = urs1 | urs2;
                            break;
                        case 0b000'0001:    // REM
                        {
                            i32 dividend = irs1;
                            i32 divisor = irs2;

                            if (divisor == 0) {
                                thread.registers[rd] = urs1;
                            } else if (dividend == INT32_MIN && divisor == -1) {
                                thread.registers[rd] = 0;     // overflow
                            } else {
                                thread.registers[rd] = static_cast<u32>(dividend % divisor);
                            }
                            break;
                        }
                        case 0b000'0101:    // MAX
                            thread.registers[rd] = static_cast<u32>(std::max(irs1, irs2));
                            break;
                    }
                    break;
                }
                case 0b111:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // AND
                            thread.registers[rd] = urs1 & urs2;
                            break;
                        case 0b000'0001:    // REMU
                        {
                            if (urs2 == 0) {
                                thread.registers[rd] = urs1;
                            } else {
                                thread.registers[rd] = urs1 % urs2;
                            }
                            break;
                        }
                        case 0b000'0100:    // PACKH
                            thread.registers[rd] = (urs1 & 0xFF) | ((urs2 & 0xFF) << 8);
                            break;
                        case 0b000'0101:    // MAXU
                            thread.registers[rd] = std::max(urs1, urs2);
                            break;
                        case 0b000'0111:    // CZERO.NEZ
                            thread.registers[rd] = (urs2 != 0) ? 0 : urs1;
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
            switch (funct3)
            {
                case 0b000:     // new.t rd, rs1
                    newT(rd, rs1);
                    break;
                case 0b001:     // yield.t
                    yieldT();
                    break;
                case 0b010:     // id.t rd
                    thread.registers[rd] = static_cast<u32>(current_thread_);
                    break;
                case 0b100:     // sleep.t rs1, rs2
                    sleepT(rs1, rs2);
                    break;
                case 0b101:     // wake.t rs1
                    wakeT(rs1);
                    break;
                case 0b111:     // end.t
                    endT();
                    break;
            }
            thread.pc += 4;
            break;
        }

        case OP_SYSTEM:     // system instructions
        {
            u32 csr = (instruction >> 20) & 0xFFF;
            u32 csr_val = readCSR(csr);

            switch (funct3)
            {
                case 0b000:     // ecall, mret, wfi ...
                    break;
                case 0b001:     // CSRRW
                {
                    writeCSR(csr, urs1);
                    thread.registers[rd] = csr_val;
                    break;
                }
                case 0b010:     // CSRRS
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val | urs1);
                    }
                    thread.registers[rd] = csr_val;
                    break;
                }
                case 0b011:     // CSRRC
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val & ~urs1);
                    }
                    thread.registers[rd] = csr_val;
                    break;
                }
                case 0b101:     // CSRRWI
                {
                    // rs1 holds the 5-bit immediate
                    writeCSR(csr, rs1);
                    thread.registers[rd] = csr_val;
                    break;
                }
                case 0b110:     // CSRRSI
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val | rs1);
                    }
                    thread.registers[rd] = csr_val;
                    break;
                }
                case 0b111:     // CSRRCI
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val & ~rs1);
                    }
                    thread.registers[rd] = csr_val;
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

    // increase the number of cycles for this thread
    thread.total_cycles++;

    // ensure x0 remains at 0 as per specification
    thread.registers[0] = 0;
}

// CSR read/write
u32 CPU::readCSR(u16 csr)
{
    if (csr >= 4096) {
        triggerException(CPU_MAIN_ILLEGAL_INSTRUCTION);
        return 0;
    }

    // handle dynamic CSRs
    switch (csr)
    {
        case 0xC00: // cycle (low 32 bits)
            return static_cast<u32>(total_cycles_);
        case 0xC80: // cycleh (high 32 bits)
            return static_cast<u32>(total_cycles_ >> 32);
        default:
            return mem_.read32(MMAP_CSR_REGISTERS + csr * sizeof(u32));
    }
}

void CPU::writeCSR(u16 csr, u32 value)
{
    // TODO: check for RO registers
    mem_.write32(MMAP_CSR_REGISTERS + csr * sizeof(u32), value);
}

} // namespace vc
