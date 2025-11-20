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
from typing import List, Union, TextIO

from packages.token import Token
from packages.specs import TokenType, RE_PATTERNS


#----- classes
class Lexer:
    """Process the assembly file and create tokens"""

    def __init__(self) -> None:
        """Constructor"""
        self._row: int = 0
        self._tokens: List[Token] = []

    def parse(self, content: Union[str, TextIO]) -> None:
        """Parse either from a file or a string"""
        if isinstance(content, str):
            # content is a string buffer
            for line in content.splitlines():
                self._row = self._row + 1
                self.tokenize(line)
        else:
            # content is a file handler
            for line in content:
                self._row = self._row + 1
                self.tokenize(line)

        # end of file
        self._tokens.append(Token(TokenType.EOF, "", self._row, 0))

    def tokenize(self, line) -> None:
        # remove the newline delimiter
        line = line.rstrip('\n')

        column = 0
        need_eol = False
        for match in RE_PATTERNS.finditer(line):
            kind = match.lastgroup
            value = match.group()
            column = match.start() + 1

            # skip comment and whitespaces
            if kind in [ TokenType.SKIP.name, TokenType.COMMENT.name ]:
                continue

            if kind:
                self._tokens.append(Token(TokenType.__getitem__(kind), value, self._row, column))
                need_eol = True

        # add the EOL
        if need_eol:
            self._tokens.append(Token(TokenType.EOL, "", self._row, len(line) + 1))

    @property
    def tokens(self) -> List[Token]:
        return self._tokens
