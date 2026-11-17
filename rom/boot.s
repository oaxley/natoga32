; -*- coding: utf-8 -*-
;
; This source file is subject to the MIT License
; that is bundled with this package in the file LICENSE.txt.
; It is also available through the Internet at this address:
; https://opensource.org/license/mit
;
; @author   Sebastien LEGRAND
; @license  MIT License
;
; @brief    NATOGA32 boot ROM: logo fade-in and "insert a cartridge" message
;
; Every routine is a leaf that only uses a0-a7 and t0-t6, so main keeps its
; state in s0-s11 and no stack is needed.

;----- Constants

.equ CSR_MTVEC,         0x305
.equ CSR_MCAUSE,        0x342

.equ VIDEO_RAM,         0x00380000      ; framebuffer A, 640x384 RGB565
.equ VIDEO_ROW_BYTES,   1280
.equ VIDEO_WORDS,       122880          ; 640 * 384 / 2
.equ HEAP,              0x00800000
.equ STACK_TOP,         0x00790000      ; top of thread 0 stack slot

.equ FADE_TABLE,        0x00800000      ; 16 halfwords: colour of each level
.equ PAIR_TABLE,        0x00800040      ; 256 words: colours of a logo byte
.equ HEX_BUFFER,        0x00800440      ; 8 hex digits + NUL

.equ COLOR_BACKGROUND,  0x0843
.equ COLOR_MESSAGE,     0xCE5A
.equ COLOR_ERROR,       0xF800
.equ COLOR_WHITE,       0xFFFF

.equ LOGO_ROW_BYTES,    280             ; 560 pixels at 4 bpp
.equ LOGO_HEIGHT,       205
.equ LOGO_DEST,         0x00392C50      ; VIDEO_RAM + 60 * 1280 + 40 * 2

.equ FADE_STEPS,        15
.equ FADE_FRAMES,       3               ; frames shown per fade step
.equ BLINK_FRAMES,      30
.equ WAITKEY_VBLANK,    0x80000001

.rom

;----- Header

.org 0x000
    .ascii "N32R"
    .half 0, 1                          ; version major, minor
    .word 0x100                         ; entry point
    .word trap_handler
    .word 0
    .word logo
    .word font

;----- Entry point

.org 0x100
_start:
    la t0, trap_handler
    csrw CSR_MTVEC, t0

    ; navy background
    li a0, COLOR_BACKGROUND
    li a1, 0
    li a2, 384
    call fill_rows

    ; logo fade-in, FADE_FRAMES frames per step
    li s0, 1
    li s1, FADE_STEPS
fade_loop:
    mv a0, s0
    call build_fade_table
    call draw_logo
    li s2, FADE_FRAMES
fade_wait:
    call wait_vblank
    addi s2, s2, -1
    bnez s2, fade_wait
    addi s0, s0, 1
    ble s0, s1, fade_loop

    ; blinking message: 30 frames on, 30 frames off
blink_loop:
    la a0, msg_cartridge
    li a1, 48
    li a2, 310
    li a3, COLOR_MESSAGE
    call draw_text
    li s2, BLINK_FRAMES
blink_on:
    call wait_vblank
    addi s2, s2, -1
    bnez s2, blink_on

    la a0, msg_cartridge
    li a1, 48
    li a2, 310
    li a3, COLOR_BACKGROUND
    call draw_text
    li s2, BLINK_FRAMES
blink_off:
    call wait_vblank
    addi s2, s2, -1
    bnez s2, blink_off
    j blink_loop

;----- wait_vblank: sleep until the next vblank event (clobbers a0, a1)

wait_vblank:
    li a0, WAITKEY_VBLANK
    sleep.t a0, zero
    ret

;----- fill_rows(a0 = colour, a1 = first row, a2 = row count)

fill_rows:
    slli t0, a0, 16
    or t0, t0, a0                       ; two pixels per word
    li t1, VIDEO_ROW_BYTES
    mul t2, a1, t1
    li t3, VIDEO_RAM
    add t2, t2, t3                      ; first word
    mul t4, a2, t1
    add t4, t4, t2                      ; end (a row is 80 x 16 bytes)
fill_rows_loop:
    sw t0, 0(t2)
    sw t0, 4(t2)
    sw t0, 8(t2)
    sw t0, 12(t2)
    addi t2, t2, 16
    bltu t2, t4, fill_rows_loop
    ret

;----- build_fade_table(a0 = step 1..15)
; level v colour: R5 = 1 + 30*v*s/225, G6 = 2 + 61*v*s/225, B5 = 3 + 28*v*s/225
; then the pair table: for each logo byte, left pixel (high nibble) in the low
; halfword, right pixel (low nibble) in the high halfword

build_fade_table:
    li t0, 0                            ; v * s
    li t1, FADE_TABLE
    addi t2, t1, 32                     ; end of the 16 halfwords
    li t3, 225
