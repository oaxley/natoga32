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
void showMenubar(States* states)
{
    if (ImGui::BeginMainMenuBar()) {
        // File Menu
        if (ImGui::BeginMenu("File")) {
            if (ImGui::MenuItem("Load Symbols")) {
            }

            ImGui::Separator();
            if (ImGui::MenuItem("Preferences")) {
            }

            ImGui::Separator();
            if (ImGui::MenuItem("Quit", "Ctrl+Q")) {
                states->quit = true;
            }
            ImGui::EndMenu();
        }

        // Edit Menu
        if (ImGui::BeginMenu("Edit")) {
            if (ImGui::MenuItem("Undo", "Ctrl+Z")) {
            }
            if (ImGui::MenuItem("Redo", "Ctrl+Y")) {
            }
            ImGui::Separator();
            if (ImGui::MenuItem("Copy", "Ctrl+C")) {
            }
            if (ImGui::MenuItem("Paste", "Ctrl+V")) {
            }
            ImGui::Separator();
            if (ImGui::MenuItem("Find", "Ctrl+F")) {
            }
            if (ImGui::MenuItem("Replace", "Ctrl+H")) {
            }
            ImGui::EndMenu();
        }

        // Window Menu
        if (ImGui::BeginMenu("Window")) {

            // memory
            if (ImGui::BeginMenu("Memory")) {
                if (ImGui::MenuItem("Heap")) {
                }
                if (ImGui::MenuItem(".BSS")) {
                }
                if (ImGui::MenuItem(".DATA")) {
                }
                if (ImGui::MenuItem(".TEXT")) {
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Explorer")) {
                }
                ImGui::EndMenu();
            }

            // Code
            if (ImGui::BeginMenu("Code")) {
                if (ImGui::MenuItem("Source")) {
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("Threads")) {
                    if (ImGui::MenuItem("Info")) {
                    }
                    if (ImGui::MenuItem("Stack")) {
                    }
                    if (ImGui::MenuItem("Local Storage")) {
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Breakpoints")) {
                }
                ImGui::EndMenu();
            }

            // CSR & Signals
            if (ImGui::BeginMenu("CSR/Signals")) {
                if (ImGui::MenuItem("Control Status Reg.")) {
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Clock")) {
                }
                if (ImGui::MenuItem("Video VSYNC")) {
                }
                if (ImGui::MenuItem("Video HSYNC")) {
                }
                if (ImGui::MenuItem("Interrupts")) {
                }
                ImGui::EndMenu();
            }

            // Video
            if (ImGui::BeginMenu("Video")) {
                if (ImGui::BeginMenu("Buffer")) {
                    if (ImGui::MenuItem("FrameBuffer A")) {
                    }
                    if (ImGui::MenuItem("FrameBuffer B")) {
                    }
                    if (ImGui::MenuItem("Z-Buffer")) {
                    }
                    if (ImGui::MenuItem("Text Layer")) {
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("Sprites")) {
                    if (ImGui::MenuItem("Viewer")) {
                    }
                    if (ImGui::MenuItem("Attributes")) {
                    }
                    if (ImGui::MenuItem("Palettes")) {
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("BG")) {
                    if (ImGui::MenuItem("Viewer")) {
                    }
                    if (ImGui::MenuItem("Map")) {
                    }
                    if (ImGui::MenuItem("Palettes")) {
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::BeginMenu("Fonts")) {
                    if (ImGui::MenuItem("Viewer")) {
                    }
                    if (ImGui::MenuItem("Palettes")) {
                    }
                    ImGui::EndMenu();
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Affine Matrices")) {
                }
                ImGui::EndMenu();
            }

            ImGui::EndMenu();
        }

        // Console
        if (ImGui::BeginMenu("Console")) {
            if (ImGui::BeginMenu("Run")) {
                if (ImGui::MenuItem("Step")) {

                }
                ImGui::Separator();
                if (ImGui::MenuItem("Pause")) {

                }
                if (ImGui::MenuItem("Resume")) {
                }
                ImGui::Separator();
                if (ImGui::MenuItem("Reset")) {

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
            }

            if (ImGui::MenuItem("Instr. Encoder")) {
            }

            ImGui::EndMenu();
        }

        // Help
        if (ImGui::BeginMenu("Help")) {
            if (ImGui::MenuItem("About")) {
                states->show_about = true;
            }
            ImGui::EndMenu();
        }

        ImGui::EndMainMenuBar();
    }
}
