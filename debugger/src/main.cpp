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
#include "debug_client.h"
#include "debug_state.h"

#include "ui/helpers.h"
#include "ui/menubar.h"
#include "ui/about.h"
#include "ui/thread_info.h"
#include "ui/memory_viewer.h"

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

    // debugger main state
    DebugState state(client);

    // initialize video backend
    GLFWwindow* window = nullptr;
    if (!initVideoBackend(window, cfg))
        return EXIT_FAILURE;

    initImGui(window, cfg);

    // UI components list
    std::vector<std::unique_ptr<IGeneric>> components;
    components.push_back(std::make_unique<UIAbout>(state, window));
    components.push_back(std::make_unique<UIThreadInfo>(state, window));
    components.push_back(std::make_unique<UIMemoryView>(state, window));

    // mainloop
    while (!glfwWindowShouldClose(window))
    {
        // get events
        glfwPollEvents();

        // start a new frame
        newFrame();

        // Global keyboard shortcuts
        ImGuiIO& io = ImGui::GetIO();
        if (io.KeyCtrl && ImGui::IsKeyPressed(ImGuiKey_Q)) {
            glfwSetWindowShouldClose(window, true);
        }

        // show the menubar
        showMenubar(state);

        // quit application
        if (state.get(QUIT_HWND)) {
            glfwSetWindowShouldClose(window, true);
        }

        // trigger the About window
        if (state.isVisible(ABOUT_HWND)) {
            ImGui::OpenPopup("About##modal");
            state.hide(ABOUT_HWND);
        }

        // render all the windows
        for (auto& component : components) {
            component->render();
        }

        // GLFW Rendering
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

