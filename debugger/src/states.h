/* -*- coding: utf-8 -*-
 * vim: filetype=cpp
 *
 * This source file is subject to the MIT License
 * that is bundled with this package in the file LICENSE.txt.
 * It is also available through the Internet at this address:
 * https://opensource.org/license/mit
 *
 * @author	Sebastien LEGRAND
 * @license	MIT License
 *
 * @brief	Debugger states values
 */
#pragma once

enum class States : uint16_t
{
    //----- Reserved
    NoState = 0x0000,

    //----- File
    LoadSymbols = 0x1000,
    Preferences = 0x1010,
    Quit = 0x1020,

    //----- Edit
    Undo = 0x2000,
    Redo = 0x2010,
    Copy = 0x2020,
    Paste = 0x2030,
    Find = 0x2040,
    Replace = 0x2050,

    //------ Window
    // Window > Memory
    Heap = 0x3000,
    Bss = 0x3010,
    Data = 0x3020,
    Text = 0x3030,
    Explorer = 0x3040,

    // Window > Code
    Source = 0x3100,
    ThreadsInfo = 0x3110,
    ThreadsStack = 0x3111,
    ThreadsStorage = 0x3112,
    Breakpoints = 0x3120,

    // Window > CSR & Signals
    CSR = 0x3200,
    Clock = 0x3210,
    VSync = 0x3220,
    HSync = 0x3230,
    Interrupts = 0x3240,

    // Window > Video > Buffer
    FrameBuffer = 0x3300,
    ZBuffer = 0x3301,
    TextLayer = 0x3302,

    // Window > Video > Sprites
    SpriteViewer = 0x3310,
    SpriteAttributes = 0x3311,
    SpritePalettes = 0x3312,

    // Window > Video > BG
    BGViewer = 0x3320,
    BGMap = 0x3321,
    BGPalettes = 0x3322,

    // Window > Video > Fonts
    FontViewer = 0x3330,
    FontPalettes = 0x3331,

    // Window > Video > Affine
    Affine = 0x3340,

    //----- Console
    // Console > Run
    RunStep = 0x3410,
    RunPause = 0x3411,
    RunResume = 0x3412,
    RunReset = 0x3413,

    // Console > Audio
    // 0x3510

    // Console > Input
    // 0x3610

    //----- Tools
    HexConverter = 0x3710,
    InstrEncoder = 0x3720,

    //----- Help
    About = 0x3810,
};
