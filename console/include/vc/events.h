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
 * @brief	Virtual Console - Events list
 */

// program-specific includes
#include "vc/types.h"

namespace vc
{

/* 32-bit waitkey space
0x00000000 - 0x01FFFFFF  // Valid RAM addresses (32MB)
0x10000000 - 0x1FFFFFFF  // Software events (user-defined)
0x80000000 - 0x8FFFFFFF  // Hardware events (bit 31 set)
*/

//----- Hardware events
constexpr u32 HW_EVENT_FLAG = 0x8000'0000;
constexpr u32 HW_VBLANK     = 0x8000'0001;
constexpr u32 HW_HBLANK     = 0x8000'0002;


} // namespace vc
