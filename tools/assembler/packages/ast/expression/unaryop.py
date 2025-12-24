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
# @brief	UnaryOp expression ast node

#----- imports
from packages.ast.node import Expression


#----- class
class UnaryOp(Expression):
    def __init__(self, op: str, expr: Expression) -> None:
        self.op = op
        self.expr = expr

    def __repr__(self) -> str:
        return f"({self.op}{self.expr})"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<UnaryOp [{self.op}]")
        self.expr.debug(indent+4)
        print(" " * indent + ">")
