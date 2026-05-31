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
#include <iostream>
#include <vector>


// program-specific includes
#include <argparse/argparse.hpp>
#include "vc/vc.h"


//----- functions

//----- main entry
int main(int argc, char* argv[])
{
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
        .default_value(std::string{"rombios.sys"})
        .required()
        .help("ROM-BIOS file");

    // name of the cartridge
    program.add_argument("file")
        .nargs(1)
        .remaining();

    // read the arguments
    try {
        program.parse_args(argc, argv);
    } catch (const std::exception& err) {
        std::cerr << err.what() << "\n";
        std::cerr << program;
        return EXIT_FAILURE;
    }

    return EXIT_SUCCESS;
}
