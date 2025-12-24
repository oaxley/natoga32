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
# @brief	Manage environment variables during pre-processing

#----- imports
from typing import List

import os

from packages.structs import Token, TokenType
from packages.classes import TokenStream


#----- functions
def envvar(tokens: List[Token]) -> List[Token]:
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
