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
 * @brief	Debug TCP/IP Client Interface
 */
#pragma once

// standard library headers
#include <string>

// program-specific includes


//----- DebugClient class
class DebugClient
{
public:
    DebugClient(std::string host, int port);
    ~DebugClient();

    bool connect();

private:
    //----- members
    std::string host_;
    int port_;

    int client_socket_ = -1;

    //----- methods
};
