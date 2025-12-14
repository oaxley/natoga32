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
# @brief	Instruction statement ast node

#----- imports
from typing import List

from packages.ast.node import Node, Statement


#----- class
class Instruction(Statement):
    def __init__(self, opcode: str, operands: List[Node]) -> None:
        self.opcode = opcode
        self.operands = operands
        self.address: int = 0

    def __repr__(self) -> str:
        ops = ", ".join(map(str, self.operands))
        return f"{self.opcode} {ops}".rstrip()

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Instruction [{self.opcode}]")
        for i in self.operands:
            i.debug(indent + 4)
        print(">")
