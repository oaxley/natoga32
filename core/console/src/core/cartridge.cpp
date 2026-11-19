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
 * @brief	Virtual Console - Cartridge validation implementation
 */
// standard library headers
#include <algorithm>
#include <cstring>
#include <iterator>

// program-specific headers
#include "cartridge.h"
#include "crc32.h"
#include "memory.h"


namespace vc {

//----- helpers

namespace {

constexpr u8 CART_MAGIC[4] = { 0x7F, 'N', '3', '2' };

constexpr u32 HEADER_OFF_VERSION = 4;
constexpr u32 HEADER_OFF_FLAGS = 6;
constexpr u32 HEADER_OFF_SIZE = 8;
constexpr u32 HEADER_OFF_CRC = 12;

constexpr u16 FLAGS_MODE_MASK = 0x0007;
constexpr u16 FLAGS_MODE_REQUIRED = 0x0008;
constexpr u16 FLAGS_RESERVED_MASK = 0xFFF0;
constexpr u16 FLAGS_MODE_MAX = 3;

constexpr u32 CHUNK_HEADER_SIZE = 8;
constexpr u32 CHUNK_OVERHEAD = 12;

constexpr u32 IHDR_OFF_REGIONS = 0x04;
constexpr u32 IHDR_OFF_RESERVED_1 = 0x05;
constexpr u32 IHDR_OFF_RISCV_ENTRY = 0x0C;
constexpr u32 IHDR_OFF_SAVE_RAM = 0x14;
constexpr u32 IHDR_OFF_RESERVED_2 = 0x16;
constexpr u32 IHDR_OFF_RESERVED_3 = 0x38;
constexpr u32 IHDR_RESERVED_3_SIZE = 8;
constexpr u8 IHDR_REGIONS_MASK = 0x07;

constexpr u32 CART_BANK_SHIFT = 12;

constexpr u32 IHDR_DATA_OFFSET = CART_HEADER_SIZE + CHUNK_HEADER_SIZE;
constexpr u32 INDX_CHUNK_OFFSET = CART_HEADER_SIZE + CHUNK_OVERHEAD + CART_IHDR_SIZE;
constexpr u32 INDX_DATA_OFFSET = INDX_CHUNK_OFFSET + CHUNK_HEADER_SIZE;

constexpr u32 INDX_ENTRY_OFF_ID = 4;
constexpr u32 INDX_ENTRY_OFF_FLAGS = 6;
constexpr u32 INDX_ENTRY_OFF_DATA = 8;
constexpr u32 INDX_ENTRY_OFF_LENGTH = 12;
constexpr u16 INDX_FLAG_RESERVED_MASK = 0xFFFE;

u16 read16(std::span<const u8> d, u64 off) {
    return static_cast<u16>(d[off] | (d[off + 1] << 8));
}

u32 read32(std::span<const u8> d, u64 off) {
    return static_cast<u32>(d[off]) | (static_cast<u32>(d[off + 1]) << 8) |
           (static_cast<u32>(d[off + 2]) << 16) | (static_cast<u32>(d[off + 3]) << 24);
}

/// Chunk type as read by read32() from the file (first letter in the lowest byte)
constexpr u32 fourcc(const char (&s)[5]) {
    return static_cast<u32>(static_cast<u8>(s[0])) | (static_cast<u32>(static_cast<u8>(s[1])) << 8) |
           (static_cast<u32>(static_cast<u8>(s[2])) << 16) | (static_cast<u32>(static_cast<u8>(s[3])) << 24);
}

constexpr u32 TYPE_IHDR = fourcc("IHDR");
constexpr u32 TYPE_INDX = fourcc("INDX");
constexpr u32 TYPE_IEND = fourcc("IEND");

constexpr u32 KNOWN_CRITICAL[] = {
    TYPE_IHDR, TYPE_INDX, TYPE_IEND, fourcc("RSCV"), fourcc("DSPC"), fourcc("TILS"), fourcc("TMAP"),
    fourcc("TPAL"), fourcc("TMAT"), fourcc("SPRS"), fourcc("SOAM"), fourcc("SPAL"), fourcc("SMAT"),
    fourcc("FNTS"), fourcc("FPAL"), fourcc("ASMD"), fourcc("ATMD"), fourcc("ATFM"), fourcc("AIFM")
};

constexpr u64 align4(u64 v) {
    return (v + 3) & ~3ull;
}

constexpr bool isLetter(u8 c) {
    return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z');
}

constexpr bool isUpper(u8 c) {
    return c >= 'A' && c <= 'Z';
}

/// Valid name, and known if it is critical (an unknown ancillary chunk is skipped)
bool isAcceptedChunkType(u32 type) {
    const u8 c0 = static_cast<u8>(type);
    const u8 c1 = static_cast<u8>(type >> 8);
    const u8 c2 = static_cast<u8>(type >> 16);
    const u8 c3 = static_cast<u8>(type >> 24);
    if (!isLetter(c0) || !isLetter(c1) || !isUpper(c2) || !isUpper(c3)) {
        return false;
    }
    if (!isUpper(c0)) {
        return true;
    }
    return std::find(std::begin(KNOWN_CRITICAL), std::end(KNOWN_CRITICAL), type) != std::end(KNOWN_CRITICAL);
}

/// CRC of the chunk at `off` (type + data) against the stored value; the chunk must fit in the file
bool chunkCrcMatches(std::span<const u8> file, u64 off, u32 length) {
    const u32 crc = crc32(file.subspan(off + 4, 4 + static_cast<u64>(length)));
    return crc == read32(file, off + CHUNK_HEADER_SIZE + align4(length));
}

bool ihdrFieldsValid(std::span<const u8> file, u64 data) {
    const u8 regions = file[data + IHDR_OFF_REGIONS];
    if (regions == 0 || (regions & ~IHDR_REGIONS_MASK) != 0 || file[data + IHDR_OFF_RESERVED_1] != 0) {
        return false;
    }

    const u32 entry = read32(file, data + IHDR_OFF_RISCV_ENTRY);
    if (entry < MMAP_CODE || entry - MMAP_CODE >= MMAP_CODE_SIZE) {
        return false;
    }

    if (read16(file, data + IHDR_OFF_SAVE_RAM) != 0 || read16(file, data + IHDR_OFF_RESERVED_2) != 0) {
        return false;
    }
    for (u32 i = 0; i < IHDR_RESERVED_3_SIZE; ++i) {
        if (file[data + IHDR_OFF_RESERVED_3 + i] != 0) {
            return false;
        }
    }
    return true;
}

/// Does INDX entry `entry` (file offset) describe the chunk with this type, position, data offset and length?
bool indxEntryMatches(std::span<const u8> file, u64 entry, u32 type, u32 id, u64 data, u32 length) {
    return read32(file, entry) == type && read16(file, entry + INDX_ENTRY_OFF_ID) == id &&
           (read16(file, entry + INDX_ENTRY_OFF_FLAGS) & INDX_FLAG_RESERVED_MASK) == 0 &&
           read32(file, entry + INDX_ENTRY_OFF_DATA) == data && read32(file, entry + INDX_ENTRY_OFF_LENGTH) == length;
}

}

//----- public functions

CartError checkCartridge(std::span<const u8> file) {
    if (file.size() < CART_HEADER_SIZE || std::memcmp(file.data(), CART_MAGIC, sizeof(CART_MAGIC)) != 0) {
        return CartError::BadMagic;
    }

    if (file[HEADER_OFF_VERSION] != CART_FORMAT_VERSION) {
        return CartError::BadVersion;
    }

    if (crc32(file.first(HEADER_OFF_CRC)) != read32(file, HEADER_OFF_CRC)) {
        return CartError::HeaderCrc;
    }

    const u16 flags = read16(file, HEADER_OFF_FLAGS);
    const u16 mode = flags & FLAGS_MODE_MASK;
    if ((flags & FLAGS_RESERVED_MASK) != 0 || mode > FLAGS_MODE_MAX ||
        (mode != 0 && (flags & FLAGS_MODE_REQUIRED) == 0)) {
        return CartError::BadFlags;
    }

    const u32 size = read32(file, HEADER_OFF_SIZE);
    if (size != file.size() || size > CART_MAX_SIZE) {
        return CartError::BadSize;
    }

    // chunk walk: IHDR, INDX, the other chunks up to IEND; the INDX comparison comes last
    const u64 total = file.size();
    u64 off = CART_HEADER_SIZE;
    u64 indx_data = 0;
    u32 indx_count = 0;
    u32 listed = 0;
    bool mismatch = false;

    for (u32 position = 0;; ++position) {
        if (off + CHUNK_HEADER_SIZE > total) {
            return position == 0 ? CartError::BadIhdr : position == 1 ? CartError::BadIndx : CartError::BadIend;
        }
        const u32 length = read32(file, off);
        const u32 type = read32(file, off + 4);
        const u64 data = off + CHUNK_HEADER_SIZE;
        const u64 end = off + CHUNK_OVERHEAD + align4(length);

        if (position == 0) {
            if (type != TYPE_IHDR || length != CART_IHDR_SIZE) {
                return CartError::BadIhdr;
            }
            if (end > total) {
                return CartError::BadIend;
            }
            if (!ihdrFieldsValid(file, data)) {
                return CartError::BadIhdr;
            }
            if (!chunkCrcMatches(file, off, length)) {
                return CartError::IhdrCrc;
            }
        } else if (position == 1) {
            if (type != TYPE_INDX || length % CART_INDX_ENTRY_SIZE != 0) {
                return CartError::BadIndx;
            }
            if (end > total) {
                return CartError::BadIend;
            }
            if (!chunkCrcMatches(file, off, length)) {
                return CartError::IndxCrc;
            }
            indx_data = data;
            indx_count = length / CART_INDX_ENTRY_SIZE;
            if (indx_count > CART_INDX_MAX_ENTRIES) {
                return CartError::IndxTooBig;
            }
        } else {
            if (end > total) {
                return CartError::BadIend;
            }
            if (type == TYPE_IHDR) {
                return CartError::BadIhdr;
            }
            if (type == TYPE_INDX) {
                return CartError::BadIndx;
            }
            if (!isAcceptedChunkType(type)) {
                return CartError::BadChunk;
            }
            if (type == TYPE_IEND) {
                if (length != 0 || end != total) {
                    return CartError::BadIend;
                }
                return (mismatch || listed != indx_count) ? CartError::IndxMismatch : CartError::None;
            }
            mismatch = mismatch || listed >= indx_count ||
                       !indxEntryMatches(file, indx_data + static_cast<u64>(listed) * CART_INDX_ENTRY_SIZE,
                                         type, listed, data, length);
            ++listed;
        }
        off = end;
    }
}

//----- Cartridge

void Cartridge::insert(std::span<const u8> file) {
    file_.assign(file.begin(), file.end());
    present_ = true;
}

void Cartridge::insert(std::vector<u8>&& file) {
    file_ = std::move(file);
    present_ = true;
}

void Cartridge::eject() {
    file_.clear();
    present_ = false;
}

bool Cartridge::isPresent() const {
    return present_;
}

size_t Cartridge::size() const {
    return file_.size();
}

void Cartridge::map(Memory& mem) {
    valid_ = false;
    if (!present_) {
        return;
    }

    const CartError error = checkCartridge(file_);
    const u32 status = CART_STATUS_PRESENT | (error == CartError::None ? CART_STATUS_VALID : static_cast<u32>(error) << 8);
    const std::span<const u8> file = file_;

    auto write32 = [&mem](u32 addr, u32 value) {
        const auto dst = mem.memview<u8>(addr, 4);
        for (u32 i = 0; i < 4; ++i) {
            dst[i] = static_cast<u8>(value >> (8 * i));
        }
    };

    write32(CART_STATUS, status);
    write32(CART_SIZE, static_cast<u32>(file.size()));
    if (error != CartError::None) {
        return;
    }
    valid_ = true;

    // the check guarantees IHDR first, then INDX, right after the header
    const u32 indx_length = read32(file, INDX_CHUNK_OFFSET);
    write32(CART_INDX_COUNT, indx_length / CART_INDX_ENTRY_SIZE);

    std::copy_n(file.begin(), CART_HEADER_SIZE, mem.memview<u8>(CART_HEADER_MIRROR, CART_HEADER_SIZE).begin());
    std::copy_n(file.begin() + IHDR_DATA_OFFSET, CART_IHDR_SIZE, mem.memview<u8>(CART_IHDR_MIRROR, CART_IHDR_SIZE).begin());
    std::copy_n(file.begin() + INDX_DATA_OFFSET, indx_length, mem.memview<u8>(CART_INDX_MIRROR, indx_length).begin());
}

bool Cartridge::isValid() const {
    return valid_;
}

u8 Cartridge::read8(u32 offset) const {
    return offset < file_.size() ? file_[offset] : 0;
}

u8 Cartridge::readWindow(u16 bank, u32 window_offset) const {
    if (!valid_) {
        return 0;
    }
    const u64 offset = (static_cast<u64>(bank) << CART_BANK_SHIFT) + window_offset;
    return offset < file_.size() ? file_[offset] : 0;
}

}
