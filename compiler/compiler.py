#!/usr/bin/env python
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

from argparse import ArgumentParser

from packages.config import Config


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
config = Config(args)

