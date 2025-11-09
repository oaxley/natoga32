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

from packages.config import Config
from packages.lexer import Lexer


# ----- globals
# ----- functions
# ----- classes

# ----- begin

# parse the command line arguments
argparse = ArgumentParser()
argparse.add_argument("file", help="assembler file to compile")
argparse.add_argument("-o", "--output", help="output file")
args = argparse.parse_args()

# initialize the configuration
try:
    config = Config(args)
except FileNotFoundError as e:
    print(e)
    sys.exit(1)

# create a new lexer
lexer = Lexer()
lexer.set_config(config)
lexer.parse()

for value in lexer.tokens():
    print(value)
