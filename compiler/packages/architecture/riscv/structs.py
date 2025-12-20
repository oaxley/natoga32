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
# @brief	RISC-V Instruction Type and Opcode dataclass

#----- imports
from dataclasses import dataclass
from enum import IntEnum, auto


#----- classes
# RISC-V instruction type
class InstrType(IntEnum):
    TYPE_I  = auto()
    TYPE_I2 = auto()
    TYPE_U  = auto()
    TYPE_S  = auto()
    TYPE_R  = auto()
    TYPE_B  = auto()
    TYPE_J  = auto()

@dataclass
class InstrOpcode:
    type    : InstrType
    opcode  : int
    rd      : int = 0
    funct3  : int = 0
    rs1     : int = 0
    rs2     : int = 0
    funct7  : int = 0
