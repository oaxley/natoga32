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
# @brief	Configuration dataclass

#----- imports
from typing import List, Set

import os

from dataclasses import dataclass, field
from argparse import Namespace

from packages.classes import SymbolTable
from .token import Token

#----- class
@dataclass(init=False)
class Config:
    """Global configuration dataclass

    Args:
        infile (str): the input filename
        outfile (str): the output filename (default: a.out)
        includes (Set[str]): includes stack to detect recursion
        tokens (List[Token]): the list of tokens after the lexer pass
        symbols (SymbolTable): the table of symbols
        row (int): current row being processed
        col (int): current col being processed
    """
    infile: str
    outfile: str
    includes: Set[str]
    tokens: List[Token] = field(default_factory=list)
    symbols: SymbolTable
    row: int
    col: int


    def __init__(self, args: Namespace) -> None:
        """Constructor

        Args:
            args (Namespace): the namespace from the Argparse parser
        """
        # set infile / outfile
        if self._is_exist(args.infile):
            self.infile = args.infile
        else:
            raise FileNotFoundError(f"Error: unable to find '{args.infile}'!")

        if args.outfile:
            self.outfile = args.outfile
        else:
            self.outfile = "a.out"

        # initialize a new SymbolTable
        self.symbols = SymbolTable()

        # track includes to avoid infinite recursion
        includes = set()

    def _is_exist(self, filename: str) -> bool:
        """Check if the file exists

        Args:
            filename (str): the filename to check

        Returns:
            bool: True if the file exists, False otherwise
        """
        return os.path.exists(filename)
