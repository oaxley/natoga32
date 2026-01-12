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
#include "cpu_thread.h"
#include "csr.h"


namespace vc
{


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
        threads_[tid].init(mem_, tid);
    }

    // reset the global cycle counter
    total_cycles_ = 0;

    // Thread 0 is the main thread
    current_thread_ = 0;
    threads_[current_thread_].setState(ThreadState::Running);
    threads_[current_thread_].setPC(RESET_DEFAULT_ADDR);

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
        if (t.getState() == ThreadState::Sleeping && t.getWaitKey() == event) {
            t.wake();
        }
    }
}

// debug access
CPUThread& CPU::getThread(int id)
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

// yield the current thread
void CPU::yieldT()
{
    CPUThread& current = threads_[current_thread_];

    // step 1 : check the canary for the current thread
    if (current.getState() == ThreadState::Running) {
        if (!current.checkStackCanary(mem_)) {
            // stack overflow detected
            triggerTrap(false, CPU_THREAD_STACK_OVERFLOW_ERROR);
            return;
        }
        current.setState(ThreadState::Ready);
    }

    // step 2 : update cycle-based sleepers
    for (auto& t : threads_) {
        if (t.getState() == ThreadState::Sleeping && t.getSleepUntil() > 0) {
            if (total_cycles_ >= t.getSleepUntil()) {
                t.wake();
            }
        }
    }

    // step 3 : find the next READY thread (round robin search)
    int start = current_thread_;
    do {
        current_thread_ = (current_thread_ + 1) % THREADS_COUNT;
        if (threads_[current_thread_].getState() == ThreadState::Ready) {
            threads_[current_thread_].setState(ThreadState::Running);
            cpu_state_ = CPUState::Running;
            return;
        }
    } while (current_thread_ != start);

    // step 4 : no READY threads - check for deadlocks
    bool has_hardware_waiters = false;
    bool has_cycle_waiters = false;

    for (const auto& t : threads_) {
        if (t.getState() == ThreadState::Sleeping) {
            if (isHardwareEvent(t.getWaitKey())) {
                has_hardware_waiters = true;
            }
            if (t.getSleepUntil() > 0) {
                has_cycle_waiters = true;
            }
        }
    }

    if (has_hardware_waiters || has_cycle_waiters) {
        // CPU idle - waiting for external events
        cpu_state_ = CPUState::Idle;
    } else {
        // deadlock situation
        triggerTrap(false, CPU_THREAD_DEADLOCK);
    }
}

// move the current thread to Sleeping status
void CPU::sleepT(u8 rs1, u8 rs2)
{
    CPUThread& t = threads_[current_thread_];
    auto& registers = t.getRegisters();
    bool yield = false;

    u32 waitkey_value = 0;
    u32 sleep_cycles = 0;

    // waitkey
    if (registers[rs1] > 0) {
        waitkey_value = registers[rs1];
        yield = true;
    }

    // sleep until
    if (registers[rs2] > 0) {
        sleep_cycles = registers[rs2];
        yield = true;
    }

    // switch to next thread
    if (yield) {
        t.sleep(total_cycles_, waitkey_value, sleep_cycles);
        yieldT();
    }
}

// wake all the threads that match the waitkey value in RS1
void CPU::wakeT(u8 rs1)
{
    u32 waitkey = threads_[current_thread_].getRegisters()[rs1];

    for (auto& t : threads_) {
        if (t.getState() == ThreadState::Sleeping && t.getWaitKey() == waitkey) {
            t.wake();
        }
    }
}

// end the current thread
void CPU::endT()
{
    // Thread 0 is immortal
    if (current_thread_ == 0) {
        triggerTrap(false, CPU_THREAD_MAIN_EXIT_ERROR);
        return;
    }

    // mark this thread as dead, and yield
    threads_[current_thread_].end();
    yieldT();
}

