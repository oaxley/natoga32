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
# @brief	Function capture_body

#----- imports
from typing import List

from packages.structs import Token, TokenType
from packages.classes import TokenStream

from packages.constants import MAX_BODY_DEPTH


#----- function

def capture_body(ts: TokenStream, begin: str, end: str) -> List[Token]:
    """Parse the token stream and extract the body between 'begin' and 'end' tokens

    Args:
        ts (TokenStream): the stream of tokens
        begin (str): value for the beginning body marker
        end (str): value for the end body marker

    Returns:
        List[Token]: the list of tokens representing the body
    """
    body: List[Token] = []
    depth = 1       # depth tracker

    while True:
        token = ts.peek()

        if token is None:
            break

        # depth tracker (for nested bodies)
        if token.type == TokenType.DIRECTIVE and token.value == begin:
            depth += 1
            if depth > MAX_BODY_DEPTH:
                raise SyntaxError("Nested body max limit exceeded!")

        if token.type == TokenType.DIRECTIVE and token.value == end:
            depth -= 1
            if depth == 0:
                break

        # add everything to the body, and move forward
        body.append(token)
        ts.advance()

    # consume "end" marker & EOL
    ts.advance()
    if ts.expect(TokenType.EOL): ts.advance()

    # return the body
    return body
