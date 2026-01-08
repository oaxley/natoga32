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
    # opcode   :             Instruction Type , Opcode    , rd     , func3, rs1    , rs2    , funct7

    # opcode 0x03 (3)
    'lb'       : InstrOpcode(InstrType.TYPE_I , 0b000_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),
    'lh'       : InstrOpcode(InstrType.TYPE_I , 0b000_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0000),
    'lw'       : InstrOpcode(InstrType.TYPE_I , 0b000_0011, 0b00000, 0b010, 0b00000, 0b00000, 0b000_0000),
    'lbu'      : InstrOpcode(InstrType.TYPE_I , 0b000_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0000),
    'lhu'      : InstrOpcode(InstrType.TYPE_I , 0b000_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x0B (11)
    'new.t'    : InstrOpcode(InstrType.TYPE_R , 0b000_1011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),
    'yield.t'  : InstrOpcode(InstrType.TYPE_I , 0b000_1011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0000),
    'id.t'     : InstrOpcode(InstrType.TYPE_I , 0b000_1011, 0b00000, 0b010, 0b00000, 0b00000, 0b000_0000),
    'sleep.t'  : InstrOpcode(InstrType.TYPE_R , 0b000_1011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0000),
    'wake.t'   : InstrOpcode(InstrType.TYPE_R , 0b000_1011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0000),
    'end.t'    : InstrOpcode(InstrType.TYPE_I , 0b000_1011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x13 (19)
    'addi'     : InstrOpcode(InstrType.TYPE_I , 0b001_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),
    'slli'     : InstrOpcode(InstrType.TYPE_I2, 0b001_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0000),
    'bseti'    : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b001_0100),
    'bclri'    : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b010_0100),
    'sext.h'   : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00101, 0b011_0000),
    'clz'      : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b011_0000),
    'ctz'      : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00001, 0b011_0000),
    'cpop'     : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00010, 0b011_0000),
    'sext.b'   : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00100, 0b011_0000),
    'binvi'    : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b011_0000),
    'slti'     : InstrOpcode(InstrType.TYPE_I , 0b001_0011, 0b00000, 0b010, 0b00000, 0b00000, 0b000_0000),
    'sltiu'    : InstrOpcode(InstrType.TYPE_I , 0b001_0011, 0b00000, 0b011, 0b00000, 0b00000, 0b000_0000),
    'xori'     : InstrOpcode(InstrType.TYPE_I , 0b001_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0000),
    'srli'     : InstrOpcode(InstrType.TYPE_I2, 0b001_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0000),
    'srai'     : InstrOpcode(InstrType.TYPE_I2, 0b001_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b010_0000),
    'bexti'    : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b010_0100),
    'rori'     : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b011_0000),
    'rev8'     : InstrOpcode(InstrType.TYPE_R , 0b001_0011, 0b00000, 0b101, 0b00000, 0b11000, 0b011_0100),
    'ori'      : InstrOpcode(InstrType.TYPE_I , 0b001_0011, 0b00000, 0b110, 0b00000, 0b00000, 0b000_0000),
    'andi'     : InstrOpcode(InstrType.TYPE_I , 0b001_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x17 (23)
    'auipc'    : InstrOpcode(InstrType.TYPE_U , 0b001_0111, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x23 (35)
    'sb'       : InstrOpcode(InstrType.TYPE_S , 0b010_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),
    'sh'       : InstrOpcode(InstrType.TYPE_S , 0b010_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0000),
    'sw'       : InstrOpcode(InstrType.TYPE_S , 0b010_0011, 0b00000, 0b010, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x33 (51)
    'add'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),
    'mul'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0001),
    'sub'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b010_0000),
    'sll'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0000),
    'mulh'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0001),
    'bset'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b001_0100),
    'bclr'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b010_0100),
    'rol'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b011_0000),
    'binv'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b011_0100),
    'slt'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b010, 0b00000, 0b00000, 0b000_0000),
    'mulhsu'   : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b010, 0b00000, 0b00000, 0b000_0001),
    'sltu'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b011, 0b00000, 0b00000, 0b000_0000),
    'mulhu'    : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b011, 0b00000, 0b00000, 0b000_0001),
    'xor'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0000),
    'div'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0001),
    'zext.h'   : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0100),
    'pack'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0100),
    'min'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0101),
    'srl'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0000),
    'divu'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0001),
    'minu'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0101),
    'czero.eqz': InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0111),
    'sra'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b010_0000),
    'bext'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b010_0100),
    'ror'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b011_0000),
    'or'       : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b110, 0b00000, 0b00000, 0b000_0000),
    'rem'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b110, 0b00000, 0b00000, 0b000_0001),
    'max'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b110, 0b00000, 0b00000, 0b000_0101),
    'and'      : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0000),
    'remu'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0001),
    'packh'    : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0100),
    'maxu'     : InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0101),
    'czero.nez': InstrOpcode(InstrType.TYPE_R , 0b011_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0111),

    # opcode 0x37 (55)
    'lui'      : InstrOpcode(InstrType.TYPE_U , 0b011_0111, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x63 (99)
    'beq'      : InstrOpcode(InstrType.TYPE_B , 0b110_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),
    'bne'      : InstrOpcode(InstrType.TYPE_B , 0b110_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0000),
    'blt'      : InstrOpcode(InstrType.TYPE_B , 0b110_0011, 0b00000, 0b100, 0b00000, 0b00000, 0b000_0000),
    'bge'      : InstrOpcode(InstrType.TYPE_B , 0b110_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0000),
    'bltu'     : InstrOpcode(InstrType.TYPE_B , 0b110_0011, 0b00000, 0b110, 0b00000, 0b00000, 0b000_0000),
    'bgeu'     : InstrOpcode(InstrType.TYPE_B , 0b110_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x67 (103)
    'jalr'     : InstrOpcode(InstrType.TYPE_I , 0b110_0111, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x6F (111)
    'jal'      : InstrOpcode(InstrType.TYPE_J , 0b110_1111, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),

    # opcode 0x73 (115)
    'ecall'    : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b000, 0b00000, 0b00000, 0b000_0000),
    'mret'     : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b000, 0b00000, 0b00010, 0b001_1000),
    'wfi'      : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b000, 0b00000, 0b00101, 0b000_1000),
    'csrrw'    : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b001, 0b00000, 0b00000, 0b000_0000),
    'csrrs'    : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b010, 0b00000, 0b00000, 0b000_0000),
    'csrrc'    : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b011, 0b00000, 0b00000, 0b000_0000),
    'csrrwi'   : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b101, 0b00000, 0b00000, 0b000_0000),
    'csrrsi'   : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b110, 0b00000, 0b00000, 0b000_0000),
    'csrrci'   : InstrOpcode(InstrType.TYPE_I , 0b111_0011, 0b00000, 0b111, 0b00000, 0b00000, 0b000_0000),
}
