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
# @brief	Semantic Analyzer data class

#----- imports
from __future__ import annotations
from typing import Dict, Optional

from dataclasses import dataclass

from packages.ast import Program
from packages.data_classes import Section
from packages.symbols import SymbolTable
from packages.architecture import Architecture


#----- class
@dataclass
class SAData:
    """Semantic Analyzer data class

    Members:
    - program (ast.Program): the AST root node
    - symbols (SymbolTable): the global symbols table
    - sections (section_t): the sections map
    - instr_size (int): the instruction size
    - entry_point (str): the label marked as entry point
    - architecture (Architecture): the selected architecture
    """
    program: Program
    symbols: SymbolTable
    sections: Dict[str, Section]

    instr_size: int = 0
    entry_point: str = ""
    arch: Optional[Architecture] = None
