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
# @brief	Type I instruction encoder

#----- imports
from typing import List, Tuple, cast

from packages import constants as const
from packages.structs import EvalResult, Relocation, RelocationType, CPU
from packages.architecture.riscv.opcode import OPCODES_T


#----- functions
def type_I_encoder(opcode: str, operands: List[EvalResult], cpu: CPU, is_uimm: bool = False) -> Tuple[bytes, List[Relocation]]:
        """Encode type I RISC-V instructions

        Args:
            opcode (str): the opcode from the source
            operands (List[EvalResult]): the list of operands
            cpu (CPU): RISC-V cpu definition
            is_uimm (bool): flag for unsigned immediate

        Returns:
            Tuple[bytes, List[Relocation]]: the encoded instruction and the list of relocations created during encoding
        """
        reloc: List[Relocation] = []
        data = OPCODES_T[opcode]

        # special case for CSR encoding
        if opcode.startswith('csr'):
            rd = cast(int, operands[0].reg)
            imm = cast(int, operands[1].value) & 0xFFF
            if opcode.endswith('i'):
                rs1 = cast(int, operands[2].value)
            else:
                rs1 = cast(int, operands[2].reg)

        else:
            # retrieve the operands
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

        return (value.to_bytes(cpu.instr_size, const.ENDIANESS), reloc)
