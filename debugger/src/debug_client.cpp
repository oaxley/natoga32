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
 * @brief	Debug TCP/IP Client Implementation
 */

// standard library headers
#include <iostream>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <netdb.h>

// program-specific includes
#include "debug_client.h"


DebugClient::DebugClient(std::string host, int port) :
    host_(host), port_(port)
{ }

DebugClient::~DebugClient()
{
    if (client_socket_ > 0) {
        close(client_socket_);
        client_socket_ = -1;
    }
}

bool DebugClient::connect()
{
    if (client_socket_ > 0) {
        return true;
    }

    // create the client socket
    client_socket_ = socket(AF_INET, SOCK_STREAM, 0);
    if (client_socket_ < 0) {
        std::cerr << "Error: failed to create client socket\n";
        return false;
    }

    // prepare the structure
    struct sockaddr_in server_addr = {};
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(port_);

    if (isdigit(host_[0])) {
        // IP address
        inet_pton(AF_INET, host_.c_str(), &server_addr.sin_addr);

    } else {
        // hostname
        struct addrinfo hints = {};
        hints.ai_family = AF_INET;

        struct addrinfo* result;
        if (getaddrinfo(host_.c_str(), nullptr, &hints, &result) != 0) {
            client_socket_ = -1;
            std::cerr << "Error: unable to resolve hostname\n";
            return false;
        }

        // copy the resolved address
        struct sockaddr_in* resolved = (struct sockaddr_in*)result->ai_addr;
        server_addr.sin_addr = resolved->sin_addr;

        // free memory
        freeaddrinfo(result);
    }

    // try to connect to host
    if (::connect(client_socket_, (struct sockaddr*)&server_addr, sizeof(server_addr)) < 0) {
        client_socket_ = -1;
        std::cerr << "Error: unable to connect to remote host\n";
        return false;
    }

    return true;
}

