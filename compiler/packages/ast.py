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
class Node:
    """A simple node"""
    pass

# support for compile time expression
class Expression(Node):
    """An compile time expression"""
    pass

class Number(Expression):
    """A number is an expression"""
    def __init__(self, value: Union[int, str]) -> None:
        self.value = value

class BinaryOp(Expression):
    """A binary operation between 2 expressions"""
    def __init__(self, op: str, left: Expression, right: Expression) -> None:
        self.op = op
        self.left = left
        self.right = right

# assembly Statements & al
class Statement(Node):
    """A simple statement"""
    pass

class Directive(Statement):
    """Assembler directive"""
    def __init__(self, name: str, args: List[str], alias: Optional[str] = None) -> None:
        self.name = name
        self.args = args
        self.alias = alias

    def __repr__(self) -> str:
        prefix = f"{self.alias}: " if self.alias else ""
        return f"{prefix}{self.name} {' '.join(map(str, self.args))}"

# custom type for Instruction class
class Instruction(Statement):
    """Assembly instruction"""
    def __init__(self, opcode: str, operands: List[Union[int, str, Expression]]) -> None:
        self.opcode = opcode
        self.operands = operands

    def __repr__(self) -> str:
        return f"{self.opcode} {', '.join(map(str, self.operands))}"

class Label(Statement):
    """Assembly label"""
    def __init__(self, name: str) -> None:
        self.name = name

    def __repr__(self) -> str:
        return f"{self.name}:"

class Program(Node):
    """A program is composed of statements"""
    def __init__(self, statements: List[Statement]) -> None:
        self.statements = statements

    def __repr__(self) -> str:
        return "\n".join(map(str, self.statements))
