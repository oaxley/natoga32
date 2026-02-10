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

        if (ImGui::Begin("ThreadInfo")) {
            ImGui::Text("Active Thread:");
            ImGui::SameLine();
            ImGui::SetNextItemWidth(80.0f);
            ImGui::InputInt("##l44", &thread_id_);

            if (thread_id_ < 0) thread_id_ = 0;
            if (thread_id_ >= ThreadCount) thread_id_ = ThreadCount - 1;
        }
        ImGui::End();
    }

private:
    int thread_id_;
};
