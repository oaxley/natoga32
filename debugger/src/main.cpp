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
 * @brief	Debugger Client
 */
// standard library headers
#include <iostream>
#include <GLFW/glfw3.h>

#include <argparse/argparse.hpp>
#include <imgui.h>
#include <imgui_impl_glfw.h>
#include <imgui_impl_opengl3.h>


// program-specific includes
#include "debug_client.h"


//----- main entry point
int main(int argc, char* argv[])
{
    //----- Read the command line
    argparse::ArgumentParser program("vcdebug", "0.1.0");

    // host
    program.add_argument("host")
        .help("IP/host to connect to");

    // port
    program.add_argument("port")
        .scan<'i', int>()
        .default_value(2600)
        .help("Host listening port");

    // read the command line arguments
    try {
        program.parse_args(argc, argv);
    } catch (const std::exception& err) {
        std::cerr << err.what() << "\n";
        std::cerr << program;
        return EXIT_FAILURE;
    }

    // retrieve the parameters and connect to host
    auto host = program.get<std::string>("host");
    auto port = program.get<int>("port");

    DebugClient client(host, port);

    if (client.connect()) {
        std::cout << "Debugger connected to host at " << host << ":" << port << "\n";
    }

    return EXIT_SUCCESS;
}

