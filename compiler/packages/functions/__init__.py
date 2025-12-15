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
# @brief	Functions package initializer

#----- imports
# helper functions
from .capture_body import capture_body
from .clone_token import clone_token
from .get_value import get_value
from .evaluate import evaluate_expr

# pre-processor functions
from .preproc import (
    handle_conditionals, define, apply_dup, envvar, handle_for_loop,
    include, parse_macro_definition, expand_macro, apply_token_pasting
)
