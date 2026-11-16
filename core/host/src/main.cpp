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
 * @brief	Virtual Console - Basic Host Program
 */

// standard library headers
#include <filesystem>
#include <iostream>
#include <string>


// program-specific includes
#include <argparse/argparse.hpp>
#include "vc/vc.h"
#include "host_config.h"
#include "console_runner.h"

//----- functions

//----- main entry
int main(int argc, char* argv[])
{
    HostConfig config;

    // the ROM-BIOS lives next to the executable by default
    const std::string defaultRom =
        (std::filesystem::path(argv[0]).parent_path() / "rombios.sys").string();

    //--- command line parser
    argparse::ArgumentParser program("vchost", "0.1.0");

    // debug flag
    program.add_argument("--debug")
        .flag()     // .default_value(false).implicit_value(true)
        .help("Enable debug server");

    // debug port
    program.add_argument("--debug-port")
        .metavar("PORT")
        .nargs(1)
        .scan<'i', int>()
        .default_value(2600)
        .help("Debug listening port");

    // rombios
    program.add_argument("--rombios")
        .metavar("FILENAME")
        .default_value(defaultRom)
        .help("ROM-BIOS file (default: rombios.sys next to the executable)");

    // host-only: window scaled x2
    program.add_argument("--superscalar")
        .flag()
        .help("Open a 1280x768 window (display scaled x2)");

    // headless run
    program.add_argument("--frames")
        .metavar("N")
        .nargs(1)
        .scan<'i', int>()
        .default_value(0)
        .help("Run N frames without a window, then exit (0 = windowed)");

    // name of the cartridge (optional: without one, the boot screen shows)
    program.add_argument("file")
        .nargs(argparse::nargs_pattern::optional)
        .default_value(std::string{})
        .help("cartridge file");

    // read the arguments
    try {
        program.parse_args(argc, argv);
    } catch (const std::exception& err) {
        std::cerr << err.what() << "\n";
        std::cerr << program;
        return EXIT_FAILURE;
    }

    // setup the configuration
    config.debug = program.get<bool>("--debug");
    config.debug_port = program.get<int>("--debug-port");
    config.rombios = program.get<std::string>("--rombios");
    config.cartridge = program.get<std::string>("file");
    config.superscalar = program.get<bool>("--superscalar");
    config.frames = program.get<int>("--frames");

    ConsoleRunner runner(config);
    return runner.run();
}
