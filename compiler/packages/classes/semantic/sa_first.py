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
import os

from typing import List, cast

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
        self.current_section = ""

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
            case '.byte' | '.half' | '.short' | '.word' | '.incbin':
                self._d_data(directive.name, directive.args)
            case '.asciz' | '.string':
                self._d_string(directive.args[0])
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
        self.config.symbols.define(label.name, section.size, SymbolType.LABEL, section.name)

    def _instruction(self, instr: ast.Instruction) -> None:
        """Process the Instruction statement

        Args:
            instr (ast.Instruction): a statement representing an instruction
        """
        section = self.config.sections[self.current_section]
        instr.address = section.size
        section.size += self.config.arch.cpu.instr_size

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

        # set the starting values in the respective section
        for name in self.config.arch.locations:
            self.config.sections[name].address = self.config.arch.locations[name]

    def _d_skip(self, arg: ast.Node) -> None:
        """Handle the .skip directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.config.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.config.symbols, section.size)

        if not result:
            raise SyntaxError(f"Error: .skip directive expects a full qualified expression")

        section.size += value

    def _d_align(self, arg: ast.Node) -> None:
        """Handle the .align directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.config.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.config.symbols, section.size)

        if not result:
            raise SyntaxError(f"Error: .align directive expects a full qualified expression")

        # compute the alignment
        delta = section.size % value
        if delta != 0:
            padding = value - delta
            section.size += padding

    def _d_org(self, arg: ast.Node) -> None:
        """Handle the .org directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.config.sections[self.current_section]
        result, value = evaluator.const_eval(arg, self.config.symbols, section.size)

        if not result:
            raise SyntaxError(f"Error: .org directive expects a full qualified expression")

        # compute the alignment
        if value < section.size:
            raise SyntaxError(f"Error: .org offset {value} is below current offset {section.size}!")

        section.size = value

    def _d_data(self, directive: str, args: List[ast.Node]) -> None:
        """Handle directives .byte, .word, .half

        Args:
            directive (str): the directive
            args (List[Node]): the arguments associated with the directive
        """
        section = self.config.sections[self.current_section]

        # special case for incbin
        if directive == '.incbin':
            # retrieve the full path for the include
            filename = cast(ast.StringLiteral, args[0])
            fullpath = os.path.join(os.path.dirname(self.config.infile), filename.text)

            # compute the size of the file
            with open(fullpath, "rb") as fh:
                size = fh.seek(0, os.SEEK_END)

            # add the size to the offset
            section.size += size
            return

        # compute the size associated with the directive
        match directive:
            case '.byte':
                size = self.config.arch.cpu.byte
            case '.half' | '.short':
                size = self.config.arch.cpu.half
            case '.word':
                size = self.config.arch.cpu.word
            case _:
                size = 1

        section.size += size * len(args)

    def _d_string(self, arg: ast.Node) -> None:
        """Handle .asciz / .string directives

        Args:
            arg (ast.Node): a string literal
        """
        assert isinstance(arg, ast.StringLiteral)

        section = self.config.sections[self.current_section]
        section.size += len(arg.text) + 1

    def _d_entry(self, arg: ast.Node) -> None:
        """Handle the .entrypoint directive

        Args:
            arg (ast.Node): an identifier representing the entrypoint label
        """
        assert isinstance(arg, ast.Identifier)
        self.config.entrypoint = arg.name