// create a new thread
void CPU::newT(u8 rd, u8 rs1)
{
    auto& thread = threads_[current_thread_];
    auto& registers = thread.getRegisters();

    // find a free slot
    for (int tid = 0; tid < THREADS_COUNT; tid++) {
        auto& t = threads_[tid];

        if (t.getState() == ThreadState::Free) {
            // initialize this thread with the user parameters
            t.init(mem_, tid, registers[rs1]);
            t.setState(ThreadState::Ready);

            // return its id
            registers[rd] = tid;
            return;
        }
    }

    // no available slot found!
    triggerTrap(false, CPU_THREAD_SPAWN_ERROR);
}

// trigger
void CPU::triggerTrap(bool is_interrupt, u32 cause, u32 trap_value)
{
    // save the current PC
    writeCSR(CSR_MEPC, threads_[current_thread_].getPC());

    // set cause
    u32 value = cause;
    if (is_interrupt) {
        value |= 0x8000'0000;       // set interrupt bit
    }
    writeCSR(CSR_MCAUSE, value);

    // set trap value
    writeCSR(CSR_MTVAL, trap_value);

    // jump to handler
    u32 addr_handler = readCSR(CSR_MTVEC);
    threads_[current_thread_].setPC(addr_handler);
}



void CPU::execute_instruction()
{
    // thread accessor
    CPUThread& thread = threads_[current_thread_];
    auto& registers = thread.getRegisters();
    u32 pc = thread.getPC();

    // fetch the instruction
    u32 instruction = mem_.read32(pc);

    // decode all the parts
    u8 opcode = decoder::opcode(instruction);
    u8 rd = decoder::rd(instruction);
    u8 rs1 = decoder::rs1(instruction);
    u8 rs2 = decoder::rs2(instruction);
    u8 funct3 = decoder::funct3(instruction);
    u8 funct7 = decoder::funct7(instruction);

    // retrieve the value for RS1 and RS2
    u32 urs1 = registers[rs1];
    u32 urs2 = registers[rs2];

    // Very BIG switch/case
    switch(opcode)
    {
        case OP_LUI:        // load upper immediate
        {
            registers[rd] = decoder::immTypeU(instruction);
            pc += 4;
            break;
        }

        case OP_AUIPC:      // add upper immediate to PC
        {
            registers[rd] = pc + decoder::immTypeU(instruction);
            pc += 4;
            break;
        }

        case OP_JAL:        // jump and link
        {
            registers[rd] = pc + 4;
            pc += decoder::immTypeJ(instruction);;
            break;
        }

        case OP_JALR:       // jump and link register
        {
            i32 imm = decoder::immTypeI(instruction);
            u32 target = (urs1 + imm) & ~1;
            registers[rd] = pc + 4;
            pc = target;
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
                pc += imm;
            } else {
                pc += 4;
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
                    registers[rd] = static_cast<i32>(static_cast<i8>(mem_.read8(addr)));
                    break;
                case 0b001:     // LH
                    registers[rd] = static_cast<i32>(static_cast<i16>(mem_.read16(addr)));
                    break;
                case 0b010:     // LW
                    registers[rd] = mem_.read32(addr);
                    break;
                case 0b100:     // LBU
                    registers[rd] = mem_.read8(addr);
                    break;
                case 0b101:     // LWU
                    registers[rd] = mem_.read16(addr);
                    break;
            }
            pc += 4;
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
            pc += 4;
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
                    registers[rd] = urs1 + imm;
                    break;
                case 0b001:     // SLLI or Zbb instruction
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLLI
                            registers[rd] = urs1 << shamt;
                            break;
                        case 0b001'0100:    // BSETI
                            registers[rd] = urs1 | (1u << shamt);
                            break;
                        case 0b010'0100:    // BCLRI
                            registers[rd] = urs1 & ~(1u << shamt);
                            break;
                        case 0b011'0000:    // SEXT.[H,B] / CLZ / CTZ / CPOP
                        {
                            switch (shamt)
                            {
                                case 0b00000:       // CLZ
                                    registers[rd] = std::countl_zero(urs1);
                                    break;
                                case 0b00001:       // CTZ
                                    registers[rd] = std::countr_zero(urs1);
                                    break;
                                case 0b00010:       // CPOP
                                    registers[rd] = std::popcount(urs1);
                                    break;
                                case 0b00100:       // SEXT.B
                                    registers[rd] = static_cast<u32>(static_cast<i8>(urs1));
                                    break;
                                case 0b00101:       // SEXT.H
                                    registers[rd] = static_cast<u32>(static_cast<i16>(urs1));
                                    break;
                            }
                            break;
                        }
                        case 0b011'0100:    // BINVI
                                registers[rd] = urs1 ^ (1u << shamt);
                            break;

                    }
                    break;
                }
                case 0b010:     // SLTI
                    registers[rd] = (irs1 < imm) ? 1 : 0;
                    break;
                case 0b011:     // SLTIU
                    registers[rd] = (urs1 < static_cast<u32>(imm)) ? 1 : 0;
                    break;
                case 0b100:     // XORI
                    registers[rd] = urs1 ^ imm;
                    break;
                case 0b101:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SRLI
                            registers[rd] = urs1 >> shamt;
                            break;
                        case 0b010'0000:    // SRAI
                            registers[rd] = irs1 >> shamt;
                            break;
                        case 0b010'0100:    // BEXTI
                            registers[rd] = (urs1 >> shamt) & 1u;
                            break;
                        case 0b011'0000:    // RORI
                            registers[rd] = std::rotr(urs1, shamt);
                            break;
                        case 0b011'0100:    // REV8
                        {
                            registers[rd] = ((urs1 >> 24) & 0xFF)
                                                 | ((urs1 >> 8 ) & 0xFF00)
                                                 | ((urs1 << 8) & 0xFF0000)
                                                 | ((urs1 << 24) & 0xFF000000);
                            break;
                        }
                    }
                    break;
                }
                case 0b110:     // ORI
                    registers[rd] = urs1 | imm;
                    break;
                case 0b111:     // ANDI
                    registers[rd] = urs1 & imm;
                    break;
            }
            pc += 4;
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
                            registers[rd] = urs1 + urs2;
                            break;
                        case 0b000'0001:    // MUL
                            registers[rd] = urs1 * urs2;
                            break;
                        case 0b010'0000:    // SUB
                            registers[rd] = urs1 - urs2;
                            break;
                    }
                    break;
                }
                case 0b001:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLL
                            registers[rd] = urs1 << shamt;
                            break;
                        case 0b000'0001:    // MULH
                        {
                            i64 a = static_cast<i64>(irs1);
                            i64 b = static_cast<i64>(irs2);
                            i64 result = a * b;
                            registers[rd] = static_cast<u32>(result >> 32);
                            break;
                        }
                        case 0b001'0100:    // BSET
                            registers[rd] = urs1 | (1u << shamt);
                            break;
                        case 0b010'0100:    // BCLR
                            registers[rd] = urs1 & ~(1u << shamt);
                            break;
                        case 0b011'0000:    // ROL
                            registers[rd] = std::rotl(urs1, shamt);
                            break;
                        case 0b011'0100:    // BINV
                            registers[rd] = urs1 ^ (1u << shamt);
                            break;
                    }
                    break;
                }
                case 0b010:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SLT
                            registers[rd] = (irs1 < irs2) ? 1 : 0;
                            break;
                        case 0b000'0001:    // MULHSU
                        {
                            i64 a = static_cast<i64>(irs1);
                            i64 b = static_cast<i64>(rs2);
                            i64 result = a * b;
                            registers[rd] = static_cast<u32>(result >> 32);
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
                            registers[rd] = (urs1 < urs2) ? 1 : 0;
                            break;
                        case 0b000'0001:    // MULHU
                        {
                            u64 a = static_cast<u64>(urs1);
                            u64 b = static_cast<u64>(urs2);
                            u64 result = a * b;
                            registers[rd] = static_cast<u32>(result >> 32);
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
                            registers[rd] = urs1 ^ urs2;
                            break;
                        case 0b000'0001:     // DIV
                        {
                            i32 dividend = irs1;
                            i32 divisor = irs2;

                            if (divisor == 0) {
                                registers[rd] = 0xFFFF'FFFF;     // -1
                            } else if (dividend == INT32_MIN && divisor == -1) {
                                registers[rd] = static_cast<u32>(INT32_MIN);     // overflow
                            } else {
                                registers[rd] = static_cast<u32>(dividend / divisor);
                            }
                            break;
                        }
                        case 0b000'0100:    // ZEXT.H & PACK
                        {
                            if (rs2 == 0) {         //< we need the register to be 0 (not its value)
                                // ZEXT.H
                                registers[rd] = urs1 & 0xFFFF;
                            } else {
                                // PACK
                                registers[rd] = (urs1 & 0xFFFF) | (urs2 << 16);
                            }
                            break;
                        }
                        case 0b000'0101:    // MIN
                            registers[rd] = static_cast<u32>(std::min(irs1, irs2));
                            break;
                    }
                    break;
                }
                case 0b101:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // SRL
                            registers[rd] = urs1 >> shamt;
                            break;
                        case 0b000'0001:    // DIVU
                        {
                            if (urs2 == 0) {
                                registers[rd] = 0xFFFF'FFFF; // max unsigned
                            } else {
                                registers[rd] = urs1 / urs2;
                            }
                            break;
                        }
                        case 0b000'0101:    // MINU
                            registers[rd] = std::min(urs1, urs2);
                            break;
                        case 0b000'0111:    // CZERO.EQZ
                            registers[rd] = (urs2 == 0) ? 0 : urs1;
                            break;
                        case 0b010'0000:    // SRA
                            registers[rd] = static_cast<u32>(irs1 >> shamt);
                            break;
                        case 0b010'0100:    // BEXT
                            registers[rd] = (urs1 >> shamt) & 1u;
                            break;
                        case 0b011'0000:    // ROR
                            registers[rd] = std::rotr(urs1, shamt);
                            break;
                    }
                    break;
                }
                case 0b110:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // OR
                            registers[rd] = urs1 | urs2;
                            break;
                        case 0b000'0001:    // REM
                        {
                            i32 dividend = irs1;
                            i32 divisor = irs2;

                            if (divisor == 0) {
                                registers[rd] = urs1;
                            } else if (dividend == INT32_MIN && divisor == -1) {
                                registers[rd] = 0;     // overflow
                            } else {
                                registers[rd] = static_cast<u32>(dividend % divisor);
                            }
                            break;
                        }
                        case 0b000'0101:    // MAX
                            registers[rd] = static_cast<u32>(std::max(irs1, irs2));
                            break;
                    }
                    break;
                }
                case 0b111:
                {
                    switch (funct7)
                    {
                        case 0b000'0000:    // AND
                            registers[rd] = urs1 & urs2;
                            break;
                        case 0b000'0001:    // REMU
                        {
                            if (urs2 == 0) {
                                registers[rd] = urs1;
                            } else {
                                registers[rd] = urs1 % urs2;
                            }
                            break;
                        }
                        case 0b000'0100:    // PACKH
                            registers[rd] = (urs1 & 0xFF) | ((urs2 & 0xFF) << 8);
                            break;
                        case 0b000'0101:    // MAXU
                            registers[rd] = std::max(urs1, urs2);
                            break;
                        case 0b000'0111:    // CZERO.NEZ
                            registers[rd] = (urs2 != 0) ? 0 : urs1;
                            break;
                    }
                    break;
                }
            }
            pc += 4;
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
                    registers[rd] = static_cast<u32>(current_thread_);
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
            pc += 4;
            break;
        }

        case OP_SYSTEM:     // system instructions
        {
            u32 csr = (instruction >> 20) & 0xFFF;
            u32 csr_val = readCSR(csr);

            switch (funct3)
            {
                case 0b000:     // ECALL/MRET
                {
                    u32 funct12 = instruction >> 20;

                    switch (funct12)
                    {
                        case 0x000:    // ECALL
                        {
                            // trigger environment call exception
                            triggerTrap(false, CPU_ECALL_FROM_M_MODE);

                            // we return early, don't do PC+4
                            return;
                        }
                        case 0x302:     // MRET
                        {
                            // restore the PC from MEPC
                            pc = readCSR(CSR_MEPC);

                            // update MSTATUS to restore interrupt enable bits
                            u32 mstatus = readCSR(CSR_MSTATUS);
                            u32 mpie = (mstatus >> 7) & 1;              // MPIE / bit 7
                            mstatus = (mstatus & ~0x8) | (mpie << 3);   // MIE = MPIE
                            mstatus |= (1 << 7);
                            writeCSR(CSR_MSTATUS, mstatus);

                            break;
                        }
                    }
                    break;
                }
                case 0b001:     // CSRRW
                {
                    writeCSR(csr, urs1);
                    registers[rd] = csr_val;
                    break;
                }
                case 0b010:     // CSRRS
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val | urs1);
                    }
                    registers[rd] = csr_val;
                    break;
                }
                case 0b011:     // CSRRC
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val & ~urs1);
                    }
                    registers[rd] = csr_val;
                    break;
                }
                case 0b101:     // CSRRWI
                {
                    // rs1 holds the 5-bit immediate
                    writeCSR(csr, rs1);
                    registers[rd] = csr_val;
                    break;
                }
                case 0b110:     // CSRRSI
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val | rs1);
                    }
                    registers[rd] = csr_val;
                    break;
                }
                case 0b111:     // CSRRCI
                {
                    if (rs1 != 0) {
                        writeCSR(csr, csr_val & ~rs1);
                    }
                    registers[rd] = csr_val;
                    break;
                }
            }
            pc += 4;
            break;
        }

        default:            // unknown instruction
            triggerTrap(false, CPU_MAIN_ILLEGAL_INSTRUCTION);
            break;
    }

    // write back the PC to the thread
    thread.setPC(pc);

    // increase the number of cycles for this thread
    thread.incrementCycles();

    // ensure x0 remains at 0 as per specification
    registers[0] = 0;
}

// CSR read/write
u32 CPU::readCSR(u16 csr)
{
    if (csr >= 4096) {
        triggerTrap(false, CPU_MAIN_ILLEGAL_INSTRUCTION);
        return 0;
    }

    // handle dynamic CSRs
    switch (csr)
    {
        case 0xC00: // cycle (low 32 bits)
            return static_cast<u32>(total_cycles_);
        case 0xC80: // cycleh (high 32 bits)
            return static_cast<u32>(total_cycles_ >> 32);

        case 0xC01: // time (low 32 bits)
            return static_cast<u32>(virtual_time_);
        case 0xC81: // timeh (high 32 bits)
            return static_cast<u32>(virtual_time_ >> 32);

        default:
            return mem_.read32(MMAP_CSR_REGISTERS + csr * sizeof(u32));
    }
}

void CPU::writeCSR(u16 csr, u32 value)
{
    // read-only CSR are the one starting with 0b11
    if ((csr >> 10) == 3) {
        triggerTrap(false, CPU_MAIN_ILLEGAL_INSTRUCTION);
        return;
    }

    // write in the register
    mem_.write32(MMAP_CSR_REGISTERS + csr * sizeof(u32), value);
}

} // namespace vc
