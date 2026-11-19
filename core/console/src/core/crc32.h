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
 * @brief	Virtual Console - CRC-32 helper header
 */
#pragma once

// standard library headers
#include <span>

// program-specific headers
#include "vc/types.h"


namespace vc {

/// CRC-32 (IEEE 802.3 / PNG / zlib). Pass the result of a previous call as
/// `crc` to continue a computation over several spans.
u32 crc32(std::span<const u8> data, u32 crc = 0);

}
