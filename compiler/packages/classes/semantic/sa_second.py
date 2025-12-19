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
# @brief	Semantic Analyzer - Second Pass

#----- imports
import os

from typing import List, cast

from packages import ast
from packages.structs import Config, EvalResult, Relocation, RelocationType
from packages.functions import evaluator


#----- class
class SASecondPass:
    """Semantic Analyzer First Pass"""

    def __init__(self, config: Config) -> None:
        """Constructor"""
        self.config = config

    def process(self) -> None:
        """Execute the second pass"""
        for stmt in self.config.program.statements:
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
            self.config.sections[self.current_section].offset = 0
            return

        match directive.name:
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
            case _:
                pass

    def _d_skip(self, arg: ast.Node) -> None:
        """Handle the .skpi directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.config.sections[self.current_section]

        _, value = evaluator.const_eval(arg, self.config.symbols, section.offset)
        section.offset += value

        if section.name in ['.text', '.data']:
            section.data.extend(b"\x00" * value)

    def _d_align(self, arg: ast.Node) -> None:
        """Handle the .align directive

        Args:
            arg (ast.Node): a const expression
        """
        assert isinstance(arg, ast.Expression)

        section = self.config.sections[self.current_section]

        _, value = evaluator.const_eval(arg, self.config.symbols, section.offset)
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

        section = self.config.sections[self.current_section]
        _, value = evaluator.const_eval(arg, self.config.symbols, section.offset)

        if value < section.offset:
            raise SyntaxError(f"Error: .org offset {value} is below current offset {section.offset}!")

        padding = value - section.offset
        section.offset = value

        if section.name in ['.text', '.data']:
            section.data.extend(b"\x00" * padding)

    def _d_data(self, directive: str, args: List[ast.Node]) -> None:
        """Handle directives .byte, .word, .half

        Args:
            directive (str): the directive
            args (List[Node]): the arguments associated with the directive
        """
        section = self.config.sections[self.current_section]

        # not supported by .bss segment
        if section.name == '.bss':
            return

        # specific case for .incbin
        if directive == '.incbin':
            # retrieve the full path for the include
            filename = cast(ast.StringLiteral, args[0])
            fullpath = os.path.join(os.path.dirname(self.config.infile), filename.text)

            # add the data to the section
            with open(fullpath, "rb") as fh:
                content = fh.read()

            section.data.extend(content)

            # add the size to the offset
            section.offset += len(content)
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

        # add each values to the data section
        symbols = self.config.symbols
        for node in args:
            assert isinstance(node, ast.Expression)
            result, value = evaluator.const_eval(node, symbols, section.offset)

            if not result:
                raise SyntaxError(f"Error: unable to process '{node.debug()}' as data.")

            if value < 0:
                section.data.extend(value.to_bytes(size, 'big', signed=True))
            else:
                section.data.extend(value.to_bytes(size, 'big', signed=False))

            section.offset += size

    def _d_string(self, arg: ast.Node) -> None:
        """Handle .asciz / .string directives

        Args:
            arg (ast.Node): a string literal
        """
        assert isinstance(arg, ast.StringLiteral)

        section = self.config.sections[self.current_section]
        section.data.extend(arg.text.encode('utf-8'))
        section.data.extend(b"\x00")
        section.offset += len(arg.text) + 1

    def _instruction(self, instr: ast.Instruction) -> None:
        """Process the instruction statement

        Args:
            instr (ast.Instruction): a statement representing an instruction
        """
        section = self.config.sections[self.current_section]

        # process the operands
        pc = instr.address
        operands: List[EvalResult] = []
        for op in instr.operands:
            result = self._op_eval(cast(ast.Expression, op), pc)
            operands.append(result)

        bincode, reloc = self.config.arch.encode(instr.opcode, operands)
        if len(reloc) > 0:
            for i in reloc:
                section.relocations.append(i)

        section.data.extend(bincode)
        section.offset += len(bincode)

    def _op_eval(self, op: ast.Expression, pc: int) -> EvalResult:
        """Evaluate an operand from an opcode

        Args:
            op (ast.Expression): the operand
            pc (int): the current program counter

        Returns:
            EvalResult: the result of the evaluation, either a value or relocation
        """
        symbols = self.config.symbols

        # a simple number
        if isinstance(op, ast.Number):
            return EvalResult(value = int(op.value))

        # current program counter
        if isinstance(op, ast.CurrentPC):
            return EvalResult(value = pc)

        # an identifier
        if isinstance(op, ast.Identifier):

            # 1. check for a register
            is_register, value = self.config.arch.is_register(op.name)
            if is_register:
                return EvalResult(reg=value)

            # 2. check for a symbol in the table
            is_symbol = symbols.exists(op.name)
            if is_symbol:
                # retrieve the symbol definition
                symbol = symbols.get(op.name)
                if symbol.defined:
                    return EvalResult(reloc=Relocation(symbol.type.name, op, 0, pc)) # type: ignore

            # for now we do not support multi-file compilation :(
            raise SyntaxError(f"Error: could not find definition for '{op.name}'")

        # relocation directives
        if isinstance(op, ast.RelocExpr):
            # lookup for the symbol
            if isinstance(op.symbol, ast.Identifier):
                if symbols.exists(op.symbol.name):
                    # for HiRel and LoRel, symbol can be anywhere
                    if isinstance(op, ast.HiRel):
                        return EvalResult(reloc=Relocation(RelocationType.RISCV_HI20, op.symbol, 0, pc))

                    if isinstance(op, ast.LoRel):
                        return EvalResult(reloc=Relocation(RelocationType.RISCV_LO12, op.symbol, 0, pc))

                    if isinstance(op, ast.PCRelHi):
                        return EvalResult(reloc=Relocation(RelocationType.RISCV_PCREL_HI20, op.symbol, 0, pc))

                    if isinstance(op, ast.PCRelLo):
                        return EvalResult(reloc=Relocation(RelocationType.RISCV_PCREL_LO12, op.symbol, 0, pc))

                    raise SyntaxError(f"Error: unknown relocation directive '{op}'")

                else:
                    raise SyntaxError(f"Error: unsupported external symbol definition")

            else:
                raise SyntaxError(f"Error: unsupported type '{op}' in relocation expression")


        # unary operator
        if isinstance(op, ast.UnaryOp):
            result, value = evaluator.const_eval(op, symbols, pc)
            if result:
                return EvalResult(value=value)
            else:
                raise SyntaxError(f"Error: unsupported UnaryOp '{op.expr}'")

        # binary operator
        if isinstance(op, ast.BinaryOp):
            # retrieve left / right
            left = self._op_eval(op.left, pc)
            right = self._op_eval(op.right, pc)

            # case 1: left is reloc, right is value
            if left.reloc and right.value is not None:
                match op.op:
                    case '+':
                        return EvalResult(reloc=Relocation(
                            left.reloc.type,
                            left.reloc.symbol,
                            left.reloc.addend + right.value,
                            pc
                        ))
                    case '-':
                        return EvalResult(reloc=Relocation(
                            left.reloc.type,
                            left.reloc.symbol,
                            left.reloc.addend - right.value,
                            pc
                        ))
                    case _:
                        raise SyntaxError(f"Error: unsupported operation '{op.op}'")

            # case 2: left is value, right is reloc
            if left.value is not None and right.reloc:
                match op.op:
                    case '+':
                        return EvalResult(reloc=Relocation(
                            right.reloc.type,
                            right.reloc.symbol,
                            right.reloc.addend + left.value,
                            pc
                        ))
                    case '-':
                        return EvalResult(reloc=Relocation(
                            right.reloc.type,
                            right.reloc.symbol,
                            right.reloc.addend - left.value,
                            pc
                        ))
                    case _:
                        raise SyntaxError(f"Error: unsupported operation '{op.op}'")

            # case 3: left is value, right is value
            if left.value and right.value:
                value = evaluator.binary_op(op.op, left.value, right.value)
                return EvalResult(value=value)

            # case 4: both side are relocation -> impossible
            raise SyntaxError(f"Error: unable to compute relocation with '{left}' and '{right}' as expression.")

        raise SyntaxError(f"Error: unsupported expression '{op}")
