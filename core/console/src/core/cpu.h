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
#include <span>
#include <memory>
#include <tuple>

// program-specific headers
#include "vc/types.h"
#include "constants.h"
#include "cpu_thread.h"

namespace vc
{
// forward declarations
class Memory;
class Cartridge;


//----- enums
enum class CPUState
{
    Running,        // CPU is actively executing instructions
    Idle,           // All threads sleeping on events/timers
    Halted          // trap without a handler (mtvec = 0), stopped until reset
};


//----- class
class CPU
{
public:
    CPU(Memory& mem);

    void reset();
    void tick();

    // fast-forward an idle CPU by up to max_cycles, returns the cycles consumed
    u64 skipIdle(u64 max_cycles);

    void wakeThreadOnEvent(u32 event);

    // cartridge seen through the window (nullptr = none, the window reads 0)
    void setCartridge(const Cartridge* cart);

    // debug access
    CPUThread& getThread(int id);
    int getCurrentThreadID() const;
    CPUState getState() const;
    u64 getTotalCycles() const;

private:
    //----- private members
    std::array<CPUThread, THREADS_COUNT> threads_;
    int current_thread_ = 0;
    u64 total_cycles_ = 0;
    u64 virtual_time_ = 0;
    bool trap_taken_ = false;       //< set by triggerTrap during the current instruction
    bool stack_overflow_ = false;   //< a stack canary was found broken: the canary is no longer checked

    Memory& mem_;
    const Cartridge* cartridge_ = nullptr;
    CPUState cpu_state_;

    //----- private methods
    // threads methods
    void yieldT();
    void resumeFromIdle();
    void wakeExpiredSleepers();
    void sleepT(u8 rs1, u8 rs2);
    void wakeT(u8 rs1);
    void endT();
    void newT(u8 rd, u8 rs1);
    bool stackOverflowed();

    // interruption / exception trigger
    void triggerTrap(bool is_interrupt, u32 cause, u32 trap_value = 0);

    // memory protection of the private thread zone and the cartridge zones
    static bool touchesZone(u32 addr, u32 size, u32 base, size_t zone_size);
    static bool insideZone(u32 addr, u32 size, u32 base, size_t zone_size);
    bool isPrivateAccess(u32 addr, u32 size) const;
    u32 readWindow(u32 addr, u32 size) const;
    static u32 accessSize(u32 funct3);

    // instruction executor
    void execute_instruction();

    // CSR read/write
    static bool isReadOnlyCSR(u16 csr);
    u32 readCSR(u16 csr);
    void writeCSR(u16 csr, u32 value);
};


} // namespace vc

