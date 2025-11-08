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
# @brief	Lexer main class

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.config import Config
from packages.token import Token, TokenTypes


#----- classes
class Lexer:
    """Process the assembly file and create tokens"""


    def __init__(self, config: Config) -> None:
        """Constructor"""
        self.config = config

