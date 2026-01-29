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
 * @brief	Debug TCP/IP Server Interface
 */
#pragma once

// standard library headers
#include <thread>
#include <atomic>
#include <vector>
#include <functional>

// program-specific includes
#include "vc/vc.h"


//----- DebugServer class
class DebugServer
{
public:
    explicit DebugServer(VC_Console_t* console, uint16_t port = 2600);
    virtual ~DebugServer();

    void start();
    void stop();

private:
    //----- members
    VC_Console_t* console_;
    uint16_t port_;

    std::atomic<bool> running_{false};

    std::thread server_thread_;
    int server_socket_ = -1;
    int client_socket_ = -1;

    //----- methods
    void serverLoop();
    void handleMessage(const uint8_t* header, const uint8_t* data);
};
