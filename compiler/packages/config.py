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
# @brief	Configuration object for the compiler

# ----- imports
from __future__ import annotations
from typing import Any, Dict, List

import os

from argparse import Namespace
from dataclasses import dataclass


# ----- classes

@dataclass(init=False)
class Config():
    """Class for keeping track of the configuration options of the compiler"""
    input_file: str
    output_file: str

    def __init__(self, args: Namespace) -> None:
        """Constructor

        Args:
            args: the Namespace from the command line parser
        """
        # input file checks
        if self.is_exist(args.file):
            self.input_file = args.file
        else:
            raise FileNotFoundError("Error: Unable to find the input file!")

        # output file checks
        if args.output is None:
            self.output_file = "a.out"
        else:
            self.output_file = args.output

    def is_exist(self, filename: str) -> bool:
        """check if the file exists"""
        return os.path.exists(filename)
