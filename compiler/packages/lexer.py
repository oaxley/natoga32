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
from typing import Any, Dict, List, ClassVar, Optional, Callable, Tuple

from packages.config import Config
from packages.token import Token, TokenType
from packages.reader import FileReader


#----- classes
class Lexer:
    """Process the assembly file and create tokens

    Notes:
        This class acts as a singleton to allow for auto-registration
    """

    # private members
    __instance: ClassVar[Optional[Lexer]] = None
    __keywords: ClassVar[Dict[str, TokenType]] = { }
    __config: ClassVar[Optional[Config]] = None

    # counters
    row: int = 1
    col: int = 1

    # queue
    __tokens: List[Token] = []

    def __new__(cls) -> Lexer:
        """Create a new instance of the Lexer or return the current one"""
        if Lexer.__instance is None:
            Lexer.__instance = object.__new__(cls)

        return Lexer.__instance

    @property
    def directives(self) -> List[str]:
        """Return a list of available directives"""
        return list(Lexer.__keywords.keys())

    def is_exist(self, name: str) -> bool:
        """Check if a directive exists"""
        return (name in Lexer.__keywords)

    @staticmethod
    def register(name: str, type: TokenType):
        """Register a new directives"""
        Lexer.__keywords[name] = type

    def set_config(self, config: Config) -> None:
        """Setup the configuration"""
        Lexer.__config = config

    def _config(self, name) -> Any:
        return Lexer.__config.__getattribute__(name)

    def _skip_comments(self, reader: FileReader) -> None:
        """skip the comments"""
        while(reader.next() != '\n'):
            pass
        self._emit_token(TokenType.EOL, "")
        Lexer.row += 1
        Lexer.col = 1

    def _get_string(self, reader: FileReader):
        """Retrieve the content of the string and emit the corresponding token"""
        string = ""
        while(True):
            c = reader.next()
            if c in ['"', '\n']:
                break
            else:
                string = string + c

        # emit the string
        self._emit_token(TokenType.STRING, string)

        # emit the last quote if it is present
        if c == '"':
            self._emit_token(TokenType.SYMBOL, c)

    def _emit_token(self, type: TokenType, string: str) -> None:
        """Add the new token to the queue"""
        Lexer.__tokens.append(Token(type, string, Lexer.row, Lexer.col))
        Lexer.col += len(string)

    def _parse_number(self, string: str) -> Tuple[bool, int]:
        """Parse a string and determine if it's a number of not"""
        try:
            # number without any sign
            if string.isdigit():
                return (True, int(string))

            # number with a sign
            if string[0] in ['-', '+']:
                if string[1:].isdigit():
                    return (True, int(string))

            # hexadecimal number: 0x or postfixed with h
            if (string[:2] == '0x'):
                return (True, int(string[2:], 16))
            if (string[-1] == 'h'):
                return (True, int(string[:-1], 16))

            # binary number: 0b or postfixed with b
            if (string[:2] == '0b'):
                return (True, int(string[2:], 2))
            if (string[-1] == 'b'):
                return (True, int(string[:-1], 2))

        except ValueError:
            pass

        return (False, 0)

    def _identify(self, string: str) -> Tuple[TokenType, str]:
        """Identify a string and returns the proper Token Type"""
        # check if we have a directive
        if self.is_exist(string):
            return (Lexer.__keywords[string], string)

        # a label end up with ':'
        if string[-1] == ':':
            return (TokenType.LABEL, string[:-1])

        # macro
        if string[0] == '%':
            return (TokenType.MACRO, string)

        # parse number
        value = self._parse_number(string)
        if value[0] is True:
            return (TokenType.NUMBER, str(value[1]))

        # default token type
        return (TokenType.IDENT, string)

    def tokens(self) -> List[Token]:
        return Lexer.__tokens

    def parse(self) -> None:
        """Parse the input file and create a flow of token"""
        reader = FileReader(self._config("input_file"))

        string = ""
        while(True):
            c = reader.next()
            if not c:
                self._emit_token(TokenType.EOF, "")
                break

            # remove the comments
            if c in [ '#', ';']:
                self._skip_comments(reader)
                continue

            # remove empty lines and white spaces
            if c in [' ', '\t', '\n'] and len(string) == 0:
                if c == '\n':
                    if Lexer.col > 1:
                        self._emit_token(TokenType.EOL, "")

                    Lexer.row += 1
                    Lexer.col = 1
                else:
                    Lexer.col += 1
                continue

            # special case for strings
            if c == '"':
                # emit the begining of the string
                self._emit_token(TokenType.SYMBOL, c)

                # retrieve the content of the string
                self._get_string(reader)

                # reset the string
                string = ""
                continue

            # check for delimiters
            if c in [' ', '\n', ',', '(', ')']:

                # try to identify the string
                if len(string) > 0:
                    i = self._identify(string)
                    self._emit_token(i[0], i[1])
                    string = ""

                # edge cases
                if c in ['(', ')']:
                    self._emit_token(TokenType.SYMBOL, c)
                    continue

                if c == '\n':
                    self._emit_token(TokenType.EOL, "")
                    Lexer.row += 1
                    Lexer.col = 1
                else:
                    Lexer.col += 1
            else:
                string = string + c


#----- begin
# load all the keywords (only once the class has been defined)
import packages.keywords
