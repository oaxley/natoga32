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
# @brief	Handle conditionals '.if / .ifdef / .ifndef'

#----- imports
from typing import List, Optional

from packages.structs import Token, TokenType
from packages.classes import TokenStream, SymbolTable

from packages.functions import evaluate_expr


#----- functions
def handle_conditionals(ts: TokenStream, symbols: SymbolTable) -> List[Token]:
    """Handle conditional compilation

    Args:
        ts (TokenStream): the current stream of tokens
        symbols (SymbolTable): the symbol table

    Returns:
        List[Token]: a new list with the tokens to consider
    """
    out: List[Token] = []

    # retrieve the conditional directive
    token = ts.advance()
    if not token:
        return out

    directive = token.value
    condition = False

    if directive == '.if':
        # collect the whole expression until EOL
        expr: List[Token] = []
        while True:
            token = ts.peek()
            if not token or token.type == TokenType.EOL:
                break
            expr.append(ts.advance()) # type: ignore

        # remove EOL
        if ts.expect(TokenType.EOL):
            ts.advance()

        # compute the expression
        value = evaluate_expr(expr, symbols)
        condition = bool(value)

    elif directive in ['.ifdef', '.ifndef']:
        # next token must be IDENT
        if not ts.expect(TokenType.IDENT):
            raise SyntaxError(".ifdef/.ifndef requires an IDENT")

        # retrieve the name and check against the symbol table
        name = ts.advance().value   # type: ignore
        is_exist = symbols.exists(name)

        # remove EOL
        if ts.expect(TokenType.EOL):
            ts.advance()

        # invert the condition for ifndef
        condition = is_exist if directive == '.ifdef' else not is_exist

    else:
        return out

    # collect True/False body
    depth = 1
    true_block: List[Token] = []
    false_block: Optional[List[Token]] = None

    while not ts.at_end():
        token = ts.peek()
        if not token:
            break

        if token.type == TokenType.DIRECTIVE:

            # increase depth
            if token.value in ['.if', '.ifdef', '.ifndef']:
                depth += 1
                true_block.append(ts.advance()) # type: ignore
                continue

            # matching .else of the outer .if
            if token.value == '.else' and depth == 1:
                ts.advance()
                false_block = []
                break

            # .endif
            if token.value == '.endif':
                depth -= 1
                ts.advance()

                # remove EOL
                if ts.expect(TokenType.EOL):
                    ts.advance()

                # end of conditional
                if depth == 0:
                    break
                else:
                    true_block.append(token)
                    continue

        true_block.append(ts.advance()) # type: ignore

    # parse the false block if needed
    if false_block is not None:
        depth = 1
        while not ts.at_end():
            token = ts.peek()
            if not token:
                break

            if token.type == TokenType.DIRECTIVE:
                if token.value in ['.if', '.ifdef', '.ifndef']:
                    depth += 1
                    false_block.append(ts.advance())    # type: ignore
                    continue

                if token.value == '.endif':
                    depth -= 1
                    ts.advance()

                    if ts.expect(TokenType.EOL):
                        ts.advance()

                    if depth == 0:
                        break
                    else:
                        false_block.append(token)
                        continue

            false_block.append(ts.advance()) # type: ignore

    # select the proper branch from the condition
    out = true_block if condition else (false_block or [])
    return out
