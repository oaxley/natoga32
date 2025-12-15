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
# @brief	MacroDefinition dataclass

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from dataclasses import dataclass

from .token import Token


#----- class
@dataclass
class MacroDefinition:
    """Macro definition

    Members:
    - name (str): the name of the macro
    - params (List[str]): the list of parameters for this macro
    - body (List[Token]): the list of tokens representing the body of the macro
    """
    name: str
    params: List[str]
    body: List[Token]
