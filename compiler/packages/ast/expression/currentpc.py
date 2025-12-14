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
# @brief	CurrentPC expression ast node

#----- imports
from packages.ast.node import Expression


#----- class
class CurrentPC(Expression):
    def __init__(self) -> None:
        pass

    def __repr__(self) -> str:
        return "$"

    def debug(self, indent: int = 0) -> None:
        print(" " * indent + f"<CurrentPC>")
