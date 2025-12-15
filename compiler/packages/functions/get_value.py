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
# @brief	Fucntion get_value

#----- imports
from typing import Optional

from packages.structs import TokenType
from packages.classes import TokenStream


#----- functions

def get_value(ts: TokenStream, ttype: TokenType, tvalue: Optional[str] = None) -> str:
    """Return next token value from the stream, only if its type correspond to ttype
    or its value is equal to tvalue

    Args:
        ts (TokenStream): the token stream
        ttype (TokenType) : the reference token type that token should match
        tvalue (Optional[str]): the reference value that token should match

    Returns:
        str: the token value
    """
    token = ts.advance()

    if not token or token.type != ttype:
        raise SyntaxError(f"Token value is either None or of the wrong type (expected {ttype})!")

    if tvalue and token.value != tvalue:
        raise SyntaxError(f"Expecting '{tvalue}', got '{token.value}'!")

    return token.value
