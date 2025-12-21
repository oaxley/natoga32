# Toolchain


## Compiler

- Risc-V and x68FP compiler


### Supported directives

#### Generic directives
- `.cpu` : select the CPU target (`risc-v` or `x68fp`)
- `.entrypoint` : Specify the beginning of the program
- `.data` : select the initialized data section
- `.bss` : select the un-initialized data section
- `.text` : select the code section
- `.org` : select the origin relative address of the section
- `.align` : align the next instruction or section on the specified value

#### Data declaration
- `.byte` : the next values until EOL are considered byte (8-bit) values
- `.word` : the next values until EOF are considered word values (1).
- `.half` : the next values until EOF are considered half-word values (1).
- `.short` : alias to `.half`

(1) The word (and half) is CPU dependent.  
For `risc-v` it's equivalent to XLEN so 32-bit, making `.half` 16-bit.  
For `x68fp`, word is 16-bit and `.half` is equivalent to `.byte`

```
.data
name: .byte 'R', 105, 115, 0x63, '-', 0b0101_0110
```


- `.asciz` : declare a null terminating string
- `.string` : alias to `.asciz`

```
.data
msg: .string "Hello, World!"
```

#### Memory allocation
- `.skip` : reserve N bytes of uninitialized space
- `.space` : alias to `.skip`

```  
.bss
array: .space 20
matrix: .skip 32
```

### Pre-processor

#### Token pasting

Token pasting is available inside the pre-processor with the special syntax `##`.


#### Environment variables

An environment variable can be referenced inside the source with the syntax
`$[<name>]`.
If the variable cannot be found, it is expanded to "" (empty string).


#### Constants definition

Constants definition can happen anywhere in the source code, as they are not
tight to any section.

- `.define` : define a new constant
- `.equ` : alias to `.define`

```
HTTP_OK: .define 200
HTTP_KO: .equ 404
```

#### Includes
- `.include` : include the source file at the current position
- `.incbin` : include the binary file at the current position

```
.data
logo: .incbin "local.png"
```

/!\ `.incbin` is not supported for `.bss` section.


#### Macros
- `.macro` : define a new macro
- `.endm` : end of the macro definition

**Syntax:**  
```
.macro <name> <arg1> <arg2> ... <argN>
[body of the macro]
.endm
```

Arguments inside the body of the macro are expanded normally.

**Example:**
```
; this is the macro definition
.macro PUSH r
  addi sp, sp, -4
  sw r, 0(sp)
.endm

...

main:
  PUSH x0   ; <-- this will be expanded by the pre-processor
  PUSH x1
```

After the pre-processing, the source become:
```
main:
  addi sp, sp, -4
  sw x0, 0(sp)
  addi sp, sp, -4
  sw x1, 0(sp)
```

#### For loops
- `.for` : define a new for loop
- `.endf` : end of the for loop definition

**Syntax:**
```
.for <var> = <init>, <end> [, step]
[body]
.endf
```

- var : the identifier for the internal counter
- init : the initial value for the counter
- end : the end value (excluded) for the counter
- step : the step value for the counter. By default 1

**Example:**
```
; save all the registers from x0 to x9
.for i = 0, 10
  addi sp, sp, -4
  sw x##i, 0(sp)        <-- use token pasting to create the register name
.endf
```

#### Conditionals
- `.if` / `.else` / `.endif` : define a new block if / else
- `.ifdef` : check if a symbol is defined
- `.ifndef` : Check if a symbol is not defined

**Syntax:**
```
.if <expr>
  <block #1>
.else
  <block #2>
.endif
```

- expr represents any expression that needs to evaluate to True to execute block #1
- The `else` block is optional
