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
 * @brief	Console Runner main class
 */
#pragma once

// standard library headers
#include <mutex>
#include <thread>
#include <atomic>


// program-specific headers
#include <vc/vc.h>


//----- constants

// target CPU frequency: 10 MHz
constexpr int CPU_FREQUENCY_HZ = 10'000'000;

//----- globals
std::atomic<bool> g_running { true };


//----- ConsoleRunner
class ConsoleRunner
{
public:
    ConsoleRunner();
    ~ConsoleRunner();

    // prevent copy/move
    ConsoleRunner(const ConsoleRunner&) = delete;
    ConsoleRunner& operator=(const ConsoleRunner&) = delete;
    ConsoleRunner(ConsoleRunner&&) = delete;
    ConsoleRunner& operator=(ConsoleRunner&&) = delete;

    bool initialize();
    void run();

private:
    VC_Console_t* console_ = nullptr;
};

