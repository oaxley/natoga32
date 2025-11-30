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
# @brief	Symbol table management

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional

from dataclasses import dataclass
from enum import IntEnum, auto

from packages.token import Token

#----- classes

# symbol types
class SymbolType(IntEnum):
    LABEL = auto()
    DEFINE = auto()
    MACRO = auto()
    UNKNOWN = auto()


# a single symbol
@dataclass
class Symbol:
    """Symbol definition

    Members:
    - name (str): the name of the symbol
    - value (Optional[int]): numeric value / address of the symbol
    - type (SymbolType): the type of the symbol
    - section (Optional[str]): the section where the symbol is defined, or None
    - defined (bool): True if the symbol has been assigned
    """
    name: str
    value: Optional[int]
    type: SymbolType
    section: Optional[str] = None
    defined: bool = False

    def __repr__(self) -> str:
        return f"""name: {self.name:<10} | value: {self.value if self.value is not None else "":<8} | type: {self.type.name:<10} | section: {self.section if self.section else "":<5} | defined: {self.defined}"""


class SymbolTable:

    def __init__(self) -> None:
        """Constructor"""
        self.symbols: Dict[str, Symbol] = {}

    def define(self, name: str, value: Optional[int], type: SymbolType, section: Optional[str] = None) -> None:
        """Define a new symbol in the table

        Args:
            name (str): the name of the symbol
            value (Optional[int]): numeric value / address of the symbol
            type (SymbolType): the type of the symbol
            section (Optional[str]): the section where the symbol is defined, or None
        """
        if name in self.symbols and self.symbols[name].defined:
            raise Exception(f"Symbol {name} redefined!")
        self.symbols[name] = Symbol(name, value, type, section, True)

    def get(self, name: str) -> Symbol:
        """Retrieve a symbol from the table

        Args:
            name (str): the name of the symbol

        Returns:
            Symbol: the associated symbol
        """
        try:
            return self.symbols[name]
        except KeyError:
            return Symbol("", None, SymbolType.UNKNOWN, None, False)

    def exists(self, name: str) -> bool:
        """Check if a symbol exists

        Args:
            name (str): the name of the symbol

        Returns:
            bool: True if the symbol exists, False otherwise
        """
        return (name in self.symbols)

    @property
    def items(self):
        for k in self.symbols.keys():
            yield (k, self.symbols[k])
