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
 * @brief	Basic widgets to show the registers
 */
#pragma once

// standard library headers
#include <array>
#include <imgui.h>

// program-specific includes
#include "generic.h"

// class definition
class Registers : public IGeneric
{
public:
    Registers(DebugClient& client, GLFWwindow* window, const std::array<uint32_t, 32>& regs) :
        IGeneric(client, window), regs_{regs}
    { }

    void render()
    {
        if (ImGui::BeginTable("Registers", 4, ImGuiTableFlags_SizingFixedFit)) {
            ImGui::TableSetupColumn("##name1", ImGuiTableColumnFlags_WidthFixed, 30.0f);
            ImGui::TableSetupColumn("##val1", ImGuiTableColumnFlags_WidthFixed, 70.0f);
            ImGui::TableSetupColumn("##name2", ImGuiTableColumnFlags_WidthFixed, 30.0f);
            ImGui::TableSetupColumn("##val2", ImGuiTableColumnFlags_WidthFixed, 70.0f);

            for (int y = 0; y < 16; y++) {
                ImGui::TableNextColumn();
                ImGui::Text("x%d", y);

                ImGui::TableNextColumn();
                // make the text selectable
                char buffer[16];
                snprintf(buffer, sizeof(buffer), "%08X", regs_[y]);
                ImGui::PushID(y);
                ImGui::SetNextItemWidth(-FLT_MIN);
                ImGui::InputText("##v1", buffer, sizeof(buffer), ImGuiInputTextFlags_ReadOnly);
                ImGui::PopID();

                ImGui::TableNextColumn();
                ImGui::Text("x%d", y+16);
                ImGui::TableNextColumn();

                snprintf(buffer, sizeof(buffer), "%08X", regs_[y+16]);
                ImGui::PushID(y + 100);
                ImGui::SetNextItemWidth(-FLT_MIN);
                ImGui::InputText("##v2", buffer, sizeof(buffer), ImGuiInputTextFlags_ReadOnly);
                ImGui::PopID();

                if (y < 15)
                    ImGui::TableNextRow();
            }

            ImGui::EndTable();
        }
    }

private:
    std::array<uint32_t, 32> regs_;

};
