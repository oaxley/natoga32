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
 * @brief	Virtual Console - Memory header
 */
#pragma once

// standard library headers
#include <vector>
#include <span>
#include <memory>

// program-specific headers
#include "vc/types.h"


namespace vc {

class Memory
{
public:
    Memory();

    void reset();

    // read operations
    u8  read8(u32 addr) const;
    u16 read16(u32 addr) const;
    u32 read32(u32 addr) const;

    // write operations
    void write8(u32 addr, u8 value);
    void write16(u32 addr, u16 value);
    void write32(u32 addr, u32 value);

    // return a RW view on a specific memory area
    template<typename T>
    std::span<T> memview(size_t offset, size_t count) {
        return { reinterpret_cast<T*>(ram_.data() + offset), count };
    }

private:
    std::vector<u8> ram_;
};

} // namespace vc
