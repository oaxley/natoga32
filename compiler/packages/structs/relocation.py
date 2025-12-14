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
# @brief	Relocation dataclass

#----- imports
from dataclasses import dataclass

from packages import ast

#----- class
@dataclass
class Relocation:
    """Class to record the relocation needed during compilation

    Members:
    - type (str): the type of relocation needed
    - symbol (str): the symbol name
    - addend (int): signed addend
    - address (int): address where the relocation should take place
    - mask (int): the bit mask to apply to retrieve the value
    """
    type: str
    symbol: ast.Node
    addend: int = 0
    address: int = 0
    mask: int = 0
