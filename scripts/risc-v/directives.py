# @file     directives.py
# @author   Sebastien LEGRAND
#
# @brief    Compiler internal directives

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from enum import IntEnum, auto


#----- globals

class Dir_Type(IntEnum):
    BSS = auto()
    CODE = auto()
    ORIGIN = auto()
    MACRO = auto()
    DATA = auto()
    UNKNOWN = auto()


DIRECTIVES = {
    '.code': {
        'type': Dir_Type.CODE
    },
    'ORG': {
        'type': Dir_Type.ORIGIN
    },
    '.data': {
        'type': Dir_Type.BSS
    },
    'db': {
        'type': Dir_Type.DATA
    },
    '%lo': {
        'type': Dir_Type.MACRO
    },
    '%hi': {
        'type': Dir_Type.MACRO
    }
}
