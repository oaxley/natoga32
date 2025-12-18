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
# @brief	X68FP Architecture

#----- imports
from dataclasses import dataclass
from enum import IntEnum, auto
from typing import Dict, List, Tuple, cast

from packages import ast
from packages.structs import CPU, EvalResult, Relocation
from packages.classes import Architecture


#----- classes

class X68fp(Architecture):
    """X68FP Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        self.cpu: CPU = CPU(2, 16, 1, 1, 2)

    def is_register(self, operand: str) -> Tuple[bool, int]:
        """Check if operand is a register"""
        return (False, 0)

    def encode(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode the instruction with its operands"""
        return (b"\x00", [])

    def expand(self, instr: ast.Instruction) -> List[ast.Instruction]:
        """Expand pseudo-instructions"""
        return [instr]
