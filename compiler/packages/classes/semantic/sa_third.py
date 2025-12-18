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
# @brief	Semantic Analysis - Third pass

#----- imports
from typing import cast, Dict

from packages import ast
from packages.structs import Config


#----- class
class SAThirdPass:
    """Semantic Analyzer Third Pass"""

    def __init__(self, config: Config) -> None:
        """Constructor"""
        self.config = config
        self.cache: Dict[str, int] = {}

    def process(self) -> None:
        """Process the relocations from the .text section"""
        text = self.config.sections['.text']

        for reloc in text.relocations:
            # symbol details
            symbol = self.config.symbols.get(cast(ast.Label, reloc.symbol).name)

            # check if the symbol is already in the cache
            if symbol.name not in self.cache:
                self.cache[symbol.name] = self.config.arch.locations[symbol.section] + symbol.value # type: ignore

            # retrieve the symbol address and the offset of the instruction
            sym_addr = self.cache[symbol.name]
            offset = reloc.address

            # delta value
            delta = (sym_addr + reloc.addend) - offset

            # current instruction
            instr = int.from_bytes(text.data[offset:offset+4], "big")

            if reloc.type == 'R_RISCV_PCREL_HI20':
                # compute the upper 20-bits
                hi20 = (delta + 0x800) >> 12

                # patch the instruction
                instr = (instr & 0xFFF) | ((hi20 & 0xFFFFF) << 12)
                text.data[offset:offset+4] = instr.to_bytes(4, 'big')

            elif reloc.type == 'R_RISCV_PCREL_LO12_I':
                lo12 = delta - ((delta + 0x800) & ~0xFFF)

                # patch the instruction
                instr &= ~(0xFFF << 20)
                instr |= (lo12 & 0xFFF) << 20
                text.data[offset:offset+4] = instr.to_bytes(4, 'big')

            elif reloc.type == 'R_RISCV_PCREL_LO12_S':
                lo12 = delta - ((delta + 0x800) & ~0xFFF)

                # patch
                instr &= ~((0x7F << 25) | (0x1F << 7))
                instr |= ((lo12 >> 5) & 0x7F) << 25     # imm[11:5] -> bits[31:25]
                instr |= (lo12 & 0x1F) << 7             # imm[4:0] -> bits[11:7]
                text.data[offset:offset+4] = instr.to_bytes(4, 'big')

            elif reloc.type == 'R_RISCV_HI20':
                delta = (sym_addr + 0x800) >> 12
                # patch the instruction
                instr = (instr & 0xFFF) | ((delta & 0xFFFFF) << 12)
                text.data[offset:offset+4] = instr.to_bytes(4, 'big')


            else:
                # R_RISCV_HI20
                # R_RISCV_LO12_I
                # R_RISCV_LO12_S
                # R_RISCV_BRANCH
                # R_RISCV_JAL
                print(f"Error: unsupported relocation type '{reloc.type}'")

