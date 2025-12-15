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
from packages.structs import Token, Config


#----- globals
def include(filename: str, config: Config) -> List[Token]:
    """Handle '.include' directive

    Args:
        filename (str): the filename to parse/include
        config (Config): the global configuration object

    Returns:
        List[Token]: a list of tokens to insert at the ".include" position
    """
    # resolve the relative path
    full_path = os.path.join(os.path.dirname(config.infile), filename)
    if full_path in config.includes:
        raise Exception(f"Error ({config.row}, {config.col}): recursive include detected for file '{full_path}'")

    # process the file
    config.includes.add(full_path)
    try:
        # open the file
        try:
            fh = open(full_path, 'r', encoding='utf-8')
        except FileNotFoundError:
            raise SyntaxError(f"Error ({config.row}, {config.col}): include file '{full_path}' cannot be found")

        # execute the lexer
        lex = lexer.Lexer()
        lex.parse(fh)

        # return the list of tokens in the include file
        return lex.tokens

    finally:
        config.includes.remove(full_path)
