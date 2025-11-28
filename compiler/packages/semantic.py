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
# @brief	Semantic Analyzer

#----- imports
from __future__ import annotations
from typing import Any, Dict, List

from packages import data_classes
from packages import ast
from packages.symbols import Symbol, SymbolTable, SymbolType
from packages import evaluator


#----- globals


#----- class
class SemanticAnalyzer:
    """Perform the 2-pass semantic analysis"""
    def __init__(self, program: ast.Program, symbols: SymbolTable, sections: Dict[str, data_classes.Section]) -> None:
        """Constructor

        Args:
            program (ast.Program): the program as an AST tree after parsing
            symbols (SymbolTable): the symbol table
            sections (Section): the Sections definition
        """
        self.program = program
        self.symbols = symbols
        self.sections = sections
        self.current_section = ""

    def first_pass(self):
        """Execute the 1st pass of the semantic analyzer"""

        for stmt in self.program.statements:
            if isinstance(stmt, ast.Directive):
                self.handle_directive(stmt)
            if isinstance(stmt, ast.Label):
                self.handle_label(stmt)

    def handle_directive(self, directive: ast.Directive):
        """Handle the directives during the first pass

        Args:
            directive (ast.Directive): the Directive statement
        """
        # section definition
        if directive.name in ['.data', '.text', '.bss']:
            self.current_section = directive.name
            return

        # handle each directives
        match directive.name:
            case '.skip':
                self._handle_skip(directive.args[0])
            case '.align':
                self._handle_align(directive.args[0])
            case '.org':
                self._handle_org(directive.args[0])
            case '.byte' | '.half' | '.word':
                self._handle_data(directive.name, directive.args)
            case '.asciz':
                self._handle_string(directive.args[0])
            case _:
                print(f"[{directive.name}]")

    def handle_label(self, label: ast.Label):
        """Handle Labels"""
        section = self.sections[self.current_section]
        self.symbols.define(label.name, section.offset, SymbolType.LABEL, self.current_section)


    def _handle_skip(self, arg: ast.Node) -> None:
        """Handle the .skip directive"""
        # select the current section
        section = self.sections[self.current_section]

        assert isinstance(arg, ast.Expression)
        result = evaluator.const_eval(arg, self.symbols, section.offset)

        if result.value:
            size = result.value
        else:
            raise SyntaxError(".skip directive expect a number or a fully qualified expression!")

        # just move the .bss pointer
        if self.current_section == ".bss":
            section.offset += size
            return

        if self.current_section in ['.text', '.data']:
            section.offset += size
            section.data.extend(b"\x00" * size)
            return

    def _handle_align(self, arg: ast.Node) -> None:
        """Handle .align directive"""
        section = self.sections[self.current_section]

        assert isinstance(arg, ast.Expression)
        result = evaluator.const_eval(arg, self.symbols, section.offset)

        if result.value:
            value = result.value
        else:
            raise SyntaxError(".align directive expect a number or a fully qualified expression!")

        delta = section.offset % value
        if delta != 0:
            padding = value - delta
            section.offset += padding
            if section.name in ['.data', '.text']:
                section.data.extend(b"\x00" * padding)

    def _handle_org(self, arg: ast.Node) -> None:
        """Handle .org directive"""
        section = self.sections[self.current_section]

        assert isinstance(arg, ast.Expression)
        result = evaluator.const_eval(arg, self.symbols, section.offset)

        if result.value:
            value = result.value
        else:
            raise SyntaxError(".org directive expect a number or a fully qualified expression!")

        if value < section.offset:
            raise SyntaxError(f"New .org offset {value} is below current offset {section.offset}!")

        padding = value - section.offset
        section.offset = value

        if section.name in ['.data', '.text']:
            section.data.extend(b"\x00" * padding)

    def _handle_data(self, directive: str, args: List[ast.Node]) -> None:
        """Handle directives like .byte, .word, .half"""
        section = self.sections[self.current_section]

        # compute the new offset
        match directive:
            case '.byte':
                size = 1
            case '.half':
                size = 2
            case '.word':
                size = 4
            case _:
                size = 1

        # add the values to the data section
        for node in args:
            assert isinstance(node, ast.Expression)
            result = evaluator.const_eval(node, self.symbols, section.offset)

            if result.reloc:
                section.relocations.append(result.reloc)
                section.data.extend(b"\x00" * size)
            else:
                assert result.value is not None
                if result.value < 0:
                    section.data.extend(result.value.to_bytes(size, 'big', signed=True))
                else:
                    section.data.extend(result.value.to_bytes(size, 'big', signed=False))

            section.offset += size

    def _handle_string(self, arg: ast.Node) -> None:
        """Handle .asciz directive"""
        section = self.sections[self.current_section]

        assert isinstance(arg, ast.StringLiteral)
        section.offset += len(arg.text) + 1
        section.data.extend(arg.text.encode('utf-8'))
        section.data.extend(b"\x00")
