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
# @brief	StringLiteral expression ast node

#----- imports
from packages.ast.node import Expression


#----- class
class StringLiteral(Expression):
    def __init__(self, text: str) -> None:
        self.text = text

    def __repr__(self) -> str:
        return f'"{self.text}"'

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<String [{self.text}]>")
