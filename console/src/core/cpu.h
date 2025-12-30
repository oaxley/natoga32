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
#include <memory>

// program-specific headers
#include "vc/types.h"

namespace vc {

class Memory;       // forward declaration

class CPU
{
public:
    CPU(Memory& mem);

    void reset();
    int step();

private:
    struct OpaqueData;
    std::unique_ptr<OpaqueData> data_{nullptr};
};


} // namespace vc

