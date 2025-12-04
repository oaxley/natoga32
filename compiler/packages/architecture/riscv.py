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
from typing import Any, Dict, List, Tuple, cast

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
    rs2: int = 0

# opcodes and their respective type
OPCODES_T: Dict[str, Opcode] = {
    'lb':    Opcode(RiscVType.TYPE_I, 0b000, 0b000_0011),
    'lh':    Opcode(RiscVType.TYPE_I, 0b001, 0b000_0011),
    'lw':    Opcode(RiscVType.TYPE_I, 0b010, 0b000_0011),
    'lbu':   Opcode(RiscVType.TYPE_I, 0b100, 0b000_0011),
    'lhu':   Opcode(RiscVType.TYPE_I, 0b101, 0b000_0011),
    'addi':  Opcode(RiscVType.TYPE_I, 0b000, 0b001_0011),
    'slli':  Opcode(RiscVType.TYPE_I2, 0b001, 0b001_0011, 0),
    'slti':  Opcode(RiscVType.TYPE_I, 0b010, 0b001_0011),
    'sltiu': Opcode(RiscVType.TYPE_I, 0b011, 0b001_0011),
    'xori':  Opcode(RiscVType.TYPE_I, 0b100, 0b001_0011),
    'srli':  Opcode(RiscVType.TYPE_I2, 0b101, 0b001_0011, 0),
    'srai':  Opcode(RiscVType.TYPE_I2, 0b101, 0b001_0011, 0b0100_000),
    'ori':   Opcode(RiscVType.TYPE_I, 0b110, 0b001_0011),
    'andi':  Opcode(RiscVType.TYPE_I, 0b111, 0b001_0011),
    'jalr':  Opcode(RiscVType.TYPE_I, 0b000, 0b110_0111),

    'csrrw': Opcode(RiscVType.TYPE_I, 0b001, 0b111_0111),
    'csrrs': Opcode(RiscVType.TYPE_I, 0b010, 0b111_0111),
    'csrrc': Opcode(RiscVType.TYPE_I, 0b011, 0b111_0111),
    'csrrwi':Opcode(RiscVType.TYPE_I, 0b101, 0b111_0111),
    'csrrsi':Opcode(RiscVType.TYPE_I, 0b110, 0b111_0111),
    'csrrci':Opcode(RiscVType.TYPE_I, 0b111, 0b111_0111),

    'auipc': Opcode(RiscVType.TYPE_U, 0, 0b001_0111),
    'lui':   Opcode(RiscVType.TYPE_U, 0, 0b011_0111),

    'add':   Opcode(RiscVType.TYPE_R, 0b000, 0b011_0011, 0b0000_000),
    'sub':   Opcode(RiscVType.TYPE_R, 0b000, 0b011_0011, 0b0100_000),
    'sll':   Opcode(RiscVType.TYPE_R, 0b001, 0b011_0011, 0b0000_000),
    'slt':   Opcode(RiscVType.TYPE_R, 0b010, 0b011_0011, 0b0000_000),
    'sltu':  Opcode(RiscVType.TYPE_R, 0b011, 0b011_0011, 0b0000_000),
    'xor':   Opcode(RiscVType.TYPE_R, 0b100, 0b011_0011, 0b0000_000),
    'srl':   Opcode(RiscVType.TYPE_R, 0b101, 0b011_0011, 0b0000_000),
    'sra':   Opcode(RiscVType.TYPE_R, 0b101, 0b011_0011, 0b0100_000),
    'or':    Opcode(RiscVType.TYPE_R, 0b110, 0b011_0011, 0b0000_000),

    'mul':   Opcode(RiscVType.TYPE_R, 0b000, 0b011_0011, 0b0000_001),
    'mulh':  Opcode(RiscVType.TYPE_R, 0b001, 0b011_0011, 0b0000_001),
    'mulhsu':Opcode(RiscVType.TYPE_R, 0b010, 0b011_0011, 0b0000_001),
    'mulhu': Opcode(RiscVType.TYPE_R, 0b011, 0b011_0011, 0b0000_001),
    'div':   Opcode(RiscVType.TYPE_R, 0b100, 0b011_0011, 0b0000_001),
    'divu':  Opcode(RiscVType.TYPE_R, 0b101, 0b011_0011, 0b0000_001),
    'rem':   Opcode(RiscVType.TYPE_R, 0b110, 0b011_0011, 0b0000_001),
    'remu':  Opcode(RiscVType.TYPE_R, 0b111, 0b011_0011, 0b0000_001),

    'clz':   Opcode(RiscVType.TYPE_R, 0b001, 0b001_0011, 0b0110_000, 0b00000),
    'ctz':   Opcode(RiscVType.TYPE_R, 0b001, 0b001_0011, 0b0110_000, 0b00001),
    'cpop':  Opcode(RiscVType.TYPE_R, 0b001, 0b001_0011, 0b0110_000, 0b00010),
    'rev8':  Opcode(RiscVType.TYPE_R, 0b101, 0b001_0011, 0b0110_100, 0b11000),
    'rol':   Opcode(RiscVType.TYPE_R, 0b001, 0b011_0011, 0b0110_000),
    'ror':   Opcode(RiscVType.TYPE_R, 0b101, 0b011_0011, 0b0110_000),
    'rori':  Opcode(RiscVType.TYPE_R, 0b101, 0b001_0011, 0b0110_000),

    'bclr':  Opcode(RiscVType.TYPE_R, 0b001, 0b011_0011, 0b0100_100),
    'bclri': Opcode(RiscVType.TYPE_R, 0b001, 0b001_0011, 0b0100_100),
    'bext':  Opcode(RiscVType.TYPE_R, 0b101, 0b011_0011, 0b0100_100),
    'bexti': Opcode(RiscVType.TYPE_R, 0b101, 0b001_0011, 0b0100_100),
    'binv':  Opcode(RiscVType.TYPE_R, 0b001, 0b011_0011, 0b0110_100),
    'binvi': Opcode(RiscVType.TYPE_R, 0b001, 0b001_0011, 0b0110_100),
    'bset':  Opcode(RiscVType.TYPE_R, 0b001, 0b011_0011, 0b0010_100),
    'bseti': Opcode(RiscVType.TYPE_R, 0b001, 0b001_0011, 0b0010_100),

    'pack':  Opcode(RiscVType.TYPE_R, 0b100, 0b011_0011, 0b0000_100),
    'packh': Opcode(RiscVType.TYPE_R, 0b111, 0b011_0011, 0b0000_100),

    'max':   Opcode(RiscVType.TYPE_R, 0b110, 0b011_0011, 0b0000_101),
    'maxu':  Opcode(RiscVType.TYPE_R, 0b111, 0b011_0011, 0b0000_101),
    'min':   Opcode(RiscVType.TYPE_R, 0b100, 0b011_0011, 0b0000_101),
    'minu':  Opcode(RiscVType.TYPE_R, 0b101, 0b011_0011, 0b0000_101),

    'sext.h':Opcode(RiscVType.TYPE_R, 0b001, 0b001_0011, 0b0110_000, 0b00101),
    'zext.h':Opcode(RiscVType.TYPE_R, 0b100, 0b011_0011, 0b0000_100, 0b00000),

    'sb':    Opcode(RiscVType.TYPE_S, 0b000, 0b010_0011),
    'sh':    Opcode(RiscVType.TYPE_S, 0b001, 0b010_0011),
    'sw':    Opcode(RiscVType.TYPE_S, 0b010, 0b010_0011),

    'beq':   Opcode(RiscVType.TYPE_B, 0b000, 0b110_0011),
    'bne':   Opcode(RiscVType.TYPE_B, 0b001, 0b110_0011),
    'blt':   Opcode(RiscVType.TYPE_B, 0b100, 0b110_0011),
    'bge':   Opcode(RiscVType.TYPE_B, 0b101, 0b110_0011),
    'bltu':  Opcode(RiscVType.TYPE_B, 0b110, 0b110_0011),
    'bgeu':  Opcode(RiscVType.TYPE_B, 0b111, 0b110_0011),

    'jal':   Opcode(RiscVType.TYPE_J, 0b000, 0b110_1111)
}

