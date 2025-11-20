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
# @brief	Add support for macros in the compiler

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from dataclasses import dataclass

from packages.token import Token, TokenStream, TokenType


#----- globals
MAX_MACRO_EXPANSION_DEPTH = 100


#----- classes

@dataclass
class MacroDefinition:
    name: str
    params: List[str]
    body: List[Token]

class MacroProcessor:
    def __init__(self) -> None:
        self.macros = { }

    def preprocess(self, tokens) -> List[Token]:
        """Pre-process the source file to build macro and expand it"""
        ts = TokenStream(tokens)
        out: List[Token] = []

        while not ts.at_end():
            token = ts.peek()

            if not token:
                raise SyntaxError("EOF encounter during macro processing!")

            # --- macro definition ---
            if token.type == TokenType.DIRECTIVE and token.value == '.macro':
                # parse the macro definition
                macro = self._parse_macro_definition(ts)
                self.macros[macro.name] = macro
                continue

            # --- macro expansion ---
            if token.type == TokenType.IDENT and token.value in self.macros:
                expanded = self._expand_macro(self.macros[token.value], ts)
                expanded = self._expand_tokens_recursive(expanded)
                out.extend(expanded)
                continue

            # --- standard token ---
            out.append(token)
            ts.advance()

        return out


    def _parse_macro_definition(self, ts: TokenStream) -> MacroDefinition:
        """Parse a macro definition and add it to the Macro table"""
        ts.advance()    # consume .macro

        # retrieve the macro name
        token = ts.advance()
        if not token:
            raise SyntaxError("EOF encounter during macro processing!")

        if token.type != TokenType.IDENT:
            raise SyntaxError("Macro name must be an identifier!")

        name = token.value

        # parse the parameters
        params: List[str] = []
        while True:
            token = ts.peek()
            if not token:
                raise SyntaxError("Error while reading macro parameters!")

            if token.type == TokenType.EOL:
                break

            if token.type == TokenType.IDENT:
                params.append(token.value)

            # move forward
            ts.advance()

        # consume EOL
        ts.advance()

        # parse the body until we reach .endm
        body: List[Token] = []
        while True:
            token = ts.peek()
            if token is None:
                raise SyntaxError("Missing .endm for macro definition")

            if token.type == TokenType.DIRECTIVE and token.value == '.endm':
                break

            # add everything to the body, and move forward
            body.append(token)
            ts.advance()

        # consume .endm
        ts.advance()

        # consume EOL if present
        token = ts.peek()
        if token and token.type == TokenType.EOL:
            ts.advance()

        return MacroDefinition(name, params, body)


    def _expand_macro(self, macro: MacroDefinition, ts: TokenStream, depth: int = 0) -> List[Token]:
        """Expand a previously defined macro"""

        if depth > MAX_MACRO_EXPANSION_DEPTH:
            raise SyntaxError("Max macro recursion depth exceeded!")

        ts.advance()        # consume macro name

        # read arguments
        args: List[Token] = []
        while True:
            token = ts.peek()
            if not token:
                raise SyntaxError("Error during macro expansion!")

            if token.type == TokenType.EOL:
                break

            if token.type != TokenType.COMMA:
                args.append(token)

            ts.advance()

        ts.advance()    # consumer EOL

        # ensure we have the same number of parameters
        if len(args) != len(macro.params):
            raise SyntaxError(f"Macro {macro.name}: argument count mismatch")

        # build mapping
        mapping = dict(zip(macro.params, args))

        # first pass
        expanded: List[Token] = []
        for token in macro.body:
            if token.type == TokenType.IDENT and token.value in mapping:
                arg = mapping[token.value]
                expanded.append(Token(arg.type, arg.value, arg.row, arg.col))
            else:
                expanded.append(Token(token.type, token.value, token.row, token.col))

        # second pass
        return self._expand_tokens_recursive(expanded, depth + 1)


    def _expand_tokens_recursive(self, tokens: List[Token], depth: int = 0) -> List[Token]:
        """Expand tokens recursively inside the macro"""
        ts = TokenStream(tokens)

        output: List[Token] = []
        while not ts.at_end():
            token = ts.peek()

            if not token:
                raise SyntaxError("Error during recursive macro expansion!")

            if token.type == TokenType.IDENT and token.value in self.macros:
                expanded = self._expand_macro(self.macros[token.value], ts, depth + 1)
                output.extend(expanded)
                continue

            output.append(token)
            ts.advance()

        return output
