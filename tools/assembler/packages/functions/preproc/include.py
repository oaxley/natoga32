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
# @brief	Manage includes during pre-processing

#----- imports
from typing import List

import os

from packages import lexer
from packages.structs import Token, Config, TokenType
from packages.classes import TokenStream
from packages.functions import get_value

#----- globals
def include(ts: TokenStream, config: Config) -> List[Token]:
    """Handle '.include' directive

    Args:
        ts (TokenStream): the current token stream
        config (Config): the global configuration object

    Returns:
        List[Token]: a list of tokens to insert at the ".include" position
    """
    # consume the .include, retrieve the filename and EOL
    ts.advance()
    string = get_value(ts, TokenType.STRING)
    ts.advance()

    filename = string[1:len(string)-1]

    # resolve the relative path
    full_path = os.path.join(os.path.dirname(config.infile), filename)
    if full_path in config.includes:
        raise Exception(f"Error: recursive include detected for file '{full_path}'")

    # process the file
    config.includes.add(full_path)
    try:
        # open the file
        try:
            fh = open(full_path, 'r', encoding='utf-8')
        except FileNotFoundError:
            raise SyntaxError(f"Error: include file '{full_path}' cannot be found")

        # execute the lexer
        lex = lexer.Lexer()
        lex.parse(fh)

        # return the list of tokens in the include file
        # minus the EOF
        return lex.tokens[:-1]

    finally:
        config.includes.remove(full_path)
