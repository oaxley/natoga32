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
# @brief	Reloc expression ast node

#----- imports
from packages.ast.node import Expression


#----- class
class RelocExpr(Expression):
    def __init__(self, expr: Expression) -> None:
        self.symbol = expr

    def debug(self, indent = 0) -> None:
        pass
