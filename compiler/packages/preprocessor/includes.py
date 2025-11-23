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
# @brief	Handle .includes directive

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Set

import os

from packages import lexer
from packages.token import TokenStream, Token, TokenType
from . import helper


#----- globals
# track the list of includes to avoid infinite recursion
include_stack : Set[str] = set()


#----- functions
def handle_include(ts: TokenStream, current_file: str) -> List[Token]:
    """Handle '.include' directive

    Args:
        ts (TokenStream): the token stream instance
        current_file (str): the current file being processed

    Returns:
        List[Token]: a list of tokens to insert at the ".include" position
    """
    ts.advance()    # consumer .include

    # retrieve the filename (STRING)
    filename = helper.get_value(ts, TokenType.STRING).strip('"')

    # consume EOL if present
    if ts.expect(TokenType.EOL): ts.advance()

    # resolve relative path
    full_path = os.path.join(os.path.dirname(current_file), filename)
    if full_path in include_stack:
        raise Exception(f"Recursive include detected: {full_path}")

    # process the file - add it to the set to avoid infinite recursion
    include_stack.add(full_path)
    try:
        # --- open file
        try:
            fh = open(full_path, 'r', encoding='utf-8')
        except FileNotFoundError:
            raise SyntaxError(f"File not found during include [{full_path}]")

        # --- lexer processing
        lex = lexer.Lexer()
        lex.parse(fh)

        return lex.tokens

    finally:
        # remove the file from the set
        include_stack.remove(full_path)
