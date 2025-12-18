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
# @brief	Semantic Analyzer

#----- imports
from typing import Any, Dict, List

from packages.structs import Config
from packages.classes.semantic.sa_first import SAFirstPass
from packages.classes.semantic.sa_second import SASecondPass
from packages.classes.semantic.sa_third import SAThirdPass


#----- class
class SemanticAnalyzer:
    """Semantic Analyzer"""

    def __init__(self, config: Config) -> None:
        """Constructor"""
        self.config = config

    def first_pass(self) -> None:
        """Execute the semantic analysis first pass"""
        obj = SAFirstPass(self.config)
        obj.process()

    def second_pass(self) -> None:
        """Execute the semantic analysis second pass"""
        obj = SASecondPass(self.config)
        obj.process()

    def third_pass(self) -> None:
        """Execute the semantic analysis third pass"""
        obj = SAThirdPass(self.config)
        obj.process()
