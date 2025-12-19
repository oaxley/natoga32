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
from packages.structs import Config, Relocation


#----- class
class RelocationEngine:
    """Perform the relocation patching of Risc-V instructions"""

    def __init__(self, config: Config) -> None:
        """Constructor

        Args:
            config (Config): the global config object containing the relocations
        """
        self.section = config.sections['.text']
        self.symbols = config.symbols
        self.address = config.arch.locations['.text']
        self.cache: Dict[str, int] = {}

    def process(self) -> None:
        """Compute all the relocation for the '.text' section"""
        for rel in self.section.relocations:
            self._compute(rel)

    def _compute(self, rel: Relocation) -> None:
        """Compute the relocation

        Args:
            rel (Relocation): the relocation to compute
        """
        # retrieve the symbol details
        symbol = self.symbols.get(cast(ast.Label, rel.symbol).name)

        # check if the symbol is already in the cache
        if symbol.name not in self.cache:
            # compute the address
            self.cache[symbol.name] = self.address + symbol.value   # type: ignore

        # retrieve the symbol address & offset where relocation should happen
        sym_addr = self.cache[symbol.name]
        offset = rel.address

        # retrieve the instruction
        inst = int.from_bytes(self.section.data[offset:offset+4], "big")

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
