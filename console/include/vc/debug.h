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
 * @brief	Debugger structures
 */
#pragma once

// program-specific includes
#include "vc/types.h"

namespace vc::debug
{

// define the message types
enum class MessageType : u16
{
    // Requests (debugger -> host)
    GetState = 0x0001,
    ReadMemory = 0x0002,
    WriteMemory = 0x0003,
    SetBreakpoint = 0x0004,
    ClearBreakpoint = 0x0005,
    StepInstruction = 0x0006,
    Continue = 0x0007,
    Pause = 0x0008,
    GetVRAM = 0x0009,
    GetThreads = 0x000A,

    // Responses (host -> debugger)
    StateUpdate = 0x1001,
    MemoryData = 0x1002,
    BreakpointHit = 0x1003,
    ThreadsInfo = 0x1004,
    VRAMData = 0x1005,

    // Events
    Reset = 0x2001,
    Paused = 0x2002,
};

// remove padding in structures as we want to stay the most efficient
#pragma pack(push, 1)
struct MessageHeader {
    u16 type;       // MessageType
    u32 size;       // Payload size
    u32 sequence;   // Sequence number for matching requets/response
};

struct CPUState {
    u32 regs[32];
    u32 pc;
    u64 cycles;
    u8 current_thread;
    u8 thread_states;
};

struct MemoryRequest {
    u32 address;
    u32 size;
};

struct MemoryData {
    u32 address;
    u32 size;
    // followed by actual data
};

struct Breakpoint {
    u32 address;
    u8 thread_id;       // 0xFF for all threads
    u8 enabled;
};

struct ThreadInfo {
    u8 id;
    u8 state;
    u32 pc;
    u64 sleep_until;
    u32 regs[32];
};
#pragma pack(pop)

} // namespace vc::debug
