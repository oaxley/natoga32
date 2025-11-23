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
# @brief	Helper functions

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional

from packages.token import (
    Token, TokenStream, TokenType
)


#----- globals
MAX_BODY_DEPTH = 50             # max body recursion depth


#----- functions
def clone_token(token: Token) -> Token:
    """Clone a Token with a Shallow Copy

    Args:
        token (Token): the token to copy

    Returns:
        Token: the shallow copy of the source token
    """
    return Token(token.type, token.value, token.row, token.col)


def capture_body(ts: TokenStream, begin: str, end: str) -> List[Token]:
    """Parse the token stream and extract the body between 'begin' and 'end' tokens

    Args:
        ts (TokenStream): the stream of tokens
        begin (str): value for the beginning body marker
        end (str): value for the end body marker

    Returns:
        List[Token]: the list of tokens representing the body
    """
    body: List[Token] = []
    depth = 1       # depth tracker

    while True:
        token = ts.peek()

        if token is None:
            break

        # depth tracker (for nested bodies)
        if token.type == TokenType.DIRECTIVE and token.value == begin:
            depth += 1
            if depth > MAX_BODY_DEPTH:
                raise SyntaxError("Nested body max limit exceeded!")

        if token.type == TokenType.DIRECTIVE and token.value == end:
            depth -= 1
            if depth == 0:
                break

        # add everything to the body, and move forward
        body.append(token)
        ts.advance()

    # consume "end" marker & EOL
    ts.advance()
    if ts.expect(TokenType.EOL): ts.advance()

    # return the body
    return body


def get_value(ts: TokenStream, ttype: TokenType, tvalue: Optional[str] = None) -> str:
    """Return next token value from the stream, only if its type correspond to ttype
    or its value is equal to tvalue

    Args:
        ts (TokenStream): the token stream
        ttype (TokenType) : the reference token type that token should match
        tvalue (Optional[str]): the reference value that token should match

    Returns:
        str: the token value
    """
    token = ts.advance()

    if not token or token.type != ttype:
        raise SyntaxError(f"Token value is either None or of the wrong type!")

    if tvalue and token.value != tvalue:
        raise SyntaxError(f"Expecting '{tvalue}', got '{token.value}'!")

    return token.value


def evaluate_expr(tokens: List[Token]) -> int:
    """Evaluate a simple expression

    Minimal expression parser that supports:
    - numbers (dec, hex, bin),
    - parantheses
    - operators (+ - * / % << >> & | ^)

    Args:
        tokens (List[Token]): the expression to parse

    Returns:
        int: the value once the expression has been evaluated
    """
    if not tokens:
        raise SyntaxError("Expression is empty!")

    # operators mapping
    op_map: Dict[TokenType, str] = {
        TokenType.PLUS: '+',
        TokenType.MINUS: '-',
        TokenType.STAR: '*',
        TokenType.SLASH: '/',
        TokenType.MODULO: '%',
        TokenType.LSHIFT: '<<',
        TokenType.RSHIFT: '>>',
        TokenType.AND: '&',
        TokenType.OR: '|',
        TokenType.XOR: '^'
    }

    parts = []
    for t in tokens:
        if t.type == TokenType.NUMBER:
            parts.append(t.value)
        elif t.type == TokenType.LPARENT:
            parts.append('(')
        elif t.type == TokenType.RPARENT:
            parts.append(')')
        elif t.type in op_map:
            parts.append(op_map[t.type])
        else:
            # default insertion
            parts.append(t.value)

    # build the expression from its part
    expr = " ".join(parts)

    # evaluate the expression, only with python standard ops
    try:
        value = eval(expr, {"__builtins__": None}, {})
    except Exception as e:
        raise SyntaxError(f"Unable to evaluate expression '{expr}': {e}!")

    return int(value)
