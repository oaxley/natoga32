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
 * @brief	Tests - Boot ROM (logo fade-in, cartridge message, trap handler)
 */

#include <catch2/catch_all.hpp>
#include <filesystem>
#include <fstream>
#include <iterator>
#include <vector>

#include "console.h"
#include "core/cpu.h"
#include "core/memory.h"
#include "core/constants.h"

using namespace vc;

namespace
{

constexpr u32 LOGO_X = 40;
constexpr u32 LOGO_Y = 60;
constexpr u32 LOGO_W = 560;
constexpr u32 LOGO_H = 205;

constexpr u16 COLOR_BACKGROUND = 0x0843;
constexpr u16 COLOR_MESSAGE = 0xCE5A;
constexpr u16 COLOR_ERROR = 0xF800;
constexpr u16 COLOR_WHITE = 0xFFFF;

// read the boot ROM image built by the rombios target (empty when unavailable)
std::vector<u8> readBootRom()
{
#ifdef VC_ROMBIOS_PATH
    std::ifstream f(VC_ROMBIOS_PATH, std::ios::binary);
    if (f) {
        return std::vector<u8>(std::istreambuf_iterator<char>(f), std::istreambuf_iterator<char>());
    }
#endif
    return {};
}

// true when the message area holds at least one message-coloured pixel
bool messageVisible(const u16* fb)
{
    for (u32 y = 310; y <= 325; ++y) {
        for (u32 x = 48; x <= 591; ++x) {
            if (fb[y * VIDEO_WIDTH + x] == COLOR_MESSAGE) {
                return true;
            }
        }
    }
    return false;
}

// a text box: draw_text glyphs are 16x16 pixels (font 8x8 scaled x2)
struct TextBox
{
    u32 x, y, chars;

    bool contains(u32 px, u32 py) const
    {
        return px >= x && px < x + chars * 16 && py >= y && py < y + 16;
    }
};

// "SYSTEM ERROR" and the 8 mcause hex digits drawn by the trap handler
constexpr TextBox ERROR_TEXT = {16, 16, 12};
constexpr TextBox ERROR_CAUSE = {16, 40, 8};
constexpr u32 ERROR_BAND_ROWS = 64;

// number of white pixels inside the box
u32 whiteInBox(const u16* fb, const TextBox& box)
{
    u32 count = 0;
    for (u32 y = box.y; y < box.y + 16; ++y) {
        for (u32 x = box.x; x < box.x + box.chars * 16; ++x) {
            if (fb[y * VIDEO_WIDTH + x] == COLOR_WHITE) {
                ++count;
            }
        }
    }
    return count;
}

// number of white pixels in the error band, outside both text boxes
u32 whiteOutsideText(const u16* fb)
{
    u32 count = 0;
    for (u32 y = 0; y < ERROR_BAND_ROWS; ++y) {
        for (u32 x = 0; x < VIDEO_WIDTH; ++x) {
            if (!ERROR_TEXT.contains(x, y) && !ERROR_CAUSE.contains(x, y) &&
                fb[y * VIDEO_WIDTH + x] == COLOR_WHITE) {
                ++count;
            }
        }
    }
    return count;
}

} // namespace


TEST_CASE("Boot ROM", "[rom]") {
    auto rom = readBootRom();
    if (rom.empty()) {
        SKIP("boot ROM image not built (rombios target unavailable)");
    }

    Console c;
    REQUIRE(c.initialize());

    SECTION("boots, fades the logo in and blinks the message") {
        REQUIRE(c.loadROM(rom.data(), rom.size()));
        c.reset();
        c.resume();

        // header
        REQUIRE(c.getMemory().read32(0x00) == 0x5232334E);      // "N32R"
        REQUIRE(c.getMemory().read32(0x08) == 0x100);

        bool seen_on = false;
        bool seen_off = false;
        for (int frame = 1; frame <= 300; ++frame) {
            c.runFrame();
            if (frame >= 120) {
                if (messageVisible(c.getFramebuffer())) {
                    seen_on = true;
                } else {
                    seen_off = true;
                }
            }
        }

        REQUIRE(c.getCPU().getState() != CPUState::Halted);
        REQUIRE(c.getState() == ConsoleState::Running);

        const u16* fb = c.getFramebuffer();
        REQUIRE(fb[0] == COLOR_BACKGROUND);

        // the first solid-ink logo pixel must be fully white after the fade
        u32 logo_addr = c.getMemory().read32(0x14);
        bool found = false;
        for (u32 y = 0; y < LOGO_H && !found; ++y) {
            for (u32 x = 0; x < LOGO_W && !found; ++x) {
                u8 b = c.getMemory().read8(logo_addr + y * (LOGO_W / 2) + x / 2);
                u8 nibble = (x & 1) ? (b & 0x0F) : (b >> 4);
                if (nibble == 15) {
                    found = true;
                    REQUIRE(fb[(LOGO_Y + y) * VIDEO_WIDTH + LOGO_X + x] == 0xFFFF);
                }
            }
        }
        REQUIRE(found);

        REQUIRE(seen_on);
        REQUIRE(seen_off);
    }

    SECTION("an illegal instruction shows the system error band") {
        REQUIRE(rom.size() > 0x110);
        rom[0x10C] = rom[0x10D] = rom[0x10E] = rom[0x10F] = 0x00;

        REQUIRE(c.loadROM(rom.data(), rom.size()));
        c.reset();
        c.resume();

        c.runFrame();
        c.runFrame();

        REQUIRE(c.getCPU().getState() != CPUState::Halted);
        REQUIRE(c.getFramebuffer()[0] == COLOR_ERROR);

        // "SYSTEM ERROR" text, then the mcause hex digits, nothing else in white
        const u16* fb = c.getFramebuffer();
        REQUIRE(whiteInBox(fb, ERROR_TEXT) > 0);
        REQUIRE(whiteInBox(fb, ERROR_CAUSE) > 0);
        REQUIRE(whiteOutsideText(fb) == 0);
    }
}
