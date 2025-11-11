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
# @brief	Parser main class

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages.token import TokenStream

#----- class
class Parser:
    """Parse the stream of tokens"""

    def __init__(self, tokens: TokenStream) -> None:
        """Constructor"""
        self.tokens = tokens

    def parse(self) -> None:
        while (token := self.tokens.next()):
            print(token)
