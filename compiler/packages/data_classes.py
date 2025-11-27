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
from typing import Any, Dict, List

from dataclasses import dataclass, field


#----- classes

@dataclass
class Relocation:
    """Class to record the relocation needed during compilation

    Members:
    - offset (int): byte offset where the relocation applies
    - type (str): the type of relocation needed
    - symbol (str): the symbol name
    - addend (int): signed addend
    - place_size (int): 4 bytes
    - place_inst (bytes): optional reference to the instruction for debug
    """
    offset: int
    type: str
    symbol: str
    addend: int
    place_size: int
    place_inst: bytes


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

    def emit_bytes(self, bs: bytes):
        self.data.extend(bs)
        self.offset += len(bs)



