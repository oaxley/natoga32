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
    GetConsoleState     = 0x0001,
    SetConsoleState     = 0x0002,
    ReadMemory          = 0x0003,
    WriteMemory         = 0x0004,
    ReadCSR             = 0x0005,
    WriteCSR            = 0x0006,
    GetThreadInfo       = 0x0007,
    SetBreakpoint       = 0x0008,
    ClearBreakpoint     = 0x0009,
    StepInstruction     = 0x000A,
    GetVRAM             = 0x000B,

    // Response (host -> debugger)
    ConsoleStateUpdate  = 0x1001,
    MemoryData          = 0x1002,
    CSRData             = 0x1003,
    ThreadInfo          = 0x1004,
    BreakpointHit       = 0x1005,

    // observer pattern subscription
    SubConsoleState     = 0x2001,
    SubMemory           = 0x2002,
    SubCSR              = 0x2003,
    SubThread           = 0x2004,
    SubVRAM             = 0x2005,

    // Events
    Reset               = 0x3001,
    Pause               = 0x3002,
    Resume              = 0x3003
};

// remove padding in structures as we want to stay the most efficient
#pragma pack(push, 1)
struct MessageHeader {
    u16 type;       // MessageType
    u32 size;       // Payload size
    u32 sequence;   // Sequence number for matching requets/response
};

//----- ConsoleState (0x0001 / 0x0002)
// Get: request
// MessageHeader with type = 0x0001, size = 0

// Get: response
struct GetConsoleStateRep {
    u64 cycles;
    u32 regs[32];
    u32 pc;
    u32 thread_states;
    u8 current_thread;
    u8 state;
};

// Set: request
struct SetConsoleStateReq {
    u8 state;
};

//----- Memory (0x0003 / 0x0004)
// Read: request
struct MemoryReadReq {
    u32 address;
    u32 size;
};

// Read: response
struct MemoryReadRep {
    u32 address;
    u32 size;
    // actual data
};

// Write: request
struct MemoryWriteReq {
    u32 address;
    u32 size;
    // data
};

//----- CSR (0x0005 / 0x0006)
// Read: request
struct CSRReadReq {
    u16 reg;
};

// Read: response
struct CSRReadRep {
    u16 reg;
    u32 value;
};

// Write: request
struct CSRWriteReq {
    u16 reg;
    u32 value;
};

//----- ThreadInfo (0x0007)
// ThreadInfo: request
struct GetThreadInfoReq {
    u8 id;
};

// ThreadInfo: response
struct GetThreadInfoRep {
    u8 id;
    u8 state;
    u32 pc;
    u32 regs[32];
    u32 waitkey;
    u64 sleep_until;
    u64 cycles;
};

//----- Breakpoint (0x0008/0x0009/0x000A)
// set
struct BreakpointReq {
    u8 thread_id;
    u8 enabled;
    u32 address;
};


#pragma pack(pop)

} // namespace vc::debug
