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
# @brief	Perform expression evaluation

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional, Tuple


from packages import ast
from packages.symbols import SymbolTable, SymbolType
from packages.data_classes import Relocation, EvalResult


#----- functions

def binary_op(op: str, a: int, b: int) -> int:
    """Evaluate a BinaryOp

    Args:
        op (str): the operation
        a (int): the left member of the operation
        b (int): the right member of the operation
    """
    match op:
        case '+':
            return (a + b)
        case '-':
            return (a - b)
        case '*':
            return (a * b)
        case '/':
            return (a // b if b != 0 else 0)
        case '%':
            return (a % b)
        case '<<':
            return (a << b)
        case '>>':
            return (a >> b)
        case '&':
            return (a & b)
        case '|':
            return (a | b)
        case '^':
            return (a ^ b)
        case _:
            raise ValueError(f"Unknown operator '{op}' for BinaryOp.")


def const_eval(expr: ast.Expression, symbols: SymbolTable, pc: int) -> Tuple[bool, int]:
    """Evaluate an expression to an integer

    Args:
        expr (ast.Expression): an expression to evaluate
        symbols (SymbolTable): the symbol table

    Returns:
        Tuple[bool, int]: True if the evaluation was done, False otherwise
    """

    # helper function to resolve a DEFINE symbol
    def resolve_define(name: str) -> Optional[int]:
        try:
            if symbols.exists(name):
                s = symbols.get(name)
                if s.defined and s.type == SymbolType.DEFINE and s.value:
                    return int(s.value)

            return None
        except ValueError:
            return None

    # expression is an number
    if isinstance(expr, ast.Number):
        return (True, expr.value)

    # expression is a simple character
    if isinstance(expr, ast.CharLiteral):
        return (True, ord(expr.ch))

    # expression is the current PC
    if isinstance(expr, ast.CurrentPC):
        return (True, pc)

    # an identifier
    if isinstance(expr, ast.Identifier):
        v = resolve_define(expr.name)
        if v is None:
            return (False, 0)
        return (True, v)

    # unary operation
    if isinstance(expr, ast.UnaryOp):
        result, value = const_eval(expr, symbols, pc)
        if not result:
            return (False, 0)

        match expr.op:
            case '+':
                return (True, value)
            case '-':
                return (True, -value)
            case '~':
                return (True, ~value)

        raise SyntaxError(f"Error: unsupported UnaryOp '{expr.op}'")

    # binary operation
    if isinstance(expr, ast.BinaryOp):

        # check left side
        result, left = const_eval(expr.left, symbols, pc)
        if not result:
            return (False, 0)

        # check right side
        result, right = const_eval(expr.right, symbols, pc)
        if not result:
            return (False, 0)

        # compute the value
        value = binary_op(expr.op, left, right)
        return (True, value)

    # default return
    return (False, 0)
