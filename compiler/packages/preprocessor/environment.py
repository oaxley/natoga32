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
# @brief	Process environment variables in the source code

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

import os

from packages.token import Token, TokenStream, TokenType


#----- functions

def handle_envvar(tokens: List[Token]) -> List[Token]:
    """Process the token list and expand the environment variables

    Args:
        tokens (List[Token]): the list of tokens from the Lexer

    Returns:
        List[Token]: a new list of tokens with the environment variables expanded
    """

    # create a new token stream from the result
    ts = TokenStream(tokens)
    out: List[Token] = []

    while not ts.at_end():
        token = ts.peek()
        if not token:
            break

        # we are interested only in envvar
        if token.type == TokenType.ENVVAR:
            name = token.value[2:-1]
            envvar = os.getenv(name, "")         # set the default value to empty
            print(f"--> {token} | name: {name} | value: '{envvar}'")

            # try to convert it to a number
            try:
                value = int(envvar, 0)
                out.append(Token(TokenType.NUMBER, str(value), token.row, token.col))
            except ValueError:
                out.append(Token(TokenType.STRING, f'"{envvar}"', token.row, token.col))

            ts.advance()
            continue

        # add the token to the list
        out.append(token)
        ts.advance()

    return out
