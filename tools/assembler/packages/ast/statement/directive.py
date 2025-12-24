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
# @brief	Directive statement ast node

#----- imports
from typing import List

from packages.ast.node import Node, Statement


#----- class
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
