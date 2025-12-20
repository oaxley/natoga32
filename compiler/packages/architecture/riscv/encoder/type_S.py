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
# @brief	Type S instruction encoder

#----- imports
from typing import List, Tuple, cast

from packages import constants as const
from packages.structs import EvalResult, Relocation, RelocationType, CPU
from packages.architecture.riscv.opcode import OPCODES_T


#----- functions
def type_S_encoder(opcode: str, operands: List[EvalResult], cpu: CPU) -> Tuple[bytes, List[Relocation]]:
        """Encode type S RISC-V instructions

        Args:
            opcode (str): the opcode from the source
            operands (List[EvalResult]): the list of operands
            cpu (CPU): RISC-V cpu definition

        Returns:
            Tuple[bytes, List[Relocation]]: the encoded instruction and the list of relocations created during encoding
        """
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

        return (value.to_bytes(cpu.instr_size, const.ENDIANESS), reloc)
