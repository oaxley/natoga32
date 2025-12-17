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
# @brief	Lexer regular expression for token extraction

#----- imports
import re

from .token_type import TokenType


#----- begin

# regular expressions to extract tokens from source code
regexp_table = [
    (TokenType.NUMBER.name, r'0x[a-zA-Z0-9_]+|0b[01_]+|\b\d+\b'),
    (TokenType.DIRECTIVE.name, r'\.[A-Za-z_][A-Za-z0-9_]*'),
    (TokenType.LABEL.name, r'[A-Za-z_][A-Za-z0-9_]*:'),
    (TokenType.IDENT.name, r'[A-Za-z_][A-Za-z0-9_\-\.]*'),
    (TokenType.ENVVAR.name, r'\$\[[^\]]*\]'),
    (TokenType.STRING.name, r'\"[^\"]*\"'),
    (TokenType.CHAR.name, r'\'.\''),

    # memory relocation operator
    (TokenType.M_ABS_LO.name, r'%lo'),
    (TokenType.M_ABS_HI.name, r'%hi'),
    (TokenType.M_PCREL_LO.name, r'%pcrel_lo'),
    (TokenType.M_PCREL_HI.name, r'%pcrel_hi'),

    # equal
    (TokenType.EQ.name, r'=='),

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
    (TokenType.COMMENT.name, r'[;#][^\n]*'),
]

# merge all the tokens under the '(?P<name>...)' pattern for re
# https://docs.python.org/3/library/re.html
all_tokens = '|'.join(f'(?P<{value}>{pattern})' for value, pattern in regexp_table)

# compute the regexp expression for fast matching
re_patterns = re.compile(all_tokens)
