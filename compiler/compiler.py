#!/usr/bin/env python3
# -*- coding: utf-8 -*-
# vim: filetype=python
#
# This source file is subject to the MIT License
# that is bundled with this package in the file LICENSE.txt.
# It is also available through the Internet at this address:
# https://opensource.org/license/mit
#
# @author	Sebastien LEGRAND
# @license	MIT License
#
# @brief	Natoga32 compiler

# ----- imports
import sys

from argparse import ArgumentParser

from packages.structs import Config
from packages.lexer import Lexer
from packages.pre_processor import PreProcessor
from packages.parser import Parser
from packages.pseudo_instr import PseudoInstruction
from packages.semantic import SemanticAnalyzer


# ----- begin

# parse the command line arguments
argparse = ArgumentParser()
argparse.add_argument("infile", help="assembler file to compile")
argparse.add_argument("-o", "--outfile", help="output file")
argparse.add_argument("-d", "--debug", type=int, help="output file")
args = argparse.parse_args()

# initialize the configuration
try:
    config = Config(args)
    print(f"Compiling '{config.infile}' to '{config.outfile}' ..." )
except FileNotFoundError as e:
    print(e)
    sys.exit(1)

try:
    # Step 1: Lexer
    step = 1
    print(f"Phase {step} : Lexer")
    lexer = Lexer()
    lexer.parse(config)
    config.tokens = lexer.tokens

    if args.debug == step:
        for i in lexer.tokens:
            print(i)
        sys.exit(0)

    # Step 2: Pre-processor
    step += 1
    print(f"Phase {step} : Pre-processor")
    preproc = PreProcessor(config)
    config.tokens = preproc.process()

    if args.debug == step:
        for i in config.tokens:
            print(i)

        sys.exit(0)

    # Step 3: Parser
    step += 1
    print(f"Phase {step} : Parser")
    parser = Parser(config.tokens)
    config.program = parser.parse_program()

    if args.debug == step:
        config.program.debug()
        sys.exit(0)

    # Step 4: Pseudo-instructions expansion
    step += 1
    print(f"Phase {step} : Pseudo-instructions expansion")
    pseudo = PseudoInstruction(config.program)
    config.program = pseudo.process()

    if args.debug == step:
        config.program.debug()
        sys.exit(0)

    # Step 5: Semantic Analyzer first pass
    step += 1
    print(f"Phase {step} : Semantic Analyzer - pass 1")
    semantic = SemanticAnalyzer(config)
    semantic.first_pass()

    if args.debug == step:
        print("==== SECTIONS ====")
        for v in config.sections.items():
            print(v[0])
            print(v[1])
            print()

        print("\n==== SYMBOLS ====")
        for k, v in config.symbols.items:
            print(v)
        sys.exit(0)

    # Step 6: Semantic Analyzer second pass
    step += 1
    print(f"Phase {step} : Semantic Analyzer - pass 2")
    semantic.second_pass()

    if args.debug == step:
        print("==== SECTIONS ====")
        for v in config.sections.items():
            print(v[0])
            print(v[1])
            print()

        print("\n==== SYMBOLS ====")
        for k, v in config.symbols.items:
            print(v)
        sys.exit(0)

    # Step 6: Semantic Analyzer second pass
    step += 1
    print(f"Phase {step} : Semantic Analyzer - pass 3")
    semantic.third_pass()

    if args.debug == step:
        print("==== SECTIONS ====")
        for v in config.sections.items():
            print(v[0])
            print(v[1])
            print()

        print("\n==== SYMBOLS ====")
        for k, v in config.symbols.items:
            print(v)
        sys.exit(0)

except SyntaxError as e:
    print(str(e))
