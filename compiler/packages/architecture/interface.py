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
from __future__ import annotations
from typing import Any, Dict, List, Tuple

from abc import ABC, abstractmethod

from packages.structs import CPU

from packages.data_classes import EvalResult, Relocation
from packages.ast import Instruction


#----- classes
class Architecture(ABC):
    """Generic Architecture"""

    def __init__(self) -> None:
        self.config: CPU = CPU()

    @abstractmethod
    def is_register(self, operand: str) -> Tuple[bool, int]:
        """Check if operand is a register"""

    @abstractmethod
    def encode(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode the instruction with its operands"""

    @abstractmethod
    def expand(self, instr: Instruction) -> List[Instruction]:
        """Expand an instruction into a list of instructions"""
