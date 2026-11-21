# n32asm preprocessor

This page describes the preprocessor of n32asm: includes, constants, conditions, loops, duplication, macros, token pasting and environment variables.

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

Each error gives the file and line of the `.include`:

| Error | Cause |
|-------|-------|
| `in.s:3: error: cannot open include file 'X': no such file or directory` | The file does not exist. |
| `in.s:3: error: .include "X": the file includes itself` | A file includes itself, directly or not. |
| `in.s:3: error: .include: expected a file name in quotes` | The name is not a string. |

To insert raw bytes, use [`.incbin`](syntax.md#data-directives).

## Constants

`.define NAME VALUE` (or `.define NAME, VALUE`) replaces every later use of `NAME` by the number `VALUE`.

```asm
.define STACK_SIZE 256
    li a0, STACK_SIZE
```

- `.define NAME` with no value defines `NAME` as `0`. Use it as a flag with `.ifdef`.
- The comma is optional: `.define A, 5` is `.define A 5`. `.define A,` with no value stops with `expected a value after ','`.
- The value is an expression, computed in 64 bits with the C rules (see [Expressions](#expressions)). `.define D 2+3*4` gives 14.
- `.define` names are not the same as `.equ` names. An `.equ` constant is invisible to `.ifdef` and `.if`. See [Syntax](syntax.md#data-directives).

## Conditions

| Directive | Meaning |
|-----------|---------|
| `.ifdef NAME` | Keep the block if `NAME` is a `.define` name. |
| `.ifndef NAME` | Keep the block if it is not. |
| `.if EXPR` | Keep the block if `EXPR` is not 0. |
| `.else` | Other branch. |
| `.endif` | End of the block. |

`.if` takes an expression (see below); `.if` with nothing after it stops with `.if expects an expression`.
A name that is not defined counts as 0, so `.if !DEBUG` keeps the block when `DEBUG` is not defined.

```asm
.if (COUNT > 3 && !FAST) || MODE == 2
    nop
.endif
```

### Expressions

`.if`, `.define`, `.dup` and `.for` use the same evaluator, with the C precedence:

| Operator | Precedence |
|----------|------------|
| `!` `-` `+` `~` (unary) | 11 (highest) |
| `*` `/` `%` | 10 |
| `+` `-` | 9 |
| `<<` `>>` | 8 |
| `<` `<=` `>` `>=` | 7 |
| `==` `!=` | 6 |
| `&` | 5 |
| `^` | 4 |
| `\|` | 3 |
| `&&` | 2 |
| `\|\|` | 1 (lowest) |

- Parentheses group: `(2+3)*4` is 20.
- Comparisons, `!`, `&&` and `||` give 1 or 0.
- `&&` and `||` stop early: in `0 && 1/0` the division is not computed.
- The values are numbers, characters, `.define` names (another name is 0) and environment variables.
- A wrong expression stops with an error at its line: `missing ')'`, `unexpected ')'`, `expected a value`, `division by zero`, `modulo by zero`, `shift amount 64 out of range (0..63)`.
- Note that `==` binds tighter than `&`, as in C: `6 & 3 == 3` is `6 & (3 == 3)`, so 0. Write `(6 & 3) == 3`.
Blocks can be nested.
A block with no `.endif` stops with `in.s:2: error: unterminated .ifdef (missing .endif)` (or `.if`, `.ifndef`), at the line of the opening directive.
`.ifdef` and `.ifndef` without a name stop with `.ifdef: expected a name`.

There is no `.elif`.

## Loops

```asm
.for i = START, END[, STEP]
    ...
.endf
```

- `END` is **exclusive**: `.for i = 0, 4` runs with `i` = 0, 1, 2, 3.
- `STEP` defaults to `1`, or to `-1` when `START` is greater than `END`. `.for k = 3, 0` gives 3, 2, 1. `.for j = 10, 0, -5` gives 10, 5.
- A step of 0 stops with `in.s:2: error: .for step cannot be zero`. A bad bound names the variable and the part: `in.s:2: error: .for i end: expected a value at the end of the expression`.
- `i` is replaced by its number wherever it appears as a name in the body. Loops can be nested.
- A missing `.endf` stops with `unterminated .for (missing .endf)`, at the line of the `.for`.

Inside a label, `i:` is not replaced, so use [token pasting](#token-pasting) with something after the variable (`handler##n##_entry:`).

## Duplication

`.dup COUNT VALUE` writes `VALUE` `COUNT` times, separated by commas.
Use it after a data directive.

```asm
    .byte .dup 8 0                  ; 0,0,0,0,0,0,0,0
    .word .dup (COUNT * 2) 0xDEADBEEF
```

`COUNT` and `VALUE` are each a single token or an expression in parentheses.
A negative count stops with `in.s:2: error: .dup count cannot be negative (-1)`. A bad expression names the part: `.dup count: ...` or `.dup value: ...`. Like `.if` and `.define`, the errors give the file and the line.

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
- `.macro` with no name stops with `in.s:2: error: .macro: expected a name`. A macro with no `.endm` stops with `unterminated .macro M (missing .endm)`, at the line of the `.macro`.
- A macro can call itself if a `.if` stops the calls (see below).
- A macro that never stops calling itself, directly or through another macro, stops with `FILE:LINE: error: macro 'NAME' expanded too deeply (more than 1000 levels): recursive macro without an end?`. `FILE:LINE` is the last call that was expanded. The limits are 1000 nested levels and 100 000 macro calls in a whole program.

```asm
.macro cnt n
.if \n > 0
    addi a0, a0, \n
    cnt \n - 1             ; stops when n reaches 0
.endif
.endm

    cnt 3                   ; addi a0, a0, 3 / 2 / 1
```

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
- `##` with nothing on one side stops with `in.s:3: error: ## needs a token on its right` (or `left`).

## Environment variables

`$[NAME]` reads the environment variable `NAME` when n32asm runs.

```asm
.define LEVEL $[N32_LEVEL]
    addi a0, zero, LEVEL
```

`N32_LEVEL=3 n32asm env.s` gives `addi a0, zero, 3`.

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
- [n32asm overview](README.md)
