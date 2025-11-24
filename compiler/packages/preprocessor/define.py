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
from __future__ import annotations
from typing import Any, Dict, List

from packages.token import Token, TokenStream, TokenType
from packages.symbols import SymbolTable

from . import helper


#----- functions
def handle_define(ts: TokenStream, symbols: SymbolTable) -> None:
    """Handle .define directive

    Args:
        ts (TokenStream): the current token stream
        symbols (SymbolTable): the global table of symbols
    """

    ts.advance()        # remove the .define

    # retrieve the name
    name = helper.get_value(ts, TokenType.IDENT)

    # grab all the tokens up to EOL
    tokens: List[Token] = []
    while True:
        token = ts.peek()
        if not token or token.type == TokenType.EOL:
            break

        tokens.append(token)
        ts.advance()

    # compute the value
    value = helper.evaluate_expr(tokens, symbols)

    # add the information to the symbol table
    symbols.add(name, str(value))
