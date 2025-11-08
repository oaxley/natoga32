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
from packages.lexer import Lexer


#----- globals

Lexer.register("<<", TokenType.OPERATOR)        # left shift
Lexer.register(">>", TokenType.OPERATOR)        # right shift
Lexer.register("+", TokenType.OPERATOR)         # plus
Lexer.register("-", TokenType.OPERATOR)         # minus
Lexer.register("*", TokenType.OPERATOR)         # multiply
Lexer.register("/", TokenType.OPERATOR)         # divide
Lexer.register("%", TokenType.OPERATOR)         # remainder
Lexer.register("&", TokenType.OPERATOR)         # bitwise and
Lexer.register("|", TokenType.OPERATOR)         # bitwise or
Lexer.register("^", TokenType.OPERATOR)         # bitwise xor
Lexer.register("~", TokenType.OPERATOR)         # bitwise not
