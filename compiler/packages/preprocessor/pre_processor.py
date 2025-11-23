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
# @brief	Compiler pre-processor main file

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

import os

from packages.token import Token, TokenStream, TokenType

from . import helper
from .macros import MacroDefinition, parse_macro_definition, expand_macro
from .includes import handle_include
from .for_loops import handle_for_loop
from .token_pasting import apply_token_pasting
from .dup import apply_dup
from .environment import handle_envvar


#----- class
class PreProcessor:
    """Compiler pre-processor"""

    def __init__(self) -> None:
        """Constructor"""
        self.macros: Dict[str, MacroDefinition] = {}

    def process(self, tokens: List[Token], current_file: str) -> List[Token]:
        """Perform pre-processing of the source file

        Args:
            tokens (List[Token]): the list of tokens from the Lexer
            current_file (str): the filename of the current file being processed

        Returns:
            List[Token]: the new list of tokens after pre-processing is done
        """
        # process the environment variables expansion
        tmp = handle_envvar(tokens)

        # create a new token stream from the result
        ts = TokenStream(tmp)
        out: List[Token] = []

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
                    continue

                # --- .include
                elif token.value == '.include':
                    included = handle_include(ts, current_file)
                    included = self.process(included, current_file)
                    included = apply_token_pasting(included)
                    included = apply_dup(included)
                    out.extend(included)
                    continue

                # --- .for
                elif token.value == '.for':
                    expanded = handle_for_loop(ts)
                    expanded = apply_token_pasting(expanded)
                    expanded = apply_dup(expanded)
                    expanded = self.process(expanded, current_file)
                    out.extend(expanded)
                    continue

            # identities token
            elif token.type == TokenType.IDENT:

                # --- macro expansion
                if token.value in self.macros:
                    expanded = expand_macro(self.macros[token.value], ts)
                    expanded = apply_token_pasting(expanded)
                    expanded = apply_dup(expanded)
                    expanded = self.process(expanded, current_file)
                    out.extend(expanded)
                    continue

            # regular token
            out.append(token)
            ts.advance()

        # apply pasting and 'dup' one last time
        out = apply_token_pasting(out)
        out = apply_dup(out)

        return out


