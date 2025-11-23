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
# @brief	Handle DUP macro

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.token import Token, TokenType
from packages.symbols import SymbolTable

from . import helper


#----- functions
def apply_dup(tokens: List[Token], symbols: SymbolTable) -> List[Token]:
    """Expands 'VALUE DUP(x)' pattern, which repeats x times the VALUE

    Args:
        tokens (List[Token]): the list of tokens from the Lexer

    Returns:
        List[Token]: a new list of tokens to replace the DUP sequence
    """
    result: List[Token] = []
    i = 0

    while i < len(tokens):
        token = tokens[i]

        # look for pattern VALUE DUP( expr )
        if (
            i + 3 < len(tokens)
            and tokens[i+1].type == TokenType.IDENT
            and tokens[i+1].value.upper() == 'DUP'
            and tokens[i+2].type == TokenType.LPARENT
        ):
            # find the matching RPARENT
            j = i + 3
            depth = 1
            while (j < len(tokens)) and (depth > 0):
                if tokens[j].type == TokenType.LPARENT:
                    depth += 1
                if tokens[j].type == TokenType.RPARENT:
                    depth -= 1
                j += 1

            if depth != 0:
                raise SyntaxError("Unmatched parenthesis in DUP expression")

            # retrieve all the tokens from the expression and evaluate it
            expr = tokens[i+3:j-1]
            count = helper.evaluate_expr(expr)

            if count < 0:
                raise SyntaxError("DUP count must be non-negative")

            # check if the token is in the symbol table
            if symbols.exists(token.value):
                value = Token(TokenType.NUMBER, symbols.value(token.value), token.row, token.col)
            else:
                value = helper.clone_token(token)

            # replace the whole expression
            for _ in range(count):
                result.append(value)
                result.append(Token(TokenType.COMMA, ",", token.row, token.col))

            # move forward
            i = j
            continue

        # keep adding normal tokens
        result.append(token)
        i += 1

    return result
