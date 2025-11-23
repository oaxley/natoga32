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
# @brief	Handle token pasting '##'

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.token import Token, TokenStream, TokenType


#----- functions

def apply_token_pasting(tokens: List[Token]) -> List[Token]:
    """Process token pasting operator '##' to merge two tokens together

    Args:
        tokens (List[Token]): the tokens representing the pasting operation (i##k)

    Returns:
        List[Token]: a new list of tokens to replace the previous pasting operator
    """
    # --- process pasting
    pasted: List[Token] = []
    i = 0
    while i < len(tokens):
        token = tokens[i]

        # look for IDENT, PASTE, IDENT (x##i)
        if (
            i + 2 < len(tokens)
            and token.type == TokenType.IDENT
            and tokens[i+1].type == TokenType.PASTE
            and tokens[i+2].type == TokenType.NUMBER
        ):
            # concatenate the left and right side
            new_value = token.value + tokens[i+2].value

            # build a new IDENT token with the same position / info that the one on the left
            pasted.append(Token(TokenType.IDENT, new_value, token.row, token.col))

            # move forward
            i += 3
            continue

        # otherwise keep adding the tokens
        pasted.append(token)
        i += 1

    return pasted
