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
# @brief	TokenType definition

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from enum import IntEnum, auto


#----- classes

class TokenType(IntEnum):
    # control characters
    EOF = auto()                        # end of file
    EOL = auto()                        # end of line '\n'

    # numbers & directives
    NUMBER = auto()                     # any numbers
    DIRECTIVE = auto()                  # assembler directive
    IDENT = auto()                      # identifier
    LABEL = auto()                      # any identifier terminated with ':'
    STRING = auto()                     # a "string"
    CHAR = auto()                       # a single char 'A'
    ENVVAR = auto()

    # comparators
    EQ = auto()                         # '=='
    NEQ = auto()                        # '!='
    GT = auto()                         # '<'
    LT = auto()                         # '>'
    GTE = auto()                        # '>='
    LTE = auto()                        # '<='

    # operators & symbols
    ASSIGN = auto()                     # '='
    LPARENT = auto()                    # '('
    RPARENT = auto()                    # ')'
    COMMA = auto()                      # ','
    PLUS = auto()                       # '+'
    MINUS = auto()                      # '-'
    STAR = auto()                       # '*'
    SLASH = auto()                      # '/'
    MODULO = auto()                     # '%'
    LSHIFT = auto()                     # '<<'
    RSHIFT = auto()                     # '>>'
    AND = auto()                        # '&'
    OR = auto()                         # '|'
    XOR = auto()                        # '^'
    DOLLAR = auto()                     # alias for current PC

    # misc
    PASTE = auto()                      # token pasting '##'
    SKIP = auto()                       # spaces and tabs
    COMMENT = auto()

    # unknown type
    UNKNOWN = auto()
