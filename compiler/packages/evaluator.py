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
from typing import Any, Dict, List, Optional

from packages.data_classes import Relocation, EvalResult

from packages import ast
from packages.symbols import Symbol, SymbolTable, SymbolType


#----- functions

def const_eval(expr: ast.Expression, symbols: SymbolTable, pc: int) -> EvalResult:
    """Try to evaluate an expression `expr` to an integer

    Args:
        expr (ast.Expression): the expression to evaluate
        symbols (SymbolTable): the symbol table
        pc (Optional[int]): the current program counter

    Returns:
        EvalResult: the result of the evaluation

    **Notes**:
    - If every identifier resolves to an integer, return EvalResult(int).
    - If the expression contains a relocation, return EvalResult(Relocation)
    """

    # helper function to resolve an identifier to an integer or None
    def resolve_ident(name: str) -> Optional[int]:
        try:
            if symbols.exists(name):
                s = symbols.get(name)
                if s.type == SymbolType.DEFINE and s.value:
                    return int(s.value)

            return None
        except ValueError:
            return None

    # expression is a number
    if isinstance(expr, ast.Number):
        return EvalResult(value=expr.value)

    if isinstance(expr, ast.CharLiteral):
        return EvalResult(value=ord(expr.ch))

    # an simple identifier
    if isinstance(expr, ast.Identifier):
        v = resolve_ident(expr.name)
        if v:
            return EvalResult(value=int(v))

        # relocation needed
        return EvalResult(reloc=Relocation(pc, 'SYMBOL', expr, 0))

    # unary operator
    if isinstance(expr, ast.UnaryOp):
        r = const_eval(expr.expr, symbols, pc)
        if r.reloc:
            # cannot fold if we have a relocation
            return EvalResult(reloc=r.reloc)

        v = r.value
        if v:
            match expr.op:
                case '+':
                    return EvalResult(value=+v)
                case '-':
                    return EvalResult(value=-v)
                case '~':
                    return EvalResult(value=~v)

        raise SyntaxError(f"Unknown UnaryOp '{expr.op}'")

    # binary operator
    if isinstance(expr, ast.BinaryOp):
        left = const_eval(expr.left, symbols, pc)
        right = const_eval(expr.right, symbols, pc)

        # if either side return a relocation, abort
        if left.reloc or right.reloc:
            return EvalResult(reloc=Relocation(pc, 'EXPR_RELOC', expr, ))

        a = left.value
        b = right.value
        assert a is not None
        assert b is not None
        match expr.op:
            case '+':
                return EvalResult(value=(a + b))
            case '-':
                return EvalResult(value=(a - b))
            case '*':
                return EvalResult(value=(a * b))
            case '/':
                return EvalResult(value=(a // b if b != 0 else 0))
            case '%':
                return EvalResult(value=(a % b))
            case '<<':
                return EvalResult(value=(a << b))
            case '>>':
                return EvalResult(value=(a >> b))
            case '&':
                return EvalResult(value=(a & b))
            case '|':
                return EvalResult(value=(a | b))
            case '^':
                return EvalResult(value=(a ^ b))

        raise SyntaxError(f"Unknown BinaryOp {expr.op}")


    # relocation node
    if isinstance(expr, ast.HiRel):
        inner = const_eval(expr.symbol, symbols, pc)
        if inner.value:
            # RISC-V hi: (value + 0x800) >> 12
            hi = (inner.value + 0x800) >> 12
            return EvalResult(hi)

        # relocation record
        return EvalResult(reloc=Relocation(pc, 'R_RISCV_HI20', expr.symbol))

    if isinstance(expr, ast.LoRel):
        inner = const_eval(expr.symbol, symbols, pc)
        if inner.value:
            lo = inner.value & 0xfff
            # I-Type immediate must preserved the sign
            return EvalResult(value=lo if lo < (1 << 11) else (lo - (1 << 12)))
        return EvalResult(reloc=Relocation(pc, 'R_RISCV_LO12_I', expr.symbol))


    # default
    raise SyntaxError("Unhandled expression type in const_eval: " + repr(expr))
