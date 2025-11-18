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
from typing import Any, Dict, List, Optional, Union

import packages.ast as ast

from packages.token import TokenStream
from packages.specs import TokenType


#----- class
class Parser:
    """Parse the stream of tokens"""

    def __init__(self, tokens: TokenStream) -> None:
        """Constructor"""
        self.ts = tokens

    def parse_program(self) -> ast.Program:
        stmts: List[ast.Statement] = []

        while not self.ts.at_end():
            # skip possible standalone EOL
            if self.ts.expect(TokenType.EOL):
                self.ts.advance()
                continue

            # end parsing on EOF
            if self.ts.expect(TokenType.EOF):
                break

            stmt = self.parse_statement()
            if stmt:
                stmts.append(stmt)

        return ast.Program(stmts)

    def parse_statement(self) -> Optional[ast.Statement]:
        token = self.ts.peek()
        if not token:
            return None

        if token.type == TokenType.EOL:
            self.ts.advance()
            return None

        if token.type == TokenType.LABEL:
            return self.parse_label()

        if token.type in [TokenType.DIRECTIVE, TokenType.IDENT]:
            # different cases:
            # IDENT followed by DIRECTIVE -> VALUE .equ ...
            # DIRECTIVE alone -> .data, .text
            # IDENT then sth else -> instruction
            if token.type == TokenType.IDENT:
                next_token = self.ts.peek(1)
                if next_token and next_token.type == TokenType.DIRECTIVE:
                    return self.parse_directive()
                else:
                    return self.parse_instruction()
            else:
                return self.parse_directive()

        raise SyntaxError(f"Unexpected token: {token.type.name} ({token.row}, {token.col})")

    def parse_label(self) -> ast.Label:
        token = self.ts.advance()
        name = token.value              # type: ignore

        # remove the trailing ':' if any
        if name.endswith(':'):
            name = name[:-1]

        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Label(name)

    def parse_directive(self) -> ast.Directive:
        # possible patterns:
        # DIRECTIVE ...
        # IDENT DIRECTIVE ...
        label_name: Optional[str] = None
        next_token = self.ts.peek(1)
        if self.ts.expect(TokenType.IDENT) and next_token and next_token.type == TokenType.DIRECTIVE:
            # form: IDENT DIRECTIVE ...
            label_name = self.ts.advance().value    # type: ignore

        dir_tok = self.ts.advance()
        if not dir_tok or dir_tok.type != TokenType.DIRECTIVE:
            raise SyntaxError(f"Expected DIRECTIVE token!")

        args: List[ast.Node] = []
        while True:
            token = self.ts.peek()

            if not token or token.type in [TokenType.EOL, TokenType.EOF]:
                break

            # remove COMMA
            if token.type == TokenType.COMMA:
                self.ts.advance()
                continue

            # parse possible token types: number, string, char, ident, comma ...
            if token.type == TokenType.NUMBER:
                t = self.ts.advance()
                args.append(ast.Number(int(t.value, 0))) # type: ignore

            elif token.type == TokenType.STRING:
                t = self.ts.advance()
                # strip the quotes around the string
                s = t.value    # type: ignore
                if len(s) >= 2 and s[0] == '"' and s[-1] == '"':
                    s = s[1:-1]
                args.append(ast.StringLiteral(s))

            elif token.type == TokenType.CHAR:
                t = self.ts.advance()
                # strip simple quotes
                s = t.value    # type: ignore
                if len(s) >= 2 and s[0] == "'" and s[-1] == "'":
                    s = s[1:-1]
                args.append(ast.CharLiteral(s))

            elif token.type == TokenType.IDENT:
                t = self.ts.advance()
                args.append(ast.Identifier(t.value)) # type: ignore

            else:
                # unknown token. Add it to the list as Identifier to consume it
                t = self.ts.advance()
                args.append(ast.Identifier(t.value)) # type: ignore

        # consume the trailing EOL if there
        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Directive(dir_tok.value, args, label_name)

    def parse_instruction(self) -> ast.Instruction:
        # instruction: IDENT <operands separated by comma>

        op_tok = self.ts.advance()
        if not op_tok or op_tok.type != TokenType.IDENT:
            raise SyntaxError(f"Expected instruction opcode (IDENT)")

        args: List[Union[ast.Node,str,int]] = []
        while True:
            token = self.ts.peek()
            if not token or token.type in [TokenType.EOL, TokenType.EOF]:
                break

            # remove COMMA
            if token.type == TokenType.COMMA:
                self.ts.advance()
                continue

            # parse possible token types: number, string, char, ident, comma ...
            if token.type == TokenType.IDENT:
                t = self.ts.advance()
                args.append(ast.Identifier(t.value)) # type: ignore

            elif token.type == TokenType.NUMBER:
                t = self.ts.advance()
                args.append(ast.Number(int(t.value, 0))) # type: ignore

            elif token.type == TokenType.STRING:
                t = self.ts.advance()
                # strip the quotes around the string
                s = t.value    # type: ignore
                if len(s) >= 2 and s[0] == '"' and s[-1] == '"':
                    s = s[1:-1]
                args.append(ast.StringLiteral(s))

            elif token.type == TokenType.CHAR:
                t = self.ts.advance()
                # strip simple quotes
                s = t.value    # type: ignore
                if len(s) >= 2 and s[0] == "'" and s[-1] == "'":
                    s = s[1:-1]
                args.append(ast.CharLiteral(s))

            else:
                # unknown token. Add it to the list as Identifier to consume it
                t = self.ts.advance()
                args.append(ast.Identifier(t.value)) # type: ignore

        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Instruction(op_tok.value, args)

