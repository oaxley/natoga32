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
# @brief	Function clone_token

#----- imports
from packages.structs import Token


#----- function
def clone_token(token: Token) -> Token:
    """Clone a Token with a Shallow Copy

    Args:
        token (Token): the token to copy

    Returns:
        Token: the shallow copy of the source token
    """
    return Token(token.type, token.value, token.row, token.col)
