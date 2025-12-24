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
# @brief	Type B instruction encoder

#----- imports
from typing import List, Tuple, cast

from packages import constants as const
from packages.structs import EvalResult, Relocation, RelocationType, CPU
from packages.architecture.riscv.opcode import OPCODES_T


#----- functions
def type_B_encoder(opcode: str, operands: List[EvalResult], cpu: CPU) -> Tuple[bytes, List[Relocation]]:
        """Encode type B RISC-V instructions

        Args:
            opcode (str): the opcode from the source
            operands (List[EvalResult]): the list of operands
            cpu (CPU): RISC-V cpu definition

        Returns:
            Tuple[bytes, List[Relocation]]: the encoded instruction and the list of relocations created during encoding
        """
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

        return (value.to_bytes(cpu.instr_size, const.ENDIANESS), reloc)
