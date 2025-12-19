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
from dataclasses import dataclass
from enum import IntEnum, auto
from typing import Dict, List, Tuple, cast

from packages import ast
from packages import constants as const
from packages.structs import CPU, EvalResult, Relocation, RelocationType
from packages.classes import Architecture


#----- globals

# RISC-V type
class RiscVType(IntEnum):
    TYPE_I  = auto()
    TYPE_I2 = auto()
    TYPE_U  = auto()
    TYPE_S  = auto()
    TYPE_R  = auto()
    TYPE_B  = auto()
    TYPE_J  = auto()

@dataclass
class Opcode:
    type    : RiscVType
    opcode  : int
    rd      : int = 0
    funct3  : int = 0
    rs1     : int = 0
    rs2     : int = 0
    funct7  : int = 0

# opcodes and their respective type
OPCODES_T: Dict[str, Opcode] = {
    'lb'    : Opcode(RiscVType.TYPE_I ,0b000_0011, 0, 0b000, 0, 0, 0          ),
    'lh'    : Opcode(RiscVType.TYPE_I ,0b000_0011, 0, 0b001, 0, 0, 0          ),
    'lw'    : Opcode(RiscVType.TYPE_I ,0b000_0011, 0, 0b010, 0, 0, 0          ),
    'lbu'   : Opcode(RiscVType.TYPE_I ,0b000_0011, 0, 0b100, 0, 0, 0          ),
    'lhu'   : Opcode(RiscVType.TYPE_I ,0b000_0011, 0, 0b101, 0, 0, 0          ),
    'addi'  : Opcode(RiscVType.TYPE_I ,0b001_0011, 0, 0b000, 0, 0, 0          ),
    'slli'  : Opcode(RiscVType.TYPE_I2,0b001_0011, 0, 0b001, 0, 0, 0          ),
    'slti'  : Opcode(RiscVType.TYPE_I ,0b001_0011, 0, 0b010, 0, 0, 0          ),
    'sltiu' : Opcode(RiscVType.TYPE_I ,0b001_0011, 0, 0b011, 0, 0, 0          ),
    'xori'  : Opcode(RiscVType.TYPE_I ,0b001_0011, 0, 0b100, 0, 0, 0          ),
    'srli'  : Opcode(RiscVType.TYPE_I2,0b001_0011, 0, 0b101, 0, 0, 0          ),
    'srai'  : Opcode(RiscVType.TYPE_I2,0b001_0011, 0, 0b101, 0, 0, 0b0100_000 ),
    'ori'   : Opcode(RiscVType.TYPE_I ,0b001_0011, 0, 0b110, 0, 0, 0          ),
    'andi'  : Opcode(RiscVType.TYPE_I ,0b001_0011, 0, 0b111, 0, 0, 0          ),
    'jalr'  : Opcode(RiscVType.TYPE_I ,0b110_0111, 0, 0b000, 0, 0, 0          ),

    'csrrw' : Opcode(RiscVType.TYPE_I ,0b111_0111, 0, 0b001, 0, 0, 0          ),
    'csrrs' : Opcode(RiscVType.TYPE_I ,0b111_0111, 0, 0b010, 0, 0, 0          ),
    'csrrc' : Opcode(RiscVType.TYPE_I ,0b111_0111, 0, 0b011, 0, 0, 0          ),
    'csrrwi': Opcode(RiscVType.TYPE_I ,0b111_0111, 0, 0b101, 0, 0, 0          ),
    'csrrsi': Opcode(RiscVType.TYPE_I ,0b111_0111, 0, 0b110, 0, 0, 0          ),
    'csrrci': Opcode(RiscVType.TYPE_I ,0b111_0111, 0, 0b111, 0, 0, 0          ),

    'auipc' : Opcode(RiscVType.TYPE_U ,0b001_0111, 0, 0b000, 0 ,0, 0          ),
    'lui'   : Opcode(RiscVType.TYPE_U ,0b011_0111, 0, 0b000, 0 ,0, 0          ),

    'add'   : Opcode(RiscVType.TYPE_R ,0b0000_000, 0, 0b000, 0, 0, 0b011_0011 ),
    'sub'   : Opcode(RiscVType.TYPE_R ,0b0100_000, 0, 0b000, 0, 0, 0b011_0011 ),
    'sll'   : Opcode(RiscVType.TYPE_R ,0b0000_000, 0, 0b001, 0, 0, 0b011_0011 ),
    'slt'   : Opcode(RiscVType.TYPE_R ,0b0000_000, 0, 0b010, 0, 0, 0b011_0011 ),
    'sltu'  : Opcode(RiscVType.TYPE_R ,0b0000_000, 0, 0b011, 0, 0, 0b011_0011 ),
    'xor'   : Opcode(RiscVType.TYPE_R ,0b0000_000, 0, 0b100, 0, 0, 0b011_0011 ),
    'srl'   : Opcode(RiscVType.TYPE_R ,0b0000_000, 0, 0b101, 0, 0, 0b011_0011 ),
    'sra'   : Opcode(RiscVType.TYPE_R ,0b0100_000, 0, 0b101, 0, 0, 0b011_0011 ),
    'or'    : Opcode(RiscVType.TYPE_R ,0b0000_000, 0, 0b110, 0, 0, 0b011_0011 ),

    'mul'   : Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b000, 0, 0, 0b011_0011 ),
    'mulh'  : Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b001, 0, 0, 0b011_0011 ),
    'mulhsu': Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b010, 0, 0, 0b011_0011 ),
    'mulhu' : Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b011, 0, 0, 0b011_0011 ),
    'div'   : Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b100, 0, 0, 0b011_0011 ),
    'divu'  : Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b101, 0, 0, 0b011_0011 ),
    'rem'   : Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b110, 0, 0, 0b011_0011 ),
    'remu'  : Opcode(RiscVType.TYPE_R ,0b0000_001, 0, 0b111, 0, 0, 0b011_0011 ),

    'sext.h': Opcode(RiscVType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0b00101, 0b0110_000 ),
    'zext.h': Opcode(RiscVType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0b00000, 0b0000_100 ),
    'clz'   : Opcode(RiscVType.TYPE_R ,0b0110_000, 0, 0b001, 0, 0b00000, 0b001_0011 ),
    'ctz'   : Opcode(RiscVType.TYPE_R ,0b0110_000, 0, 0b001, 0, 0b00001, 0b001_0011 ),
    'cpop'  : Opcode(RiscVType.TYPE_R ,0b0110_000, 0, 0b001, 0, 0b00010, 0b001_0011 ),
    'rev8'  : Opcode(RiscVType.TYPE_R ,0b0110_100, 0, 0b101, 0, 0b11000, 0b001_0011 ),

    'rol'   : Opcode(RiscVType.TYPE_R ,0b0110_000, 0, 0b001, 0, 0, 0b011_0011 ),
    'ror'   : Opcode(RiscVType.TYPE_R ,0b0110_000, 0, 0b101, 0, 0, 0b011_0011 ),
    'rori'  : Opcode(RiscVType.TYPE_R ,0b0110_000, 0, 0b101, 0, 0, 0b001_0011 ),
    'bclr'  : Opcode(RiscVType.TYPE_R ,0b0100_100, 0, 0b001, 0, 0, 0b011_0011 ),
    'bclri' : Opcode(RiscVType.TYPE_R ,0b0100_100, 0, 0b001, 0, 0, 0b001_0011 ),
    'bext'  : Opcode(RiscVType.TYPE_R ,0b0100_100, 0, 0b101, 0, 0, 0b011_0011 ),
    'bexti' : Opcode(RiscVType.TYPE_R ,0b0100_100, 0, 0b101, 0, 0, 0b001_0011 ),
    'binv'  : Opcode(RiscVType.TYPE_R ,0b0110_100, 0, 0b001, 0, 0, 0b011_0011 ),
    'binvi' : Opcode(RiscVType.TYPE_R ,0b0110_100, 0, 0b001, 0, 0, 0b001_0011 ),
    'bset'  : Opcode(RiscVType.TYPE_R ,0b0010_100, 0, 0b001, 0, 0, 0b011_0011 ),
    'bseti' : Opcode(RiscVType.TYPE_R ,0b0010_100, 0, 0b001, 0, 0, 0b001_0011 ),

    'pack'  : Opcode(RiscVType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0, 0b0000_100 ),
    'packh' : Opcode(RiscVType.TYPE_R ,0b011_0011, 0, 0b111, 0, 0, 0b0000_100 ),
    'max'   : Opcode(RiscVType.TYPE_R ,0b011_0011, 0, 0b110, 0, 0, 0b0000_101 ),
    'maxu'  : Opcode(RiscVType.TYPE_R ,0b011_0011, 0, 0b111, 0, 0, 0b0000_101 ),
    'min'   : Opcode(RiscVType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0, 0b0000_101 ),
    'minu'  : Opcode(RiscVType.TYPE_R ,0b011_0011, 0, 0b101, 0, 0, 0b0000_101 ),

    'sb'    : Opcode(RiscVType.TYPE_S ,0b010_0011, 0, 0b000, 0, 0, 0),
    'sh'    : Opcode(RiscVType.TYPE_S ,0b010_0011, 0, 0b001, 0, 0, 0),
    'sw'    : Opcode(RiscVType.TYPE_S ,0b010_0011, 0, 0b010, 0, 0, 0),

    'beq'   : Opcode(RiscVType.TYPE_B ,0b110_0011, 0, 0b000, 0, 0, 0),
    'bne'   : Opcode(RiscVType.TYPE_B ,0b110_0011, 0, 0b001, 0, 0, 0),
    'blt'   : Opcode(RiscVType.TYPE_B ,0b110_0011, 0, 0b100, 0, 0, 0),
    'bge'   : Opcode(RiscVType.TYPE_B ,0b110_0011, 0, 0b101, 0, 0, 0),
    'bltu'  : Opcode(RiscVType.TYPE_B ,0b110_0011, 0, 0b110, 0, 0, 0),
    'bgeu'  : Opcode(RiscVType.TYPE_B ,0b110_0011, 0, 0b111, 0, 0, 0),

    'jal'   : Opcode(RiscVType.TYPE_J ,0b110_1111, 0, 0b000, 0, 0, 0),

    'ecall' : Opcode(RiscVType.TYPE_I, 0b111_0011, 0, 0b000, 0, 0, 0),
    'mret'  : Opcode(RiscVType.TYPE_I, 0b111_0011, 0, 0b000, 0, 0b00010, 0),
    'wfi'   : Opcode(RiscVType.TYPE_I, 0b111_0011, 0, 0b000, 0, 0b00101, 0b0001_000),

    'new.t'  : Opcode(RiscVType.TYPE_R, 0b000_1011, 0, 0b000, 0, 0, 0),
    'yield.t': Opcode(RiscVType.TYPE_I, 0b000_1011, 0, 0b001, 0, 0, 0),
    'id.t'   : Opcode(RiscVType.TYPE_I, 0b000_1011, 0, 0b010, 0, 0, 0),
    'sleep.t': Opcode(RiscVType.TYPE_I, 0b000_1011, 0, 0b100, 0, 0, 0),
    'wake.t' : Opcode(RiscVType.TYPE_I, 0b000_1011, 0, 0b101, 0, 0, 0),
    'end.t'  : Opcode(RiscVType.TYPE_I, 0b000_1011, 0, 0b111, 0, 0, 0),
}

