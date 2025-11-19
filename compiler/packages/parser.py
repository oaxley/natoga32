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
from packages.specs import TokenType, OP_INFO


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

    def parse_operands(self) -> List[ast.Node]:
        args: List[ast.Node] = []
        while True:
            token = self.ts.peek()
            if not token or token.type in [TokenType.EOL, TokenType.EOF]:
                break

            # remove COMMA
            if token.type == TokenType.COMMA:
                self.ts.advance()
                continue

            expr = self.parse_expression()
            args.append(expr)

        return args


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

        args = self.parse_operands()

        # consume the trailing EOL if there
        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Directive(dir_tok.value, args, label_name)

    def parse_instruction(self) -> ast.Instruction:
        # instruction: IDENT <operands separated by comma>

        op_tok = self.ts.advance()
        if not op_tok or op_tok.type != TokenType.IDENT:
            raise SyntaxError(f"Expected instruction opcode (IDENT)")

        args = self.parse_operands()

        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Instruction(op_tok.value, args)

    def parse_expression(self, min_prec=1) -> ast.Expression:
        # parse the left part of the expression
        left = self.parse_primary()

        # parse binary operators
        while True:
            token = self.ts.peek()
            if token is None:
                break

            # look for the token in the list of operators
            op = None
            op_prec = -1
            for sym, (prec, ttype) in OP_INFO.items():
                if token.type == ttype:
                    op = sym
                    op_prec = prec
                    break

            if op is None:      # this is not an operator
                break

            # operator precedence check
            if op_prec < min_prec:
                break

            # consume the operator
            self.ts.advance()

            # parse the right hand side of the expression, with higher precedence
            right = self.parse_expression(op_prec + 1)

            # create the BinaryOp Node (recursively)
            left = ast.BinaryOp(op, left, right)

        return left

    def parse_primary(self) -> ast.Expression:
        token = self.ts.advance()
        if token is None:
            raise SyntaxError("Unexpected EOF expression!")

        if token.type == TokenType.NUMBER:
            return ast.Number(int(token.value, 0))

        if token.type == TokenType.IDENT:
            return ast.Identifier(token.value)

        if token.type == TokenType.STRING:
            # strip the quotes around the string
            s = token.value    # type: ignore
            if len(s) >= 2 and s[0] == '"' and s[-1] == '"':
                s = s[1:-1]
            return ast.StringLiteral(s)

        if token.type == TokenType.CHAR:
            # strip simple quotes
            s = token.value    # type: ignore
            if len(s) >= 2 and s[0] == "'" and s[-1] == "'":
                s = s[1:-1]
            return ast.CharLiteral(s)

        if token.type == TokenType.LPARENT:
            expr = self.parse_expression()
            self.ts.expect(TokenType.RPARENT)           # a closing ')' must match the initial '('
            self.ts.advance()
            return expr

        # unary operator
        if token.type in [TokenType.PLUS, TokenType.MINUS]:
            op = '+' if token.type == TokenType.PLUS else '-'
            right = self.parse_primary()
            return ast.UnaryOp(op, right)

        raise SyntaxError(f"Unexpected token in expression: {token}")
