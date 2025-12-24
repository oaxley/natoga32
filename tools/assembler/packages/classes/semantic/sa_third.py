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

    def process(self) -> None:
        """Process the relocations from the .text section"""
        engine = self.config.arch.reloc_engine
        engine.process(self.config.sections, self.config.symbols)
