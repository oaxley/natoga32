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
# @brief	Type J instruction encoder

#----- imports
from typing import List, Tuple, cast

from packages import constants as const
from packages.structs import EvalResult, Relocation, RelocationType, CPU
from packages.architecture.riscv.opcode import OPCODES_T


#----- functions
def type_J_encoder(opcode: str, operands: List[EvalResult], cpu: CPU) -> Tuple[bytes, List[Relocation]]:
        """Encode type J RISC-V instructions

        Args:
            opcode (str): the opcode from the source
            operands (List[EvalResult]): the list of operands
            cpu (CPU): RISC-V cpu definition

        Returns:
            Tuple[bytes, List[Relocation]]: the encoded instruction and the list of relocations created during encoding
        """
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

        return (value.to_bytes(cpu.instr_size, const.ENDIANESS), reloc)

