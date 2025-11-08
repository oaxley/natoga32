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
from typing import Any, Dict, List, ClassVar, Optional, Callable

from packages.config import Config
from packages.token import Token, TokenType

#----- classes
class Lexer:
    """Process the assembly file and create tokens

    Notes:
        This class acts as a singleton to allow for auto-registration
    """

    __instance: ClassVar[Optional[Lexer]] = None
    __directives: ClassVar[Dict[str, TokenType]] = { }
    __config: ClassVar[Optional[Config]] = None

    def __new__(cls) -> Lexer:
        """Create a new instance of the Lexer or return the current one"""
        if Lexer.__instance is None:
            Lexer.__instance = object.__new__(cls)

        return Lexer.__instance

    @property
    def directives(self) -> List[str]:
        """Return a list of available directives"""
        return list(Lexer.__directives.keys())

    def is_exist(self, name: str) -> bool:
        """Check if a directive exists"""
        return (name in Lexer.__directives)

    @staticmethod
    def register(name: str, type: TokenType):
        """Register a new directives"""
        Lexer.__directives[name] = type

    def set_config(self, config: Config) -> None:
        """Setup the configuration"""
        Lexer.__config = config


#----- begin
# load all the directives (only once the class has been defined)
import packages.directives
