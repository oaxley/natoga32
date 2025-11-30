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
# @brief	X68FP Architecture

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Tuple

from .interface import Architecture, Encoder, CPU


#----- classes

class X68FPEncoder(Encoder):
    """X68FP encoder"""

    def __init__(self) -> None:
        super().__init__()

class X68fp(Architecture):
    """X68FP Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        self.encoder: Encoder = X68FPEncoder()
        self.config: CPU = CPU(2, 1, 1, 2)

    def is_register(self, operand: str) -> Tuple[bool, int]:
        """Check if operand is a register"""
        return (False, 0)
