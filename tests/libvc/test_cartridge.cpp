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
 * @brief	Tests - Cartridge validation
 */

#include <catch2/catch_all.hpp>
#include <vector>

#include "cartridge_builder.h"
#include "core/cartridge.h"

using namespace vc;

namespace
{

CartError checkWithFlags(u16 flags) {
    test::CartBuilder b;
    b.headerFlags = flags;
    return checkCartridge(b.build());
}

/// Builder with an RSCV chunk followed by a TILS chunk
test::CartBuilder twoChunks() {
    test::CartBuilder b;
    b.chunks.push_back({ "TILS", std::vector<u8>(8, 0), 0 });
    return b;
}

/// Cartridge with `count` empty TILS chunks
CartError checkWithChunkCount(size_t count) {
    test::CartBuilder b;
    b.chunks.assign(count, { "TILS", {}, 0 });
    return checkCartridge(b.build());
}

/// Cartridge with an extra chunk of the given type after RSCV
CartError checkWithChunkType(const char* type) {
    test::CartBuilder b;
    b.chunks.push_back({ type, std::vector<u8>(4, 0), 0 });
    return checkCartridge(b.build());
}

/// File offset of the data of the INDX chunk
u32 indxData(const std::vector<u8>& file) {
    return test::chunkOffset(file, 1) + 8;
}

}

TEST_CASE("Cartridge header checks", "[cartridge]")
{
    SECTION("valid default cartridge") {
        CHECK(checkCartridge(test::CartBuilder{}.build()) == CartError::None);
    }

    SECTION("magic") {
        auto file = test::CartBuilder{}.build();
        CHECK(checkCartridge(std::span<const u8>(file.data(), 15)) == CartError::BadMagic);
        CHECK(checkCartridge(std::span<const u8>()) == CartError::BadMagic);
        for (u32 i = 0; i < 4; ++i) {
            auto bad = file;
            bad[i] ^= 0x01;
            CHECK(checkCartridge(bad) == CartError::BadMagic);
        }
    }

    SECTION("format version") {
        auto file = test::CartBuilder{}.build();
        file[4] = 0x11;
        test::fixHeaderCrc(file);
        CHECK(checkCartridge(file) == CartError::BadVersion);
    }

    SECTION("header CRC") {
        auto file = test::CartBuilder{}.build();
        file[5] = 0x01;
        CHECK(checkCartridge(file) == CartError::HeaderCrc);
    }

    SECTION("flags") {
        CHECK(checkWithFlags(0x0000) == CartError::None);
        CHECK(checkWithFlags(0x0010) == CartError::BadFlags);
        CHECK(checkWithFlags(0x8000) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0002) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0008 | 4) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0008 | 7) == CartError::BadFlags);
        CHECK(checkWithFlags(0x0008 | 3) == CartError::None);
        CHECK(checkWithFlags(0x0008) == CartError::None);
    }

    SECTION("total size") {
        auto file = test::CartBuilder{}.build();

        auto bigger = file;
        test::put32(bigger, 8, static_cast<u32>(file.size()) + 4);
        test::fixHeaderCrc(bigger);
        CHECK(checkCartridge(bigger) == CartError::BadSize);

        auto huge = file;
        test::put32(huge, 8, 0xFFFF'FFFF);
        test::fixHeaderCrc(huge);
        CHECK(checkCartridge(huge) == CartError::BadSize);

        // the size matches, so the header is fine (the trailing bytes fail the IEND check)
        auto extra = file;
        extra.insert(extra.end(), 4, 0);
        test::fixFileSize(extra);
        CHECK(checkCartridge(extra) == CartError::BadIend);
    }
}

