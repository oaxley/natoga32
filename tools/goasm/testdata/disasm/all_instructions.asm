; all_instructions.asm - Test file for disassembler round-trip testing
;
; This file contains all instruction types for round-trip testing.
; Note: Uses assembler-specific syntax for load/store (not standard RISC-V)

.text

; =============================================================================
; RV32I Base Integer Instructions
; =============================================================================

; --- R-type Arithmetic ---
r_type_arith:
    add  t0, t1, t2
    sub  t3, t4, t5
    sll  t0, a0, a1
    slt  a0, a1, a2
    sltu a3, a4, a5
    xor  s0, s1, s2
    srl  s3, s4, s5
    sra  s6, s7, s8
    or   s9, s10, s11
    and  t3, t4, t5

; --- I-type Arithmetic ---
i_type_arith:
    addi t0, t1, 100
    addi t2, t3, -50
    slti a0, a1, 10
    sltiu a2, a3, 20
    xori a4, a5, 0xFF
    ori  a6, a7, 0x55
    andi s0, s1, 0xAA

; --- I2-type Shifts ---
i2_type_shifts:
    slli t0, t1, 5
    srli t2, t3, 10
    srai t4, t5, 15

; --- Load Instructions (I-type) ---
; Standard RISC-V Syntax: lw rd, offset(rs1)
load_instrs:
    lb   t0, 0(a0)
    lh   t1, 2(a1)
    lw   t2, 4(a2)
    lbu  t3, 0(a3)
    lhu  t4, 2(a4)
    lw   t5, -4(sp)       ; Negative offset
    lw   a0, 0(zero)      ; Load from address 0

; --- Store Instructions (S-type) ---
; Standard RISC-V Syntax: sw rs2, offset(rs1)
store_instrs:
    sb   s0, 0(a0)
    sh   s1, 2(a1)
    sw   s2, 4(a2)
    sw   s3, -8(sp)       ; Negative offset
    sw   zero, 0(a3)      ; Store zero

; --- U-type ---
u_type:
    lui  t0, 0x12345
    auipc t1, 0x1000

; --- Branch Instructions ---
branch_instrs:
    beq  t0, t1, branch_target
    bne  t2, t3, branch_target
    blt  t4, t5, branch_target
    bge  a0, a1, branch_target
    bltu a2, a3, branch_target
    bgeu a4, a5, branch_target
branch_target:
    add  zero, zero, zero

; --- Jump Instructions ---
jump_instrs:
    jal  ra, call_target
    jalr t0, t1, 0
    jalr ra, a0, 4
call_target:
    add  zero, zero, zero

; =============================================================================
; RV32M Extension
; =============================================================================

m_extension:
    mul    t0, t1, t2
    mulh   t3, t4, t5
    mulhsu a0, a1, a2
    mulhu  a3, a4, a5
    div    s0, s1, s2
    divu   s3, s4, s5
    rem    s6, s7, s8
    remu   s9, s10, s11

; =============================================================================
; Zbb Extension - Basic Bit-manipulation
; =============================================================================

zbb_extension:
    ; Count instructions (rd, rs1) - unary Zbb
    clz    t0, a0         ; Count leading zeros
    ctz    t1, a1         ; Count trailing zeros
    cpop   t2, a2         ; Population count (count set bits)

    ; Sign extend (rd, rs1) - unary Zbb
    sext.b t3, a3         ; Sign-extend byte
    sext.h t4, a4         ; Sign-extend halfword

    ; Zero extend uses pack (zext.h is alias for pack rd, rs1, zero)
    pack   t5, a5, zero   ; Zero-extend halfword (pack low halves)

    ; Min/max (rd, rs1, rs2)
    min    s0, s1, s2     ; Signed minimum
    minu   s3, s4, s5     ; Unsigned minimum
    max    s6, s7, s8     ; Signed maximum
    maxu   s9, s10, s11   ; Unsigned maximum

    ; Rotate (rd, rs1, rs2/shamt)
    rol    t0, t1, t2     ; Rotate left
    ror    t3, t4, t5     ; Rotate right
    rori   a0, a1, 7      ; Rotate right immediate

    ; Byte reverse (rd, rs1) - unary Zbb
    rev8   a2, a3         ; Reverse bytes in word

