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
from __future__ import annotations
from typing import Any, Dict, List, Optional

from packages import ast
from packages.architecture import Architecture, Riscv, X68fp


#----- class

class PseudoInstruction:

    def __init__(self, program: ast.Program) -> None:
        """Constructor"""
        self.program = program
        self.arch: Optional[Architecture] = None

    def process(self) -> ast.Program:
        """Process the AST Tree and expand the pseudo-instructions"""

        statements: List[ast.Statement] = []

        for stmt in self.program.statements:
            if isinstance(stmt, ast.Directive):
                self._handle_directive(stmt)
                statements.append(stmt)
            elif isinstance(stmt, ast.Instruction):
                if self.arch is not None:
                    stmts = self.arch.expand(stmt)
                    statements.extend(stmts)
                else:
                    statements.append(stmt)
            else:
                # add the statement to the program
                statements.append(stmt)

        return ast.Program(statements)

    def _handle_directive(self, directive: ast.Directive) -> None:
        """Process the directives"""
        if directive.name == '.cpu':
            if isinstance(directive.args[0], ast.Identifier):
                name = directive.args[0].name
                if name == 'risc-v':
                    self.arch = Riscv()
                    return
                elif name == 'x68fp':
                    self.arch = X68fp()
                    return

            raise SyntaxError(f"Error: unknown architecture '{directive.args[0]}'")
