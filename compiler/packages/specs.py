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
# @brief	Token Specifications as regexp

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

import re
from enum import IntEnum, auto

#----- class
# token types
class TokenType(IntEnum):
    # control characters
    EOF = auto()                        # end of file
    EOL = auto()                        # end of line '\n'

    # numbers & directives
    NUMBER = auto()                     # any numbers
    DIRECTIVE = auto()                  # assembler directive
    IDENT = auto()                      # identifier
    LABEL = auto()                      # any identifier terminated with ':'
    MACRO = auto()                      # assembler macro
    FUNCTION = auto()                   # macro function
    STRING = auto()                     # a "string"
    CHAR = auto()                       # a single char 'A'

    # operators & symbols
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
    QUOTE = auto()                      # '"'
    DOLLAR = auto()                     # alias for current PC

    # misc
    SKIP = auto()                       # spaces and tabs
    COMMENT = auto()

    UNKNOWN = auto()                    # unknown word


#----- globals
TOKENS_SPECS = [
    (TokenType.NUMBER.name, r'0x[a-zA-Z0-9_]+|0b[01_]+|\b\d+\b'),
    (TokenType.DIRECTIVE.name, r'\.[A-Za-z_][A-Za-z0-9_]*'),
    (TokenType.FUNCTION.name, r'[A-Za-z_][A-Za-z0-9_]*\('),
    (TokenType.LABEL.name, r'[A-Za-z_][A-Za-z0-9_]*:'),
    (TokenType.IDENT.name, r'[A-Za-z_][A-Za-z0-9_-]*'),
    (TokenType.MACRO.name, r'\%[A-Za-z_]+'),
    (TokenType.STRING.name, r'\"[^\"]*\"'),
    (TokenType.CHAR.name, r'\'.\''),


    # operators & symbols
    (TokenType.LPARENT.name, r'\('),
    (TokenType.RPARENT.name, r'\)'),
    (TokenType.COMMA.name, r'\,'),
    (TokenType.PLUS.name, r'\+'),
    (TokenType.MINUS.name, r'-'),
    (TokenType.STAR.name, r'\*'),
    (TokenType.SLASH.name, r'/'),
    (TokenType.MODULO.name, r'%'),
    (TokenType.LSHIFT.name, r'<<'),
    (TokenType.RSHIFT.name, r'>>'),
    (TokenType.AND.name, r'&'),
    (TokenType.OR.name, r'\|'),
    (TokenType.XOR.name, r'\^'),
    (TokenType.QUOTE.name, r'\"'),
    (TokenType.DOLLAR.name, r'\$'),

    # misc
    (TokenType.SKIP.name, r'[ \t]+'),
    (TokenType.COMMENT.name, r';[^\n]*'),
]

# compile the regexp
ALL_TOKENS = '|'.join(f'(?P<{value}>{pattern})' for value, pattern in TOKENS_SPECS)
RE_PATTERNS = re.compile(ALL_TOKENS)
