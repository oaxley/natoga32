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
# @brief	Type R instruction encoder

#----- imports
from typing import List, Tuple, cast

from packages import constants as const
from packages.structs import EvalResult, Relocation, RelocationType, CPU
from packages.architecture.riscv.opcode import OPCODES_T


#----- functions
def type_R_encoder(opcode: str, operands: List[EvalResult], cpu: CPU) -> Tuple[bytes, List[Relocation]]:
        """Encode type R RISC-V instructions

        Args:
            opcode (str): the opcode from the source
            operands (List[EvalResult]): the list of operands
            cpu (CPU): RISC-V cpu definition

        Returns:
            Tuple[bytes, List[Relocation]]: the encoded instruction and the list of relocations created during encoding
        """
        data = OPCODES_T[opcode]

        # retrieve the operands
        xlen = cpu.xlen - 1
        rd = cast(int, operands[0].reg) & xlen
        rs1 = cast(int, operands[1].reg) & xlen

        if len(operands) == 3:
            if operands[2].reg is not None:
                rs2 = operands[2].reg & xlen
            elif operands[2].value is not None:
                rs2 = operands[2].value & xlen
            else:
                raise SyntaxError("Error: Type R instruction does not support relocation")
        else:
            rs2 = data.rs2

        funct3 = data.funct3
        funct7 = data.funct7
        value = (
            (funct7 << 25) |
            (rs2    << 20) |
            (rs1    << 15) |
            (funct3 << 12) |
            (rd     << 7 ) |
            data.opcode
        )

        return (value.to_bytes(cpu.instr_size, const.ENDIANESS), [])

