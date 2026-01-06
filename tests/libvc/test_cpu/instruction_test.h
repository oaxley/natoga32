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
    }

    ThreadContext& getCtx(int tid = 0) {
        return cpu.getThreadContext(tid);
    }

    void loadInstruction(u32 instr, int tid = 0) {
        mem.write32(MMAP_CODE, instr);
        getCtx(tid).pc = old_pc = MMAP_CODE;
    }

    void execute() {
        cpu.tick();
        new_pc = getPC();
    }

    u32 getReg(int r, int tid = 0) {
        return getCtx(tid).registers[r];
    }

    void setReg(int r, u32 value, int tid = 0) {
        getCtx(tid).registers[r] = value;
    }

    u32 getPC(int tid = 0) {
        return getCtx(tid).pc;
    }

    u32 deltaPC() {
        return new_pc - old_pc;
    }
};
