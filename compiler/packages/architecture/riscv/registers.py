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
# @brief	RISC-V Registers definition

#----- imports
from packages.structs import CPU


#----- function
def define_registers(cpu: CPU) -> None:
    """Define the CPU registers and their aliases

    Args:
        cpu (CPU): the CPU dataclass
    """
    # define the standard x0 ... x31
    for i in range(32):
        cpu.registers[f'x{i}'] = i

    # aliases
    count = 0
    for r in ["zero", "ra", "sp", "gp", "tp", "t0", "t1", "t2","s0", "s1"]:
        cpu.registers[r] = f'x{count}'
        count += 1

    # a0 ... a7
    for r in range(8):
        cpu.registers[f'a{r}'] = f'x{10 + r}'

    # s2 ... s11
    for r in range(2, 12):
        cpu.registers[f's{r}'] = f'x{16 + r}'

    # t3 ... t6
    cpu.registers['t3'] = 'x28'
    cpu.registers['t4'] = 'x29'
    cpu.registers['t5'] = 'x30'
    cpu.registers['t6'] = 'x31'

    # fp
    cpu.registers['fp'] = 'x8'
