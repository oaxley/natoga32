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
# @brief	Symbols table

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, ClassVar, Optional

from packages.token import TokenType


#----- class
class Symbols:
    """Symbols table

    Notes:
        This class acts as a singleton
    """

    # private members
    __instance: ClassVar[Optional[Symbols]] = None
    __keywords: ClassVar[Dict[str, TokenType]] = { }

    def __new__(cls) -> Symbols:
        """Create a new instance or return the current one"""
        if Symbols.__instance is None:
            Symbols.__instance = object.__new__(cls)

        return Symbols.__instance

    @property
    def keywords(self) -> List[str]:
        """Return the list of defined keywords"""
        return list(Symbols.__keywords.keys())

    def is_exist(self, name: str) -> bool:
        """Check if a keyword exists in the list"""
        return (name in Symbols.__keywords)

    def get_type(self, name: str) -> TokenType:
        """Return the type of a token"""
        if self.is_exist(name):
            return self.__keywords[name]
        else:
            return TokenType.UNKNOWN

    @staticmethod
    def register(name: str, type: TokenType) -> None:
        """Register a new keyword"""
        Symbols.__keywords[name] = type


#----- begin
# load all the keywords
import packages.keywords
