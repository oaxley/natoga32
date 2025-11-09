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
# @brief	List of all available operators

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.token import TokenType
from packages.symbols import Symbols


#----- globals

Symbols.register("<<", TokenType.OPERATOR)        # left shift
Symbols.register(">>", TokenType.OPERATOR)        # right shift
Symbols.register("+", TokenType.OPERATOR)         # plus
Symbols.register("-", TokenType.OPERATOR)         # minus
Symbols.register("*", TokenType.OPERATOR)         # multiply
Symbols.register("/", TokenType.OPERATOR)         # divide
Symbols.register("%", TokenType.OPERATOR)         # remainder
Symbols.register("&", TokenType.OPERATOR)         # bitwise and
Symbols.register("|", TokenType.OPERATOR)         # bitwise or
Symbols.register("^", TokenType.OPERATOR)         # bitwise xor
Symbols.register("~", TokenType.OPERATOR)         # bitwise not
