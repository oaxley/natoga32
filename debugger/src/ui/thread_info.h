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

// local constants
constexpr int space_size = 10;


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
            // compute the column values
            computeColumns();

            // retrieve the thread registers values
            ThreadInfo info = state_.getThreadInfo(thread_id_);

            // Thread selector
            ImGui::Text("Active Thread:");
            ImGui::SameLine();
            ImGui::SetNextItemWidth(80.0f);
            ImGui::InputInt("##l44", &thread_id_);

            if (thread_id_ < 0) thread_id_ = 0;
            if (thread_id_ >= ThreadCount) thread_id_ = ThreadCount - 1;

            // print thread info
            printInfo(info);

            // registers table
            printRegisters(info);

        }
        ImGui::End();
    }

private:
    //----- members
    int thread_id_;
    int first_column_ = 80;            // first column adjustment
    int second_column_ = 0;            // second column adjustment
    bool init_ = false;

    //----- methods
    std::string stateToText(ThreadState state)
    {
        std::string value;

        switch (state)
        {
            case ThreadState::Free:
                value = "Free";
                break;
            case ThreadState::Ready:
                value = "Ready";
                break;
            case ThreadState::Running:
                value = "Running";
                break;
            case ThreadState::Sleeping:
                value = "Sleeping";
                break;
            case ThreadState::Dead:
                value = "Dead";
                break;
        }
        return value;
    }

    void computeColumns()
    {
        if (init_)
            return;

        // first column
        {
            int max_size = ImGui::CalcTextSize("Waitkey:").x;
            if ((first_column_ - space_size - max_size) < 0)
            first_column_ += (first_column_ - max_size) + space_size;
        }

        // second column is computed from the first one
        {
            int max_size = ImGui::CalcTextSize("Sleep until:").x;
            int min_size = first_column_ + ImGui::CalcTextSize("00000000").x;
            second_column_ = min_size + space_size + max_size + space_size;
        }

        // initialization is done
        init_ = true;
    }

    void printInfo(ThreadInfo& info)
    {
        int pos;

        // Thread status
        ImGui::SameLine(second_column_);
        ImGui::Text("%s", stateToText(info.state).c_str());

        // first line
        pos = first_column_ - space_size - ImGui::CalcTextSize("PC:").x;
        ImGui::SetCursorPosX(pos);
        ImGui::Text("PC:");
        ImGui::SameLine(first_column_);
        ImGui::Text("%08X", info.pc);

        pos = second_column_ - space_size - ImGui::CalcTextSize("Cycles:").x;
        ImGui::SameLine(pos);
        ImGui::Text("Cycles:");
        ImGui::SameLine(second_column_);
        ImGui::Text("%ld", info.cycles);

        // second line
        pos = first_column_ - space_size - ImGui::CalcTextSize("Waitkey:").x;
        ImGui::SetCursorPosX(pos);
        ImGui::Text("Waitkey:");
        ImGui::SameLine(first_column_);
        ImGui::Text("%08X", info.waitkey);

        pos = second_column_ - space_size - ImGui::CalcTextSize("Sleep until:").x;
        ImGui::SameLine(pos);
        ImGui::Text("Sleep until:");
        ImGui::SameLine(second_column_);
        ImGui::Text("%ld", info.sleep_until);
    }

    void printRegisters(ThreadInfo& info)
    {
        int width_value = ImGui::CalcTextSize("00000000 ").x;
        int width_text = ImGui::CalcTextSize("x31").x;

        ImGui::SeparatorText("Registers");
        ImGui::SetCursorPosX(40);
        ImGui::BeginTable("Registers", 4, ImGuiTableFlags_SizingFixedFit);

        ImGui::TableSetupColumn("##name1", ImGuiTableColumnFlags_WidthFixed, width_text);
        ImGui::TableSetupColumn("##val1", ImGuiTableColumnFlags_WidthFixed, width_value);
        ImGui::TableSetupColumn("##name2", ImGuiTableColumnFlags_WidthFixed, width_text);
        ImGui::TableSetupColumn("##val2", ImGuiTableColumnFlags_WidthFixed, width_value);

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
};
