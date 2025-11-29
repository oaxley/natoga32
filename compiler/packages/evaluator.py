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

def _eval_binop(op: str, a: int, b: int) -> int:
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


def const_eval(expr: ast.Expression, symbols: SymbolTable, pc: int) -> EvalResult:
    """Try to evaluate an expression `expr` to an integer

    Args:
        expr (ast.Expression): the expression to evaluate
        symbols (SymbolTable): the symbol table
        pc (int): the current program counter

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
                if s.type in [SymbolType.DEFINE, SymbolType.LABEL] and s.value:
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
        return EvalResult(reloc=Relocation('SYMBOL', expr, 0, pc))

    # current PC
    if isinstance(expr, ast.CurrentPC):
        return EvalResult(value=pc)

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

        # case 1 : both are constants
        if left.reloc is None and right.reloc is None:
            return EvalResult(value=_eval_binop(expr.op, left.value, right.value)) # type: ignore

        # case 2 : left side is reloc, right side is constant
        if left.reloc and right.reloc is None:
            addend = right.value
            # only + or - supported
            if expr.op == '+':
                return EvalResult(reloc=Relocation(
                    left.reloc.type,
                    left.reloc.symbol,
                    left.reloc.addend + addend,  # type: ignore
                    pc
                ))
            if expr.op == '-':
                return EvalResult(reloc=Relocation(
                    left.reloc.type,
                    left.reloc.symbol,
                    left.reloc.addend - addend,  # type: ignore
                    pc
                ))
            raise SyntaxError(f"Error: unsupported reloc combination with '{expr.op}'")

        # case 3 : left side is constant, right side is reloc
        if left.reloc is None and right.reloc:
            addend = left.value
            # only + or - supported
            if expr.op == '+':
                return EvalResult(reloc=Relocation(
                    right.reloc.type,
                    right.reloc.symbol,
                    right.reloc.addend + addend,  # type: ignore
                    pc
                ))
            if expr.op == '-':
                return EvalResult(reloc=Relocation(
                    right.reloc.type,
                    right.reloc.symbol,
                    addend,  # type: ignore
                    pc
                ))
            raise SyntaxError(f"Error: unsupported reloc combination with '{expr.op}'")

        # case 4 : both side reloc -> impossible
        raise SyntaxError(f"Error: cannot create relocation with two symbols")

    # relocation node
    if isinstance(expr, ast.HiRel):
        inner = const_eval(expr.symbol, symbols, pc)
        if inner.value:
            # RISC-V hi: (value + 0x800) >> 12
            hi = (inner.value + 0x800) >> 12
            return EvalResult(hi)

        # relocation record
        return EvalResult(reloc=Relocation('R_RISCV_HI20', expr.symbol, 0, pc))

    if isinstance(expr, ast.LoRel):
        inner = const_eval(expr.symbol, symbols, pc)
        if inner.value:
            lo = inner.value & 0xfff
            # I-Type immediate must preserved the sign
            return EvalResult(value=lo if lo < (1 << 11) else (lo - (1 << 12)))

        assert pc is not None
        return EvalResult(reloc=Relocation('R_RISCV_LO12_I', expr.symbol, 0, pc))

    if isinstance(expr, ast.PCRelHi):
        inner = const_eval(expr.symbol, symbols, pc)
        if inner.value and pc is not None:
            diff = inner.value - pc
            hi = (diff + 0x800) >> 12
            return EvalResult(value=hi)

        return EvalResult(reloc=Relocation('R_RISCV_PCREL_HI20', expr.symbol, 0, pc))

    if isinstance(expr, ast.PCRelLo):
        inner = const_eval(expr.symbol, symbols, pc)
        if inner.value and pc is not None:
            diff = inner.value - pc
            lo = diff & 0xfff
            return EvalResult(value=lo if lo < (1 << 11) else (lo - (1 << 12)))

        return EvalResult(reloc=Relocation('R_RISCV_PCREL_LO12_I', expr.symbol, 0, pc))


    # default
    raise SyntaxError("Unhandled expression type in const_eval: " + repr(expr))
