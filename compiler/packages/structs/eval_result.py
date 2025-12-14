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
# @brief	EvalResult dataclass

#----- imports
from typing import Optional
from dataclasses import dataclass

from .relocation import Relocation


#----- class
@dataclass
class EvalResult:
    """Result of a constant evaluation, that can be either a number or a relocation

    Members:
    - value (Optional[int]): an integer if the value is known, None otherwise
    - reloc (Optional[Relocation]): a relocation structure if value cannot be computed
    - reg (Optional[int]): an integer representing the register number
    """
    value: Optional[int] = None
    reloc: Optional[Relocation] = None
    reg: Optional[int] = None
