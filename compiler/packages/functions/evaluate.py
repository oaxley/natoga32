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
# @brief	Function evaluate_expr

#----- imports
from typing import Dict, List

from packages.structs import Token, TokenType
from packages.classes import SymbolTable


#----- function
def evaluate_expr(tokens: List[Token], symbols: SymbolTable) -> int:
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
        elif t.type == TokenType.IDENT and symbols.exists(t.value):
            parts.append(symbols.get(t.value))
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
