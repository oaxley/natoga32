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
# @brief	Relocation Engine interface

#----- imports
from typing import Dict
from abc import ABC, abstractmethod

from packages.structs import Section
from packages.classes import SymbolTable


#----- classes
class RelocationEngine(ABC):
    """Relocation Engine abstract interface"""

    def __init__(self) -> None:
        """Constructor"""
        self.cache: Dict[str, int] = {}

    @abstractmethod
    def process(self, sections: Dict[str, Section], symbols: SymbolTable) -> None:
        """Process all the relocations in the '.text' section"""


class DefaultRelocEngine(RelocationEngine):
    """Default Relocation Engine"""

    def __init__(self) -> None:
        super().__init__()

    def process(self, sections: Dict[str, Section], symbols: SymbolTable) -> None:
        pass
