# @file     tokenizer.py
# @author   Sebastien LEGRAND
#
# @brief    Compiler tokenizer

#----- imports
from __future__ import annotations
from typing import TextIO, Tuple, Set, Any

from enum import IntEnum, auto

from architecture import REGISTERS, OPCODES
from directives import DIRECTIVES

#----- globals

# token IDs
class TOKEN_ID(IntEnum):
    OPCODE = auto()
    REGISTER = auto()
    IMMEDIATE = auto()
    LABEL = auto()
    DIRECTIVE = auto()
    UNKNOWN = (1 << 8) - 1

#----- class
class Tokenizer:
    def __init__(self, filename: str) -> None:
        self.filename: str = filename
        self.fh: TextIO = None
        self.labels: Set[str] = set()

    def openFile(self) -> None:
        self.fh = open(self.filename, "r")

    def next(self) -> str:
        return self.fh.read(1)

    def peek(self) -> str:
        c = self.fh.read(1)
        self.fh.seek(-1, 1)
        return c

    def identify(self, string) -> Tuple[TOKEN_ID, str, Any]:
        if string in OPCODES:
            return (TOKEN_ID.OPCODE, string, OPCODES[string])
        elif string in REGISTERS:
            return (TOKEN_ID.REGISTER, string, REGISTERS[string])
        elif string[-1] == ':':
            self.labels.add(string[:-1])                # keep track of labels
            return (TOKEN_ID.LABEL, string[:-1], None)
        elif string in self.labels:
            return (TOKEN_ID.LABEL, string, None)
        elif string in DIRECTIVES:
            return (TOKEN_ID.DIRECTIVE, string, DIRECTIVES[string])

        # check for numbers
        (is_valid, value) = self.parseNumber(string)
        if is_valid:
            return (TOKEN_ID.IMMEDIATE, value, None)

        return (TOKEN_ID.LABEL, string, None)

    def parseNumber(self, string) -> Tuple[bool, int|str]:
        try:
            # number without any sign
            if string.isdigit():
                return (True, int(string))

            # number with a sign
            if string[0] in ['-', '+']:
                if string[1:].isdigit():
                    return (True, int(string))

            # hex number (0x) / '....h'
            if (string[:2] == '0x') and (string[2:].isdigit()):
                return (True, int(string[2:], 16))
            if (string[-1] == 'h') and (string[:-1].isdigit()):
                return (True, int(string[:-1], 16))

            # binary number (0b) / '....b'
            if (string[:2] == '0b') and (string[2:].isdigit()):
                return (True, int(string[2:], 2))
            if (string[-1] == 'b') and (string[:-1].isdigit()):
                return (True, int(string[:-1], 2))

        except ValueError:
            pass

        return (False, string)
