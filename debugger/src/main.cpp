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
#include <iostream>
#include <vector>
#include <memory>
#include <optional>
#include <argparse/argparse.hpp>

// program-specific includes
#include "config.h"
#include "states.h"
#include "debug_client.h"
#include "ui/helpers.h"
#include "ui/generic.h"
#include "ui/menubar.h"
#include "ui/about.h"


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
        IGeneric(client, window), name_{name}
    {
        addChild(std::make_unique<MyButton>(client, window, "clicker #1"));
        addChild(std::make_unique<MyButton>(client, window, "clicker #2"));
    }

    virtual ~ClickMe()
    { }

    void render()
    {
        ImGui::Begin(name_.c_str());

        IGeneric::render();

        ImGui::End();
    }

private:
    int count_ = 0;
    std::string name_;
};

//----- main entry point
int main(int argc, char* argv[])
{
    //----- Read the command line
    argparse::ArgumentParser program("vcdebug", "0.1.0");

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

    initImGui(window);

    // ImGui widgets list
    std::vector<std::unique_ptr<IGeneric>> widgets;
    widgets.push_back(std::make_unique<ClickMe>(client, window, "Click Window"));
    widgets.push_back(std::make_unique<About>(client, window));

    // state variables
    States states;

    // mainloop
    while (!glfwWindowShouldClose(window))
    {
        // get events
        glfwPollEvents();

        // Start ImGui frame
        newFrame();

        // Global keyboard shortcuts
        ImGuiIO& io = ImGui::GetIO();
        if (io.KeyCtrl && ImGui::IsKeyPressed(ImGuiKey_Q)) {
            glfwSetWindowShouldClose(window, true);
        }

        // show the menubar
        showMenubar(window, &states);

        // about modal window
        if (states.show_about) {
            ImGui::OpenPopup("About##modal");
            states.show_about = false;
        }

        // render all the widgets
        for (auto& widget : widgets) {
            widget->render();
        }

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

