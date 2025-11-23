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
from typing import Any, Dict, List

from dataclasses import dataclass

from packages.token import Token

#----- classes

# a single symbol
@dataclass
class Symbol:
    """Symbol definition

    Members:
    - name (str): the name of the symbol
    - value (str): the value of the symbol
    """
    name: str
    value: str

class SymbolTable:

    def __init__(self) -> None:
        """Constructor"""
        self.symbols: Dict[str, Symbol] = {}

    def add(self, name: str, value: str) -> None:
        """Add a new symbol to the table

        Args:
            name (str): the name of the symbol to add
            value (str): the associated value
        """
        self.symbols[name] = Symbol(name, value)

    def exists(self, name: str) -> bool:
        """Check if a symbol exists

        Args:
            name (str): the name of the symbol

        Returns:
            bool: True if the symbol exists, False otherwise
        """
        return (name in self.symbols)

    def value(self, name: str) -> str:
        """Return the value of a symbol

        Args:
            name (str): the name of the symbol

        Returns:
            str: the associate value of the symbol
        """
        if name in self.symbols:
            return self.symbols[name].value
        else:
            return ""
