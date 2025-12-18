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
# @brief	Generic Encoder Interface

#----- imports
from typing import List, Tuple, Dict
from abc import ABC, abstractmethod

from packages import ast
from packages.structs import CPU, Relocation, EvalResult


#----- classes
class Architecture(ABC):
    """Architecture Abstract Interface"""

    def __init__(self) -> None:
        self.cpu: CPU = CPU()
        self.locations: Dict[str, int] = {}

    @abstractmethod
    def is_register(self, operand: str) -> Tuple[bool, int]:
        """Check if operand is a register"""

    @abstractmethod
    def encode(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode the instruction with its operands"""

    @abstractmethod
    def expand(self, instr: ast.Instruction) -> List[ast.Instruction]:
        """Expand an instruction into a list of instructions"""

class DefaultArch(Architecture):
    """Default Architecture"""
    def __init__(self) -> None:
        super().__init__()

    def is_register(self, operand: str) -> Tuple[bool, int]:
        return (False, 0)

    def encode(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        return (b"\x00", [])

    def expand(self, instr: ast.Instruction) -> List[ast.Instruction]:
        return []
