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
# @brief	RelocationEngine class

#----- imports
from typing import Dict, cast

from packages import ast
from packages import constants as const
from packages.structs import Relocation, RelocationType, Section
from packages.classes import RelocationEngine, SymbolTable


#----- global
Sections = Dict[str, Section]


#----- class
class RiscvRelocEngine(RelocationEngine):
    """Perform the relocation patching of Risc-V instructions"""

    def __init__(self) -> None:
        """Constructor"""
        super().__init__()

    def process(self, sections: Sections, symbols: SymbolTable) -> None:
        """Compute all the relocation for the '.text' section
        Args:
            sections (Dict[str, Section]): the assembly sections
            symbols (SymbolTable): the symbol table
        """
        for rel in sections['.text'].relocations:
            self._compute(rel, sections, symbols)

    def _compute(self, rel: Relocation, sections: Sections, symbols: SymbolTable) -> None:
        """Compute the relocation

        Args:
            rel (Relocation): the relocation to compute
        """
        # we work only on section '.text'
        section = sections['.text']

        # retrieve the symbol details
        symbol = symbols.get(cast(ast.Label, rel.symbol).name)
        assert symbol.value is not None

        # check if the symbol is already in the cache
        if symbol.name not in self.cache:
            address = sections[symbol.section].address     # type: ignore
            self.cache[symbol.name] = address + symbol.value

        # retrieve the symbol address
        sym_addr = self.cache[symbol.name]

        # compute the current PC (respective to the section starting address)
        # and the offset, where the relocation should take place
        offset = rel.address
        instr_pc = section.address + offset

        # retrieve the instruction
        inst = int.from_bytes(section.data[offset:offset+4], const.ENDIANESS)

        # process the relocation
        match rel.type:
            case RelocationType.RISCV_HI20:
                hi20 = self._hi20(sym_addr + rel.addend)
                inst = self._patch_hi20(inst, hi20)

            case RelocationType.RISCV_LO12_I:
                hi20 = self._hi20(sym_addr + rel.addend)
                lo12 = self._lo12(sym_addr + rel.addend, hi20)
                inst = self._patch_lo12_i(inst, lo12)

            case RelocationType.RISCV_LO12_S:
                hi20 = self._hi20(sym_addr + rel.addend)
                lo12 = self._lo12(sym_addr + rel.addend, hi20)
                inst = self._patch_lo12_s(inst, lo12)

            case RelocationType.RISCV_PCREL_HI20:
                delta = sym_addr + rel.addend - instr_pc
                hi20 = self._hi20(delta)
                inst = self._patch_hi20(inst, hi20)

            case RelocationType.RISCV_PCREL_LO12_I:
                delta = sym_addr + rel.addend - instr_pc
                hi20 = self._hi20(delta)
                lo12 = self._lo12(delta, hi20)
                inst = self._patch_lo12_i(inst, lo12)

            case RelocationType.RISCV_PCREL_LO12_S:
                delta = sym_addr + rel.addend - instr_pc
                hi20 = self._hi20(delta)
                lo12 = self._lo12(delta, hi20)
                inst = self._patch_lo12_s(inst, lo12)

            case RelocationType.RISCV_BRANCH:
                delta = sym_addr + rel.addend - instr_pc
                inst = self._patch_branch(inst, delta)

            case RelocationType.RISCV_JAL:
                delta = sym_addr + rel.addend - instr_pc
                inst = self._patch_jal(inst, delta)

            case _:
                raise NotImplementedError(rel.type)

        # replace the instruction
        section.data[offset:offset+4] = inst.to_bytes(4, const.ENDIANESS)

    def _hi20(self, value: int) -> int:
        """Compute HI20 value"""
        return (value + 0x800) >> 12

    def _lo12(self, value:int, hi20:int) -> int:
        """Compute LO12 value"""
        return value - (hi20 << 12)

    def _patch_hi20(self, inst: int, hi20: int) -> int:
        """Patch Type-U instruction"""
        inst &= 0xFFF
        inst |= (hi20 & 0xFFFFF) << 12
        return inst

    def _patch_lo12_i(self, inst: int, imm12: int) -> int:
        """Patch Type-I instruction"""
        inst &= ~(0xFFF << 12)
        inst |= (imm12 & 0xFFF) << 20
        return inst

    def _patch_lo12_s(self, inst: int, imm12: int) -> int:
        """Patch Type-S instruction"""
        inst &= ~((0x7F << 25) | (0x1F << 7))
        inst |= ((imm12 >> 5) & 0x7F) << 25     # imm[11:5] -> bits[31:25]
        inst |= (imm12 & 0x1F) << 7             # imm[4:0] -> bits[11:7]
        return inst

    def _patch_branch(self, inst: int, delta: int) -> int:
        """Patch Type-B instruction"""
        if delta & 0x1:
            raise SyntaxError("Error: branch target is not aligned")

        imm = delta >> 1
        inst &= ~((1 << 31) | (0x3F << 25) | (0xF << 8) | (1 << 7))

        inst |= ((imm >> 11) & 0x01) << 31
        inst |= ((imm >>  5) & 0x3F) << 25
        inst |= ((imm >>  1) & 0x0F) << 8
        inst |= ((imm >> 10) & 0x01) << 7
        return inst

    def _patch_jal(self, inst: int, delta: int) -> int:
        """Patch Type-J instruction"""
        if delta & 0x1:
            raise SyntaxError("Error: jal target is not aligned")

        imm = delta >> 1
        inst &= ~((1 << 31) | (0x3FF << 21) | (1 << 20) | (0xFF << 12))

        inst |= ((imm >> 19 ) & 0x01 ) << 31
        inst |= ((imm >>  9 ) & 0x3FF) << 21
        inst |= ((imm >>  8 ) & 0x01 ) << 20
        inst |=  (imm & 0xFF) << 12

        return inst
