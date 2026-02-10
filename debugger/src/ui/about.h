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
 * @brief	About window
 */
#pragma once

// standard library headers
#include <GLFW/glfw3.h>
#include <imgui.h>

// program-specific includes
#include "constants.h"
#include "generic.h"


// about window
class About : public IGeneric
{
public:
    About(DebugState& state, GLFWwindow* window) :
        IGeneric(state, window, ABOUT_HWND)
    { }

    void render()
    {
        if (ImGui::BeginPopupModal("About##modal", nullptr, ImGuiWindowFlags_AlwaysAutoResize)) {
            ImGui::Text("VCDebug");
            ImGui::Separator();

            ImGui::Text("Version: %s", VERSION.c_str());
            ImGui::Text("Dear ImGui: %s", IMGUI_VERSION);
            ImGui::Text("GLFW: %s", glfwGetVersionString());

            ImGui::Spacing();
            ImGui::Separator();
            ImGui::Spacing();

            // center the OK button
            float button_width = 120.0f;
            float avail_width = ImGui::GetContentRegionAvail().x;
            ImGui::SetCursorPosX((avail_width - button_width) / 2.0f + ImGui::GetCursorPosX());

            if (ImGui::Button("OK", ImVec2(button_width, 0)) ||
                ImGui::IsKeyPressed(ImGuiKey_Escape)) {
                    ImGui::CloseCurrentPopup();
            }

            ImGui::EndPopup();
        }
    }
};
