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
# @brief	Identifier expression ast node

#----- imports
from packages.ast.node import Expression


#----- class
class Identifier(Expression):
    def __init__(self, name: str) -> None:
        self.name = name

    def __repr__(self) -> str:
        return self.name

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<Identifier [{self.name}]>")
