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
from dataclasses import dataclass

from packages.data_classes import EvalResult, Relocation
from packages.ast import Instruction


#----- classes
@dataclass
class CPU:
    """CPU Configuration

    Members:
    - instr_size (int): the size of an instruction in byte
    - xlen (int): the size of an instruction in bit
    - byte (int): the size associated with the ".byte" directive
    - half (int): the size associated with the ".half" directive
    - word (int): the size associated with the ".word" directive
    """
    instr_size: int = 0
    xlen: int = 0
    byte: int = 0
    half: int = 0
    word: int = 0


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