TEST_CASE("Cartridge chunk checks", "[cartridge]")
{
    SECTION("IHDR") {
        const auto valid = test::CartBuilder{}.build();

        auto renamed = valid;
        test::patchChunkType(renamed, 0, "RSCV");
        CHECK(checkCartridge(renamed) == CartError::BadIhdr);

        auto length = valid;
        test::put32(length, test::chunkOffset(length, 0), 60);
        CHECK(checkCartridge(length) == CartError::BadIhdr);

        auto crc = valid;
        test::breakChunkCrc(crc, 0);
        CHECK(checkCartridge(crc) == CartError::IhdrCrc);

        for (u8 regions : { 0x00, 0x08, 0x80, 0x0F }) {
            auto bad = valid;
            test::patchChunkByte(bad, 0, 0x04, regions);
            CHECK(checkCartridge(bad) == CartError::BadIhdr);
        }
        for (u8 regions : { 0x01, 0x02, 0x04, 0x07 }) {
            auto ok = valid;
            test::patchChunkByte(ok, 0, 0x04, regions);
            CHECK(checkCartridge(ok) == CartError::None);
        }

        // save RAM size and the reserved bytes
        for (u32 off : { 0x05u, 0x14u, 0x15u, 0x16u, 0x17u, 0x38u, 0x3Fu }) {
            auto bad = valid;
            test::patchChunkByte(bad, 0, off, 1);
            CHECK(checkCartridge(bad) == CartError::BadIhdr);
        }

        // RISC-V entry point must be inside .text
        for (u32 entry : { 0x0140'0000u, 0x0200'0000u, 0x019F'FFFFu, 0u }) {
            auto bad = valid;
            test::patchChunk32(bad, 0, 0x0C, entry);
            CHECK(checkCartridge(bad) == CartError::BadIhdr);
        }
        for (u32 entry : { 0x01A0'0000u, 0x01FF'FFFFu }) {
            auto ok = valid;
            test::patchChunk32(ok, 0, 0x0C, entry);
            CHECK(checkCartridge(ok) == CartError::None);
        }

        // DSP entry point: any value
        for (u32 entry : { 0u, 1u, 0xFFFF'FFFFu }) {
            auto ok = valid;
            test::patchChunk32(ok, 0, 0x10, entry);
            CHECK(checkCartridge(ok) == CartError::None);
        }
    }

    SECTION("INDX") {
        const auto valid = twoChunks().build();

        auto swapped = valid;
        test::swapChunks(swapped, 1);
        CHECK(checkCartridge(swapped) == CartError::BadIndx);

        auto length = valid;
        test::put32(length, test::chunkOffset(length, 1), 20);
        CHECK(checkCartridge(length) == CartError::BadIndx);

        auto crc = valid;
        test::breakChunkCrc(crc, 1);
        CHECK(checkCartridge(crc) == CartError::IndxCrc);

        CHECK(checkWithChunkCount(vc::CART_INDX_MAX_ENTRIES) == CartError::None);
        CHECK(checkWithChunkCount(vc::CART_INDX_MAX_ENTRIES + 1) == CartError::IndxTooBig);
    }

    SECTION("chunk chain") {
        const auto valid = twoChunks().build();

        auto no_iend = valid;
        no_iend.resize(no_iend.size() - 12);
        test::fixFileSize(no_iend);
        CHECK(checkCartridge(no_iend) == CartError::BadIend);

        auto iend_data = valid;
        iend_data.insert(iend_data.end() - 4, 4, 0);
        const u32 iend = test::chunkOffset(iend_data, 4);
        test::put32(iend_data, iend, 4);
        test::fixChunkCrc(iend_data, iend);
        test::fixFileSize(iend_data);
        CHECK(checkCartridge(iend_data) == CartError::BadIend);

        auto trailing = valid;
        trailing.insert(trailing.end(), 4, 0);
        test::fixFileSize(trailing);
        CHECK(checkCartridge(trailing) == CartError::BadIend);

        auto not_last = valid;
        test::swapChunks(not_last, 3);
        CHECK(checkCartridge(not_last) == CartError::BadIend);

        // a huge length must not read outside the file
        for (int chunk : { 0, 1, 2, 3 }) {
            auto huge = valid;
            test::put32(huge, test::chunkOffset(huge, chunk), 0xFFFF'FFFF);
            CHECK(checkCartridge(huge) != CartError::None);
        }
        auto huge = valid;
        test::put32(huge, test::chunkOffset(huge, 2), 0xFFFF'FFFF);
        CHECK(checkCartridge(huge) == CartError::BadIend);
        test::put32(huge, test::chunkOffset(huge, 2), 0xFFFF'FFFD);
        CHECK(checkCartridge(huge) == CartError::BadIend);

        // file cut in the middle of a chunk, or between chunks
        for (u32 cut : { 1u, 4u, 8u, 10u, 11u, 20u, 27u }) {
            auto bad = valid;
            bad.resize(test::chunkOffset(bad, 2) + cut);
            test::fixFileSize(bad);
            CHECK(checkCartridge(bad) == CartError::BadIend);
        }
        auto cut_indx = valid;
        cut_indx.resize(test::chunkOffset(cut_indx, 2));
        test::fixFileSize(cut_indx);
        CHECK(checkCartridge(cut_indx) == CartError::BadIend);

        // not even a complete IHDR or INDX
        auto cut_ihdr = valid;
        cut_ihdr.resize(20);
        test::fixFileSize(cut_ihdr);
        CHECK(checkCartridge(cut_ihdr) == CartError::BadIhdr);
        auto header_only = valid;
        header_only.resize(vc::CART_HEADER_SIZE);
        test::fixFileSize(header_only);
        CHECK(checkCartridge(header_only) == CartError::BadIhdr);
        auto cut_indx_header = valid;
        cut_indx_header.resize(test::chunkOffset(cut_indx_header, 1) + 4);
        test::fixFileSize(cut_indx_header);
        CHECK(checkCartridge(cut_indx_header) == CartError::BadIndx);
    }

    SECTION("chunk names") {
        CHECK(checkWithChunkType("TIL5") == CartError::BadChunk);
        CHECK(checkWithChunkType("TIls") == CartError::BadChunk);
        CHECK(checkWithChunkType("TI S") == CartError::BadChunk);
        CHECK(checkWithChunkType("ABCD") == CartError::BadChunk);
        CHECK(checkWithChunkType("abCd") == CartError::BadChunk);
        CHECK(checkWithChunkType("abCD") == CartError::None);
        CHECK(checkWithChunkType("dSYM") == CartError::None);
        CHECK(checkWithChunkType("TILS") == CartError::None);
        CHECK(checkWithChunkType("IHDR") == CartError::BadIhdr);
        CHECK(checkWithChunkType("INDX") == CartError::BadIndx);

        // every known critical name is accepted
        for (const char* type : { "RSCV", "DSPC", "TILS", "TMAP", "TPAL", "TMAT", "SPRS", "SOAM", "SPAL", "SMAT",
                                  "FNTS", "FPAL", "ASMD", "ATMD", "ATFM", "AIFM" }) {
            CHECK(checkWithChunkType(type) == CartError::None);
        }
    }

    SECTION("INDX consistency") {
        const auto valid = twoChunks().build();
        const u32 entry1 = indxData(valid) + vc::CART_INDX_ENTRY_SIZE;

        auto type = valid;
        type[entry1] = 'M';
        test::fixChunkCrc(type, test::chunkOffset(type, 1));
        CHECK(checkCartridge(type) == CartError::IndxMismatch);

        auto id = valid;
        id[entry1 + 4] = 0;
        test::fixChunkCrc(id, test::chunkOffset(id, 1));
        CHECK(checkCartridge(id) == CartError::IndxMismatch);

        auto flags = valid;
        flags[entry1 + 7] = 0x01;
        test::fixChunkCrc(flags, test::chunkOffset(flags, 1));
        CHECK(checkCartridge(flags) == CartError::IndxMismatch);

        auto boot = valid;
        boot[entry1 + 6] = 0x01;
        test::fixChunkCrc(boot, test::chunkOffset(boot, 1));
        CHECK(checkCartridge(boot) == CartError::None);

        auto offset = valid;
        test::put32(offset, entry1 + 8, test::get32(offset, entry1 + 8) + 4);
        test::fixChunkCrc(offset, test::chunkOffset(offset, 1));
        CHECK(checkCartridge(offset) == CartError::IndxMismatch);

        auto length = valid;
        test::put32(length, entry1 + 12, test::get32(length, entry1 + 12) + 1);
        test::fixChunkCrc(length, test::chunkOffset(length, 1));
        CHECK(checkCartridge(length) == CartError::IndxMismatch);

        auto fewer = twoChunks();
        fewer.unlistedChunks = 1;
        CHECK(checkCartridge(fewer.build()) == CartError::IndxMismatch);

        // one entry more than the chunks: drop the TILS chunk but keep its entry
        auto more = valid;
        const u32 tils = test::chunkOffset(more, 3);
        more.erase(more.begin() + tils, more.begin() + test::chunkOffset(more, 4));
        test::fixFileSize(more);
        CHECK(checkCartridge(more) == CartError::IndxMismatch);
    }

    SECTION("other chunk CRC is not checked") {
        auto file = test::CartBuilder{}.build();
        test::breakChunkCrc(file, 2);
        CHECK(checkCartridge(file) == CartError::None);
    }
}
