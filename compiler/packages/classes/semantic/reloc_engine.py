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

