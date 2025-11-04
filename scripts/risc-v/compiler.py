#!/usr/bin/env python
# @author   Sebastien LEGRAND
#
# @brief    Risc-V tiny compiler (for testing purpose only)

#----- imports
from __future__ import annotations
from typing import TextIO, Tuple, Dict, List, Any

import sys
import argparse

from architecture import REGISTERS
from tokenizer import Tokenizer, TOKEN_ID


#----- class

class Lexer:
    def __init__(self) -> None:
        pass

    def parse(self, token: str) -> Tuple[int, str]:
        pass



#----- functions
def printToken(token: Tuple[TOKEN_ID, str, Any]) -> None:
    (t_id, t_string, t_value) = token
    print(f"token_id = {t_id:<3d} | token = {t_string:<10} | value = {t_value}")

#----- begin

# print(REGISTERS)
# sys.exit(0)

# read the command line arguments
parser = argparse.ArgumentParser()
parser.add_argument("-f","--file",required=True, help="assembly file to parse")
args = parser.parse_args()

tokenizer = Tokenizer(args.file)
tokenizer.openFile()

string = ""
while(True):
    c = tokenizer.next()
    if not c:
        break

    if c in [' ', '\n', '\t'] and len(string) == 0:
        continue

    # special case for label
    if c == ':':
        string = string + c
        token = tokenizer.identify(string)
        printToken(token)
        string = ""

    elif c in [' ',',','(',')','\n']:
        token = tokenizer.identify(string)
        printToken(token)
        string = ""

    else:
        string = string + c



