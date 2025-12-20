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
# @brief	Type U instruction encoder

#----- imports
from typing import List, Tuple, cast

from packages import constants as const
from packages.structs import EvalResult, Relocation, RelocationType, CPU
from packages.architecture.riscv.opcode import OPCODES_T


#----- functions
def type_U_encoder(opcode: str, operands: List[EvalResult], cpu: CPU) -> Tuple[bytes, List[Relocation]]:
        """Encode type U RISC-V instructions

        Args:
            opcode (str): the opcode from the source
            operands (List[EvalResult]): the list of operands
            cpu (CPU): RISC-V cpu definition

        Returns:
            Tuple[bytes, List[Relocation]]: the encoded instruction and the list of relocations created during encoding
        """
        reloc: List[Relocation] = []

        # retrieve the operands
        rd = cast(int, operands[0].reg)

        if operands[1].value is not None:
            imm = (operands[1].value >> 12) & 0xFFFFF
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

        return (value.to_bytes(cpu.instr_size, const.ENDIANESS), reloc)
