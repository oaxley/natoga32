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
# @brief	LoRel relocation expression ast node

#----- imports
from packages.ast.node.expression import Expression
from packages.ast.expression.reloc.reloc import RelocExpr


#----- class
class LoRel(RelocExpr):
    def __init__(self, expr: Expression) -> None:
        super().__init__(expr)

    def __repr__(self) -> str:
        return f"%lo({self.symbol})"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<LoRel ")
        self.symbol.debug(indent + 4)
        print(" " * indent + ">")
