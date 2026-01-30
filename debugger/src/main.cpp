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
#include <argparse/argparse.hpp>

// program-specific includes
#include "debug_client.h"
#include "ui/helpers.h"
#include "ui/generic.h"

class ClickMe : public IGeneric
{
public:
    ClickMe(DebugClient& client, std::string name) :
        IGeneric(client), name_{name}
    { }

    virtual ~ClickMe()
    { }

    void render()
    {
        ImGui::Begin(name_.c_str());

        if (ImGui::Button("Click Me!")) {
            count_++;
        }
        ImGui::SameLine();
        ImGui::Text("Count: %d", count_);

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

    // host
    program.add_argument("host")
        .help("IP/host to connect to");

    // port
    program.add_argument("port")
        .scan<'i', int>()
        .default_value(2600)
        .help("Host listening port");

    // read the command line arguments
    try {
        program.parse_args(argc, argv);
    } catch (const std::exception& err) {
        std::cerr << err.what() << "\n";
        std::cerr << program;
        return EXIT_FAILURE;
    }

    // retrieve the parameters and connect to host
    auto host = program.get<std::string>("host");
    auto port = program.get<int>("port");

    DebugClient client(host, port);

    if (client.connect()) {
        std::cout << "Debugger connected to host at " << host << ":" << port << "\n";
    }

    // initialize video backend
    GLFWwindow* window = nullptr;
    if (!initVideoBackend(window))
        return EXIT_FAILURE;

    initImGui(window);

    // ImGui widgets list
    std::vector<std::unique_ptr<IGeneric>> widgets;
    widgets.push_back(std::make_unique<ClickMe>(client, "Button #1"));
    widgets.push_back(std::make_unique<ClickMe>(client, "Button #2"));

    // Application state
    // bool show_demo = false;
    // float slider_value = 0.5f;
    // int counter = 0;
    // float color[3] = {0.4f, 0.7f, 0.2f};


    // mainloop
    while (!glfwWindowShouldClose(window))
    {
        // get events
        glfwPollEvents();

        // Start ImGui frame
        newFrame();

        // render all the widgets
        for (auto& widget : widgets) {
            widget->render();
        }

        // ImGui::Text("Welcome to Dear ImGui!");
        // ImGui::Spacing();

        // if (ImGui::Button("Click me!")) {
        //     counter++;
        // }
        // ImGui::SameLine();
        // ImGui::Text("Count: %d", counter);

        // ImGui::SliderFloat("Slider", &slider_value, 0.0f, 1.0f);
        // ImGui::ColorEdit3("Color", color);

        // ImGui::Checkbox("Show Demo Window", &show_demo);

        // ImGui::Text("FPS: %.1f", io.Framerate);


        // Render
        render(window);
    }

    // clean ImGui + video backend
    cleanImGui();
    cleanVideoBackend(window);

    return EXIT_SUCCESS;
}

