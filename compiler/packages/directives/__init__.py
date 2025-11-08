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
# @brief	__init__ file for assembler directives

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from importlib import import_module, resources


#----- begin

# register all the files present in the current directory
for name in resources.contents(__name__):
    # do not process this file (avoid recursion hell!)
    if name.endswith(".py") and name != "__init__.py":
        import_module(f"{__name__}.{name[:-3]}")
