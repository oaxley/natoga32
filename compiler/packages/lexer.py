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
from typing import List

from packages.config import Config
from packages.token import Token

from packages.specs import TokenType, RE_PATTERNS


#----- classes
class Lexer:
    """Process the assembly file and create tokens"""

    def __init__(self, config: Config) -> None:
        """Constructor"""
        # current row/col
        self._row: int = 1
        self._col: int = 1

        self._config = config
        self._tokens: List[Token] = []

    def parse(self) -> None:
        with open(self._config.input_file, "r") as fh:
            for line in fh:
                self._row = self._row + 1
                self.tokenize(line)

            # end of file
            self._tokens.append(Token(TokenType.EOF, "", self._row, self._col))

    def tokenize(self, line) -> None:
        # remove the newline delimiter
        line = line.rstrip('\n')

        for match in RE_PATTERNS.finditer(line):
            kind = match.lastgroup
            value = match.group()
            self._col = match.start() + 1

            # skip comment and whitespaces
            if kind in [ TokenType.SKIP.name, TokenType.COMMENT.name ]:
                continue

            print(f'kind = {kind} | value = [{value}]')

    @property
    def tokens(self) -> List[Token]:
        return self._tokens
