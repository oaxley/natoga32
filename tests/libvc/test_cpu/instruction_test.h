// Helper class for instruction testing
#pragma once
#include <array>

#include "core/memory.h"
#include "core/cpu.h"

using namespace vc;

struct ThreadTest
{
    u32 size;           // code size
    u32 addr;           // current code address
    u32 old_pc;         // previous PC before execution
};

class InstructionTest
{
private:
    std::array<ThreadTest, THREADS_COUNT> thread_;

public:
    Memory mem;
    CPU cpu;

    InstructionTest() :
        thread_{}, mem(), cpu(mem)
    {
        for (int tid = 0; tid < THREADS_COUNT; tid++) {
            thread_[tid].size = 0;
            thread_[tid].addr = MMAP_CODE;
            thread_[tid].old_pc = 0;
        }
    }

    ThreadContext& getCtx(int tid = 0)
    {
        return cpu.getThreadContext(tid);
    }

    void loadInstruction(u32 instr, u32 offset = 0, int tid = 0)
    {
        u32 addr = thread_[tid].addr + offset + (thread_[tid].size * sizeof(u32));
        mem.write32(addr, instr);
        if (thread_[tid].size == 0) {
            getCtx(tid).pc = addr;
            thread_[tid].old_pc = addr;
        }
        thread_[tid].size++;
    }

    void reset(int tid = 0)
    {
        thread_[tid].size = 0;
        thread_[tid].addr = MMAP_CODE;
        thread_[tid].old_pc = MMAP_CODE;
    }

    u32 getPC(int tid = 0) {
        return getCtx(tid).pc;
    }

    void execute()
    {
        cpu.tick();
    }

    u32 getReg(int r, int tid = 0) {
        return getCtx(tid).registers[r];
    }

    void setReg(int r, u32 value, int tid = 0) {
        getCtx(tid).registers[r] = value;
    }

    u32 deltaPC(int tid = 0) {
        u32 new_pc = getCtx(tid).pc;
        u32 delta = new_pc - thread_[tid].old_pc;
        thread_[tid].old_pc = new_pc;
        return delta;
    }
};
