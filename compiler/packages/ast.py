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
    def debug(self, indent: int = 0) -> None:
        pass

class Statement(Node):
    def debug(self, indent: int = 0) -> None:
        pass

class Program(Node):
    def __init__(self, statements: List[Statement]) -> None:
        self.statements = statements

    def __repr__(self) -> str:
        return "\n".join(map(str, self.statements))

    def debug(self, indent: int = 0) -> None:
        for i in self.statements:
            i.debug(0)

# Expressions

class Expression(Node):
    def debug(self, indent: int = 0) -> None:
        pass

class BinaryOp(Expression):
    def __init__(self, op: str, left: Expression, right: Expression) -> None:
        self.op = op
        self.left = left
        self.right = right

    def __repr__(self) -> str:
        return f"{self.left} {self.op} {self.right}"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<BinaryOp [{self.op}]")
        print(" " * (indent + 2) + "left:")
        self.left.debug(indent+4)
        print(" " * (indent + 2) + "right:")
        self.right.debug(indent+4)
        print(" " * indent + ">")

class UnaryOp(Expression):
    def __init__(self, op: str, expr: Expression) -> None:
        self.op = op
        self.expr = expr

    def __repr__(self) -> str:
        return f"({self.op}{self.expr})"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<UnaryOp [{self.op}]")
        print(" " * (indent + 2) + "expr:")
        self.expr.debug(indent+4)
        print(" " * indent + ">")

class Number(Expression):
    def __init__(self, value: int) -> None:
        self.value = value

    def __repr__(self) -> str:
        return hex(self.value)

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Number [{self.value}]>")

class Identifier(Expression):
    def __init__(self, name: str) -> None:
        self.name = name

    def __repr__(self) -> str:
        return self.name

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Identifier [{self.name}]>")

class StringLiteral(Expression):
    def __init__(self, text: str) -> None:
        self.text = text

    def __repr__(self) -> str:
        return f'"{self.text}"'

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<String [{self.text}]>")

class CharLiteral(Expression):
    def __init__(self, ch: str) -> None:
        self.ch = ch

    def __repr__(self) -> str:
        return f"'{self.ch}'"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Char [{self.ch}]>")


# Relocation nodes
class RelocExpr(Expression):
    def debug(self, indent = 0) -> None:
        pass

class HiRel(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%hi({self.symbol})"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<HiRel ")
        self.symbol.debug(indent + 4)
        print(" " * indent + ">")

class LoRel(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%lo({self.symbol})"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<LoRel ")
        self.symbol.debug(indent + 4)
        print(" " * indent + ">")

class PCRelHi(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%pcrel_hi({self.symbol})"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<PCRelHi ")
        self.symbol.debug(indent + 4)
        print(" " * indent + ">")

class PCRelLo(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def __repr__(self) -> str:
        return f"%pcrel_lo({self.symbol})"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<PCRelLo ")
        self.symbol.debug(indent + 4)
        print(" " * indent + ">")


# Statements

class Label(Statement):
    def __init__(self, name: str) -> None:
        self.name = name

    def __repr__(self) -> str:
        return f"{self.name}:"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Label [{self.name}]>")

class Directive(Statement):
    def __init__(self, name: str, args: List[Node]) -> None:
        self.name = name
        self.args = args

    def __repr__(self) -> str:
        args_s = ", ".join(map(str, self.args)) if self.args else ""
        return f"{self.name} {args_s}".rstrip()

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Directive [{self.name}]")
        for i in self.args:
            i.debug(indent + 4)
        print(">")

class Instruction(Statement):
    def __init__(self, opcode: str, operands: List[Node]) -> None:
        self.opcode = opcode
        self.operands = operands

    def __repr__(self) -> str:
        ops = ", ".join(map(str, self.operands))
        return f"{self.opcode} {ops}".rstrip()

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Instruction [{self.opcode}]")
        for i in self.operands:
            i.debug(indent + 4)
        print(">")
