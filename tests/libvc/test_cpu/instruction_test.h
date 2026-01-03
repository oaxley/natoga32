// Helper class for instruction testing
#pragma once

#include "core/memory.h"
#include "core/cpu.h"

using namespace vc;

class InstructionTest
{
public:
    Memory mem;
    CPU cpu;
    int tid;
    u32 old_pc = 0;
    u32 new_pc = 0;

    InstructionTest() :
        mem(), cpu(mem), tid{0} {
        cpu.reset();
    }

    void loadInstruction(u32 instr) {
        mem.write32(MMAP_CODE, instr);
        auto& thread = cpu.getThreadContext(tid);
        thread.pc = MMAP_CODE;
        old_pc = thread.pc;
    }

    void execute() {
        cpu.tick();
        new_pc = getPC();
    }

    u32 getReg(int r) {
        auto& thread = cpu.getThreadContext(tid);
        return thread.registers[r];
    }

    void setReg(int r, u32 value) {
        auto& thread = cpu.getThreadContext(tid);
        thread.registers[r] = value;
    }

    u32 getPC() {
        auto& thread = cpu.getThreadContext(tid);
        return thread.pc;
    }

    u32 deltaPC() {
        return new_pc - old_pc;
    }
};
