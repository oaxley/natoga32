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
# @brief	Program ast node

#----- imports
from typing import List

from .node import Node
from .statement import Statement


#----- class
class Program(Node):
    def __init__(self, statements: List[Statement]) -> None:
        self.statements = statements

    def __repr__(self) -> str:
        return "\n".join(map(str, self.statements))

    def debug(self, indent: int = 0) -> None:
        for i in self.statements:
            i.debug(0)
