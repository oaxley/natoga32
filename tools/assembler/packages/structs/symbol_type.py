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
# @brief	SymbolType enum

#----- imports
from enum import IntEnum, auto


#----- class
class SymbolType(IntEnum):
    """Define all the types that a symbol can have"""
    LABEL = auto()
    DEFINE = auto()
    MACRO = auto()
    UNKNOWN = auto()
