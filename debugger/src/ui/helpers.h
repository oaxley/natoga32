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
 * @brief	GLFW/ImGui helper functions
 */
#pragma once

// standard library headers
#include <iostream>
#include <GLFW/glfw3.h>
#include <imgui.h>
#include <imgui_impl_glfw.h>
#include <imgui_impl_opengl3.h>

// program-specific includes
#include "config.h"


//----- functions
bool initVideoBackend(GLFWwindow*& window, const Config::Config& cfg)
{
    // initialize GLFW
    if (!glfwInit()) {
        std::cerr << "Error: failed to initialize GLFW\n";
        window = nullptr;
        return false;
    }

    // OpenGL 3.3 Core
    glfwWindowHint(GLFW_CONTEXT_VERSION_MAJOR, 3);
    glfwWindowHint(GLFW_CONTEXT_VERSION_MINOR, 3);
    glfwWindowHint(GLFW_OPENGL_PROFILE, GLFW_OPENGL_CORE_PROFILE);

    // window is hidden first so we can position it
    glfwWindowHint(GLFW_VISIBLE, GLFW_FALSE);

    // create window
    window = glfwCreateWindow(cfg.window.width, cfg.window.height, "vcdebug", nullptr, nullptr);
    if (!window) {
        std::cerr << "Error: failed to create window\n";
        glfwTerminate();
        return false;
    }

    // position the window
    glfwSetWindowPos(window, cfg.window.x, cfg.window.y);

    // make this window the current context
    glfwMakeContextCurrent(window);

    // activate VSYNC every frame
    glfwSwapInterval(1);

    // show the window
    glfwShowWindow(window);

    return true;
}

void cleanVideoBackend(GLFWwindow* window)
{
    if (window)
        glfwDestroyWindow(window);

    glfwTerminate();
}

void initImGui(GLFWwindow* window, const Config::Config& cfg)
{
    if (!window)
        return;

    // setup Dear ImGui
    IMGUI_CHECKVERSION();
    ImGui::CreateContext();

    ImGuiIO& io = ImGui::GetIO();
    io.ConfigFlags |= ImGuiConfigFlags_NavEnableKeyboard;

    if (!cfg.font.path.empty()) {
        io.Fonts->AddFontFromFileTTF(cfg.font.path.c_str(), cfg.font.size);
    }

    ImGui::StyleColorsDark();

    // Platform/Renderer backends
    ImGui_ImplGlfw_InitForOpenGL(window, true);
    ImGui_ImplOpenGL3_Init("#version 330");
}

void cleanImGui()
{
    ImGui_ImplOpenGL3_Shutdown();
    ImGui_ImplGlfw_Shutdown();
    ImGui::DestroyContext();
}

void newFrame()
{
    ImGui_ImplOpenGL3_NewFrame();
    ImGui_ImplGlfw_NewFrame();
    ImGui::NewFrame();
}

void render(GLFWwindow* window)
{
    // ImGui rendering
    ImGui::Render();

    // retrieve the framebuffer dimensions
    int display_w, display_h;
    glfwGetFramebufferSize(window, &display_w, &display_h);

    // clear the screen with RGB(0,0,0) - black
    glViewport(0, 0, display_w, display_h);
    glClearColor(0, 0, 0, 1.0f);
    glClear(GL_COLOR_BUFFER_BIT);

    // backend rendering
    ImGui_ImplOpenGL3_RenderDrawData(ImGui::GetDrawData());

    // swap buffers
    glfwSwapBuffers(window);
}
