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
# @brief	BinaryOp expression ast node

#----- imports
from packages.ast.node import Expression


#----- class
class BinaryOp(Expression):
    def __init__(self, op: str, left: Expression, right: Expression) -> None:
        self.op = op
        self.left = left
        self.right = right

    def __repr__(self) -> str:
        return f"{self.left} {self.op} {self.right}"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<BinaryOp [{self.op}]")
        self.left.debug(indent+4)
        self.right.debug(indent+4)
        print(" " * indent + ">")
