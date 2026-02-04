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
 * @brief	Debugger Client
 */
// standard library headers
#include <array>
#include <iostream>
#include <vector>
#include <memory>
#include <optional>
#include <argparse/argparse.hpp>

// program-specific includes
#include "constants.h"
#include "config.h"
#include "states.h"
#include "debug_client.h"
#include "ui/helpers.h"
#include "ui/generic.h"
#include "ui/menubar.h"
#include "ui/about.h"
#include "ui/registers.h"


class MyButton : public IGeneric
{
public:
    MyButton(DebugClient& client, GLFWwindow* window, std::string label) :
        IGeneric(client, window), label_{label}
    { }

    void render()
    {
        if (ImGui::Button(label_.c_str())) {
            count_++;
        }
        ImGui::SameLine();
        ImGui::Text("Count: %d", count_);
    }

private:
    int count_ = 0;
    std::string label_;

};

class ClickMe : public IGeneric
{
public:
    ClickMe(DebugClient& client, GLFWwindow* window, std::string name) :
        IGeneric(client, window, true), name_{name}
    {
        std::array<uint32_t, 32> regs = {0};
        regs[12] = 0xDEADC0DE;
        regs[20] = 0xCAFEBABE;

        addChild(std::make_unique<Registers>(client, window, regs));
        // addChild(std::make_unique<MyButton>(client, window, "clicker #1"));
        // addChild(std::make_unique<MyButton>(client, window, "clicker #2"));
    }

    virtual ~ClickMe()
    { }

    void render()
    {
        if (!isVisible()) return;

        ImGui::Begin(name_.c_str());

        IGeneric::render();

        ImGui::End();
    }

private:
    int count_ = 0;
    std::string name_;
};

class FrameBuffer : public IGeneric
{
public:
    FrameBuffer(DebugClient& client, GLFWwindow* window, std::string name) :
        IGeneric(client, window), name_{name}
    {}

    void render()
    {
        if (!isVisible()) return;

        ImGui::PushStyleColor(ImGuiCol_WindowBg, IM_COL32(128, 128, 128, 255));

        // Set fixed size before Begin
        ImGui::SetNextWindowSize(ImVec2(640, 384), ImGuiCond_Always);
        ImGui::SetNextWindowSizeConstraints(ImVec2(640, 384), ImVec2(640, 384));  // Prevent resizing

        ImGuiWindowFlags flags = ImGuiWindowFlags_NoResize | ImGuiWindowFlags_NoScrollbar;

        if (ImGui::Begin(name_.c_str(), &visible_, flags)) {
            // If you have a texture ID from OpenGL/Vulkan
            // ImGui::Image((ImTextureID)(intptr_t)texture_id, ImVec2(640, 384));
            ImGui::Dummy(ImVec2(640, 384));  // Reserve space
        }
        ImGui::End();
        ImGui::PopStyleColor();
    }
private:
    std::string name_;
};

//----- main entry point
int main(int argc, char* argv[])
{
    //----- Read the command line
    argparse::ArgumentParser program("vcdebug", VERSION);

    // configuration file (optional)
    program.add_argument("-c", "--config")
        .help("Path to configuration file")
        .default_value(std::string(""));

    // read the command line arguments
    try {
        program.parse_args(argc, argv);
    } catch (const std::exception& err) {
        std::cerr << err.what() << "\n";
        std::cerr << program;
        return EXIT_FAILURE;
    }

    // retrieve the parameters if any
    auto config_path = program.get<std::string>("--config");


    // convert the "user_config" to optional for loadConfig
    std::optional<std::string> user_config;
    if (!config_path.empty()) {
        user_config = config_path;
    }

    // load the configuration
    auto result = Config::loadConfig(user_config);
    if (!result) {
        std::cerr << "Error: " << result.error << "\n";
        return EXIT_FAILURE;
    }
    Config::Config cfg = *result;


    // connect to the host server
    DebugClient client(cfg.connection.host, cfg.connection.port);

    // disable for now
    // if (!client.connect()) {
    //     return EXIT_FAILURE;
    // }

    // initialize video backend
    GLFWwindow* window = nullptr;
    if (!initVideoBackend(window, cfg))
        return EXIT_FAILURE;

    initImGui(window, cfg);

    // new window
    ClickMe clickme(client, window, "Click Me");
    FrameBuffer fb(client, window, "FrameBuffer");

    // widgets are by default visible
    IGeneric widgets(client, window, true);
    widgets.addChild(clickme);
    widgets.addChild(std::make_unique<About>(client, window));
    widgets.addChild(fb);

    // mainloop
    States state;
    while (!glfwWindowShouldClose(window))
    {
        // get events
        glfwPollEvents();

        // reset the state and start a new frame
        state = States::NoState;
        newFrame();

        // Global keyboard shortcuts
        ImGuiIO& io = ImGui::GetIO();
        if (io.KeyCtrl && ImGui::IsKeyPressed(ImGuiKey_Q)) {
            glfwSetWindowShouldClose(window, true);
        }

        // show the menubar
        showMenubar(state);

        switch (state)
        {
            case States::About:
                ImGui::OpenPopup("About##modal");
                break;

            case States::LoadSymbols:
                clickme.toggleVisible();
                break;

            case States::FrameBuffer:
                fb.toggleVisible();
                break;

            case States::Quit:
                glfwSetWindowShouldClose(window, true);
                break;

            default:
                break;
        }

        widgets.render();

        // Render
        render(window);
    }

    //----- save the configuration

    // retrieve the window dimensions
    glfwGetWindowPos(window, &cfg.window.x, &cfg.window.y);
    glfwGetWindowSize(window, &cfg.window.width, &cfg.window.height);

    Config::saveConfig(cfg, user_config);

    // clean ImGui + video backend
    cleanImGui();
    cleanVideoBackend(window);

    return EXIT_SUCCESS;
}

