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
# @brief	Handle .dup macro

#----- imports
from typing import List

from packages.structs import Token, TokenType
from packages.classes import SymbolTable

from packages.functions import evaluate_expr, clone_token


#----- functions
def apply_dup(tokens: List[Token], symbols: SymbolTable) -> List[Token]:
    """Expand the .dup directive

    Args:
        tokens (List[Token]): the input list of tokens from the Lexer
        symbols (SymbolTable): the symbol table

    Returns:
        List[Token]: the new list of tokens with the '.dup' expanded
    """
    result: List[Token] = []
    i = 0
    while i < len(tokens):
        token = tokens[i]

        if (token.type == TokenType.DIRECTIVE) and (token.value.upper() == '.DUP'):
            # expression
            if tokens[i+1].type == TokenType.LPARENT:
                # find the matching pair
                j = i+2
                depth = 1
                while (j < len(tokens)) and (depth > 0):
                    if tokens[j].type == TokenType.LPARENT:
                        depth += 1
                    if tokens[j].type == TokenType.RPARENT:
                        depth -= 1
                    j += 1

                if depth != 0:
                    raise SyntaxError("Error: unmatched parenthesis in .dup expression")

                # compute the expression
                expr = tokens[i+2:j-1]
                count = evaluate_expr(expr, symbols)

            elif tokens[i+1].type == TokenType.NUMBER:
                count = int(tokens[i+1].value)
                j = i+2

            else:
                raise SyntaxError("Error: wrong expression type in .dup directive")

            if count < 0:
                raise SyntaxError("Error: negative count in .dup expression")

            # retrieve the symbol to duplicate
            token = tokens[j]

            # check the symbol table for this token
            if symbols.exists(token.value):
                symbol = symbols.get(token.value)
                value = Token(TokenType.NUMBER, str(symbol.value), token.row, token.col)
            else:
                value = clone_token(token)

            # replace the initial expression
            comma = Token(TokenType.COMMA, ",", token.row, token.col)
            for _ in range(count):
                result += [ value, comma ]

            i = j + 1
            continue

        # add other tokens without modifying them
        result.append(token)
        i += 1

    return result
