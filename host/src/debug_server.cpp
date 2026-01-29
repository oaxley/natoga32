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
 * @brief	Debug TCP/IP Server Implementation
 */

// standard library headers
#include <cstring>
#include <iostream>
#include <unistd.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>


// program-specific includes
#include "vc/vc.h"
#include "vc/debug.h"
#include "debug_server.h"


DebugServer::DebugServer(VC_Console_t* console, uint16_t port) :
    console_(console), port_(port)
{ }

DebugServer::~DebugServer()
{
    stop();
}

void DebugServer::start()
{
    if (running_)
        return;

    // create a new socket
    server_socket_ = socket(AF_INET, SOCK_STREAM, 0);
    if (server_socket_ < 0) {
        throw std::runtime_error("Failed to create server socket");
    }

    // enable the reuse of address/port
    int opt = 1;
    setsockopt(server_socket_, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));

    // prepare socket structure
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(port_);

    // bind the socket
    if (bind(server_socket_, (sockaddr*)&addr, sizeof(addr)) < 0) {
        throw std::runtime_error("Failed to bind socket");
    }

    // wait for connection and start the listening thread
    listen(server_socket_, 1);
    running_ = true;
    server_thread_ = std::thread(&DebugServer::serverLoop, this);
    std::cout << "Debug server listening on port " << port_ << "\n";
}

void DebugServer::stop()
{
    running_ = false;

    // close the socket
    if (server_socket_ > 0) {
        close(server_socket_);
        server_socket_ = -1;
    }

    // wait for the thread
    if (server_thread_.joinable()) {
        server_thread_.join();
    }
}

void DebugServer::serverLoop()
{
    while (running_)
    {
        sockaddr_in client_addr{};
        socklen_t client_len = sizeof(client_addr);

        // accept incomming connection
        client_socket_ = accept(server_socket_, (sockaddr*)&client_addr, &client_len);
        if (client_socket_ < 0)
            continue;

        std::cout << "New connection from " << inet_ntoa(client_addr.sin_addr) << "\n";

        // read message from clients (inner loop)
        while (running_)
        {
            // read the message header
            vc::debug::MessageHeader header;
            ssize_t n = recv(client_socket_, &header, sizeof(header), MSG_WAITALL);
            if (n != sizeof(header))
                break;

            // read the payload
            std::vector<uint8_t> payload(header.size);
            if (header.size > 0) {
                n = recv(client_socket_, payload.data(), header.size, MSG_WAITALL);
                if (n != header.size)
                    break;
            }

            // handle the mesasage
            handleMessage(reinterpret_cast<uint8_t*>(&header), payload.data());
        }

        // close the connection with the client
        close(client_socket_);
        client_socket_ = -1;
        std::cout << "Client disconnected\n";
    }
}

void DebugServer::handleMessage(const uint8_t* header, const uint8_t* data)
{
    // retrieve the header
    auto* msg_header = reinterpret_cast<const vc::debug::MessageHeader*>(header);

    switch (static_cast<vc::debug::MessageType>(msg_header->type))
    {
        case vc::debug::MessageType::GetConsoleState:
            break;

        case vc::debug::MessageType::SetConsoleState:
            break;

        case vc::debug::MessageType::ReadMemory:
            break;

        case vc::debug::MessageType::WriteMemory:
            break;

        case vc::debug::MessageType::ReadCSR:
            break;

        case vc::debug::MessageType::WriteCSR:
            break;

        case vc::debug::MessageType::GetThreadInfo:
            break;

        case vc::debug::MessageType::SetBreakpoint:
            break;

        case vc::debug::MessageType::ClearBreakpoint:
            break;

        case vc::debug::MessageType::StepInstruction:
            break;

        case vc::debug::MessageType::GetVRAM:
            break;

        case vc::debug::MessageType::Pause:
            vc_pause(console_);
            break;

        case vc::debug::MessageType::Resume:
            vc_resume(console_);
            break;

        case vc::debug::MessageType::Reset:
            vc_reset(console_);
            break;

        default:
        break;
    }
}
