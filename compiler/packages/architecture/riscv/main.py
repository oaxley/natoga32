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
from typing import Dict, List, Tuple, cast

from packages import ast
from packages import constants as const
from packages.structs import CPU, EvalResult, Relocation, RelocationType
from packages.classes import Architecture

from .opcode import OPCODES_T
from .pseudo import instr_expand
from .registers import define_registers
from .structs import InstrType

from .encoder import (
    type_B_encoder, type_I_encoder, type_J_encoder,
    type_R_encoder, type_S_encoder, type_U_encoder
)


#----- class
class Riscv(Architecture):
    """RISC-V Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        # CPU definition
        self.cpu: CPU = CPU(4, 32, 1, 2, 4)
        define_registers(self.cpu)

        # memory map
        self.mmap = {
            '.bss' : 0x0100_0000,
            '.data': 0x0120_0000,
            '.text': 0x01A0_0000
        }

    def encode(self, opcode: str, operands: List[EvalResult]) -> Tuple[bytes, List[Relocation]]:
        """Encode the instruction with its operands"""

        if opcode in OPCODES_T:
            match OPCODES_T[opcode].type:
                case InstrType.TYPE_I:
                    return type_I_encoder(opcode, operands, self.cpu, False)
                case InstrType.TYPE_I2:
                    return type_I_encoder(opcode, operands, self.cpu, True)
                case InstrType.TYPE_U:
                    return type_U_encoder(opcode, operands, self.cpu)
                case InstrType.TYPE_R:
                    return type_R_encoder(opcode, operands, self.cpu)
                case InstrType.TYPE_S:
                    return type_S_encoder(opcode, operands, self.cpu)
                case InstrType.TYPE_B:
                    return type_B_encoder(opcode, operands, self.cpu)
                case InstrType.TYPE_J:
                    return type_J_encoder(opcode, operands, self.cpu)

        raise SyntaxError(f"Error: unknown opcode '{opcode}'")

    def expand(self, instr: ast.Instruction) -> List[ast.Instruction]:
        return instr_expand(self.cpu, instr)
