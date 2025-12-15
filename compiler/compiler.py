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
from __future__ import annotations
from typing import Any, Dict, List

import sys

from argparse import ArgumentParser

from packages.structs import Config
from packages.lexer import Lexer
from packages.pre_processor import PreProcessor

# from packages.parser import Parser
# from packages.preprocessor import PreProcessor
# from packages.symbols import SymbolTable
# from packages.semantic import SemanticAnalyzer
# from packages.data_classes import Section
# from packages.pseudo_instr import PseudoInstruction


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
    out = preproc.process()

    if args.debug == step:
        for i in out:
            print(i)

        sys.exit(0)

except SyntaxError as e:
    print(str(e))


# # open the file
# fh = open(config.input_file, "r", encoding="utf-8")

# try:
#     # global symbol table
#     symbols = SymbolTable()

#     # create a new lexer
#     step = 1
#     print(f"Phase {step} : Lexer")
#     lexer = Lexer()
#     lexer.parse(fh)

#     if args.debug == step:
#         for i in lexer.tokens:
#             print(i)
#         sys.exit(0)

#     # pre-processor
#     step += 1
#     print(f"Phase {step} : Pre-processor")
#     preproc = PreProcessor(symbols)
#     tokens = preproc.process(lexer.tokens, config.input_file)

#     if args.debug == step:
#         for i in tokens:
#             print(i)
#         sys.exit(0)

#     # parser
#     step += 1
#     print(f"Phase {step} : Parser")
#     parser = Parser(tokens)
#     program = parser.parse_program()

#     if args.debug == step:
#         program.debug()
#         sys.exit(0)


#     # pseudo instruction expansion
#     step += 1
#     pseudo = PseudoInstruction(program)
#     program = pseudo.process()

#     if args.debug == step:
#         program.debug()
#         sys.exit(0)


#     # define the assembly sections
#     sections: Dict[str, Section] = {
#         ".text": Section(".text"),
#         ".data": Section(".data"),
#         ".bss": Section(".bss")
#     }

#     # semantic analyzer
#     step += 1
#     print(f"Phase {step} : Semantic Analyzer - First Pass")
#     semantic = SemanticAnalyzer(program, symbols, sections)
#     semantic.first_pass()

#     if args.debug == step:
#         print("==== SECTIONS ====")
#         for v in sections.items():
#             print(v[0])
#             print(v[1])
#             print()

#         print("\n==== SYMBOLS ====")
#         for k, v in symbols.items:
#             print(v)

#         sys.exit(0)

#     # semantic phase 2
#     step += 1
#     print(f"Phase {step} : Semantic Analyzer - Second Pass")
#     semantic.second_pass()

#     if args.debug == step:
#         print("==== SECTIONS ====")
#         for v in sections.items():
#             print(v[0])
#             print(v[1])
#             print()

#         print("\n==== SYMBOLS ====")
#         for k, v in symbols.items:
#             print(v)

#         sys.exit(0)


# except SyntaxError as e:
#     print(str(e))
