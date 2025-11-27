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
        self.pc = 0

    def first_pass(self):
        """Execute the 1st pass of the semantic analyzer"""

        for stmt in self.program.statements:
            if isinstance(stmt, ast.Directive):
                self.handle_directive(stmt)


    def handle_directive(self, directive: ast.Directive):
        """Handle the directives during the first pass

        Args:
            directive (ast.Directive): the Directive statement
        """
        # section definition
        if directive.name in ['.data', '.text', '.bss']:
            self.current_section = directive.name
            self.pc = 0
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
            case _:
                print(f"[{directive.name}]")


    def _handle_skip(self, arg: ast.Node) -> None:
        """Handle the .skip directive"""
        if isinstance(arg, ast.Number):
            size = arg.value
        else:
            raise SyntaxError("Directive '.skip' needs a number!")

        # select the current section
        section = self.sections[self.current_section]

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
        if isinstance(arg, ast.Number):
            value = arg.value
        else:
            raise SyntaxError("Directive '.align' needs a number!")

        section = self.sections[self.current_section]
        delta = section.offset % value
        if delta != 0:
            padding = value - delta
            section.offset += padding
            if section.name in ['.data', '.text']:
                section.data.extend(b"\x00" * padding)

    def _handle_org(self, arg: ast.Node) -> None:
        """Handle .org directive"""
        if isinstance(arg, ast.Number):
            value = arg.value
        else:
            raise SyntaxError("Directive '.align' needs a number!")

        section = self.sections[self.current_section]

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

        # increase offset
        section.offset += size * len(args)

        # add the values to the data section
        for node in args:
            if isinstance(node, ast.Number):
                if node.value < 0:
                    section.data.extend(node.value.to_bytes(size, 'big', signed=True))
                else:
                    section.data.extend(node.value.to_bytes(size, 'big', signed=False))

            elif isinstance(node, ast.CharLiteral):
                if len(node.ch) > 1:
                    value = node.ch[0]
                else:
                    value = node.ch

                section.data.append(ord(value))

            else:
                print(f"? {type(node)}")
