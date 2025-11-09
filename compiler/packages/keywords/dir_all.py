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
from packages.symbols import Symbols


#----- begin

# generics
Symbols.register(".cpu", TokenType.DIRECTIVE)
Symbols.register(".org", TokenType.DIRECTIVE)
Symbols.register(".text", TokenType.DIRECTIVE)
Symbols.register(".data", TokenType.DIRECTIVE)
Symbols.register(".bss", TokenType.DIRECTIVE)
Symbols.register(".include", TokenType.DIRECTIVE)
Symbols.register(".global", TokenType.DIRECTIVE)

# memory allocation
Symbols.register(".ascii", TokenType.DIRECTIVE)
Symbols.register(".asciiz", TokenType.DIRECTIVE)
Symbols.register(".db", TokenType.DIRECTIVE)
Symbols.register(".dw", TokenType.DIRECTIVE)
Symbols.register(".dl", TokenType.DIRECTIVE)
Symbols.register(".alloc.b", TokenType.DIRECTIVE)
Symbols.register(".alloc.w", TokenType.DIRECTIVE)
Symbols.register(".alloc.l", TokenType.DIRECTIVE)
Symbols.register(".align", TokenType.DIRECTIVE)
Symbols.register(".dup", TokenType.DIRECTIVE)
Symbols.register(".equ", TokenType.DIRECTIVE)

# macros
Symbols.register(".macro", TokenType.DIRECTIVE)
Symbols.register(".endm", TokenType.DIRECTIVE)
Symbols.register(".repeat", TokenType.DIRECTIVE)
Symbols.register(".endr", TokenType.DIRECTIVE)

# "if" and associates
Symbols.register(".if", TokenType.DIRECTIVE)
Symbols.register(".ifdef", TokenType.DIRECTIVE)
Symbols.register(".ifndef", TokenType.DIRECTIVE)
Symbols.register(".else", TokenType.DIRECTIVE)
Symbols.register(".endif", TokenType.DIRECTIVE)

# dollar
Symbols.register("$", TokenType.DOLLAR)
