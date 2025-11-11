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
from typing import Any, Dict, List, Union

#----- classes
class Node:
    """A simple node"""
    pass

# support for compile time expression
class Expression(Node):
    """An compile time expression"""
    def __init__(self, value: int) -> None:
        self.value = value

class BinaryOp(Node):
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
    def __init__(self, name: str, args: List[str]) -> None:
        self.name = name
        self.args = args

# custom type for Instruction class
class Instruction(Statement):
    """Assembly instruction"""
    def __init__(self, opcode: str, operands: List[Union[int, str, Expression]]) -> None:
        self.opcode = opcode
        self.operands = operands

class Label(Statement):
    """Assembly label"""
    def __init__(self, name: str) -> None:
        self.name = name

class Program(Node):
    """A program is composed of statements"""
    def __init__(self, statements: List[Statement]) -> None:
        self.statements = statements
