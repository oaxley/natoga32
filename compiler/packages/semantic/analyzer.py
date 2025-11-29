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
# @brief	Semantic Analyzer main file

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages import ast
from packages.symbols import SymbolTable
from packages import data_classes as dc

from .first_pass import FirstPass
from .second_pass import SecondPass

#----- class

class SemanticAnalyzer:
    """Semantic Analyzer"""

    def __init__(self, program: ast.Program, symbols: SymbolTable, sections: dc.section_t) -> None:
        """Constructor

        Args:
            program (ast.Program): the program as an AST tree after parsing
            symbols (SymbolTable): the symbol table
            sections (section_t): the Sections definition
        """
        self.data = dc.SAData(program, symbols, sections)

    def first_pass(self) -> None:
        """Execute the semantic analyzer first pass"""
        sa_pass = FirstPass(self.data)
        sa_pass.process()

    def second_pass(self) -> None:
        """Execute the semantic analyzer first pass"""
        sa_pass = SecondPass(self.data)
        sa_pass.process()
