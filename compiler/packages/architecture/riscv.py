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
# @brief	RISC-V architecture

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from .interface import Encoder, Architecture, CPU

#----- class

class RiscVEncoder(Encoder):
    """RISC-V encoder"""

    def __init__(self) -> None:
        """Constructor"""
        super().__init__()

class Riscv(Architecture):
    """RISC-V Architecture"""

    def __init__(self) -> None:
        """Constructor"""
        self.encoder: Encoder = RiscVEncoder()
        self.config: CPU = CPU(4, 1, 2, 4)