; =============================================================================
; Zbs Extension - Single-bit Instructions
; =============================================================================

zbs_extension:
    ; Bit set (rd, rs1, rs2/shamt)
    bset   t0, t1, t2     ; Bit set
    bseti  t3, t4, 5      ; Bit set immediate

    ; Bit clear (rd, rs1, rs2/shamt)
    bclr   t5, a0, a1     ; Bit clear
    bclri  a2, a3, 10     ; Bit clear immediate

    ; Bit invert (rd, rs1, rs2/shamt)
    binv   a4, a5, a6     ; Bit invert
    binvi  a7, s0, 15     ; Bit invert immediate

    ; Bit extract (rd, rs1, rs2/shamt)
    bext   s1, s2, s3     ; Bit extract
    bexti  s4, s5, 20     ; Bit extract immediate

; =============================================================================
; Zicond Extension - Conditional Operations
; =============================================================================

zicond_extension:
    czero.eqz t0, t1, t2  ; Zero if rs2 == 0
    czero.nez t3, t4, t5  ; Zero if rs2 != 0

; =============================================================================
; Zbkb Extension - Bit-manipulation for Cryptography
; =============================================================================

zbkb_extension:
    pack   s0, s1, s2     ; Pack low halves
    packh  s3, s4, s5     ; Pack low bytes

; =============================================================================
; CSR Instructions
; =============================================================================

; Syntax: csrrw rd, rs1, csr (assembler format)
csr_instrs:
    csrrw  t0, a0, 0x300      ; Read/write mstatus
    csrrs  t1, a1, 0x304      ; Read/set mie
    csrrc  t2, a2, 0x305      ; Read/clear mtvec
    csrrw  t3, zero, 0x340    ; Write mscratch (read discarded)
    csrrs  t4, zero, 0x341    ; Read mepc (no bits set)
    csrrc  t5, zero, 0x342    ; Read mcause (no bits cleared)

; CSR immediate variants - uimm is encoded as register number
csr_imm_instrs:
    csrrwi a0, x0, 0x300      ; uimm=0
    csrrsi a1, x5, 0x304      ; uimm=5
    csrrci a2, x15, 0x305     ; uimm=15
    csrrwi a3, x31, 0x340     ; uimm=31 (max)

; =============================================================================
; System Instructions (no operands)
; =============================================================================

system_instrs:
    ecall
    mret
    wfi

; =============================================================================
; Custom Thread Instructions
; =============================================================================

thread_instrs:
    new.t  t0, t1
    yield.t
    id.t   t2
    sleep.t ra, t3
    wake.t  ra, t4
    end.t

; =============================================================================
; Pseudo-Instructions (will be expanded by assembler)
; =============================================================================

pseudo_instrs:
    nop
    li   a0, 42
    li   a1, -100
    mv   t0, a0
    not  t1, a1
    neg  t2, a2
    j    pseudo_target
    jr   t0
    ret
    beqz t0, pseudo_target
    bnez t1, pseudo_target
    blez t2, pseudo_target
    bgez t3, pseudo_target
    bltz t4, pseudo_target
    bgtz t5, pseudo_target
    bgt  a0, a1, pseudo_target
    ble  a2, a3, pseudo_target
    bgtu a4, a5, pseudo_target
    bleu a6, a7, pseudo_target
pseudo_target:
    seqz t0, a0
    snez t1, a1
    sltz t2, a2
    sgtz t3, a3

; =============================================================================
; Edge Cases
; =============================================================================

edge_cases:
    ; Max/min immediate values for I-type
    addi t0, zero, 2047      ; Max positive 12-bit signed
    addi t1, zero, -2048     ; Min negative 12-bit signed

    ; All zeros instruction (nop)
    addi zero, zero, 0

    ; Using numeric register names
    add  x0, x1, x2
    add  x3, x4, x5
    add  x6, x7, x8
    add  x9, x10, x11
    add  x12, x13, x14
    add  x15, x16, x17
    add  x18, x19, x20
    add  x21, x22, x23
    add  x24, x25, x26
    add  x27, x28, x29
    add  x30, x31, x0

end_marker:
    ecall
