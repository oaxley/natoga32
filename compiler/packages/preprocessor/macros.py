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
# @brief	Handle macros in the source code

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from dataclasses import dataclass

from packages.preprocessor import helper
from packages.token import TokenStream, Token, TokenType


#----- class
@dataclass
class MacroDefinition:
    """Macro definition

    Members:
    - name (str): the name of the macro
    - params (List[str]): the list of parameters for this macro
    - body (List[Token]): the list of tokens representing the body of the macro
    """
    name: str
    params: List[str]
    body: List[Token]


#----- functions
def parse_macro_definition(ts: TokenStream) -> MacroDefinition:
    """Parse a macro definition and add it to the Macro table

    Args:
        ts (TokenStream): the token stream instance

    Returns
        MacroDefinition: an instance of MacroDefinition representing the macro
    """
    ts.advance()    # consume .macro

    # retrieve the macro name
    name = helper.get_value(ts, TokenType.IDENT)

    # parse the parameters
    params: List[str] = []
    while True:
        token = ts.peek()
        if not token or token.type == TokenType.EOL:
            ts.advance()        # consumer EOL
            break

        if token.type == TokenType.IDENT:
            params.append(token.value)

        # move forward
        ts.advance()

    # parse the body until we reach .endm
    body = helper.capture_body(ts, '.macro', '.endm')

    return MacroDefinition(name, params, body)


def expand_macro(macro: MacroDefinition, ts: TokenStream) -> List[Token]:
    """Expand macro invocation at current stream position

    Args:
        macro (MacroDefinition): the macro detected during processing
        ts (TokenStream): the current token stream

    Returns:
        List[Token]: A list of tokens after the macro has been expanded
    """
    # consume the macro name (MacroDefinition is already properly populated)
    ts.advance()

    # parse the parameters. if EOL, means no parameters.
    params: List[List[Token]] = []
    if ts.peek() and ts.peek().type == TokenType.EOL:   # type: ignore
        ts.advance()    # consume EOL

    else:
        # we are trying to see if the macro parameters are simple tokens
        # or a complete expression that will have to be expanded later
        while not ts.at_end():
            # collect tokens until COMMA or EOL
            arg_tokens: List[Token] = []
            while not ts.at_end():
                token = ts.peek()
                if token and token.type in [TokenType.COMMA, TokenType.EOL]:
                    break

                arg_tokens.append(ts.advance())     # type: ignore

            params.append(arg_tokens)

            # consume COMMA if present
            if ts.peek() and ts.peek().type == TokenType.COMMA:     # type: ignore
                ts.advance()
                continue

            # we have reach EOL
            if ts.peek() and ts.peek().type == TokenType.EOL:       # type: ignore
                ts.advance()

            break

    # Map parameters -> token lists (fill with empty lists if fewer args)
    mapping: Dict[str, List[Token]] = {}
    for i, pname in enumerate(macro.params):
        mapping[pname] = params[i] if i < len(params) else []

    # expand the macro body with substitution
    expanded: List[Token] = []
    for bt in macro.body:
        # token is a parameter
        if bt.type == TokenType.IDENT and bt.value in mapping:
            for t in mapping[bt.value]:
                expanded.append(helper.clone_token(t))
            continue

        # normal token
        expanded.append(helper.clone_token(bt))

    return expanded
