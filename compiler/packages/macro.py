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


#----- functions
def clone_token(token: Token) -> Token:
    """Clone a Token with Shallow Copy"""
    return Token(token.type, token.value, token.row, token.col)

def capture_body(ts: TokenStream, begin: str, end: str) -> List[Token]:
    """Parse the token stream and extract the body between 'begin' and 'end' tokens

    Args:
        ts    : the stream of tokens
        begin : token for the beginning marker
        end   : token for the end marker

    Returns:
        The token list corresponding to the body
    """
    body: List[Token] = []
    depth = 1       # depth tracker

    while True:
        token = ts.peek()

        if token is None:
            break

        # depth tracker (for nested bodies)
        if token.type == TokenType.DIRECTIVE and token.value == begin:
            depth += 1
            if depth > MAX_BODY_DEPTH:
                raise SyntaxError("Nested body max limit exceeded!")

        if token.type == TokenType.DIRECTIVE and token.value == end:
            depth -= 1
            if depth == 0:
                break

        # add everything to the body, and move forward
        body.append(token)
        ts.advance()

    # consume "end" marker & EOL
    ts.advance()
    if ts.peek() and ts.peek().type == TokenType.EOL:   # type: ignore
        ts.advance()

    # return the body
    return body


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
        """Perform macro pre-processing

        Args:
            tokens      : the list of tokens from the Lexer
            current_file: the filename of the current file processed

        Returns:
            A new list of tokens, with the macro expanded.
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
            ts : the token stream instance
        Returns
            A MacroDefintion with the name, params and body
        """
        ts.advance()    # consume .macro

        # retrieve the macro name
        token = ts.advance()
        if not token or token.type != TokenType.IDENT:
            raise SyntaxError("Macro definition requires a name")

        name = token.value

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
        body = capture_body(ts, '.macro', '.endm')

        return MacroDefinition(name, params, body)

    def _expand_macro(self, macro: MacroDefinition, ts: TokenStream, current_file: str) -> List[Token]:
        """Expand macro invocation at current stream position

        Args:
            macro           : the macro detected during processin
            ts              : the current token stream
            current_file    : the current file being processed

        Returns:
            A new token list, with the macro expanded
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
                        expanded.append(clone_token(t))
                    continue

                # normal token
                expanded.append(clone_token(bt))

            return expanded

        finally:
            # reduce the recursion depth
            self.expansion_depth -= 1

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
            # consume the comma and retrieve the next token
            ts.advance()
            token = ts.peek()

            if token and token.type == TokenType.NUMBER:
                step_val = int(token.value, 0)

                # consume the number and retrieve the next token
                ts.advance()
                token = ts.peek()
            else:
                raise SyntaxError(".for step must be a number")
        else:
            # step not provided -> default to +1
            step_val = 1

        # consume EOL if present
        if token and token.type == TokenType.EOL:
            ts.advance()

        # --- capture body
        body = capture_body(ts, '.for', '.endf')

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

        tokens = self._apply_token_pasting(tokens)
        return tokens

    def _apply_token_pasting(self, tokens: List[Token]) -> List[Token]:
        """Process pasting to allow form 'x##i'"""
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
            tokens: the list of tokens from the Lexer
        Returns:
            a new list of tokens with the DUP expanded
        """
        result: List[Token] = []
        i = 0

        while i < len(tokens):
            token = tokens[i]

            # look for pattern VALUE DUP( NUMBER )
            if (
                i + 4 < len(tokens)
                and tokens[i+1].type == TokenType.IDENT
                and tokens[i+1].value.upper() == 'DUP'
                and tokens[i+2].type == TokenType.LPARENT
            ):
                # ensure we have a right parenthesis
                if tokens[i+4].type != TokenType.RPARENT:
                    raise SyntaxError("Unmatched parenthesis in DUP expression")

                # ensure we have a number
                if tokens[i+3].type not in [TokenType.NUMBER, TokenType.CHAR, TokenType.IDENT]:
                    raise SyntaxError(f"DUP argument must be NUMBER, got {tokens[i+3].type.name}")

                count = int(tokens[i+3].value, 0)

                # replace the whole expression
                for _ in range(count):
                    result.append(Token(token.type, token.value, token.row, token.col))
                    result.append(Token(TokenType.COMMA, ",", token.row, token.col))

                # move forward
                i += 5
                continue

            # keep adding normal tokens
            result.append(token)
            i += 1

        return result
