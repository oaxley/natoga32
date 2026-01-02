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
# @brief	RISC-V Opcode definition

#----- imports
from typing import Dict
from .structs import InstrType, InstrOpcode


#----- globals
OPCODES_T: Dict[str, InstrOpcode] = {
    'lb'    : InstrOpcode(InstrType.TYPE_I ,0b000_0011, 0, 0b000, 0, 0, 0          ),
    'lh'    : InstrOpcode(InstrType.TYPE_I ,0b000_0011, 0, 0b001, 0, 0, 0          ),
    'lw'    : InstrOpcode(InstrType.TYPE_I ,0b000_0011, 0, 0b010, 0, 0, 0          ),
    'lbu'   : InstrOpcode(InstrType.TYPE_I ,0b000_0011, 0, 0b100, 0, 0, 0          ),
    'lhu'   : InstrOpcode(InstrType.TYPE_I ,0b000_0011, 0, 0b101, 0, 0, 0          ),
    'addi'  : InstrOpcode(InstrType.TYPE_I ,0b001_0011, 0, 0b000, 0, 0, 0          ),
    'slli'  : InstrOpcode(InstrType.TYPE_I2,0b001_0011, 0, 0b001, 0, 0, 0          ),
    'slti'  : InstrOpcode(InstrType.TYPE_I ,0b001_0011, 0, 0b010, 0, 0, 0          ),
    'sltiu' : InstrOpcode(InstrType.TYPE_I ,0b001_0011, 0, 0b011, 0, 0, 0          ),
    'xori'  : InstrOpcode(InstrType.TYPE_I ,0b001_0011, 0, 0b100, 0, 0, 0          ),
    'srli'  : InstrOpcode(InstrType.TYPE_I2,0b001_0011, 0, 0b101, 0, 0, 0          ),
    'srai'  : InstrOpcode(InstrType.TYPE_I2,0b001_0011, 0, 0b101, 0, 0, 0b0100_000 ),
    'ori'   : InstrOpcode(InstrType.TYPE_I ,0b001_0011, 0, 0b110, 0, 0, 0          ),
    'andi'  : InstrOpcode(InstrType.TYPE_I ,0b001_0011, 0, 0b111, 0, 0, 0          ),
    'jalr'  : InstrOpcode(InstrType.TYPE_I ,0b110_0111, 0, 0b000, 0, 0, 0          ),

    'csrrw' : InstrOpcode(InstrType.TYPE_I ,0b111_0011, 0, 0b001, 0, 0, 0          ),
    'csrrs' : InstrOpcode(InstrType.TYPE_I ,0b111_0011, 0, 0b010, 0, 0, 0          ),
    'csrrc' : InstrOpcode(InstrType.TYPE_I ,0b111_0011, 0, 0b011, 0, 0, 0          ),
    'csrrwi': InstrOpcode(InstrType.TYPE_I ,0b111_0011, 0, 0b101, 0, 0, 0          ),
    'csrrsi': InstrOpcode(InstrType.TYPE_I ,0b111_0011, 0, 0b110, 0, 0, 0          ),
    'csrrci': InstrOpcode(InstrType.TYPE_I ,0b111_0011, 0, 0b111, 0, 0, 0          ),

    'auipc' : InstrOpcode(InstrType.TYPE_U ,0b001_0111, 0, 0b000, 0 ,0, 0          ),
    'lui'   : InstrOpcode(InstrType.TYPE_U ,0b011_0111, 0, 0b000, 0 ,0, 0          ),

    'add'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b000, 0, 0, 0b0000_000 ),
    'sub'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b000, 0, 0, 0b0100_000 ),
    'sll'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b001, 0, 0, 0b0000_000 ),
    'slt'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b010, 0, 0, 0b0000_000 ),
    'sltu'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b011, 0, 0, 0b0000_000 ),
    'xor'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0, 0b0000_000 ),
    'srl'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b101, 0, 0, 0b0000_000 ),
    'sra'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b101, 0, 0, 0b0100_000 ),
    'or'    : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b110, 0, 0, 0b0000_000 ),
    'and'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b111, 0, 0, 0b0000_000 ),

    'mul'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b000, 0, 0, 0b0000_001 ),
    'mulh'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b001, 0, 0, 0b0000_001 ),
    'mulhsu': InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b010, 0, 0, 0b0000_001 ),
    'mulhu' : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b011, 0, 0, 0b0000_001 ),
    'div'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0, 0b0000_001 ),
    'divu'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b101, 0, 0, 0b0000_001 ),
    'rem'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b110, 0, 0, 0b0000_001 ),
    'remu'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b111, 0, 0, 0b0000_001 ),

    'sext.h': InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0b00101, 0b0110_000 ),
    'zext.h': InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0b00000, 0b0000_100 ),
    'clz'   : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0b00000, 0b0110_000 ),
    'ctz'   : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0b00001, 0b0110_000 ),
    'cpop'  : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0b00010, 0b0110_000 ),
    'rev8'  : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b101, 0, 0b11000, 0b0110_100 ),

    'rol'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b001, 0, 0, 0b0110_000 ),
    'ror'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b101, 0, 0, 0b0110_000 ),
    'rori'  : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b101, 0, 0, 0b0110_000 ),
    'bclr'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b001, 0, 0, 0b0100_100 ),
    'bclri' : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0, 0b0100_100 ),
    'bext'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b101, 0, 0, 0b0100_100 ),
    'bexti' : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b101, 0, 0, 0b0100_100 ),
    'binv'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b001, 0, 0, 0b0110_100 ),
    'binvi' : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0, 0b0110_100 ),
    'bset'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b001, 0, 0, 0b0010_100 ),
    'bseti' : InstrOpcode(InstrType.TYPE_R ,0b001_0011, 0, 0b001, 0, 0, 0b0010_100 ),

    'pack'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0, 0b0000_100 ),
    'packh' : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b111, 0, 0, 0b0000_100 ),
    'max'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b110, 0, 0, 0b0000_101 ),
    'maxu'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b111, 0, 0, 0b0000_101 ),
    'min'   : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b100, 0, 0, 0b0000_101 ),
    'minu'  : InstrOpcode(InstrType.TYPE_R ,0b011_0011, 0, 0b101, 0, 0, 0b0000_101 ),

    'sb'    : InstrOpcode(InstrType.TYPE_S ,0b010_0011, 0, 0b000, 0, 0, 0),
    'sh'    : InstrOpcode(InstrType.TYPE_S ,0b010_0011, 0, 0b001, 0, 0, 0),
    'sw'    : InstrOpcode(InstrType.TYPE_S ,0b010_0011, 0, 0b010, 0, 0, 0),

    'beq'   : InstrOpcode(InstrType.TYPE_B ,0b110_0011, 0, 0b000, 0, 0, 0),
    'bne'   : InstrOpcode(InstrType.TYPE_B ,0b110_0011, 0, 0b001, 0, 0, 0),
    'blt'   : InstrOpcode(InstrType.TYPE_B ,0b110_0011, 0, 0b100, 0, 0, 0),
    'bge'   : InstrOpcode(InstrType.TYPE_B ,0b110_0011, 0, 0b101, 0, 0, 0),
    'bltu'  : InstrOpcode(InstrType.TYPE_B ,0b110_0011, 0, 0b110, 0, 0, 0),
    'bgeu'  : InstrOpcode(InstrType.TYPE_B ,0b110_0011, 0, 0b111, 0, 0, 0),

    'jal'   : InstrOpcode(InstrType.TYPE_J ,0b110_1111, 0, 0b000, 0, 0, 0),

    'ecall' : InstrOpcode(InstrType.TYPE_I, 0b111_0011, 0, 0b000, 0, 0, 0),
    'mret'  : InstrOpcode(InstrType.TYPE_I, 0b111_0011, 0, 0b000, 0, 0b00010, 0),
    'wfi'   : InstrOpcode(InstrType.TYPE_I, 0b111_0011, 0, 0b000, 0, 0b00101, 0b0001_000),

    'new.t'  : InstrOpcode(InstrType.TYPE_R, 0b000_1011, 0, 0b000, 0, 0, 0),
    'yield.t': InstrOpcode(InstrType.TYPE_I, 0b000_1011, 0, 0b001, 0, 0, 0),
    'id.t'   : InstrOpcode(InstrType.TYPE_I, 0b000_1011, 0, 0b010, 0, 0, 0),
    'sleep.t': InstrOpcode(InstrType.TYPE_R, 0b000_1011, 0, 0b100, 0, 0, 0),
    'wake.t' : InstrOpcode(InstrType.TYPE_R, 0b000_1011, 0, 0b101, 0, 0, 0),
    'end.t'  : InstrOpcode(InstrType.TYPE_I, 0b000_1011, 0, 0b111, 0, 0, 0),
}
