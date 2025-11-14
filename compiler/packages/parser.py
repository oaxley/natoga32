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

from packages.token import TokenStream, Token
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

    def parse(self):
        # retrieve all the tokens for this statement
        tokens: List[Token] = []
        while True:
            t = self.tokens.next()
            if not t or t.type in [TokenType.EOL, TokenType.EOF]:
                break
            else:
                tokens.append(t)

        # print(f"{' '.join(map(str, tokens))} {len(tokens)}")
        if len(tokens) == 0:
            return None

        # single directive or label
        if len(tokens) == 1:
            t = tokens[0]
            if t.type == TokenType.LABEL:
                return ast.Label(t.value[:-1])
            elif t.type == TokenType.DIRECTIVE:
                return ast.Directive(t.value, None, None)
            elif t.type == TokenType.IDENT:
                return ast.Instruction(t.value, [])
            else:
                raise SyntaxError(f"Unknown token [{t.value}] at ({t.row}, {t.col})")

        return None



