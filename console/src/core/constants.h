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
 * @brief	Virtual Console - Constants
 */
#pragma once

// standard library includes
#include <cstring>

// program-specific includes
#include "vc/types.h"

namespace vc
{

// memory map segments
constexpr u32 MMAP_ROM_BOOT_LOADER   = 0x0000'0000;
constexpr u32 MMAP_INTERRUPT_VECTOR  = 0x0020'0000;
constexpr u32 MMAP_CSR_REGISTERS     = 0x0020'1000;
constexpr u32 MMAP_THREADS_REG_BASE  = 0x0020'5000;
constexpr u32 MMAP_THREADS_TLS_BASE  = 0x0020'6000;
constexpr u32 MMAP_CARTRIDGE_IO      = 0x0022'0000;
constexpr u32 MMAP_HOST_S_RAM        = 0x0026'0000;
constexpr u32 MMAP_AUDIO_RAM         = 0x002E'0000;
constexpr u32 MMAP_VIDEO_RAM         = 0x0038'0000;
constexpr u32 MMAP_HEAP              = 0x0078'0000;
constexpr u32 MMAP_STACK_BASE        = 0x00F8'0000;
constexpr u32 MMAP_BSS               = 0x0100'0000;
constexpr u32 MMAP_DATA              = 0x0120'0000;
constexpr u32 MMAP_CODE              = 0x01A0'0000;

// segments size
constexpr size_t MMAP_ROM_SIZE       = 2 * 1024 * 1024;
constexpr size_t MMAP_IV_SIZE        = 128;
constexpr size_t MMAP_CSR_SIZE       = 16 * 1024;
constexpr size_t MMAP_TREG_SIZE      = 1024;
constexpr size_t MMAP_TTLS_SIZE      = 2 * 1024;
constexpr size_t MMAP_CARTIO_SIZE    = 256 * 1024;
constexpr size_t MMAP_SRAM_SIZE      = 512 * 1024;
constexpr size_t MMAP_ARAM_SIZE      = 512 * 1024;
constexpr size_t MMAP_VRAM_SIZE      = 4 * 1024 * 1024;
constexpr size_t MMAP_HEAP_SIZE      = 8 * 1024 * 1024;
constexpr size_t MMAP_STACK_SIZE     = 512 * 1024;
constexpr size_t MMAP_BSS_SIZE       = 2 * 1024 * 1024;
constexpr size_t MMAP_DATA_SIZE      = 8 * 1024 * 1024;
constexpr size_t MMAP_CODE_SIZE      = 6 * 1024 * 1024;
constexpr size_t MMAP_UNUSED_SIZE    = 1024 * (3 + 102 + 128) + 3968;

// total ram size
constexpr size_t MMAP_RAM_SIZE       = 32 * 1024 * 1024;

// console default address on reset
constexpr u32 RESET_DEFAULT_ADDR     = MMAP_ROM_BOOT_LOADER + 0x100;

// threads count + number of registers
constexpr int THREADS_COUNT = 8;
constexpr int THREADS_REGISTERS = 32;

// stack canary value
constexpr u32 STACK_CANARY_VALUE = 0xBAADF00D;

// RISC-V Opcodes
constexpr u32 OP_LOAD   = 0b000'0011;       // 0x03
constexpr u32 OP_STORE  = 0b010'0011;       // 0x23
constexpr u32 OP_BRANCH = 0b110'0011;       // 0x63

constexpr u32 OP_AUIPC  = 0b001'0111;       // 0x17
constexpr u32 OP_LUI    = 0b011'0111;       // 0x37
constexpr u32 OP_JALR   = 0b110'0111;       // 0x67
constexpr u32 OP_JAL    = 0b110'1111;       // 0x6F

constexpr u32 OP_OP_IMM = 0b001'0011;       // 0x13
constexpr u32 OP_OP     = 0b011'0011;       // 0x33

constexpr u32 OP_SYSTEM = 0b111'0011;       // 0x73
constexpr u32 OP_CUSTOM = 0b000'1011;       // 0x0B


} // namespace vc
