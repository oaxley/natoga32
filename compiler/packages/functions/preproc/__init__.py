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
# @brief	Pre-Processor functions initializer

#----- imports
from .conditionals import handle_conditionals
from .define import define
from .dup import apply_dup
from .environment import envvar
from .for_loops import handle_for_loop
from .include import include
from .macro import parse_macro_definition, expand_macro
from .token_pasting import apply_token_pasting
