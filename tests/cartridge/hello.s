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
; @brief    Sample cartridge: one section of each kind and an asset id from assets.inc

.include "assets.inc"

.rodata
greeting:
    .word 0x48454C4F

.data
counter:
    .word 0

.bss
buffer:
    .space 16

.text
.global main
main:
    li a0, ASSET_LEVEL1_TILES
loop:
    j loop
