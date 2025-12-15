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
# @brief	PreProcessor class

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.structs import (
    Token, TokenType, SymbolType,
    Config, MacroDefinition
)
from packages.classes import TokenStream, SymbolTable

from packages.functions import (
    envvar, define, parse_macro_definition, expand_macro,
    apply_dup, apply_token_pasting, include, handle_conditionals, handle_for_loop
)


#----- class
class PreProcessor:
    """Compiler pre-processor"""

    def __init__(self, config: Config) -> None:
        """Constructor"""
        self.config = config
        self.macros: Dict[str, MacroDefinition] = {}


    def process(self) -> List[Token]:
        """Perform the pre-processing of the source file

        Returns:
            List[Token]: the new list of tokens after pre-processing
        """
        out = self._process_tokens(self.config.tokens, self.config.symbols)
        return out


    def _process_tokens(self, tokens: List[Token], symbols: SymbolTable) -> List[Token]:
        """Process the tokens recursively, expand macros, conditionals and symbols

        Args:
            tokens (List[Token]): the input list of tokens
            symbols (SymbolTable): the symbol table

        Returns:
            List[Token]: the list of tokens expanded
        """
        # process the environment variables expansion
        tmp = envvar(tokens)

        # create a new token stream with the result
        ts = TokenStream(tmp)
        out: List[Token] = []

        # parse all the tokens one by one
        while not ts.at_end():
            token = ts.peek()
            if not token:
                break

            # take care of directives
            if token.type == TokenType.DIRECTIVE:

                # --- .macro
                if token.value == '.macro':
                    macro = parse_macro_definition(ts)
                    self.macros[macro.name] = macro
                    symbols.define(macro.name, None, SymbolType.MACRO, None)
                    continue

                # --- .include
                elif token.value == '.include':
                    ts.advance()
                    filename = ts.peek()
                    included = include(filename.value, self.config)
                    included = self._process_tokens(included, symbols)
                    included = apply_token_pasting(included)
                    out.extend(included)
                    continue

                # --- .for
                elif token.value == '.for':
                    expanded = handle_for_loop(ts, symbols)
                    expanded = apply_token_pasting(expanded)
                    expanded = self._process_tokens(expanded, symbols)
                    out.extend(expanded)
                    continue

                # --- .define
                elif token.value in ['.define', '.equ']:
                    define(ts, symbols)
                    continue

                # --- .if / .ifdef / .ifndef
                elif token.value in [ '.if', '.ifdef', '.ifndef']:
                    block = handle_conditionals(ts, symbols)
                    block = apply_token_pasting(block)
                    block = self._process_tokens(block, symbols)
                    out.extend(block)
                    continue

            # identities token
            elif token.type == TokenType.IDENT:

                # --- macro expansion
                if token.value in self.macros:
                    expanded = expand_macro(self.macros[token.value], ts)
                    expanded = apply_token_pasting(expanded)
                    expanded = self._process_tokens(expanded, symbols)
                    out.extend(expanded)
                    continue

                # --- define
                if symbols.exists(token.value):
                    out.append(self._expand_define(token))
                    ts.advance()
                    continue

            # regular token
            out.append(token)
            ts.advance()

        # apply pasting and 'dup' one last time
        out = apply_token_pasting(out)
        out = apply_dup(out, symbols)

        return out

    def _expand_define(self, token: Token) -> Token:
        """Expand an IDENT that exists as .define

        Args:
            token (Token): the IDENT token

        Returns:
            Token: the new Token to replace the identifier
        """
        symbol = self.config.symbols.get(token.value)
        return Token(TokenType.NUMBER, str(symbol.value), token.row, token.col)
