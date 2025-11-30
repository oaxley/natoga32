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
# @brief	RISC-V architecture

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Tuple

from packages.data_classes import EvalResult
from .interface import Architecture, CPU

#----- class

class Riscv(Architecture):
    """RISC-V Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        self.config: CPU = CPU(4, 1, 2, 4)

        # registers list and their aliases
        self.reg_list: List[str] = []
        for i in range(32):
            self.reg_list.append(f"x{i}")

        self.reg_alias: List[str] = [
            "zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2",
            "s0", "s1"
        ]
        for i in range(8):
            self.reg_alias.append(f"a{i}")
        for i in range(2, 12):
            self.reg_alias.append(f"s{i}")

        self.reg_alias.append("t3")
        self.reg_alias.append("t4")
        self.reg_alias.append("t5")
        self.reg_alias.append("t6")
        self.reg_alias.append("t7")


    def is_register(self, operand: str) -> Tuple[bool, int]:
        """Check if operand is a register"""
        if operand in self.reg_list:
            return (True, self.reg_list.index(operand))

        if operand in self.reg_alias:
            return (True, self.reg_alias.index(operand))

        # special case fp <-> x8
        if operand == "fp":
            return (True, 8)

        return (False, 0)

    def encode(self, opcode: str, operands: List[EvalResult]) -> bytes:
        """Encode the instruction with its operands"""
        return b"\x00"
