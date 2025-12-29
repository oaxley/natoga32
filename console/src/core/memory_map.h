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
 * @brief	Virtual Console - Memroy map
 */
#pragma once

// standard library includes
#include <cstring>

// program-specific includes
#include "vc/types.h"

namespace vc
{

// memory map segments
constexpr u32 ROM_BOOT_LOADER   = 0x0000'0000;
constexpr u32 INTERRUPT_VECTOR  = 0x0020'0000;
constexpr u32 CSR_REGISTERS     = 0x0020'1000;
constexpr u32 THREADS_REG_BASE  = 0x0020'5000;
constexpr u32 THREADS_META_BASE = 0x0020'6000;
constexpr u32 CARTRIDGE_IO      = 0x0022'0000;
constexpr u32 HOST_S_RAM        = 0x0026'0000;
constexpr u32 AUDIO_RAM         = 0x002E'0000;
constexpr u32 VIDEO_RAM         = 0x0038'0000;
constexpr u32 HEAP              = 0x0078'0000;
constexpr u32 STACK_TOP_BASE    = 0x00F8'0000;
constexpr u32 BSS               = 0x0100'0000;
constexpr u32 DATA              = 0x0120'0000;
constexpr u32 CODE              = 0x01A0'0000;

// segments size
constexpr size_t ROM_SIZE       = 2 * 1024 * 1024;
constexpr size_t IV_SIZE        = 128;
constexpr size_t CSR_SIZE       = 16 * 1024;
constexpr size_t TREG_SIZE      = 1024;
constexpr size_t TMETA_SIZE     = 2 * 1024;
constexpr size_t CARTIO_SIZE    = 256 * 1024;
constexpr size_t SRAM_SIZE      = 512 * 1024;
constexpr size_t ARAM_SIZE      = 512 * 1024;
constexpr size_t VRAM_SIZE      = 4 * 1024 * 1024;
constexpr size_t HEAP_SIZE      = 8 * 1024 * 1024;
constexpr size_t STACK_SIZE     = 512 * 1024;
constexpr size_t BSS_SIZE       = 2 * 1024 * 1024;
constexpr size_t DATA_SIZE      = 8 * 1024 * 1024;
constexpr size_t CODE_SIZE      = 6 * 1024 * 1024;

// total ram size
constexpr size_t RAM_SIZE       = 32 * 1024 * 1024;

} // namespace vc
