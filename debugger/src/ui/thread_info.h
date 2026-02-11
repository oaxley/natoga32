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
 * @brief	ThreadInfo window
 */
#pragma once

// standard library headers
#include <imgui.h>
#include <iostream>

// program-specific includes
#include "generic.h"
#include "../constants.h"
#include "../debug_state.h"


// class definition
class UIThreadInfo : public IGeneric
{
public:
    UIThreadInfo(DebugState& state, GLFWwindow* window) :
        IGeneric(state, window, THREADINFO_HWND), thread_id_{state_.getActiveThreadID()}
    { }

    ~UIThreadInfo()
    { }

    void render()
    {
        if (!state_.isVisible(id_)) return;

        if (ImGui::Begin("ThreadInfo", &state_.get(id_))) {
            // retrieve the thread registers values
            ThreadInfo info = state_.getThreadInfo(thread_id_);

            ImGui::Text("Active Thread:");
            ImGui::SameLine();
            ImGui::SetNextItemWidth(80.0f);
            ImGui::InputInt("##l44", &thread_id_);
            ImGui::SameLine(250);
            ImGui::Text("%s", stateToText(info.state).c_str());

            if (thread_id_ < 0) thread_id_ = 0;
            if (thread_id_ >= ThreadCount) thread_id_ = ThreadCount - 1;

            ImGui::SetCursorPosX(48);
            ImGui::Text("PC:");
            ImGui::SameLine(80);
            ImGui::Text("%08X", info.pc);

            ImGui::SameLine(198);
            ImGui::Text("Cycles:");
            ImGui::SameLine();
            ImGui::Text("%ld", info.cycles);

            ImGui::Text("Waitkey:");
            ImGui::SameLine(80);
            ImGui::Text("%08X", info.waitkey);

            ImGui::SameLine(160);
            ImGui::Text("Sleep until:");
            ImGui::SameLine();
            ImGui::Text("%ld", info.sleep_until);

            ImGui::SeparatorText("Registers");

            // registers table
            ImGui::SetCursorPosX(40);
            ImGui::BeginTable("Registers", 4, ImGuiTableFlags_SizingFixedFit);

            ImGui::TableSetupColumn("##name1", ImGuiTableColumnFlags_WidthFixed, 30.0f);
            ImGui::TableSetupColumn("##val1", ImGuiTableColumnFlags_WidthFixed, 70.0f);
            ImGui::TableSetupColumn("##name2", ImGuiTableColumnFlags_WidthFixed, 40.0f);
            ImGui::TableSetupColumn("##val2", ImGuiTableColumnFlags_WidthFixed, 70.0f);

            char buffer[16];
            for (int i = 0; i < (RegisterCount >> 1); i++)
            {
                // first column: register #1 name
                ImGui::TableNextColumn();
                ImGui::Text("x%d", i);

                // second column: register #1 value
                // make the text selectable
                ImGui::TableNextColumn();
                snprintf(buffer, sizeof(buffer), "%08X", info.regs[i]);
                ImGui::PushID(thread_id_ * 100 + i);
                ImGui::SetNextItemWidth(-FLT_MIN);
                ImGui::InputText("##v1", buffer, sizeof(buffer), ImGuiInputTextFlags_ReadOnly);
                ImGui::PopID();

                // first column: register #2 name
                ImGui::TableNextColumn();
                ImGui::Text("x%d", i+16);

                // second column: register #2 value
                // make the text selectable
                ImGui::TableNextColumn();
                snprintf(buffer, sizeof(buffer), "%08X", info.regs[i + 16]);
                ImGui::PushID(thread_id_ * 100 + (i+16));
                ImGui::SetNextItemWidth(-FLT_MIN);
                ImGui::InputText("##v2", buffer, sizeof(buffer), ImGuiInputTextFlags_ReadOnly);
                ImGui::PopID();
            }
            ImGui::EndTable();
        }
        ImGui::End();
    }

private:
    //----- members
    int thread_id_;

    //----- methods
    std::string stateToText(ThreadState state)
    {
        std::string value;

        switch (state)
        {
            case ThreadState::Free:
                value="Free";
                break;
            case ThreadState::Ready:
                value="Ready";
                break;
            case ThreadState::Running:
                value="Running";
                break;
            case ThreadState::Sleeping:
                value="Sleeping";
                break;
            case ThreadState::Dead:
                value="Dead";
                break;
        }
        return value;
    }
};
