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
# @brief	Pseudo-Instruction expansion

#----- imports
from typing import List, cast

from packages import ast
from packages.structs import CPU

#----- functions
def instr_expand(cpu: CPU, instr: ast.Instruction) -> List[ast.Instruction]:
    """Expland the pseudo-instruction into standard Risc-V instructions

    Args:
        instr (ast.Instruction): the instruction to expand

    Returns:
        List[ast.Instruction]: the list of instruction following the expansion
    """
    instructions: List[ast.Instruction] = []

    #--- lambdas definition
    # create the label identifier from an operand
    label = lambda x: ast.Identifier(cast(ast.Identifier, instr.operands[x]).name)

    # create the register identifier
    x = lambda n: ast.Identifier(f"x{n}")

    # definition of the number '0'
    zero = ast.Number(0)

    # groups of aliases to expand
    instr_0arg = ['nop', 'ret']
    instr_1arg = ['call', 'j', 'jr']
    instr_2arg = ['mv', 'seqz', 'snez', 'sltz', 'sgtz', 'not', 'neg']
    branch_1 = ['beqz', 'bnez', 'blez', 'bgez', 'bltz', 'bgtz']
    branch_2 = ['bgt', 'ble', 'bgtu', 'bleu']

    if instr.opcode == "li":
        rd, imm = instr.operands

        if isinstance(imm, ast.Number):
            if -2048 <= imm.value <= 2047:
                return [
                    ast.Instruction("addi", [rd, x(0), imm])
                ]

            # big constant
            upper = (imm.value + 0x800) >> 12
            lower = imm.value - (upper << 12)

            return [
                ast.Instruction("lui", [rd, ast.Number(upper)]),
                ast.Instruction("addi", [rd, rd, ast.Number(lower)])
            ]

        if isinstance(imm, ast.Expression):
            return [
                ast.Instruction("lui", [rd, ast.HiRel(imm)]),
                ast.Instruction("addi", [rd, rd, ast.LoRel(imm)])
            ]

    elif instr.opcode == "la":
        rd = instr.operands[0]
        offset = label(1)
        return [
            ast.Instruction('auipc', [rd, ast.PCRelHi(offset)]),
            ast.Instruction('addi', [rd, rd, ast.PCRelLo(offset)])
        ]

    elif instr.opcode.startswith("csr"):
        match instr.opcode:
            case "csrr":
                rd, csr = instr.operands
                return [ ast.Instruction('csrrs', [rd, csr, x(0)])]
            case "csrw":
                csr, rs1 = instr.operands
                return [ ast.Instruction('csrrw', [x(0), csr, rs1])]
            case "csrs":
                csr, rs1 = instr.operands
                return [ ast.Instruction('csrrs', [x(0), csr, rs1])]
            case "csrc":
                csr, rs1 = instr.operands
                return [ ast.Instruction('csrrc', [x(0), csr, rs1])]
            case "csrwi":
                csr, imm = instr.operands
                return [ ast.Instruction('csrrwi', [x(0), csr, imm])]
            case "csrsi":
                csr, imm = instr.operands
                return [ ast.Instruction('csrrsi', [x(0), csr, imm])]
            case "csrci":
                csr, imm = instr.operands
                return [ ast.Instruction('csrrci', [x(0), csr, imm])]

    elif instr.opcode in instr_0arg:
        match instr.opcode:
            case "nop":
                return [ ast.Instruction('addi', [x(0), x(0), zero]) ]
            case "ret":
                return [ast.Instruction("jalr", [ x(0), x(1), zero])]

    elif instr.opcode in instr_1arg:
        offset = label(0)
        match instr.opcode:
            case "call":
                return [
                    ast.Instruction('auipc', [x(1), ast.PCRelHi(offset)]),
                    ast.Instruction('jalr', [x(1), ast.PCRelLo(offset), x(1)])
                ]
            case "j":
                return [ ast.Instruction('jal', [x(0), offset]) ]
            case "jr":
                return [ ast.Instruction('jal', [x(1), offset])]

    elif instr.opcode in instr_2arg:
        rd = instr.operands[0]
        rs = instr.operands[1]

        match instr.opcode:
            case "mv":
                return [ast.Instruction('addi', [rd, rs, zero])]
            case "seqz":
                return [ast.Instruction('sltiu', [rd, rs, ast.Number(1)])]
            case "snez":
                return [ast.Instruction('sltu', [rd, x(0), rs])]
            case "sltz":
                return [ast.Instruction('slt', [rd, rs, x(0)])]
            case "sgtz":
                return [ast.Instruction('slt', [rd, x(0), rs])]
            case "not":
                return [ast.Instruction('xori', [rd, rs, ast.Number(-1)])]
            case "neg":
                return [ast.Instruction('sub', [rd, x(0), rs])]

    elif instr.opcode in branch_1:
        rs = instr.operands[0]
        offset = label(1)

        match instr.opcode:
            case 'beqz':
                return [ast.Instruction('beq', [rs, x(0), offset])]
            case 'bnez':
                return [ast.Instruction('bne', [rs, x(0), offset])]
            case 'blez':
                return [ast.Instruction('bge', [x(0), rs, offset])]
            case 'bgez':
                return [ast.Instruction('bge', [rs, x(0), offset])]
            case 'bltz':
                return [ast.Instruction('blt', [rs, x(0), offset])]
            case 'bgtz':
                return [ast.Instruction('blt', [x(0), rs, offset])]

    elif instr.opcode in branch_2:
        rs = instr.operands[0]
        rt = instr.operands[1]
        offset = label(2)

        match instr.opcode:
            case 'bgt':
                return [ast.Instruction('blt', [rt, rs, offset])]
            case 'ble':
                return [ast.Instruction('bge', [rt, rs, offset])]
            case 'bgtu':
                return [ast.Instruction('bltu', [rt, rs, offset])]
            case 'bleu':
                return [ast.Instruction('bgeu', [rt, rs, offset])]

    # misc pseudo-instruction
    elif instr.opcode == "syscall":
        param = instr.operands[0]
        if isinstance(param, ast.Number):
            return [
                ast.Instruction("addi", [x(17), x(0), param]),
                ast.Instruction("ecall", [])
            ]
        elif isinstance(param, ast.Identifier):
            if (param.name in cpu.registers):
                return [
                    ast.Instruction("add", [x(17), x(0), param]),
                    ast.Instruction("ecall", [])
                ]
            else:
                raise SyntaxError(f"Error: {param.name} is not a valid register in syscall")

    else:
        instructions.append(instr)

    return instructions
