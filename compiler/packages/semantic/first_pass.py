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
# @brief	Semantic Analyzer - First pass

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional

from packages import ast, evaluator
from packages.symbols import SymbolType
from packages.architecture import Riscv, X68fp

from .struct import SAData


#----- class

class FirstPass:
    """Semantic Analyzer - First Pass"""

    def __init__(self, data: SAData) -> None:
        """Constructor

        Args:
            data (dc.SAData): the semantic analyzer structure
        """
        self.data = data
        self.current_section = ""

    def process(self) -> None:
        """Execute the first pass"""
        for stmt in self.data.program.statements:
            if isinstance(stmt, ast.Directive):
                self._directive(stmt)
            elif isinstance(stmt, ast.Label):
                self._label(stmt)
            elif isinstance(stmt, ast.Instruction):
                self._instruction(stmt)
            else:
                raise SyntaxError(f"Unknown statement ({stmt})")

    def _directive(self, directive: ast.Directive) -> None:
        """Process the Directive statement

        Args:
            directive (ast.Directive): a statement representing a directive
        """
        # select the proper section
        if directive.name in ['.text', '.data', '.bss']:
            self.current_section = directive.name
            return

        # handle each directives
        match directive.name:
            case '.cpu':
                self._d_cpu(directive.args[0])
            case '.skip':
                self._d_skip(directive.args[0])
            case '.align':
                self._d_align(directive.args[0])
            case '.org':
                self._d_org(directive.args[0])
            case '.byte' | '.half' | '.short' | '.word':
                self._d_data(directive.name, directive.args)
            case '.asciz':
                self._d_string(directive.args[0], True)
            case '.string':
                self._d_string(directive.args[0], False)
            case '.entrypoint':
                self._d_entry(directive.args[0])
            case _:
                pass

    def _label(self, label: ast.Label) -> None:
        """Process the Label statement

        Args:
            label (ast.Label): a statement representing a label
        """
        section = self.data.sections[self.current_section]
        self.data.symbols.define(label.name, section.offset, SymbolType.LABEL, section.name)

    def _instruction(self, instr: ast.Instruction) -> None:
        """Process the Instruction statement

        Args:
            instr (ast.Instruction): a statement representing an instruction
        """
        section = self.data.sections[self.current_section]
        instr.address = section.offset
        section.offset += self.data.instr_size

    def _d_cpu(self, arg: ast.Node) -> None:
        """Handle the .cpu directive

        Args:
            arg (ast.Node): the cpu type Identifier
        """
        assert isinstance(arg, ast.Identifier)
        match arg.name:
            case "risc-v":
                self.data.arch = Riscv()
            case "x68fp":
                self.data.arch = X68fp()
            case _:
                raise SyntaxError(f"Error: unknown architecture '{arg.name}'")

        self.data.instr_size = self.data.arch.config.instr_size

    def _d_skip(self, arg: ast.Node) -> None:
        """Handle the .skip directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.data.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.data.symbols, section.offset)

        if not result:
            raise SyntaxError(f"Error: .skip directive expects a full qualified expression")

        section.offset += value

    def _d_align(self, arg: ast.Node) -> None:
        """Handle the .align directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.data.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.data.symbols, section.offset)

        if not result:
            raise SyntaxError(f"Error: .align directive expects a full qualified expression")

        # compute the alignment
        delta = section.offset % value
        if delta != 0:
            padding = value - delta
            section.offset += padding

    def _d_org(self, arg: ast.Node) -> None:
        """Handle the .org directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.data.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.data.symbols, section.offset)

        if not result:
            raise SyntaxError(f"Error: .org directive expects a full qualified expression")

        # compute the alignment
        if value < section.offset:
            raise SyntaxError(f"Error: .org offset {value} is below current offset {section.offset}!")

        section.offset = value

    def _d_data(self, directive: str, args: List[ast.Node]) -> None:
        """Handle directives .byte, .word, .half

        Args:
            directive (str): the directive
            args (List[Node]): the arguments associated with the directive
        """
        section = self.data.sections[self.current_section]

        # compute the size associated with the directive
        assert self.data.arch is not None
        arch = self.data.arch
        match directive:
            case '.byte':
                size = arch.config.byte
            case '.half' | '.short':
                size = arch.config.half
            case '.word':
                size = arch.config.word
            case _:
                size = 1

        section.offset += size * len(args)

    def _d_string(self, arg: ast.Node, is_nul: bool) -> None:
        """Handle .asciz / .string directives

        Args:
            arg (ast.Node): a string literal
            is_nul (bool): True if the string should be terminated with '\0'
        """
        assert isinstance(arg, ast.StringLiteral)

        section = self.data.sections[self.current_section]
        section.offset += len(arg.text)
        if is_nul:
            section.offset += 1

    def _d_entry(self, arg: ast.Node) -> None:
        """Handle the .entrypoint directive

        Args:
            arg (ast.Node): an identifier representing the entrypoint label
        """
        assert isinstance(arg, ast.Identifier)
        self.data.entry_point = arg.name
