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
# @brief	CPU dataclass

#----- imports
from dataclasses import dataclass


#----- class
@dataclass
class CPU:
    """CPU Configuration

    Members:
    - instr_size (int): the size of an instruction in byte
    - xlen (int): the size of an instruction in bit
    - byte (int): the size associated with the ".byte" directive
    - half (int): the size associated with the ".half" directive
    - word (int): the size associated with the ".word" directive
    """
    instr_size: int = 0
    xlen: int = 0
    byte: int = 0
    half: int = 0
    word: int = 0
