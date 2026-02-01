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
void showMenubar(GLFWwindow* window, States* states)
{
    if (ImGui::BeginMainMenuBar()) {
        if (ImGui::BeginMenu("File")) {
            if (ImGui::MenuItem("Open", "Ctrl+O")) {
                // handle open
            }
            if (ImGui::MenuItem("Save", "Ctrl+S")) {
                // handle save
            }
            ImGui::Separator();
            if (ImGui::MenuItem("Quit", "Ctrl+Q")) {
                glfwSetWindowShouldClose(window, true);
            }
            ImGui::EndMenu();
        }
        if (ImGui::BeginMenu("View")) {
            if (ImGui::MenuItem("Registers")) {
                // toggle registers window
            }
            if (ImGui::MenuItem("Memory")) {
                // toggle memory window
            }
            if (ImGui::MenuItem("Disassembly")) {
                // toggle disassembly window
            }
            ImGui::EndMenu();
        }
        if (ImGui::BeginMenu("Help")) {
            if (ImGui::MenuItem("About")) {
                states->show_about = true;
            }
            ImGui::EndMenu();
        }
        ImGui::EndMainMenuBar();
    }
}
