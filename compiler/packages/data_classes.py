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
# @brief	Dataclasses used in the compiler

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional

from dataclasses import dataclass, field

from packages import ast, symbols


#----- classes

@dataclass
class Relocation:
    """Class to record the relocation needed during compilation

    Members:
    - type (str): the type of relocation needed
    - symbol (str): the symbol name
    - addend (int): signed addend
    - size (int): instruction size
    """
    type: str
    symbol: ast.Node
    addend: int = 0
    size: int = 0

@dataclass
class Section:
    """Define a simple assembly section like '.text', '.data' or '.bss'

    Members:
    - name (str): the name of the section
    - data (bytearray): the data to store in this section
    - relocations (list): the relocation list
    - offset (int): the current write cursor
    - start (int): the starting address for the section
    """
    name: str
    data: bytearray = field(default_factory=bytearray)
    relocations: List[Relocation] = field(default_factory=list)
    offset: int = 0
    align: int = 1
    address: int = 0

    def __repr__(self) -> str:
        return f"""data: {self.data}\noffset: {self.offset}\nrelocations: {self.relocations}"""


@dataclass
class EvalResult:
    """Result of a constant evaluation, that can be either a number or a relocation

    Members:
    - value (Optional[int]): an integer if the value is known, None otherwise
    - reloc (Optional[Relocation]): a relocation structure if value cannot be computed
    """
    value: Optional[int] = None
    reloc: Optional[Relocation] = None


section_t = Dict[str, Section]
@dataclass
class SAData:
    """Semantic Analyzer data class

    Members:
    - program (ast.Program): the AST root node
    - symbols (SymbolTable): the global symbols table
    - sections (section_t): the sections map
    - instr_size (int): the instruction size
    - entry_point (str): the label marked as entry point
    """
    program: ast.Program
    symbols: symbols.SymbolTable
    sections: section_t

    instr_size: int = 0
    entry_point: str = ""
