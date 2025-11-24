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
# @brief	Handle for loops

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.token import Token, TokenStream, TokenType

from . import helper


#----- functions
def handle_for_loop(ts: TokenStream) -> List[Token]:
    """Handle '.for/.endf' directives

    Args:
        ts (TokenStream): the current token stream

    Returns:
        List[Token]: a list of tokens to insert at the ".for" loop position
    """
    ts.advance()    # consume .for

    # parse <IDENT> = <NUMBER>, <NUMBER>
    var_name = helper.get_value(ts, TokenType.IDENT)
    helper.get_value(ts, TokenType.ASSIGN)
    start_value = int(helper.get_value(ts, TokenType.NUMBER), 0)
    ts.advance()
    end_value = int(helper.get_value(ts, TokenType.NUMBER), 0)

    # 6. Optional step
    step_value = 1
    if ts.expect(TokenType.COMMA):
        ts.advance()    # remove comma

        # detect negative/positive sign if any
        sign = 1
        if ts.expect(TokenType.MINUS):
            sign = -1
            ts.advance()
        elif ts.expect(TokenType.PLUS):
            ts.advance()

        # retrieve the step value
        step_value = sign * int(helper.get_value(ts, TokenType.NUMBER), 0)
        if step_value == 0:
            raise SyntaxError("Step value cannot be 0")

    # consume EOL if present
    if ts.expect(TokenType.EOL): ts.advance()

    # --- capture body
    body = helper.capture_body(ts, '.for', '.endf')

    # --- body expansion
    tokens: List[Token] = []
    current = start_value

    # loop condition
    if step_value > 0:
        cond = lambda c: c < end_value
    else:
        cond = lambda c: c > end_value

    while cond(current):

        iteration_body: List[Token] = []
        for token in body:
            # substitute var_name with NUMBER token
            if token.type == TokenType.IDENT and token.value == var_name:
                iteration_body.append(Token(TokenType.NUMBER, str(current), token.row, token.col))
            else:
                # clone the token to avoid messing up with the loop body
                iteration_body.append(helper.clone_token(token))

        # add those tokens to the global list
        tokens.extend(iteration_body)

        # current = current +/- step_val
        current += step_value

    return tokens
