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
# @brief	Relocation dataclass

#----- imports
from enum import IntEnum, auto
from dataclasses import dataclass

from packages import ast

#----- classes

class RelocationType(IntEnum):
    """Supported types for Relocation"""
    RISCV_PCREL_HI20 = auto()
    RISCV_PCREL_LO12 = auto()       # temporary before deciding for _I or _S
    RISCV_PCREL_LO12_I = auto()
    RISCV_PCREL_LO12_S = auto()
    RISCV_HI20 = auto()
    RISCV_LO12 = auto()             # temporary before deciding for _I or _S
    RISCV_LO12_I = auto()
    RISCV_LO12_S = auto()
    RISCV_BRANCH = auto()
    RISCV_JAL = auto()

    CODE_LABEL = auto()             # label in the .text section
    DATA_LABEL = auto()             # label in the .data section
    BSS_LABEL = auto()              # label in the .bss section

    UNKNOWN = auto()


@dataclass
class Relocation:
    """Class to record the relocation needed during compilation

    Members:
    - type (RelocationType): the type of relocation needed
    - symbol (str): the symbol name
    - addend (int): signed addend
    - address (int): address where the relocation should take place
    """
    type: RelocationType
    symbol: ast.Node
    addend: int = 0
    address: int = 0
