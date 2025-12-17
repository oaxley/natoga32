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
from typing import List, Optional

import packages.ast as ast

from packages.structs import Token, TokenType
from packages.classes import TokenStream


#----- globals
# operators precedence
op_rank = {
    '*' : (6, TokenType.STAR),
    '/' : (6, TokenType.SLASH),
    '%' : (6, TokenType.MODULO),
    '+' : (5, TokenType.PLUS),
    '-' : (5, TokenType.MINUS),
    '<<': (4, TokenType.LSHIFT),
    '>>': (4, TokenType.RSHIFT),
    '&' : (3, TokenType.AND),
    '^' : (2, TokenType.XOR),
    '|' : (1, TokenType.OR)
}


#----- class
class Parser:
    """Parse the stream of tokens"""

    def __init__(self, tokens: List[Token]) -> None:
        """Constructor"""
        self.ts = TokenStream(tokens)

    def parse_program(self) -> ast.Program:
        """Parse the 'Program'

        Returns:
            ast.Program: the AST Program root node
        """
        stmts: List[ast.Statement] = []

        while not self.ts.at_end():
            # skip possible standalone EOL
            if self.ts.expect(TokenType.EOL):
                self.ts.advance()
                continue

            # end parsing on EOF
            if self.ts.expect(TokenType.EOF):
                break

            stmt = self._parse_statement()
            if stmt:
                stmts.append(stmt)

        return ast.Program(stmts)

    def _parse_statement(self) -> Optional[ast.Statement]:
        """Parse a single statement

        Returns:
            Optional[ast.Statement]: None if no statement were found, otherwise the corresponding AST Statement
        """
        token = self.ts.peek()

        match token.type:
            case TokenType.EOF:
                return None

            case TokenType.EOL:
                self.ts.advance()
                return None

            case TokenType.LABEL:
                return self._parse_label()

            case TokenType.IDENT:
                return self._parse_instruction()

            case TokenType.DIRECTIVE:
                return self._parse_directive()

        raise SyntaxError(f"Unexpected token: {token.type.name} ({token.row}, {token.col})")

    def _parse_label(self) -> ast.Label:
        """Process a label

        Returns:
            ast.Label: the string that represents the label, without the ':' if it is present
        """
        token = self.ts.advance()
        name = token.value

        # remove the trailing ':' if any
        if name.endswith(':'):
            name = name[:-1]

        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Label(name)

    def _parse_operands(self) -> List[ast.Node]:
        """Parse the operands / arguments

        Returns:
            List[ast.Node]: a list of AST Node (or derivates) that represents the operands of the root node
        """
        args: List[ast.Node] = []
        while True:
            token = self.ts.peek()
            if token.type in [TokenType.EOL, TokenType.EOF]:
                break

            # remove COMMA
            if token.type == TokenType.COMMA:
                self.ts.advance()
                continue

            expr = self._parse_expression()
            args.append(expr)

        return args

    def _parse_directive(self) -> ast.Directive:
        """Parse a Directive

        Returns:
            ast.Directive: a node that represents the Directive and its operands in the source code
        """
        # retrieve the directive name
        dir_name = self.ts.advance().value

        # parse all the operands
        args = self._parse_operands()

        # consume the trailing EOL
        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Directive(dir_name, args)

    def _parse_instruction(self) -> ast.Instruction:
        """Parse a simple instruction in the source code

        Returns:
            ast.Instruction: a node that represents an Assembly Instruction
        """
        # instruction: IDENT <operands separated by comma>

        op_tok = self.ts.advance()
        if op_tok.type != TokenType.IDENT:
            raise SyntaxError(f"Expected instruction opcode (IDENT)")

        args = self._parse_operands()

        if self.ts.expect(TokenType.EOL):
            self.ts.advance()

        return ast.Instruction(op_tok.value, args)

    def _parse_expression(self, min_prec: int = 1) -> ast.Expression:
        """Parse a complex expression

        Args:
            min_prec (int): the current minimal precedence

        Returns:
            ast.Expression: a node that represents the AST Expression
        """
        # parse the left part of the expression
        left = self._parse_primary()

        # parse binary operators
        while True:
            token = self.ts.peek()
            if token.type == TokenType.EOF:
                break

            # look for the token in the list of operators
            op = None
            op_prec = -1
            for sym, (prec, ttype) in op_rank.items():
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
            right = self._parse_expression(op_prec + 1)

            # create the BinaryOp Node (recursively)
            left = ast.BinaryOp(op, left, right)

        return left

    def _parse_primary(self) -> ast.Expression:
        """Parse the left side of an expression

        Returns:
            ast.Expression: the left side of an expression
        """
        token = self.ts.advance()
        if token.type == TokenType.EOF:
            raise SyntaxError("Unexpected EOF expression!")

        if token.type == TokenType.NUMBER:
            return ast.Number(int(token.value, 0))

        if token.type == TokenType.IDENT:
            return ast.Identifier(token.value)

        if token.type == TokenType.DOLLAR:
            return ast.CurrentPC()

        if token.type == TokenType.STRING:
            # strip the quotes around the string
            s = token.value
            if len(s) >= 2 and s[0] == '"' and s[-1] == '"':
                s = s[1:-1]
            return ast.StringLiteral(s)

        if token.type == TokenType.CHAR:
            # strip simple quotes
            s = token.value
            if len(s) >= 2 and s[0] == "'" and s[-1] == "'":
                s = s[1:-1]
            return ast.CharLiteral(s)

        if token.type in [TokenType.M_ABS_LO, TokenType.M_ABS_HI, TokenType.M_PCREL_LO, TokenType.M_PCREL_HI]:
            if not self.ts.expect(TokenType.LPARENT):
                raise SyntaxError("Missing parenthesis")

            expr = self._parse_expression()
            if self.ts.expect(TokenType.RPARENT):
                self.ts.advance()

            match token.type:
                case TokenType.M_ABS_LO:
                    return ast.LoRel(expr)
                case TokenType.M_ABS_HI:
                    return ast.HiRel(expr)
                case TokenType.M_PCREL_LO:
                    return ast.PCRelLo(expr)
                case TokenType.M_PCREL_HI:
                    return ast.PCRelHi(expr)

        if token.type == TokenType.LPARENT:
            expr = self._parse_expression()
            if self.ts.expect(TokenType.RPARENT):           # a closing ')' must match the initial '('
                self.ts.advance()
            return expr

        # unary operator
        if token.type in [TokenType.PLUS, TokenType.MINUS]:
            op = '+' if token.type == TokenType.PLUS else '-'
            right = self._parse_primary()
            return ast.UnaryOp(op, right)

        raise SyntaxError(f"Unexpected token in expression: {token}")
