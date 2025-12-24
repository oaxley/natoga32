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
# @brief	Handle '.define' directive

#----- imports
from typing import List

from packages.structs import TokenType, Token, SymbolType
from packages.classes import TokenStream, SymbolTable

from packages.functions import get_value, evaluate_expr


#----- functions
def define(ts: TokenStream, symbols: SymbolTable) -> None:
    """Handle .define directive

    Args:
        ts (TokenStream): the current token stream
        symbols (SymbolTable): the global table of symbols
    """

    ts.advance()        # remove the .define

    # retrieve the name
    name = get_value(ts, TokenType.IDENT)

    # grab all the tokens up to EOL
    tokens: List[Token] = []
    while True:
        token = ts.peek()
        if not token or token.type == TokenType.EOL:
            ts.advance()
            break

        tokens.append(token)
        ts.advance()

    # compute the value
    value = evaluate_expr(tokens, symbols)

    # add the information to the symbol table
    symbols.define(name, value, SymbolType.DEFINE, None)
