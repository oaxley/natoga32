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
 * @brief	Virtual Console - Generic helper functions
 */

#pragma once

// program-specific headers
#include "vc/types.h"
#include "vc/events.h"

namespace vc {

// check if a waitkey is an hardware event
inline bool isHardwareEvent(u32 waitkey)
{
    return (waitkey & HW_EVENT_FLAG) != 0;
}


} // namespace vc
