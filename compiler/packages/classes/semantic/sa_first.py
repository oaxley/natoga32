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
# @brief	Semantic Analyzer - First Pass

#----- imports
from typing import List

from packages import ast
from packages.structs import Config, SymbolType
from packages.architecture import Riscv, X68fp
from packages.functions import evaluator

#----- class
class SAFirstPass:
    """Semantic Analyzer First Pass"""

    def __init__(self, config: Config) -> None:
        """Constructor"""
        self.config = config
        self.instr_size = 0

    def process(self) -> None:
        """Process the program and compute data offsets / size"""
        for stmt in self.config.program.statements:
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
        section = self.config.sections[self.current_section]
        self.config.symbols.define(label.name, section.offset, SymbolType.LABEL, section.name)

    def _instruction(self, instr: ast.Instruction) -> None:
        """Process the Instruction statement

        Args:
            instr (ast.Instruction): a statement representing an instruction
        """
        section = self.config.sections[self.current_section]
        instr.address = section.offset
        section.offset += self.instr_size

    def _d_cpu(self, arg: ast.Node) -> None:
        """Handle the .cpu directive

        Args:
            arg (ast.Node): the cpu type Identifier
        """
        assert isinstance(arg, ast.Identifier)
        match arg.name:
            case "risc-v":
                self.config.arch = Riscv()
            case "x68fp":
                self.config.arch = X68fp()
            case _:
                raise SyntaxError(f"Error: unknown architecture '{arg.name}'")

        self.instr_size = self.config.arch.config.instr_size

    def _d_skip(self, arg: ast.Node) -> None:
        """Handle the .skip directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.config.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.config.symbols, section.offset)

        if not result:
            raise SyntaxError(f"Error: .skip directive expects a full qualified expression")

        section.offset += value

    def _d_align(self, arg: ast.Node) -> None:
        """Handle the .align directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.config.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.config.symbols, section.offset)

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

        section = self.config.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.config.symbols, section.offset)

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
        section = self.config.sections[self.current_section]

        # compute the size associated with the directive
        match directive:
            case '.byte':
                size = self.config.arch.config.byte
            case '.half' | '.short':
                size = self.config.arch.config.half
            case '.word':
                size = self.config.arch.config.word
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

        section = self.config.sections[self.current_section]
        section.offset += len(arg.text)
        if is_nul:
            section.offset += 1

    def _d_entry(self, arg: ast.Node) -> None:
        """Handle the .entrypoint directive

        Args:
            arg (ast.Node): an identifier representing the entrypoint label
        """
        assert isinstance(arg, ast.Identifier)
        self.config.entrypoint = arg.name
