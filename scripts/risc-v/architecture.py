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
    'ori': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b110
    },
    'andi': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b001_0011,
        'funct3': 0b111
    },
    'auipc': {
        'type': OP_Format.TYPE_U,
        'opcode': 0b001_0111
    },
    'sb': {
        'type': OP_Format.TYPE_S,
        'opcode': 0b010_0011,
        'funct3': 0b000
    },
    'sh': {
        'type': OP_Format.TYPE_S,
        'opcode': 0b010_0011,
        'funct3': 0b001
    },
    'sw': {
        'type': OP_Format.TYPE_S,
        'opcode': 0b010_0011,
        'funct3': 0b010
    },
    'add': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b000,
        'tail': 0b0000_000
    },
    'sub': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b000,
        'tail': 0b0100_000
    },
    'sll': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b001,
        'tail': 0b0000_000
    },
    'slt': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b010,
        'tail': 0b0000_000
    },
    'sltu': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b011,
        'tail': 0b0000_000
    },
    'xor': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b100,
        'tail': 0b0000_000
    },
    'srl': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b101,
        'tail': 0b0000_000
    },
    'sra': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b101,
        'tail': 0b0100_000
    },
    'or': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b110,
        'tail': 0b0000_000
    },
    'and': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b111,
        'tail': 0b0000_000
    },
    'lui': {
        'type': OP_Format.TYPE_U,
        'opcode': 0b011_0111
    },
    'beq': {
        'type': OP_Format.TYPE_B,
        'opcode': 0b110_0011,
        'funct3': 0b000
    },
    'bne': {
        'type': OP_Format.TYPE_B,
        'opcode': 0b110_0011,
        'funct3': 0b001
    },
    'blt': {
        'type': OP_Format.TYPE_B,
        'opcode': 0b110_0011,
        'funct3': 0b100
    },
    'bge': {
        'type': OP_Format.TYPE_B,
        'opcode': 0b110_0011,
        'funct3': 0b101
    },
    'bltu': {
        'type': OP_Format.TYPE_B,
        'opcode': 0b110_0011,
        'funct3': 0b110
    },
    'bgeu': {
        'type': OP_Format.TYPE_B,
        'opcode': 0b110_0011,
        'funct3': 0b111
    },
    'jalr': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b110_0111,
        'funct3': 0b000
    },
    'jal': {
        'type': OP_Format.TYPE_J,
        'opcode': 0b110_1111,
    },
    'mul': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b000,
        'tail': 0b0000_001
    },
    'mulh': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b001,
        'tail': 0b0000_001
    },
    'mulhsu': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b010,
        'tail': 0b0000_001
    },
    'mulhu': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b011,
        'tail': 0b0000_001
    },
    'div': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b100,
        'tail': 0b0000_001
    },
    'divu': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b101,
        'tail': 0b0000_001
    },
    'rem': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b110,
        'tail': 0b0000_001
    },
    'remu': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b111,
        'tail': 0b0000_001
    },
    'csrrw': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b111_0011,
        'funct3': 0b001
    },
    'csrrs': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b111_0011,
        'funct3': 0b010
    },
    'csrrc': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b111_0011,
        'funct3': 0b011
    },
    'csrrwi': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b111_0011,
        'funct3': 0b101
    },
    'csrrsi': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b111_0011,
        'funct3': 0b110
    },
    'csrrci': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b111_0011,
        'funct3': 0b111
    },

    'clz': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b001,
        'tail': 0b0110_0000_0000
    },
    'ctz': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b001,
        'tail': 0b0110_0000_0001
    },
    'cpop': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b001,
        'tail': 0b0110_0000_0010
    },

    'rol': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b001,
        'tail': 0b0110_000
    },
    'ror': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b101,
        'tail': 0b0110_000
    },
    'rori': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b101,
        'tail': 0b0110_000
    },

    'rev8': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b101,
        'tail': 0b0110_1001_1000
    },

    'bclr': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b001,
        'tail': 0b0100_100
    },
    'bclri': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b001,
        'tail': 0b0100_100
    },
    'bext': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b101,
        'tail': 0b0100_100
    },
    'bexti': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b101,
        'tail': 0b0100_100
    },
    'binv': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b001,
        'tail': 0b0110_100
    },
    'binvi': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b001,
        'tail': 0b0110_100
    },
    'bset': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b001,
        'tail': 0b0010_100
    },
    'bseti': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b001_0011,
        'funct3': 0b001,
        'tail': 0b0010_100
    },

    'pack': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b100,
        'tail': 0b0000_100
    },
    'packh': {
        'type': OP_Format.TYPE_R,
        'opcode': 0b011_0011,
        'funct3': 0b111,
        'tail': 0b0000_100
    },

    'ecall': {
        'type': OP_Format.TYPE_I,
        'opcode': 0b0000_0001_0000_0000_0000_0000_0111_0011
    },
}