#----- class

class Riscv(Architecture):
    """RISC-V Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        self.config: CPU = CPU(4, 32, 1, 2, 4)

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
                case RiscVType.TYPE_R:
                    return self._type_R(opcode, operands)
                case RiscVType.TYPE_S:
                    return self._type_S(opcode, operands)
                case RiscVType.TYPE_B:
                    return self._type_B(opcode, operands)
                case RiscVType.TYPE_J:
                    return self._type_J(opcode, operands)

        raise SyntaxError(f"Error: unknown opcode '{opcode}'")

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
                        i.reloc.mask = 0x01F
                    else:
                        i.reloc.mask = 0xFFF

                    # add the _I marker for relocation
                    i.reloc.type += '_I'
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
        mask = 0x0FFFFF

        # retrieve the operands
        rd = cast(int, operands[0].reg)

        if operands[1].value is not None:
            imm = (operands[1].value >> 12) & mask
        else:
            imm = 0

            assert operands[1].reloc is not None
            r = operands[1].reloc

            if r.type not in ['R_RISCV_PCREL_HI20', 'R_RISCV_HI20']:
                raise SyntaxError(f"Error: auipc/lui support only %pcrel_hi or %hi relocations.")

            r.mask = mask
            reloc.append(r)

        value = (
            (imm << 12) |
            (rd << 7) |
            OPCODES_T[opcode].opcode
        )

        return (value.to_bytes(4, 'big'), reloc)

    def _type_R(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type R RISC-V instructions"""

        # retrieve the operands
        xlen = self.config.xlen - 1
        rd = cast(int, operands[0].reg) & xlen
        rs1 = cast(int, operands[1].reg) & xlen

        if len(operands) == 3:
            if operands[2].reg is not None:
                rs2 = operands[2].reg & xlen
            elif operands[2].value is not None:
                rs2 = operands[2].value & xlen
            else:
                raise SyntaxError("Error: Type_R cannot be relocation")
        else:
            rs2 = OPCODES_T[opcode].rs2

        op = OPCODES_T[opcode]
        value = (
            (op.funct7 << 25) |
            (rs2 << 20) |
            (rs1 << 15) |
            (op.funct3 << 12) |
            (rd << 7) |
            op.opcode
        )

        return (value.to_bytes(4, 'big'), [])

    def _type_S(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type S RISC-V instructions"""
        reloc: List[Relocation] = []

        # retrieve the operands
        rs2 = None
        rs1 = None
        imm = None
        for i in operands:
            if i.reg is not None:
                if rs2 is None:
                    rs2 = i.reg
                else:
                    rs1 = i.reg
            else:
                if i.value is not None:
                    imm = i.value & 0x0FFF
                else:
                    r = i.reloc
                    assert r is not None
                    imm = 0

                    # rewrite the relocation type depending on the instruction
                    if r.type == 'LABEL':
                        r.type = 'R_RISCV_LO12'

                    r.mask = 0x0FFF
                    r.type += '_S'
                    reloc.append(r)

        # build the instruction
        assert rs2 is not None
        assert rs1 is not None
        assert imm is not None

        op = OPCODES_T[opcode]
        imm4_0 = imm & 0b11111
        imm11_5 = imm >> 5

        value = (
            (imm11_5 << 25) |
            (rs2 << 20) |
            (rs1 << 15) |
            (op.funct3 << 12) |
            (imm4_0 << 7) |
            op.opcode
        )

        return (value.to_bytes(4, 'big'), reloc)

    def _type_B(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type B RISC-V instructions"""
        reloc: List[Relocation] = []

        # retrieve the operands
        rs1 = None
        rs2 = None
        imm = None
        for i in operands:
            if i.reg is not None:
                if rs1 is None:
                    rs1 = i.reg
                else:
                    rs2 = i.reg
            else:
                assert i.reloc is not None
                i.reloc.type = 'R_RISCV_BRANCH'
                reloc.append(i.reloc)
                imm = 0

        # build the instruction
        assert rs2 is not None
        assert rs1 is not None
        assert imm is not None

        op = OPCODES_T[opcode]
        imm4_0 = imm & 0b11111
        imm11_5 = imm >> 5

        value = (
            (imm11_5 << 25) |
            (rs2 << 20) |
            (rs1 << 15) |
            (op.funct3 << 12) |
            (imm4_0 <<  7) |
            op.opcode
        )

        return (value.to_bytes(4, 'big'), reloc)

    def _type_J(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type J RISC-V instructions"""
        reloc: List[Relocation] = []

        # retrieve operands
        rd = cast(int, operands[0].reg)

        if operands[1].value is not None:
            imm = operands[1].value
            if imm & 0b1 or imm & 0b10:
                raise SyntaxError("Error: JAL offset must be 4-byte aligned")

            if imm < -(1 << 20) or imm >= (1 << 20):
                raise SyntaxError("Error: JAL offset out of range")

        else:
            imm = 0

            assert operands[1].reloc is not None
            r = operands[1].reloc
            r.type = 'R_RISCV_JAL'
            reloc.append(r)

        # encode the instruction
        imm20    = (imm >> 20) & 0x1
        imm10_1  = (imm >> 1)  & 0x3FF
        imm11    = (imm >> 11) & 0x1
        imm19_12 = (imm >> 12) & 0xFF

        value = (
            (imm20    << 31) |
            (imm10_1  << 21) |
            (imm11    << 20) |
            (imm19_12 << 12) |
            (rd       << 7 ) |
            OPCODES_T[opcode].opcode
        )

        return (value.to_bytes(4, 'big'), reloc)
