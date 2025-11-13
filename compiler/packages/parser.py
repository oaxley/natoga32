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
# @brief	Parser main class

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional

import packages.ast as ast

from packages.token import TokenStream
from packages.specs import TokenType

#----- class
class Parser:
    """Parse the stream of tokens"""

    def __init__(self, tokens: TokenStream) -> None:
        """Constructor"""
        self.tokens = tokens

    def process(self) -> ast.Program:
        """Process the tokens and create the AST"""
        statements: List[ast.Statement] = []
        while (not self.tokens.end()):
            stmt = self.parse()
            if stmt:
                statements.append(stmt)

        return ast.Program(statements)

    def parse(self) -> Optional[ast.Statement]:
        """Process tokens and build a stateemt"""
        # retrieve the next token
        token = self.tokens.peek()
        if token is None:
            return None

        # check for directives
        if token.type == TokenType.DIRECTIVE:
            return self._parse_directive()

        # unknown token
        self.tokens.next()
        return None

    def _parse_directive(self) -> Optional[ast.Directive]:
        """Parse the tokens and build a directive"""
        # retrieve the next token
        token = self.tokens.next()
        if not token:
            return None

        # ensure a directive starts with an '.' at the begining
        if token.type == TokenType.DIRECTIVE and token.value[0] != '.':
            raise SyntaxError(f"Error ({token.row},{token.col}): '{token.value}' recognized as DIRECTIVE without starting with '.'")

        # retrieve the args for this directive
        args = []
        t = self.tokens.next()
        while (t and t.type != TokenType.EOL):
            args.append(t.value)
            t = self.tokens.next()

        return ast.Directive(token.value, args)

