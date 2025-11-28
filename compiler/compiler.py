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

import packages.ast as ast

from packages.config import Config
from packages.lexer import Lexer
from packages.parser import Parser
from packages.preprocessor import PreProcessor
from packages.symbols import SymbolTable
from packages.semantic import SemanticAnalyzer
from packages.data_classes import Section


# ----- begin

# parse the command line arguments
argparse = ArgumentParser()
argparse.add_argument("file", help="assembler file to compile")
argparse.add_argument("-o", "--output", help="output file")
argparse.add_argument("-d", "--debug", type=int, help="output file")
args = argparse.parse_args()

# initialize the configuration
try:
    config = Config(args)
except FileNotFoundError as e:
    print(e)
    sys.exit(1)

# open the file
fh = open(config.input_file, "r", encoding="utf-8")

try:
    # global symbol table
    symbols = SymbolTable()

    # create a new lexer
    lexer = Lexer()
    lexer.parse(fh)

    if args.debug == 1:
        for i in lexer.tokens:
            print(i)
        sys.exit(0)

    # pre-processor
    preproc = PreProcessor(symbols)
    tokens = preproc.process(lexer.tokens, config.input_file)

    if args.debug == 2:
        for i in tokens:
            print(i)
        sys.exit(0)

    # parser
    parser = Parser(tokens)
    program = parser.parse_program()

    if args.debug == 3:
        program.debug()
        sys.exit(0)

    # define the assembly sections
    sections: Dict[str, Section] = {
        ".text": Section(".text"),
        ".data": Section(".data"),
        ".bss": Section(".bss")
    }

    # semantic analyzer
    semantic = SemanticAnalyzer(program, symbols, sections)
    semantic.first_pass()

    # print("==== AST ====")
    # print(program)
    print(sections)

except SyntaxError as e:
    print(str(e))
