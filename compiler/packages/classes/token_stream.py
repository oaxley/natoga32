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
# @brief	TokenStream class

#----- imports
from typing import List, Optional, Union

from packages.structs import Token, TokenType


#----- class

class TokenStream:
    """Class to hold a stream of tokens"""
    def __init__(self, tokens: List[Token]) -> None:
        """Constructor"""
        self.tokens = tokens
        self.pos = 0

    def peek(self) -> Optional[Token]:
        """Return the next token, without consuming it

        Returns:
            Optional[Token]: None if we are beyond boundaries, the next token otherwise
        """
        if self.pos >= len(self.tokens):
            return None
        return self.tokens[self.pos]

    def advance(self) -> Optional[Token]:
        """Return the current token and move the cursor forward to the next token

        Returns:
            Optional[Token]: None if no token can be found, the current token otherwise
        """
        token = self.peek()
        if token is None:
            return None
        self.pos += 1
        return token

    def expect(self, what: Union[TokenType, str]) -> bool:
        """Check the current token against conditions

        - If what is a TokenType, token is tested against its type.
        - If what is a str, token is tested against its value.

        Args:
            what (Union[TokenType, str]): condition to test the current token against

        Returns:
            bool: True if the current token satisfies all the conditions, False otherwise
        """
        token = self.peek()
        if token:
            if isinstance(what, TokenType) and token.type == what:
                return True
            elif isinstance(what, str) and token.value == what:
                return True

        return False

    def at_end(self) -> bool:
        """Check if we are at the end of the token stream

        Returns:
            bool: True if we are at the end of the stream, False otherwise
        """
        token = self.peek()
        if (token is None) or (token.type == TokenType.EOF):
            return True

        return False
