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
# @brief	Temporary file to store all the assembler directives

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.token import TokenType
from packages.lexer import Lexer


#----- begin

# generics
Lexer.register(".cpu", TokenType.DIRECTIVE)
Lexer.register(".org", TokenType.DIRECTIVE)
Lexer.register(".text", TokenType.DIRECTIVE)
Lexer.register(".data", TokenType.DIRECTIVE)
Lexer.register(".bss", TokenType.DIRECTIVE)
Lexer.register(".include", TokenType.DIRECTIVE)
Lexer.register(".global", TokenType.DIRECTIVE)

# memory allocation
Lexer.register(".ascii", TokenType.DIRECTIVE)
Lexer.register(".asciiz", TokenType.DIRECTIVE)
Lexer.register(".db", TokenType.DIRECTIVE)
Lexer.register(".dw", TokenType.DIRECTIVE)
Lexer.register(".dl", TokenType.DIRECTIVE)
Lexer.register(".alloc.b", TokenType.DIRECTIVE)
Lexer.register(".alloc.w", TokenType.DIRECTIVE)
Lexer.register(".alloc.l", TokenType.DIRECTIVE)
Lexer.register(".align", TokenType.DIRECTIVE)
Lexer.register(".dup", TokenType.DIRECTIVE)
Lexer.register(".equ", TokenType.DIRECTIVE)

# macros
Lexer.register(".macro", TokenType.DIRECTIVE)
Lexer.register(".endm", TokenType.DIRECTIVE)
Lexer.register(".repeat", TokenType.DIRECTIVE)
Lexer.register(".endr", TokenType.DIRECTIVE)

# "if" and associates
Lexer.register(".if", TokenType.DIRECTIVE)
Lexer.register(".ifdef", TokenType.DIRECTIVE)
Lexer.register(".ifndef", TokenType.DIRECTIVE)
Lexer.register(".else", TokenType.DIRECTIVE)
Lexer.register(".endif", TokenType.DIRECTIVE)

# dollar
Lexer.register("$", TokenType.DOLLAR)
