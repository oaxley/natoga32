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
 * @brief	Frame pacing when vsync is not available
 */
#pragma once

// standard library headers
#include <cstdint>


//----- FramePacer
//
// Frame N ends at start + N / fps seconds (exact, no drift). A frame a bit
// late is caught up by the next ones; after a stall longer than one frame the
// cadence restarts from now instead of running frames back to back.
// Times are in nanoseconds, from any monotonic clock.
class FramePacer
{
public:
    explicit FramePacer(uint32_t fps) : fps_(fps) {}

    // start the cadence at now_ns
    void start(uint64_t now_ns)
    {
        start_ns_ = now_ns;
        frame_ = 0;
    }

    // a frame has been rendered at now_ns: returns how long to wait (ns)
    uint64_t delayAfterFrame(uint64_t now_ns)
    {
        ++frame_;
        const uint64_t deadline = deadlineOf(frame_);
        if (now_ns <= deadline) {
            return deadline - now_ns;
        }

        // late by more than one frame: resynchronise
        if (now_ns - deadline > NS_PER_SECOND / fps_) {
            start(now_ns);
        }
        return 0;
    }

private:
    static constexpr uint64_t NS_PER_SECOND = 1'000'000'000;

    uint64_t deadlineOf(uint64_t frame) const
    {
        return start_ns_ + frame * NS_PER_SECOND / fps_;
    }

    uint32_t fps_;
    uint64_t start_ns_ = 0;
    uint64_t frame_ = 0;
};
