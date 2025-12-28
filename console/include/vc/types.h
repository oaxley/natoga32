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
 * @brief	Virtual Console types definition
 */
#pragma once

// standard library headers
#include <cstdlib>


namespace vc {

// basic types
using u8 = u_int8_t;
using u16 = u_int16_t;
using u32 = u_int32_t;
using u64 = u_int64_t;

using i8 = int8_t;
using i16 = int16_t;
using i32 = int32_t;
using i64 = int64_t;

// memory size constants
constexpr size_t RAM_SIZE = 32 * 1024 * 1024;       // 32 MiB
constexpr size_t DSP_RAM_SIZE = 64 * 1024;          // 64 KiB

} // namespace vc
