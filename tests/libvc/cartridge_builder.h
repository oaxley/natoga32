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
 * @brief	Tests - Cartridge file builder helper
 */
#pragma once

// standard library headers
#include <algorithm>
#include <array>
#include <cstring>
#include <span>
#include <stdexcept>
#include <string>
#include <vector>

// program-specific headers
#include "vc/types.h"
#include "core/cartridge.h"
#include "core/crc32.h"


namespace test {

using vc::u8;
using vc::u16;
using vc::u32;

//----- little-endian helpers

inline void put16(std::vector<u8>& d, u32 off, u16 v) {
    d[off] = static_cast<u8>(v);
    d[off + 1] = static_cast<u8>(v >> 8);
}

inline void put32(std::vector<u8>& d, u32 off, u32 v) {
    for (u32 i = 0; i < 4; ++i) {
        d[off + i] = static_cast<u8>(v >> (8 * i));
    }
}

inline u32 get32(const std::vector<u8>& d, u32 off) {
    return static_cast<u32>(d[off]) | (static_cast<u32>(d[off + 1]) << 8) |
           (static_cast<u32>(d[off + 2]) << 16) | (static_cast<u32>(d[off + 3]) << 24);
}

inline u32 align4(u32 v) {
    return (v + 3) & ~3u;
}

//----- file fix-up functions

/// Recompute the header CRC (bytes 12-15) after a test modified bytes 0-11
inline void fixHeaderCrc(std::vector<u8>& file) {
    put32(file, 12, vc::crc32(std::span<const u8>(file.data(), 12)));
}

/// Recompute the CRC of the chunk starting at `chunk_offset` (its length field)
inline void fixChunkCrc(std::vector<u8>& file, u32 chunk_offset) {
    const u32 length = get32(file, chunk_offset);
    const u32 crc = vc::crc32(std::span<const u8>(file.data() + chunk_offset + 4, 4 + length));
    put32(file, chunk_offset + 8 + align4(length), crc);
}

/// File offset of the index-th chunk (IHDR = 0, INDX = 1...), i.e. of its length field
inline u32 chunkOffset(const std::vector<u8>& file, int index) {
    u32 off = vc::CART_HEADER_SIZE;
    for (int i = 0; i < index; ++i) {
        off += 12 + align4(get32(file, off));
    }
    return off;
}

/// Recompute the total size field and the header CRC after a test changed the file length
inline void fixFileSize(std::vector<u8>& file) {
    put32(file, 8, static_cast<u32>(file.size()));
    fixHeaderCrc(file);
}

/// Overwrite one byte of the data of the index-th chunk and fix the chunk CRC
inline void patchChunkByte(std::vector<u8>& file, int index, u32 data_offset, u8 value) {
    const u32 chunk = chunkOffset(file, index);
    file[chunk + 8 + data_offset] = value;
    fixChunkCrc(file, chunk);
}

/// Overwrite 4 bytes (little-endian) of the data of the index-th chunk and fix the chunk CRC
inline void patchChunk32(std::vector<u8>& file, int index, u32 data_offset, u32 value) {
    const u32 chunk = chunkOffset(file, index);
    put32(file, chunk + 8 + data_offset, value);
    fixChunkCrc(file, chunk);
}

/// Overwrite the type of the index-th chunk and fix the chunk CRC
inline void patchChunkType(std::vector<u8>& file, int index, const char (&type)[5]) {
    const u32 chunk = chunkOffset(file, index);
    std::memcpy(&file[chunk + 4], type, 4);
    fixChunkCrc(file, chunk);
}

/// Flip a bit of the stored CRC of the index-th chunk
inline void breakChunkCrc(std::vector<u8>& file, int index) {
    const u32 chunk = chunkOffset(file, index);
    file[chunk + 8 + align4(get32(file, chunk))] ^= 0x01;
}

/// Swap the index-th chunk with the next one (the size and the CRCs do not change)
inline void swapChunks(std::vector<u8>& file, int index) {
    const u32 first = chunkOffset(file, index);
    const u32 second = first + 12 + align4(get32(file, first));
    const u32 end = second + 12 + align4(get32(file, second));
    std::rotate(file.begin() + first, file.begin() + second, file.begin() + end);
}

//----- builder

struct CartChunk {
    std::string type;
    std::vector<u8> data;
    u16 flags = 0;
};

class CartBuilder {
public:
    u16 headerFlags = 0;
    std::array<u8, 64> ihdr{};
    std::vector<CartChunk> chunks;
    size_t unlistedChunks = 0;      // number of chunks, at the end of `chunks`, left out of INDX

    CartBuilder() {
        ihdr[0x04] = 0x07;
        put32le(0x0C, 0x01A0'0000);
        std::memcpy(&ihdr[0x18], "TEST", 4);
        chunks.push_back({ "RSCV", std::vector<u8>(16, 0), 1 });
    }

    std::vector<u8> build() const {
        std::vector<u8> file(vc::CART_HEADER_SIZE, 0);

        // header: size and CRC are filled in at the end
        static constexpr u8 magic[4] = { 0x7F, 'N', '3', '2' };
        std::memcpy(file.data(), magic, sizeof(magic));
        file[4] = vc::CART_FORMAT_VERSION;
        put16(file, 6, headerFlags);

        appendChunk(file, "IHDR", ihdr);

        // chunk data starts 8 bytes after the chunk, which begins after IHDR and INDX
        const size_t listed = chunks.size() - unlistedChunks;
        u32 offset = static_cast<u32>(file.size()) + 12 + static_cast<u32>(listed) * vc::CART_INDX_ENTRY_SIZE;
        std::vector<u8> index;
        for (size_t i = 0; i < listed; ++i) {
            const CartChunk& c = chunks[i];
            const u32 length = static_cast<u32>(c.data.size());
            const size_t at = index.size();
            index.resize(at + vc::CART_INDX_ENTRY_SIZE, 0);
            copyType(&index[at], c.type);
            put16(index, static_cast<u32>(at + 4), static_cast<u16>(i));
            put16(index, static_cast<u32>(at + 6), c.flags);
            put32(index, static_cast<u32>(at + 8), offset + 8);
            put32(index, static_cast<u32>(at + 12), length);
            offset += 12 + align4(length);
        }
        appendChunk(file, "INDX", index);

        for (const CartChunk& c : chunks) {
            appendChunk(file, c.type, c.data);
        }
        appendChunk(file, "IEND", {});

        put32(file, 8, static_cast<u32>(file.size()));
        fixHeaderCrc(file);
        return file;
    }

private:
    void put32le(u32 off, u32 v) {
        for (u32 i = 0; i < 4; ++i) {
            ihdr[off + i] = static_cast<u8>(v >> (8 * i));
        }
    }

    static void copyType(u8* dst, const std::string& type) {
        if (type.size() != 4) {
            throw std::invalid_argument("chunk type must be 4 characters: " + type);
        }
        std::memcpy(dst, type.data(), 4);
    }

    static void appendChunk(std::vector<u8>& file, const std::string& type, std::span<const u8> data) {
        const u32 start = static_cast<u32>(file.size());
        const u32 length = static_cast<u32>(data.size());
        file.resize(start + 12 + align4(length), 0);
        put32(file, start, length);
        copyType(&file[start + 4], type);
        if (length != 0) {
            std::memcpy(&file[start + 8], data.data(), length);
        }
        fixChunkCrc(file, start);
    }
};

}
