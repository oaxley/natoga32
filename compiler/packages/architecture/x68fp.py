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
from __future__ import annotations
from typing import Any, Dict, List, Tuple

from compiler.packages.ast import Instruction

from .interface import Architecture, CPU
from packages.data_classes import EvalResult, Relocation


#----- classes

class X68fp(Architecture):
    """X68FP Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        self.config: CPU = CPU(2, 16, 1, 1, 2)

    def is_register(self, operand: str) -> Tuple[bool, int]:
        """Check if operand is a register"""
        return (False, 0)

    def encode(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode the instruction with its operands"""
        return (b"\x00", [])

    def expand(self, instr: Instruction) -> List[Instruction]:
        """Expand pseudo-instructions"""
        return [instr]
