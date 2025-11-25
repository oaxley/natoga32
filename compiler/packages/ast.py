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
# @brief	AST Nodes

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Union, Optional

#----- classes

# Nodes
class Node:
    pass

class Statement(Node):
    pass

class Program(Node):
    def __init__(self, statements: List[Statement]) -> None:
        self.statements = statements

    def __repr__(self) -> str:
        return "\n".join(map(str, self.statements))

# Expressions

class Expression(Node):
    pass

class BinaryOp(Expression):
    def __init__(self, op: str, left: Expression, right: Expression) -> None:
        self.op = op
        self.left = left
        self.right = right

    def __repr__(self) -> str:
        return f"{self.left} {self.op} {self.right}"

class UnaryOp(Expression):
    def __init__(self, op: str, expr: Expression) -> None:
        self.op = op
        self.expr = expr

    def __repr__(self) -> str:
        return f"({self.op}{self.expr})"

class Number(Expression):
    def __init__(self, value: int) -> None:
        self.value = value

    def __repr__(self) -> str:
        return hex(self.value)

class Identifier(Expression):
    def __init__(self, name: str) -> None:
        self.name = name

    def __repr__(self) -> str:
        return self.name

class StringLiteral(Expression):
    def __init__(self, text: str) -> None:
        self.text = text

    def __repr__(self) -> str:
        return f'"{self.text}"'

class CharLiteral(Expression):
    def __init__(self, ch: str) -> None:
        self.ch = ch

    def __repr__(self) -> str:
        return f"'{self.ch}'"


# Relocation nodes
class RelocExpr(Expression):
    pass

class HiRel(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%hi({self.symbol})"

class LoRel(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%lo({self.symbol})"

class PCRelHi(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%pcrel_hi({self.symbol})"

class PCRelLo(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%pcrel_lo({self.symbol})"


# Statements

class Label(Statement):
    def __init__(self, name: str) -> None:
        self.name = name

    def __repr__(self) -> str:
        return f"{self.name}:"

class Directive(Statement):
    def __init__(self, name: str, args: List[Node], label: Optional[str] = None) -> None:
        self.name = name
        self.args = args
        self.label = label          # optional identifier placed before directive

    def __repr__(self) -> str:
        prefix = f"{self.label} " if self.label else ""
        args_s = ", ".join(map(str, self.args)) if self.args else ""
        return f"{prefix}{self.name} {args_s}".rstrip()

class Instruction(Statement):
    def __init__(self, opcode: str, operands: List[Node]) -> None:
        self.opcode = opcode
        self.operands = operands

    def __repr__(self) -> str:
        ops = ", ".join(map(str, self.operands))
        return f"{self.opcode} {ops}".rstrip()
