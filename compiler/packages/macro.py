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

from packages import helper

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
        self.macros: Dict[str, MacroDefinition] = { }
        self.includes = set()
        self.expansion_depth = 0

    def preprocess(self, tokens: List[Token], current_file: str) -> List[Token]:
        """Perform pre-processing of the source file

        Args:
            tokens (List[Token]): the list of tokens from the Lexer
            current_file (str): the filename of the current file being processed

        Returns:
            List[Token]: the new list of tokens after pre-processing is done
        """
        ts = TokenStream(tokens)
        out: List[Token] = []

        while not ts.at_end():
            token = ts.peek()
            if not token:
                break

            # take care of directives
            if token.type == TokenType.DIRECTIVE:

                # --- .macro
                if token.value == '.macro':
                    macro = self._parse_macro_definition(ts)
                    self.macros[macro.name] = macro
                    continue

                # --- .include
                elif token.value == '.include':
                    included = self._handle_include(ts, current_file)
                    included = self.preprocess(included, current_file)
                    included = self._apply_token_pasting(included)
                    included = self._apply_dup(included)
                    out.extend(included)
                    continue

                # --- .for
                elif token.value == '.for':
                    expanded = self._handle_for_loop(ts, current_file)
                    expanded = self._apply_token_pasting(expanded)
                    expanded = self._apply_dup(expanded)
                    expanded = self.preprocess(expanded, current_file)
                    out.extend(expanded)
                    continue

            # identities token
            elif token.type == TokenType.IDENT:

                # --- macro expansion
                if token.value in self.macros:
                    expanded = self._expand_macro(self.macros[token.value], ts, current_file)
                    expanded = self._apply_token_pasting(expanded)
                    expanded = self._apply_dup(expanded)
                    expanded = self.preprocess(expanded, current_file)
                    out.extend(expanded)
                    continue

            # environment variables replacement
            elif token.type == TokenType.ENVVAR:
                name = token.value[2:-1]
                envvar = os.getenv(name)
                if envvar != None:
                    try:
                        value = int(envvar, 0)
                        out.append(Token(TokenType.NUMBER,hex(value),token.row,token.col))
                    except ValueError:
                        out.append(Token(TokenType.STRING,f'"{envvar}"',token.row,token.col))
                else:
                    raise SyntaxError(f"Could not find environment variable '{name}'")

                ts.advance()
                continue

            # regular token
            out.append(token)
            ts.advance()

        # apply pasting and 'dup' one last time
        out = self._apply_token_pasting(out)
        out = self._apply_dup(out)

        return out

    def _parse_macro_definition(self, ts: TokenStream) -> MacroDefinition:
        """Parse a macro definition and add it to the Macro table

        Args:
            ts (TokenStream): the token stream instance

        Returns
            MacroDefinition: an instance of MacroDefinition representing the macro
        """
        ts.advance()    # consume .macro

        # retrieve the macro name
        name = helper.get_value(ts, TokenType.IDENT)

        # parse the parameters
        params: List[str] = []
        while True:
            token = ts.peek()
            if not token or token.type == TokenType.EOL:
                ts.advance()        # consumer EOL
                break

            if token.type == TokenType.IDENT:
                params.append(token.value)

            # move forward
            ts.advance()

        # parse the body until we reach .endm
        body = helper.capture_body(ts, '.macro', '.endm')

        return MacroDefinition(name, params, body)

    def _expand_macro(self, macro: MacroDefinition, ts: TokenStream, current_file: str) -> List[Token]:
        """Expand macro invocation at current stream position

        Args:
            macro (MacroDefinition): the macro detected during processing
            ts (TokenStream): the current token stream
            current_file (str): the current file being processed

        Returns:
            List[Token]: A list of tokens after the macro has been expanded
        """
        # recursion depth check
        self.expansion_depth += 1
        try:
            if self.expansion_depth >= MAX_MACRO_EXPANSION_DEPTH:
                raise SyntaxError("Macro expansion recursion limit exceeded")

            # consume the macro name (MacroDefinition is already properly populated)
            ts.advance()

            # parse the parameters. if EOL, means no parameters.
            params: List[List[Token]] = []
            if ts.peek() and ts.peek().type == TokenType.EOL:   # type: ignore
                ts.advance()    # consume EOL

            else:
                # we are trying to see if the macro parameters are simple tokens
                # or a complete expression that will have to be expanded later
                while not ts.at_end():
                    # collect tokens until COMMA or EOL
                    arg_tokens: List[Token] = []
                    while not ts.at_end():
                        token = ts.peek()
                        if token and token.type in [TokenType.COMMA, TokenType.EOL]:
                            break

                        arg_tokens.append(ts.advance())     # type: ignore

                    params.append(arg_tokens)

                    # consume COMMA if present
                    if ts.peek() and ts.peek().type == TokenType.COMMA:     # type: ignore
                        ts.advance()
                        continue

                    # we have reach EOL
                    if ts.peek() and ts.peek().type == TokenType.EOL:       # type: ignore
                        ts.advance()

                    break

            # Map parameters -> token lists (fill with empty lists if fewer args)
            mapping: Dict[str, List[Token]] = {}
            for i, pname in enumerate(macro.params):
                mapping[pname] = params[i] if i < len(params) else []

            # expand the macro body with substitution
            expanded: List[Token] = []
            for bt in macro.body:
                # token is a parameter
                if bt.type == TokenType.IDENT and bt.value in mapping:
                    for t in mapping[bt.value]:
                        expanded.append(helper.clone_token(t))
                    continue

                # normal token
                expanded.append(helper.clone_token(bt))

            return expanded

        finally:
            # reduce the recursion depth
            self.expansion_depth -= 1

    def _handle_include(self, ts: TokenStream, current_file: str) -> List[Token]:
        """Handle '.include' directive

        Args:
            ts (TokenStream): the token stream instance
            current_file (str): the current file being processed

        Returns:
            List[Token]: a list of tokens to insert at the ".include" position
        """
        ts.advance()    # consumer .include

        # retrieve the filename (STRING)
        filename = helper.get_value(ts, TokenType.STRING).strip('"')

        # consume EOL if present
        if ts.expect(TokenType.EOL): ts.advance()

        # resolve relative path
        full_path = os.path.join(os.path.dirname(current_file), filename)
        if full_path in self.includes:
            raise Exception(f"Recursive include detected: {full_path}")

        # process the file - add it to the set to avoid infinite recursion
        self.includes.add(full_path)
        try:
            # --- open file
            try:
                fh = open(full_path, 'r', encoding='utf-8')
            except FileNotFoundError:
                raise SyntaxError(f"File not found during include [{full_path}]")

            # --- lexer processing
            lexer = Lexer()
            lexer.parse(fh)

            return lexer.tokens

        finally:
            # remove the file from the set
            self.includes.remove(full_path)

    def _handle_for_loop(self, ts: TokenStream, current_file: str) -> List[Token]:
        """Handle '.for/.endf' directives

        Args:
            ts (TokenStream): the current token stream
            current_file (str): the filename of the file being processed

        Returns:
            List[Token]: a list of tokens to insert at the ".for" loop position
        """
        ts.advance()    # consume .for

        # parse <IDENT> = <NUMBER>, <NUMBER>
        var_name = helper.get_value(ts, TokenType.IDENT)
        helper.get_value(ts, TokenType.EQUAL)
        start_value = int(helper.get_value(ts, TokenType.NUMBER), 0)
        ts.advance()
        end_value = int(helper.get_value(ts, TokenType.NUMBER), 0)

        # 6. Optional step
        step_value = 1
        if ts.expect(TokenType.COMMA):
            ts.advance()    # remove comma
            step_value = int(helper.get_value(ts, TokenType.NUMBER), 0)
            if step_value == 0:
                raise SyntaxError("Step value cannot be 0")

        # consume EOL if present
        if ts.expect(TokenType.EOL): ts.advance()

        # --- capture body
        body = helper.capture_body(ts, '.for', '.endf')

        # --- body expansion
        tokens: List[Token] = []
        current = start_value

        # loop condition
        if step_value > 0:
            cond = lambda c: c < end_value
        else:
            cond = lambda c: c > end_value

        while cond(current):

            iteration_body: List[Token] = []
            for token in body:
                # substitute var_name with NUMBER token
                if token.type == TokenType.IDENT and token.value == var_name:
                    iteration_body.append(Token(TokenType.NUMBER, str(current), token.row, token.col))
                else:
                    # clone the token to avoid messing up with the loop body
                    iteration_body.append(helper.clone_token(token))

            # pre-process to allow macros inside loops
            iteration_body = self.preprocess(iteration_body, current_file)

            # add those tokens to the global list
            tokens.extend(iteration_body)

            # current = current +/- step_val
            current += step_value

        return tokens

    def _apply_token_pasting(self, tokens: List[Token]) -> List[Token]:
        """Process token pasting operator '##' to merge two tokens together

        Args:
            tokens (List[Token]): the tokens representing the pasting operation (i##k)

        Returns:
            List[Token]: a new list of tokens to replace the previous pasting operator
        """
        # --- process pasting
        pasted: List[Token] = []
        i = 0
        while i < len(tokens):
            token = tokens[i]

            # look for IDENT, PASTE, IDENT (x##i)
            if (
                i + 2 < len(tokens)
                and token.type == TokenType.IDENT
                and tokens[i+1].type == TokenType.PASTE
                and tokens[i+2].type == TokenType.NUMBER
            ):
                # concatenate the left and right side
                new_value = token.value + tokens[i+2].value

                # build a new IDENT token with the same position / info that the one on the left
                pasted.append(Token(TokenType.IDENT, new_value, token.row, token.col))

                # move forward
                i += 3
                continue

            # otherwise keep adding the tokens
            pasted.append(token)
            i += 1

        return pasted

    def _apply_dup(self, tokens: List[Token]) -> List[Token]:
        """Expands 'VALUE DUP(x)' pattern, which repeats x times the VALUE

        Args:
            tokens (List[Token]): the list of tokens from the Lexer

        Returns:
            List[Token]: a new list of tokens to replace the DUP sequence
        """
        result: List[Token] = []
        i = 0

        while i < len(tokens):
            token = tokens[i]

            # look for pattern VALUE DUP( expr )
            if (
                i + 3 < len(tokens)
                and tokens[i+1].type == TokenType.IDENT
                and tokens[i+1].value.upper() == 'DUP'
                and tokens[i+2].type == TokenType.LPARENT
            ):
                # find the matching RPARENT
                j = i + 3
                depth = 1
                while (j < len(tokens)) and (depth > 0):
                    if tokens[j].type == TokenType.LPARENT:
                        depth += 1
                    if tokens[j].type == TokenType.RPARENT:
                        depth -= 1
                    j += 1

                if depth != 0:
                    raise SyntaxError("Unmatched parenthesis in DUP expression")

                # retrieve all the tokens from the expression and evaluate it
                expr = tokens[i+3:j-1]
                value = helper.evaluate_expr(expr)

                if value < 0:
                    raise SyntaxError("DUP count must be non-negative")

                # replace the whole expression
                for _ in range(value):
                    result.append(helper.clone_token(token))
                    result.append(Token(TokenType.COMMA, ",", token.row, token.col))

                # move forward
                i = j
                continue

            # keep adding normal tokens
            result.append(token)
            i += 1

        return result
