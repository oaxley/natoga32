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
 * @brief	Tests - host frame pacing without vsync
 */

#include <catch2/catch_all.hpp>

#include "frame_pacer.h"

constexpr uint64_t MS = 1'000'000;      // 1 ms in ns

TEST_CASE("Frame pacer waits for the rest of the frame", "[host][pacer]") {
    FramePacer pacer(60);
    pacer.start(0);

    // the frame took 5 ms: wait until 1/60 s
    REQUIRE(pacer.delayAfterFrame(5 * MS) == 16'666'666 - 5 * MS);
}

TEST_CASE("Frame pacer keeps an exact 60 Hz cadence", "[host][pacer]") {
    FramePacer pacer(60);
    pacer.start(0);

    // each frame takes 1 ms: after 60 frames we are exactly 1 s later
    uint64_t now = 0;
    for (int i = 0; i < 60; ++i) {
        now += 1 * MS;
        now += pacer.delayAfterFrame(now);
    }
    REQUIRE(now == 1'000'000'000);
}

TEST_CASE("Frame pacer catches up after a slightly late frame", "[host][pacer]") {
    FramePacer pacer(60);
    pacer.start(0);

    // 20 ms frame: no wait, the next frame keeps the original cadence
    REQUIRE(pacer.delayAfterFrame(20 * MS) == 0);
    REQUIRE(pacer.delayAfterFrame(21 * MS) == 33'333'333 - 21 * MS);
}

TEST_CASE("Frame pacer resynchronises after a long stall", "[host][pacer]") {
    FramePacer pacer(60);
    pacer.start(0);

    // 100 ms stall (window dragged, debugger...): do not run frames back to back
    REQUIRE(pacer.delayAfterFrame(100 * MS) == 0);
    REQUIRE(pacer.delayAfterFrame(101 * MS) == 16'666'666 - 1 * MS);
}