fade_level_loop:
    li t4, 30
    mul t4, t4, t0
    divu t4, t4, t3
    addi t4, t4, 1
    slli t5, t4, 11                     ; R5

    li t4, 61
    mul t4, t4, t0
    divu t4, t4, t3
    addi t4, t4, 2
    slli t4, t4, 5
    or t5, t5, t4                       ; G6

    li t4, 28
    mul t4, t4, t0
    divu t4, t4, t3
    addi t4, t4, 3
    or t5, t5, t4                       ; B5

    sh t5, 0(t1)
    add t0, t0, a0
    addi t1, t1, 2
    bne t1, t2, fade_level_loop

    li t0, 0                            ; logo byte value
    li t1, FADE_TABLE
    li t2, PAIR_TABLE
    li t3, 256
fade_pair_loop:
    srli t4, t0, 4
    slli t4, t4, 1
    add t4, t4, t1
    lhu t4, 0(t4)                       ; left pixel
    andi t5, t0, 15
    slli t5, t5, 1
    add t5, t5, t1
    lhu t5, 0(t5)                       ; right pixel
    slli t5, t5, 16
    or t4, t4, t5
    sw t4, 0(t2)
    addi t2, t2, 4
    addi t0, t0, 1
    bne t0, t3, fade_pair_loop
    ret

;----- draw_logo: draw the logo at (40, 60) with the current pair table

draw_logo:
    la a0, logo                         ; source
    li a1, LOGO_DEST                    ; destination
    li a2, PAIR_TABLE
    li a3, LOGO_HEIGHT
draw_logo_row:
    addi a4, a0, LOGO_ROW_BYTES         ; end of the source row
draw_logo_loop:
    lbu t0, 0(a0)
    lbu t1, 1(a0)
    slli t0, t0, 2
    slli t1, t1, 2
    add t0, t0, a2
    add t1, t1, a2
    lw t0, 0(t0)
    lw t1, 0(t1)
    sw t0, 0(a1)
    sw t1, 4(a1)
    addi a0, a0, 2
    addi a1, a1, 8
    bne a0, a4, draw_logo_loop

    addi a1, a1, 160                    ; 1280 - 560 * 2
    addi a3, a3, -1
    bnez a3, draw_logo_row
    ret

;----- draw_text(a0 = string, a1 = x, a2 = y, a3 = colour)
; font8x8 glyphs 0x20-0x5F scaled x2, only the set pixels are drawn
; (x must be even so that two pixels fit a word)

draw_text:
    slli t6, a3, 16
    or t6, t6, a3                       ; two pixels per word
    li t0, VIDEO_ROW_BYTES
    mul a2, a2, t0
    slli a1, a1, 1
    add a1, a1, a2
    li t0, VIDEO_RAM
    add a1, a1, t0                      ; top-left of the current glyph
    la a4, font
draw_text_char:
    lbu t0, 0(a0)
    beqz t0, draw_text_done
    addi t0, t0, -0x20
    slli t0, t0, 3
    add t0, t0, a4                      ; glyph
    mv t1, a1                           ; destination row
    addi t2, t0, 8                      ; end of the glyph
draw_text_row:
    lbu t3, 0(t0)                       ; bit 0 = leftmost pixel
    mv t4, t1
draw_text_pixel:
    andi t5, t3, 1
    beqz t5, draw_text_next
    sw t6, 0(t4)
    sw t6, VIDEO_ROW_BYTES(t4)
draw_text_next:
    srli t3, t3, 1
    addi t4, t4, 4
    bnez t3, draw_text_pixel
    addi t0, t0, 1
    addi t1, t1, VIDEO_ROW_BYTES        ; two framebuffer rows
    addi t1, t1, VIDEO_ROW_BYTES
    bne t0, t2, draw_text_row
    addi a0, a0, 1
    addi a1, a1, 32                     ; 16 pixels
    j draw_text_char
draw_text_done:
    ret

;----- Trap handler: red band, "SYSTEM ERROR" and mcause, then idle forever

trap_handler:
    li sp, STACK_TOP

    li a0, COLOR_ERROR
    li a1, 0
    li a2, 64
    call fill_rows

    la a0, msg_error
    li a1, 16
    li a2, 16
    li a3, COLOR_WHITE
    call draw_text

    ; mcause as 8 hex digits
    csrr t0, CSR_MCAUSE
    li t1, HEX_BUFFER
    addi t2, t1, 8
    li t3, 10
trap_hex_loop:
    srli t4, t0, 28
    slli t0, t0, 4
    addi t5, t4, '0'
    blt t4, t3, trap_hex_store
    addi t5, t4, 'A' - 10
trap_hex_store:
    sb t5, 0(t1)
    addi t1, t1, 1
    bne t1, t2, trap_hex_loop
    sb zero, 0(t1)

    li a0, HEX_BUFFER
    li a1, 16
    li a2, 40
    li a3, COLOR_WHITE
    call draw_text

trap_idle:
    call wait_vblank
    j trap_idle

;----- Data

msg_cartridge:
    .string "INSERT A CARTRIDGE AND PRESS RESET"
msg_error:
    .string "SYSTEM ERROR"

.align 4
logo:
    .incbin "logo.raw"

.align 4
font:
    .incbin "font.raw"
