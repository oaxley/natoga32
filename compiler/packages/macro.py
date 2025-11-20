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
# @brief	Add support for macros in the compiler

#----- imports
from __future__ import annotations
from typing import Any, Dict, List, Optional

import os

from dataclasses import dataclass

from packages.token import Token, TokenStream, TokenType
from packages.lexer import Lexer


#----- globals
MAX_MACRO_EXPANSION_DEPTH = 100


#----- classes

@dataclass
class MacroDefinition:
    name: str
    params: List[str]
    body: List[Token]

class MacroProcessor:
    def __init__(self) -> None:
        self.macros = { }
        self.includes = set()

    def preprocess(self, tokens: List[Token], current_file: str) -> List[Token]:
        """Pre-process the source file to build macro and expand it"""
        ts = TokenStream(tokens)
        out: List[Token] = []

        while not ts.at_end():
            token = ts.peek()

            if not token:
                raise SyntaxError("EOF encounter during macro processing!")

            # --- macro definition ---
            if token.type == TokenType.DIRECTIVE and token.value == '.macro':
                # parse the macro definition
                macro = self._parse_macro_definition(ts)
                self.macros[macro.name] = macro
                continue

            # --- macro expansion ---
            if token.type == TokenType.IDENT and token.value in self.macros:
                expanded = self._expand_macro(self.macros[token.value], ts, current_file)
                expanded = self._expand_tokens_recursive(expanded, current_file)
                out.extend(expanded)
                continue

            # --- include ---
            if token.type == TokenType.DIRECTIVE and token.value == '.include':
                included = self._handle_include(ts, current_file)
                included = self.preprocess(included, current_file)
                out.extend(included)
                continue

            # --- for loops ---
            if token.type == TokenType.DIRECTIVE and token.value == '.for':
                expanded = self._handle_for_loop(ts, current_file)
                out.extend(expanded)
                continue

            # --- standard token ---
            out.append(token)
            ts.advance()

        return out

    def _parse_macro_definition(self, ts: TokenStream) -> MacroDefinition:
        """Parse a macro definition and add it to the Macro table"""
        ts.advance()    # consume .macro

        # retrieve the macro name
        token = ts.advance()
        if not token:
            raise SyntaxError("EOF encounter during macro processing!")

        if token.type != TokenType.IDENT:
            raise SyntaxError("Macro name must be an identifier!")

        name = token.value

        # parse the parameters
        params: List[str] = []
        while True:
            token = ts.peek()
            if not token:
                raise SyntaxError("Error while reading macro parameters!")

            if token.type == TokenType.EOL:
                break

            if token.type == TokenType.IDENT:
                params.append(token.value)

            # move forward
            ts.advance()

        # consume EOL
        ts.advance()

        # parse the body until we reach .endm
        body: List[Token] = []
        while True:
            token = ts.peek()
            if token is None:
                raise SyntaxError("Missing .endm for macro definition")

            if token.type == TokenType.DIRECTIVE and token.value == '.endm':
                break

            # add everything to the body, and move forward
            body.append(token)
            ts.advance()

        # consume .endm
        ts.advance()

        # consume EOL if present
        token = ts.peek()
        if token and token.type == TokenType.EOL:
            ts.advance()

        return MacroDefinition(name, params, body)

    def _expand_macro(self, macro: MacroDefinition, ts: TokenStream, current_file: str, depth: int = 0) -> List[Token]:

        if depth > MAX_MACRO_EXPANSION_DEPTH:
            raise SyntaxError("Max macro recursion depth exceeded!")

        ts.advance()        # consume macro name

        # --- parse arguments as a List of Tokens
        args: List[List[Token]] = []
        for i, param in enumerate(macro.params):
            # collect tokens until COMMA or EOL
            arg_tokens = []
            while not ts.at_end():
                token = ts.peek()
                if token.type in [TokenType.COMMA, TokenType.EOL]: # type: ignore
                    break
                arg_tokens.append(token)
                ts.advance()

            args.append(arg_tokens)

            # consume the comma if present
            if ts.peek() and ts.peek().type == TokenType.COMMA: # type: ignore
                ts.advance()
                continue

            # break if no comma
            break

        # consumer final EOL
        if ts.peek() and ts.peek().type == TokenType.EOL:   # type: ignore
            ts.advance()

        # --- create substitution map
        mapping = dict(zip(macro.params, args))

        # --- expand the macro body with substitution
        expanded = []
        for token in macro.body:
            # token in a parameter
            if token.type == TokenType.IDENT and token.value in mapping:
                for arg_tok in mapping[token.value]:
                    # clone token to avoid mutating the original
                    expanded.append(Token(arg_tok.type, arg_tok.value, arg_tok.row, arg_tok.col))
                continue

            # normal token
            expanded.append(Token(token.type, token.value, token.row, token.col))

        # --- allow for nested macro expansion
        expanded = self.preprocess(expanded, current_file)
        return expanded

    def _expand_tokens_recursive(self, tokens: List[Token], current_file: str, depth: int = 0) -> List[Token]:
        """Expand tokens recursively inside the macro"""
        ts = TokenStream(tokens)

        output: List[Token] = []
        while not ts.at_end():
            token = ts.peek()

            if not token:
                raise SyntaxError("Error during recursive macro expansion!")

            if token.type == TokenType.IDENT and token.value in self.macros:
                expanded = self._expand_macro(self.macros[token.value], ts, current_file, depth + 1)
                output.extend(expanded)
                continue

            output.append(token)
            ts.advance()

        return output

    def _handle_include(self, ts: TokenStream, current_file: str) -> List[Token]:
        """Handle include of other files"""
        ts.advance()    # consumer .include

        # next token must be a string
        token = ts.advance()
        if not token:
            raise SyntaxError("Error while processing .include directive!")

        if token.type != TokenType.STRING:
            raise SyntaxError(f".include directive expects a string literal, got {token.type.name}")

        # extract the filename
        filename = token.value.strip('"')

        # consume EOL if present
        if ts.peek() and ts.peek().type == TokenType.EOL:   # type: ignore
            ts.advance()

        # resolve relative path
        full_path = os.path.join(os.path.dirname(current_file), filename)

        # detect recursion
        if full_path in self.includes:
            raise Exception(f"Recursive include detected: {full_path}")

        # process the file - add it to the set to avoid infinite recursion
        self.includes.add(full_path)
        try:
            # --- open file
            fh = open(full_path, 'r', encoding='utf-8')

            # --- lexer processing
            lexer = Lexer()
            lexer.parse(fh)

            # --- macro pre-processing
            tokens = self.preprocess(lexer.tokens, full_path)

            return tokens

        finally:
            # remove the file from the set
            self.includes.remove(full_path)

    def _handle_for_loop(self, ts: TokenStream, current_file: str) -> List[Token]:
        """Handle for loops"""
        ts.advance()    # consume .for

        # 1. loop variable name
        token = ts.advance()
        if not token:
            raise SyntaxError("Error while processing .for directive!")

        if token.type == TokenType.IDENT:
            var_name = token.value
        else:
            raise SyntaxError(".for expects: .for <IDENT> = start, end, step")

        # 2. ensure we have '='
        token = ts.advance()
        if token and token.value != '=':
            raise SyntaxError(".for syntax error: expected '='")

        # 3. start value
        token = ts.advance()
        if token and token.type == TokenType.NUMBER:
            start_value = int(token.value, 0)
        else:
            raise SyntaxError(".for start value must be a number")

        # 4. ','
        ts.advance()

        # 5. end value
        token = ts.advance()
        if token and token.type == TokenType.NUMBER:
            end_value = int(token.value, 0)
        else:
            raise SyntaxError(".for end value must be a number")

        # 6. Optional step
        token = ts.peek()
        if token and token.type == TokenType.COMMA:
            # step is provided
            ts.advance()

            token = ts.advance()
            if token and token.type == TokenType.NUMBER:
                step_val = int(token.value, 0)
            else:
                raise SyntaxError(".for step must be a number")
        else:
            # step not provided -> default to +1
            step_val = 1

        # consume EOL
        if token and token.type == TokenType.EOL:
            ts.advance()

        # --- capture body
        body: List[Token] = []
        while True:
            token = ts.peek()
            if token is None:
                raise SyntaxError("Missing .endf for macro definition")

            if token.type == TokenType.DIRECTIVE and token.value == '.endf':
                break

            # add everything to the body, and move forward
            body.append(token)
            ts.advance()

        # consumer .endf and EOL
        ts.advance()
        if ts.peek() and ts.peek().type == TokenType.EOL:   # type: ignore
            ts.advance()

        # --- body expansion
        tokens: List[Token] = []
        current = start_value
        while (step_val > 0 and current < end_value) or (step_val < 0 and current > end_value):

            iteration_body: List[Token] = []
            for token in body:
                # substitute var_name with NUMBER token
                if token.type == TokenType.IDENT and token.value == var_name:
                    iteration_body.append(Token(TokenType.NUMBER, str(current), token.row, token.col))
                else:
                    # clone the token to avoid messing up with the loop body
                    iteration_body.append(Token(token.type, token.value, token.row, token.col))

            # pre-process to allow macros inside loops
            iteration_body = self.preprocess(iteration_body, current_file)

            # add those tokens to the global list
            tokens.extend(iteration_body)

            # current = current +/- step_val
            current += step_val

        return tokens
