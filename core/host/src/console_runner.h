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

// program-specific headers
#include <vc/vc.h>
#include "host_config.h"


//----- ConsoleRunner
class ConsoleRunner
{
public:
    explicit ConsoleRunner(const HostConfig& config);
    ~ConsoleRunner();

    // prevent copy/move
    ConsoleRunner(const ConsoleRunner&) = delete;
    ConsoleRunner& operator=(const ConsoleRunner&) = delete;
    ConsoleRunner(ConsoleRunner&&) = delete;
    ConsoleRunner& operator=(ConsoleRunner&&) = delete;

    // load the ROM, then run headless (frames > 0) or in a window;
    // returns the process exit code
    int run();

private:
    bool boot();
    int runHeadless();
    int runWindowed();

    HostConfig config_;
    VC_Console_t* console_ = nullptr;
};