#----- class

class Riscv(Architecture):
    """RISC-V Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        # CPU definition
        self.cpu: CPU = CPU(4, 32, 1, 2, 4)

        # sections location
        self.locations = {
            '.bss' : 0x0100_0000,
            '.data': 0x0120_0000,
            '.text': 0x01A0_0000
        }

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
        data = OPCODES_T[opcode]

        rd = None
        rs1 = None
        imm = 0
        for i in operands:
            if i.reg is not None:
                if rd is None:
                    rd = i.reg
                else:
                    rs1 = i.reg
            else:
                if i.value is not None:
                    # immediate is signed, between -2048 ... +2047
                    if is_uimm:
                        imm = (data.funct7 << 5) | (i.value & 0x01F)
                    else:
                        imm = i.value & 0xFFF

                elif i.reloc is not None:
                    if i.reloc.type == RelocationType.RISCV_LO12:
                        i.reloc.type = RelocationType.RISCV_LO12_I

                    elif i.reloc.type == RelocationType.RISCV_PCREL_LO12:
                        i.reloc.type = RelocationType.RISCV_PCREL_LO12_I

                    else:
                        raise SyntaxError(f"Error: unsupported relocation '{i.reloc.type.name}' for type I")

                    reloc.append(i.reloc)

        # ensure all the value are not None
        rd = data.rd if rd is None else rd
        rs1 = data.rs1 if rs1 is None else rs1

        # build the instruction
        funct3 = data.funct3
        value = (
            (imm    << 20) |
            (rs1    << 15) |
            (funct3 << 12) |
            (rd     << 7 ) |
            data.opcode
        )

        return (value.to_bytes(4, const.ENDIANESS), reloc)

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

            if r.type not in [RelocationType.RISCV_PCREL_HI20, RelocationType.RISCV_HI20]:
                raise SyntaxError(f"Error: auipc/lui support only %pcrel_hi or %hi relocations.")

            reloc.append(r)

        value = (
            (imm << 12) |
            (rd  << 7 ) |
            OPCODES_T[opcode].opcode
        )

        return (value.to_bytes(4, const.ENDIANESS), reloc)

    def _type_R(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type R RISC-V instructions"""

        # retrieve the operands
        xlen = self.cpu.xlen - 1
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

        funct3 = OPCODES_T[opcode].funct3
        funct7 = OPCODES_T[opcode].funct7
        value = (
            (funct7 << 25) |
            (rs2    << 20) |
            (rs1    << 15) |
            (funct3 << 12) |
            (rd     << 7 ) |
            OPCODES_T[opcode].opcode
        )

        return (value.to_bytes(4, const.ENDIANESS), [])

    def _type_S(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type S RISC-V instructions"""
        reloc: List[Relocation] = []

        # retrieve the operands
        rs2 = None
        rs1 = None
        imm = 0
        for i in operands:
            if i.reg is not None:
                if rs2 is None:
                    rs2 = i.reg
                else:
                    rs1 = i.reg
            else:
                if i.value is not None:
                    imm = i.value & 0x0FFF

                elif i.reloc is not None:
                    if i.reloc.type == RelocationType.RISCV_LO12:
                        i.reloc.type = RelocationType.RISCV_LO12_S

                    elif i.reloc.type == RelocationType.RISCV_PCREL_LO12:
                        i.reloc.type = RelocationType.RISCV_PCREL_LO12_S

                    else:
                        raise SyntaxError(f"Error: unsupported relocation '{i.reloc.type.name}' for type S")

                    reloc.append(i.reloc)

        # build the instruction
        data = OPCODES_T[opcode]
        rs1 = data.rs1 if rs1 is None else rs1
        rs2 = data.rs2 if rs2 is None else rs2

        imm4_0 = imm & 0b11111
        imm11_5 = imm >> 5
        funct3 = data.funct3

        value = (
            (imm11_5 << 25) |
            (rs2     << 20) |
            (rs1     << 15) |
            (funct3  << 12) |
            (imm4_0  << 7 ) |
            data.opcode
        )

        return (value.to_bytes(4, const.ENDIANESS), reloc)

    def _type_B(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type B RISC-V instructions"""
        reloc: List[Relocation] = []

        # retrieve the operands
        rs1 = None
        rs2 = None
        imm = 0
        for i in operands:
            if i.reg is not None:
                if rs1 is None:
                    rs1 = i.reg
                else:
                    rs2 = i.reg
            else:
                if i.value is not None:
                    imm = i.value
                elif i.reloc is not None:
                    i.reloc.type = RelocationType.RISCV_BRANCH
                    reloc.append(i.reloc)

        # build the instruction
        data = OPCODES_T[opcode]
        rs1 = data.rs1 if rs1 is None else rs1
        rs2 = data.rs2 if rs2 is None else rs2

        # check boundaries
        if imm & 0b1:
            raise SyntaxError(f"Error: branch offset must be 2-byte aligned")

        if imm < -(1 << 12) or imm >= (1 << 12):
            raise SyntaxError(f"Error: branch offset out of range")

        imm12   = (imm >> 12) & 0x1
        imm10_5 = (imm >> 5 ) & 0x3F
        imm4_1  = (imm >> 1 ) & 0xF
        imm11   = (imm >> 11) & 0x1

        funct3 = data.funct3
        value = (
            (imm12   << 31) |
            (imm10_5 << 25) |
            (rs2     << 20) |
            (rs1     << 15) |
            (funct3  << 12) |
            (imm4_1  << 8 ) |
            (imm11   << 7 ) |
            data.opcode
        )

        return (value.to_bytes(4, const.ENDIANESS), reloc)

    def _type_J(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode type J RISC-V instructions"""
        reloc: List[Relocation] = []

        # retrieve operands
        rd = cast(int, operands[0].reg)
        imm = 0

        if operands[1].value is not None:
            imm = operands[1].value
            if imm & 0b1 or imm & 0b10:
                raise SyntaxError("Error: JAL offset must be 4-byte aligned")

            if imm < -(1 << 20) or imm >= (1 << 20):
                raise SyntaxError("Error: JAL offset out of range")

        elif operands[1].reloc is not None:
            # set the relocation type
            r = operands[1].reloc
            r.type = RelocationType.RISCV_JAL
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

        return (value.to_bytes(4, const.ENDIANESS), reloc)

    def expand(self, instr: ast.Instruction) -> List[ast.Instruction]:
        """Expland the pseudo-instruction into their full instructions"""
        instructions: List[ast.Instruction] = []

        #--- lambdas definition
        # create the label identifier from an operand
        label = lambda x: ast.Identifier(cast(ast.Identifier, instr.operands[x]).name)

        # create the register identifier
        x = lambda n: ast.Identifier(f"x{n}")

        # definition of the number '0'
        zero = ast.Number(0)

        # groups of aliases to expand
        instr_0arg = ['nop', 'ret']
        instr_1arg = ['call', 'j', 'jr']
        instr_2arg = ['mv', 'seqz', 'snez', 'sltz', 'sgtz', 'not', 'neg']
        branch_1 = ['beqz', 'bnez', 'blez', 'bgez', 'bltz', 'bgtz']
        branch_2 = ['bgt', 'ble', 'bgtu', 'bleu']

        if instr.opcode == "li":
            rd, imm = instr.operands

            if isinstance(imm, ast.Number):
                if -2048 <= imm.value <= 2047:
                    return [
                        ast.Instruction("addi", [rd, x(0), imm])
                    ]

                # big constant
                upper = (imm.value + 0x800) >> 12
                lower = imm.value - (upper << 12)

                return [
                    ast.Instruction("lui", [rd, ast.Number(upper)]),
                    ast.Instruction("addi", [rd, rd, ast.Number(lower)])
                ]

            if isinstance(imm, ast.Expression):
                return [
                    ast.Instruction("lui", [rd, ast.HiRel(imm)]),
                    ast.Instruction("addi", [rd, rd, ast.LoRel(imm)])
                ]

        elif instr.opcode == "la":
            rd = instr.operands[0]
            offset = label(1)
            return [
                ast.Instruction('auipc', [rd, ast.PCRelHi(offset)]),
                ast.Instruction('addi', [rd, rd, ast.PCRelLo(offset)])
            ]

        elif instr.opcode in instr_0arg:
            match instr.opcode:
                case "nop":
                    return [ ast.Instruction('addi', [x(0), x(0), zero]) ]
                case "ret":
                    return [ast.Instruction("jalr", [ x(0), x(1), zero])]

        elif instr.opcode in instr_1arg:
            offset = label(0)
            match instr.opcode:
                case "call":
                    return [
                        ast.Instruction('auipc', [x(1), ast.PCRelHi(offset)]),
                        ast.Instruction('jalr', [x(1), ast.PCRelLo(offset), x(1)])
                    ]
                case "j":
                    return [ ast.Instruction('jal', [x(0), offset]) ]
                case "jr":
                    return [ ast.Instruction('jal', [x(1), offset])]

        elif instr.opcode in instr_2arg:
            rd = instr.operands[0]
            rs = instr.operands[1]

            match instr.opcode:
                case "mv":
                    return [ast.Instruction('addi', [rd, rs, zero])]
                case "seqz":
                    return [ast.Instruction('sltiu', [rd, rs, ast.Number(1)])]
                case "snez":
                    return [ast.Instruction('sltu', [rd, x(0), rs])]
                case "sltz":
                    return [ast.Instruction('slt', [rd, rs, x(0)])]
                case "sgtz":
                    return [ast.Instruction('slt', [rd, x(0), rs])]
                case "not":
                    return [ast.Instruction('xori', [rd, rs, ast.Number(-1)])]
                case "neg":
                    return [ast.Instruction('sub', [rd, x(0), rs])]

        elif instr.opcode in branch_1:
            rs = instr.operands[0]
            offset = label(1)

            match instr.opcode:
                case 'beqz':
                    return [ast.Instruction('beq', [rs, x(0), offset])]
                case 'bnez':
                    return [ast.Instruction('bne', [rs, x(0), offset])]
                case 'blez':
                    return [ast.Instruction('bge', [x(0), rs, offset])]
                case 'bgez':
                    return [ast.Instruction('bge', [rs, x(0), offset])]
                case 'bltz':
                    return [ast.Instruction('blt', [rs, x(0), offset])]
                case 'bgtz':
                    return [ast.Instruction('blt', [x(0), rs, offset])]

        elif instr.opcode in branch_2:
            rs = instr.operands[0]
            rt = instr.operands[1]
            offset = label(2)

            match instr.opcode:
                case 'bgt':
                    return [ast.Instruction('blt', [rt, rs, offset])]
                case 'ble':
                    return [ast.Instruction('bge', [rt, rs, offset])]
                case 'bgtu':
                    return [ast.Instruction('bltu', [rt, rs, offset])]
                case 'bleu':
                    return [ast.Instruction('bgeu', [rt, rs, offset])]

        # misc pseudo-instruction
        elif instr.opcode == "syscall":
            param = instr.operands[0]
            if isinstance(param, ast.Number):
                return [
                    ast.Instruction("addi", [x(1), x(0), param]),
                    ast.Instruction("ecall", [])
                ]
            elif isinstance(param, ast.Identifier):
                r, _ = self.is_register(param.name)
                if r:
                    return [
                        ast.Instruction("add", [x(1), x(0), param]),
                        ast.Instruction("ecall", [])
                    ]
                else:
                    raise SyntaxError(f"Error: {param.name} is not a valid register in syscall")

        else:
            instructions.append(instr)

        return instructions
