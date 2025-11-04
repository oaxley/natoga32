# @file     architecture.py
# @author   Sebastien LEGRAND
#
# @brief    RISC-V Architecture file

#----- imports
from __future__ import annotations
from typing import Any, List, Dict

from enum import IntEnum, auto


#----- globals

# registers map with aliases
REGISTERS = {}
REGISTERS = { f'x{i}': i for i in range(0, 32) }
REGISTERS['zero'] = 0
REGISTERS['ra'] = 1
REGISTERS['sp'] = 2
REGISTERS['gp'] = 3
REGISTERS['tp'] = 4
REGISTERS['t0'] = 5
REGISTERS['t1'] = 6
REGISTERS['t2'] = 7
REGISTERS['s0'] = 8
REGISTERS['fp'] = 8
REGISTERS['s1'] = 9
REGISTERS['a0'] = 10
REGISTERS['a1'] = 11
for i in range (2, 8):
    REGISTERS[f'a{i}'] = i + 10

for i in range (2, 12):
    REGISTERS[f's{i}'] = i + 16

for i in range (3, 7):
    REGISTERS[f't{i}'] = i + 25


# opcode formats
class OP_Format(IntEnum):
    TYPE_I = auto()
    TYPE_U = auto()
    TYPE_S = auto()
    TYPE_R = auto()
    TYPE_B = auto()
    TYPE_J = auto()

# opcodes map
OPCODES = {
    'lb': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b000_0011,
        'funct3': 0b000
    },
    'lh': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b000_0011,
        'funct3': 0b001
    },
    'lw': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b000_0011,
        'funct3': 0b010
    },
    'lbu': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b000_0011,
        'funct3': 0b100
    },
    'lhu': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b000_0011,
        'funct3': 0b101
    },
    'addi': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b000
    },
    'slli': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b001,
        'tail': 0b0000_000
    },
    'slti': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b010
    },
    'sltiu': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b011
    },
    'xori': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b100
    },
    'srli': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b101,
        'tail': 0b0000_000
    },
    'srai': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b101,
        'tail': 0b0100_000
    },
}
