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

from packages.token import TokenType


#----- globals
TOKENS_SPECS = [
    (TokenType.NUMBER.name, r'0x[a-zA-Z0-9_]+|0b[01_]+|\b\d+\b'),
    (TokenType.DIRECTIVE.name, r'\.[A-Za-z_][A-Za-z0-9_]*'),
    (TokenType.LABEL.name, r'[A-Za-z_][A-Za-z0-9_]*:'),
    (TokenType.IDENT.name, r'[A-Za-z_][A-Za-z0-9_-]*'),
    (TokenType.ENVVAR.name, r'\$\[[^\]]*\]'),
    (TokenType.STRING.name, r'\"[^\"]*\"'),
    (TokenType.CHAR.name, r'\'.\''),

    # '<' '<<' operators
    (TokenType.LSHIFT.name, r'<<'),
    (TokenType.LTE.name, r'<='),
    (TokenType.LT.name, r'<'),

    (TokenType.RSHIFT.name, r'>>'),
    (TokenType.GTE.name, r'>='),
    (TokenType.GT.name, r'>'),

    # operators & symbols
    (TokenType.ASSIGN.name, r'='),
    (TokenType.LPARENT.name, r'\('),
    (TokenType.RPARENT.name, r'\)'),
    (TokenType.COMMA.name, r'\,'),
    (TokenType.PLUS.name, r'\+'),
    (TokenType.MINUS.name, r'-'),
    (TokenType.STAR.name, r'\*'),
    (TokenType.SLASH.name, r'/'),
    (TokenType.MODULO.name, r'%'),
    (TokenType.AND.name, r'&'),
    (TokenType.OR.name, r'\|'),
    (TokenType.XOR.name, r'\^'),
    (TokenType.DOLLAR.name, r'\$'),



    # misc
    (TokenType.PASTE.name, r'##'),
    (TokenType.SKIP.name, r'[ \t]+'),
    (TokenType.COMMENT.name, r';[^\n]*'),
]

# compile the regexp
ALL_TOKENS = '|'.join(f'(?P<{value}>{pattern})' for value, pattern in TOKENS_SPECS)
RE_PATTERNS = re.compile(ALL_TOKENS)

# operators precedence
# "Please Excuse My Dear Aunt Sally"
OP_INFO = {
    '*' : (6, TokenType.STAR),
    '/' : (6, TokenType.SLASH),
    '%' : (6, TokenType.MODULO),
    '+' : (5, TokenType.PLUS),
    '-' : (5, TokenType.MINUS),
    '<<': (4, TokenType.LSHIFT),
    '>>': (4, TokenType.RSHIFT),
    '&' : (3, TokenType.AND),
    '^' : (2, TokenType.XOR),
    '|' : (1, TokenType.OR)
}
