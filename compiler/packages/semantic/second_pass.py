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
# @brief	Semasntic Analyzer second pass

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional

from packages import data_classes as dc
from packages import ast
from packages import evaluator

from packages.architecture import Architecture, Riscv, X68fp


#----- class

class SecondPass:
    """Semantic Analyzer - Second Pass"""

    def __init__(self, data: dc.SAData) -> None:
        """Constructor

        Args:
            data (dc.SAData): the semantic analyzer structure
        """
        self.data = data
        self.current_section = ""
        self.arch: Optional[Architecture] = None

    def process(self) -> None:
        """Execute the second pass"""
        for stmt in self.data.program.statements:
            if isinstance(stmt, ast.Directive):
                self._directive(stmt)
            elif isinstance(stmt, ast.Instruction):
                self._instruction(stmt)

    def _directive(self, directive: ast.Directive) -> None:
        """Process the Directive statement

        Args:
            directive (ast.Directive): a statement representing a directive
        """
        if directive.name in ['.text', '.data', '.bss']:
            self.current_section = directive.name
            self.data.sections[self.current_section].offset = 0
            return

        match directive.name:
            case '.cpu':
                self._d_cpu(directive.args[0])
            case '.skip':
                self._d_skip(directive.args[0])
            case '.align':
                self._d_align(directive.args[0])
            case '.org':
                self._d_org(directive.args[0])
            case '.byte' | '.half' | '.word':
                self._d_data(directive.name, directive.args)
            case '.asciz':
                self._d_string(directive.args[0], True)
            case '.ascii':
                self._d_string(directive.args[0], False)
            case _:
                pass

    def _d_cpu(self, arg: ast.Node) -> None:
        """Handle the .cpu directive

        Args:
            arg (ast.Node): the cpu type Identifier
        """
        assert isinstance(arg, ast.Identifier)
        if arg.name == "risc-v":
            self.arch = Riscv()
        elif arg.name == "x68fp":
            self.arch = X68fp()
        else:
            raise SyntaxError(f"Error: unknown architecture '{arg.name}'")

    def _d_skip(self, arg: ast.Node) -> None:
        """Handle the .cpu directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.data.sections[self.current_section]

        # evaluate the skip value
        result = evaluator.const_eval(arg, self.data.symbols, section.offset)

        if result.reloc:
            raise SyntaxError(f"Error: .skip directive expects a full qualified expression")

        # move the pointer
        assert result.value is not None
        section.offset += result.value

        if section.name in ['.text', '.data']:
            section.data.extend(b"\x00" * result.value)

    def _d_align(self, arg: ast.Node) -> None:
        """Handle the .align directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.data.sections[self.current_section]

        # evaluate the skip value
        result = evaluator.const_eval(arg, self.data.symbols, section.offset)
        if result.reloc:
            raise SyntaxError(f"Error: .align directive expects a full qualified expression")

        # compute the alignment
        assert result.value is not None
        value = result.value

        delta = section.offset % value
        if delta != 0:
            padding = value - delta
            section.offset += padding
            if section.name in ['.text', '.data']:
                section.data.extend(b"\x00" * padding)

    def _d_org(self, arg: ast.Node) -> None:
        """Handle the .org directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.data.sections[self.current_section]

        # evaluate the skip value
        result = evaluator.const_eval(arg, self.data.symbols, section.offset)
        if result.reloc:
            raise SyntaxError(f"Error: .align directive expects a full qualified expression")

        # compute the alignment
        assert result.value is not None
        value = result.value

        if value < section.offset:
            raise SyntaxError(f"Error: .org offset {value} is below current offset {section.offset}!")

        padding = value = section.offset
        section.offset = value

        if section.name in ['.text', '.data']:
            section.data.extend(b"\x00" * padding)

    def _d_data(self, directive: str, args: List[ast.Node]) -> None:
        """Handle directives .byte, .word, .half

        Args:
            directive (str): the directive
            args (List[Node]): the arguments associated with the directive
        """
        section = self.data.sections[self.current_section]

        # not supported by .bss segment
        if section.name == '.bss':
            return

        # compute the size associated with the directive
        assert self.arch is not None
        match directive:
            case '.byte':
                size = self.arch.config.byte
            case '.half':
                size = self.arch.config.half
            case '.word':
                size = self.arch.config.word
            case _:
                size = 1

        # add each values to the data section
        symbols = self.data.symbols
        for node in args:
            assert isinstance(node, ast.Expression)
            result = evaluator.const_eval(node, symbols, section.offset)

            if result.reloc:
                section.relocations.append(result.reloc)
                section.data.extend(b"\x00" * size)
            else:
                assert result.value
                if result.value < 0:
                    section.data.extend(result.value.to_bytes(size, 'big', signed=True))
                else:
                    section.data.extend(result.value.to_bytes(size, 'big', signed=False))

            section.offset += size

    def _d_string(self, arg: ast.Node, is_nul: bool) -> None:
        """Handle .asciz / .ascii directives

        Args:
            arg (ast.Node): a string literal
            is_nul (bool): True if the string should be terminated with '\0'
        """
        assert isinstance(arg, ast.StringLiteral)

        section = self.data.sections[self.current_section]
        section.offset += len(arg.text)
        section.data.extend(arg.text.encode('utf-8'))
        if is_nul:
            section.offset += 1
            section.data.extend(b"\x00")

    def _instruction(self, instr: ast.Instruction) -> None:
        """Process the Instruction statement

        Args:
            instr (ast.Instruction): a statement representing an instruction
        """
        section = self.data.sections[self.current_section]

        pc = instr.address
        operands = []
        for op in instr.operands:
            print(op)
