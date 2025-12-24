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
# @brief	Structures for Symbol Table

#----- imports
from typing import Optional
from dataclasses import dataclass

from .symbol_type import SymbolType


#----- classes
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

