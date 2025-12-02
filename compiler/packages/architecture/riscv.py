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

from enum import IntEnum, auto
from dataclasses import dataclass

from packages.data_classes import EvalResult, Relocation
from .interface import Architecture, CPU


#----- globals

# RISC-V type
class RiscVType(IntEnum):
    TYPE_I = auto()
    TYPE_I2 = auto()
    TYPE_U = auto()
    TYPE_S = auto()
    TYPE_R = auto()
    TYPE_B = auto()
    TYPE_J = auto()

@dataclass
class Opcode:
    type: RiscVType
    funct3: int
    opcode: int
    funct7: int = 0

# opcodes and their respective type
OPCODES_T: Dict[str, Opcode] = {
    'lb': Opcode(RiscVType.TYPE_I, 0b000, 0b000_0011),
    'lh': Opcode(RiscVType.TYPE_I, 0b001, 0b000_0011),
    'lw': Opcode(RiscVType.TYPE_I, 0b010, 0b000_0011),
    'lbu': Opcode(RiscVType.TYPE_I, 0b100, 0b000_0011),
    'lhu': Opcode(RiscVType.TYPE_I, 0b101, 0b000_0011),
    'addi': Opcode(RiscVType.TYPE_I, 0b000, 0b001_0011),
    'slli': Opcode(RiscVType.TYPE_I2, 0b001, 0b001_0011, 0),
    'slti': Opcode(RiscVType.TYPE_I, 0b010, 0b001_0011),
    'sltiu': Opcode(RiscVType.TYPE_I, 0b011, 0b001_0011),
    'xori': Opcode(RiscVType.TYPE_I, 0b100, 0b001_0011),
    'srli': Opcode(RiscVType.TYPE_I2, 0b101, 0b001_0011, 0),
    'srai': Opcode(RiscVType.TYPE_I2, 0b101, 0b001_0011, 0b0100_000),
    'ori': Opcode(RiscVType.TYPE_I, 0b110, 0b001_0011),
    'andi': Opcode(RiscVType.TYPE_I, 0b111, 0b001_0011),
    'auipc': Opcode(RiscVType.TYPE_U, 0, 0b001_0111),
    'lui': Opcode(RiscVType.TYPE_U, 0, 0b011_0111),
    'jalr': Opcode(RiscVType.TYPE_I, 0b000, 0b110_0111)
}

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

    def encode(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode the instruction with its operands"""

        # operands
        print(operands)

        if opcode in OPCODES_T:
            match OPCODES_T[opcode].type:
                case RiscVType.TYPE_I:
                    return self._type_I(opcode, operands)
                case RiscVType.TYPE_I2:
                    return self._type_I(opcode, operands, True)
                case RiscVType.TYPE_U:
                    return self._type_U(opcode, operands)

        return (b"\x00", [])

    def _type_I(self, opcode: str, operands: List[EvalResult], is_uimm: bool = False) -> Tuple[bytes, List[Relocation]]:
        """Encode type I RISC-V instructions"""

        reloc: List[Relocation] = []

        # retrieve the operands
        rd = None
        rs1 = None
        imm = None
        for i in operands:
            if i.reg is not None:
                if rd is None:
                    rd = i.reg
                else:
                    rs1 = i.reg
            else:
                if i.value is not None:
                    imm = i.value & 0x0FFF
                else:
                    assert i.reloc is not None
                    imm = 0
                    if is_uimm:
                        i.reloc.size = 5
                        i.reloc.position = 20
                    reloc.append(i.reloc)

        # ensure all the value are not None
        assert rd is not None
        assert rs1 is not None
        assert imm is not None

        # build the instruction
        data = OPCODES_T[opcode]

        if is_uimm:
            imm = (data.funct7 << 5) | (imm & 31)

        value = (
            (imm << 20) |
            (rs1 << 15) |
            (data.funct3 << 12) |
            (rd << 7) |
            data.opcode
        )

        return (value.to_bytes(4, 'big'), reloc)

    def _type_U(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type U RISC-V instructions"""
        reloc: List[Relocation] = []

        # retrieve the operands
        rd = operands[0].reg

        if operands[1].value is not None:
            imm = (operands[1].value >> 12) & 0x0FFFFF
        else:
            assert operands[1].reloc is not None
            imm = 0
            r = operands[1].reloc
            r.size = 20
            r.position = 12
            reloc.append(r)
            print(reloc)

        assert rd is not None

        value = (
            (imm << 12) |
            (rd << 7) |
            OPCODES_T[opcode].opcode
        )

        return (value.to_bytes(4, 'big'), reloc)
