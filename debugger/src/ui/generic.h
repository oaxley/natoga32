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
 * @brief	UI Generic Interface
 */
#pragma once

// standard library headers
#include <vector>
#include <memory>
#include <GLFW/glfw3.h>

// program-specific includes
#include "debug_client.h"

//----- Class definition
class IGeneric
{
public:
    IGeneric(DebugClient& client, GLFWwindow* window) :
        client_{client}, childs_{}, window_{window}
    { }

    virtual ~IGeneric() { }

    // render the UI element
    virtual void render() {
        for (auto& child : childs_) {
            child->render();
        }
    }

    // add a new child to this element
    void addChild(std::unique_ptr<IGeneric> child) {
        childs_.push_back(std::move(child));
    }

protected:
    DebugClient& client_;
    std::vector<std::unique_ptr<IGeneric>> childs_;
    GLFWwindow* window_;
};
