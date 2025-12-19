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
# @brief	Section dataclass

#----- imports
from typing import List
from dataclasses import dataclass, field

from .relocation import Relocation


#----- class
@dataclass
class Section:
    """Define a simple assembly section like '.text', '.data' or '.bss'

    Members:
    - name (str): the name of the section
    - data (bytearray): the data to store in this section
    - relocations (list): the relocation list
    - offset (int): the current write cursor
    - address (int): section start address
    """
    name: str
    data: bytearray = field(default_factory=bytearray)
    relocations: List[Relocation] = field(default_factory=list)
    offset: int = 0
    address: int = 0

    def __repr__(self) -> str:
        return f"""data: {self.data}\noffset: {self.offset}\nrelocations: {self.relocations}"""
