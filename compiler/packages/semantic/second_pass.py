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
# @brief	Semasntic Analyzer second pass

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages import data_classes as dc

#----- class

class SecondPass:
    """Semantic Analyzer - Second Pass"""

    def __init__(self, data: dc.SAData) -> None:
        """Constructor

        Args:
            data (dc.SAData): the semantic analyzer structure
        """
        self.data = data

    def process(self) -> None:
        """Execute the second pass"""
