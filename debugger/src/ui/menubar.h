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
 * @brief	Main Menu bar
 */
#pragma once

// standard library headers
#include <GLFW/glfw3.h>
#include <imgui.h>

// program-specific includes
#include "generic.h"
#include "states.h"


// menubar definition
void showMenubar(States& state)
{
    if (ImGui::BeginMainMenuBar()) {
        // File Menu
        if (ImGui::BeginMenu("File")) {
            if (ImGui::MenuItem("Load Symbols")) {
                state = States::LoadSymbols;
            }

            ImGui::Separator();
            if (ImGui::MenuItem("Preferences")) {
                state = States::Preferences;
            }

            ImGui::Separator();
            if (ImGui::MenuItem("Quit", "Ctrl+Q")) {
                state = States::Quit;
            }
            ImGui::EndMenu();
        }

        // Edit Menu
        if (ImGui::BeginMenu("Edit")) {
            if (ImGui::MenuItem("Undo", "Ctrl+Z")) {
                state = States::Undo;
            }
            if (ImGui::MenuItem("Redo", "Ctrl+Y")) {
                state = States::Redo;
            }
            ImGui::Separator();
            if (ImGui::MenuItem("Copy", "Ctrl+C")) {
                state = States::Copy;
            }
            if (ImGui::MenuItem("Paste", "Ctrl+V")) {
                state = States::Paste;
            }
            ImGui::Separator();
            if (ImGui::MenuItem("Find", "Ctrl+F")) {
                state = States::Find;
            }
            if (ImGui::MenuItem("Replace", "Ctrl+H")) {
                state = States::Replace;
            }
            ImGui::EndMenu();
        }

        // Window Menu
        if (ImGui::BeginMenu("Window")) {

            // memory
            if (ImGui::BeginMenu("Memory")) {
                if (ImGui::MenuItem("Heap")) {
                    state = States::Heap;
                }
                if (ImGui::MenuItem(".BSS")) {
                    state = States::Bss;
                }
                if (ImGui::MenuItem(".DATA")) {
                    state = States::Data;
                }
                if (ImGui::MenuItem(".TEXT")) {
                    state = States::Text;
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Explorer")) {
                    state = States::Explorer;
                }
                ImGui::EndMenu();
            }

            // Code
            if (ImGui::BeginMenu("Code")) {
                if (ImGui::MenuItem("Source")) {
                    state = States::Source;
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("Threads")) {
                    if (ImGui::MenuItem("Info")) {
                        state = States::ThreadsInfo;
                    }
                    if (ImGui::MenuItem("Stack")) {
                        state = States::ThreadsStack;
                    }
                    if (ImGui::MenuItem("Local Storage")) {
                        state = States::ThreadsStorage;
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Breakpoints")) {
                    state = States::Breakpoints;
                }
                ImGui::EndMenu();
            }

            // CSR & Signals
            if (ImGui::BeginMenu("CSR/Signals")) {
                if (ImGui::MenuItem("Control Status Reg.")) {
                    state = States::CSR;
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Clock")) {
                    state = States::Clock;
                }
                if (ImGui::MenuItem("Video VSYNC")) {
                    state = States::VSync;
                }
                if (ImGui::MenuItem("Video HSYNC")) {
                    state = States::HSync;
                }
                if (ImGui::MenuItem("Interrupts")) {
                    state = States::Interrupts;
                }
                ImGui::EndMenu();
            }

            // Video
            if (ImGui::BeginMenu("Video")) {
                if (ImGui::BeginMenu("Buffer")) {
                    if (ImGui::MenuItem("FrameBuffer A")) {
                        state = States::FrameBufferA;
                    }
                    if (ImGui::MenuItem("FrameBuffer B")) {
                        state = States::FrameBufferB;
                    }
                    if (ImGui::MenuItem("Z-Buffer")) {
                        state = States::ZBuffer;
                    }
                    if (ImGui::MenuItem("Text Layer")) {
                        state = States::TextLayer;
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("Sprites")) {
                    if (ImGui::MenuItem("Viewer")) {
                        state = States::SpriteViewer;
                    }
                    if (ImGui::MenuItem("Attributes")) {
                        state = States::SpriteAttributes;
                    }
                    if (ImGui::MenuItem("Palettes")) {
                        state = States::SpritePalettes;
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("BG")) {
                    if (ImGui::MenuItem("Viewer")) {
                        state = States::BGViewer;
                    }
                    if (ImGui::MenuItem("Map")) {
                        state = States::BGMap;
                    }
                    if (ImGui::MenuItem("Palettes")) {
                        state = States::BGPalettes;
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("Fonts")) {
                    if (ImGui::MenuItem("Viewer")) {
                        state = States::FontViewer;
                    }
                    if (ImGui::MenuItem("Palettes")) {
                        state = States::FontPalettes;
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Affine Matrices")) {
                    state = States::Affine;
                }
                ImGui::EndMenu();
            }

            ImGui::EndMenu();
        }

        // Console
        if (ImGui::BeginMenu("Console")) {
            if (ImGui::BeginMenu("Run")) {
                if (ImGui::MenuItem("Step")) {
                    state = States::RunStep;
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Pause")) {
                    state = States::RunPause;
                }
                if (ImGui::MenuItem("Resume")) {
                    state = States::RunResume;
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Reset")) {
                    state = States::RunReset;
                }
                ImGui::EndMenu();
            }
            if (ImGui::MenuItem("Audio")) {
            }
            if (ImGui::MenuItem("Input")) {
            }

            ImGui::EndMenu();
        }

        // Tools
        if (ImGui::BeginMenu("Tools")) {
            if (ImGui::MenuItem("Hex Converter")) {
                state = States::HexConverter;
            }
            if (ImGui::MenuItem("Instr. Encoder")) {
                state = States::InstrEncoder;
            }

            ImGui::EndMenu();
        }

        // Help
        if (ImGui::BeginMenu("Help")) {
            if (ImGui::MenuItem("About")) {
                state = States::About;
            }
            ImGui::EndMenu();
        }

        ImGui::EndMainMenuBar();
    }
}
