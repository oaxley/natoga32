# goasm preprocessor

This page describes the preprocessor of goasm: includes, constants, conditions, loops, duplication, macros, token pasting and environment variables.

The preprocessor runs on tokens, before the parser (phase 2).
It replaces text; it does not know labels or sections.

A complete example:

```asm
.define COUNT 4
.define DEBUG
.if COUNT > 3
    .define MODE 1
.else
    .define MODE 2
.endif
.ifdef DEBUG
    nop
.endif
.ifndef RELEASE
    nop
.endif

.macro push reg
    addi sp, sp, -4
    sw   \reg, 0(sp)
.endmacro
.macro pop reg
    lw   \reg, 0(sp)
    addi sp, sp, 4
.endmacro

.text
    push ra
    addi a0, zero, MODE
    pop  ra
    ret

.for i = 0, COUNT
    addi x##i, zero, i
.endf

.for n = 0, 3
handler##n##_entry:
    nop
.endf

.data
zeros:
    .byte .dup 8 0
words:
    .word .dup (COUNT * 2) 0xDEADBEEF
```

## Include

```asm
.include "defs.inc"
```

The path is relative to the directory of the **input file**, also inside a nested include.
The included text replaces the line.
Its `.define` names and macros stay visible after it.

| Error | Cause |
|-------|-------|
| `cannot open include file 'X': ...` | The file does not exist. |
| `recursive include detected: X` | A file includes itself, directly or not. |

To insert raw bytes, use [`.incbin`](syntax.md#data-directives).

## Constants

`.define NAME VALUE` replaces every later use of `NAME` by the number `VALUE`.

```asm
.define STACK_SIZE 256
    li a0, STACK_SIZE
```

- `.define NAME` with no value defines `NAME` as `0`. Use it as a flag with `.ifdef`.
- Write `.define NAME VALUE` without a comma. `.define A, 5` fails with `cannot evaluate token: ,`.
- The value is evaluated **from left to right**, with no precedence and no parentheses. `.define D 2+3*4` gives 20.
- It accepts numbers, characters, other `.define` names, environment variables, a leading `-`, and these operators: `+ - * / & | ^ << >> == != < > <= >=`.
- `.define` names are not the same as `.equ` names. An `.equ` constant is invisible to `.ifdef` and `.if`. See [Syntax](syntax.md#data-directives).

## Conditions

| Directive | Meaning |
|-----------|---------|
| `.ifdef NAME` | Keep the block if `NAME` is a `.define` name. |
| `.ifndef NAME` | Keep the block if it is not. |
| `.if EXPR` | Keep the block if `EXPR` is not 0. |
| `.else` | Other branch. |
| `.endif` | End of the block. |

`.if` uses the same left-to-right evaluator as `.define`.
`&&`, `||` and parentheses are **not** supported, and give `cannot evaluate token: ...`.
`!` is a trap: the lexer drops a `!` that is not followed by `=`, so `.if !X` is read as `.if X` and no error is given. Write `.if X == 0` instead.
A name that is not defined counts as 0.
Blocks can be nested.
A block with no `.endif` stops with `unterminated conditional block`.

There is no `.elif`.

## Loops

```asm
.for i = START, END[, STEP]
    ...
.endf
```

- `END` is **exclusive**: `.for i = 0, 4` runs with `i` = 0, 1, 2, 3.
- `STEP` defaults to `1`, or to `-1` when `START` is greater than `END`. `.for k = 3, 0` gives 3, 2, 1. `.for j = 10, 0, -5` gives 10, 5.
- A step of 0 stops with `.for step cannot be zero`.
- `i` is replaced by its number wherever it appears as a name in the body. Loops can be nested.
- A missing `.endf` stops with an error.

Inside a label, `i:` is not replaced, so use [token pasting](#token-pasting) with something after the variable (`handler##n##_entry:`).

## Duplication

`.dup COUNT VALUE` writes `VALUE` `COUNT` times, separated by commas.
Use it after a data directive.

```asm
    .byte .dup 8 0                  ; 0,0,0,0,0,0,0,0
    .word .dup (COUNT * 2) 0xDEADBEEF
```

`COUNT` and `VALUE` are each a single token or an expression in parentheses.
A negative count is an error.

## Macros

```asm
.macro NAME param1, param2
    ... \param1 ... \param2 ...
.endmacro
```

- `.endm` is a short form of `.endmacro`.
- A parameter is used as `\name` in the body.
- A call is the macro name followed by arguments separated by commas: `push ra`.
- An argument such as `4(sp)` is one argument.
- A missing argument is replaced by nothing. The instruction that uses it usually fails to assemble, and the error points at that instruction, not at the missing argument.
- A macro can call another macro, also one defined later, and macros can be nested.
- A macro name hides an instruction of the same name.
- `.macro` with no name stops with `expected macro name after .macro at line N`.
- A macro that calls itself crashes goasm with a Go stack overflow. Do not write recursive macros.

## Token pasting

`a##b` joins two tokens into one name.
It is mostly used with a `.for` variable or a `.define` name:

```asm
.for i = 0, 4
    addi x##i, zero, i      ; x0, x1, x2, x3 as register names
.endf
```

- A `.define` name or loop variable next to `##` is replaced by its number first, so `l##i##_x:` gives `l0_x:`, `l1_x:`, ...
- A label token that ends with `:` is not replaced. In `handler##COUNT:` the text `COUNT` stays as it is. Put a suffix after the variable.
- `##` with nothing on one side stops with `## requires a token ...`.

## Environment variables

`$[NAME]` reads the environment variable `NAME` when goasm runs.

```asm
.define LEVEL $[N32_LEVEL]
    addi a0, zero, LEVEL
```

`N32_LEVEL=3 goasm env.s` gives `addi a0, zero, 3`.

| Variable | Result |
|----------|--------|
| Number (`10`, `0x1F`, `0b11`, `0o7`) | A number token. |
| Other text, used alone | A string token. |
| Other text, used in `.define` | Error `strconv.ParseInt: parsing "abc": invalid syntax`. |
| Unset or empty | No token. A `.define` of it has the value 0. |

## See also

- [Syntax](syntax.md)
- [Instructions](instructions.md)
- [Quick reference](quick-reference.md)
- [goasm overview](README.md)
