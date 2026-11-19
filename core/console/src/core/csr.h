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
 * @brief	Control and Status Registers definition
 */

// program specific includes
#include "vc/types.h"

namespace vc
{

// trap handling CSRs
constexpr u16 CSR_MTVEC    = 0x305;      // Trap vector base address
constexpr u16 CSR_MEPC     = 0x341;      // Exception program counter
constexpr u16 CSR_MCAUSE   = 0x342;      // Trap cause
constexpr u16 CSR_MTVAL    = 0x343;      // Trap value

// status and control
constexpr u16 CSR_MSTATUS  = 0x300;      // machine status
constexpr u16 CSR_MIE      = 0x304;      // interrupt enable
constexpr u16 CSR_MIP      = 0x344;      // interrupt pending

// scratch register
constexpr u16 CSR_MSCRATCH = 0x340;      // scratch register

// counters (read-only)
constexpr u16 CSR_CYCLE    = 0xC00;      // CPU global cycle counter, low 32 bits
constexpr u16 CSR_TIME     = 0xC01;      // virtual time, low 32 bits
constexpr u16 CSR_INSTRET  = 0xC02;      // completed instructions of the current thread, low 32 bits
constexpr u16 CSR_CYCLEH   = 0xC80;      // high 32 bits of cycle
constexpr u16 CSR_TIMEH    = 0xC81;      // high 32 bits of time
constexpr u16 CSR_INSTRETH = 0xC82;      // high 32 bits of instret

// cartridge (per thread, stored in the Thread Control Block)
constexpr u16 CSR_CARTBANK = 0x7C0;      // cartridge bank register, 16 bits


} // namespace vc
