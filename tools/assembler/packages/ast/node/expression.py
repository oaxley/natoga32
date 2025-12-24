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
# @brief	Expression ast node

#----- imports
from .node import Node


#----- class
class Expression(Node):
    def debug(self, indent: int = 0) -> None:
        pass
