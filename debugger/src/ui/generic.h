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
    IGeneric(DebugClient& client, GLFWwindow* window, bool visible = false) :
        client_{client}, window_{window}, visible_{visible}, children_{}, owned_{}
    { }

    virtual ~IGeneric() { }

    // render the UI element
    virtual void render() {
        if (!isVisible()) return;

        for (auto* child : children_) {
            child->render();
        }
    }

    // add a new child and own it
    void addChild(std::unique_ptr<IGeneric> child) {
        children_.push_back(child.get());
        owned_.push_back(std::move(child));
    }

    // add a new child (don't own it)
    void addChild(IGeneric& child) {
        children_.push_back(&child);
    }

    // visibility flag accessors
    bool isVisible() const { return visible_; }
    void show() { visible_ = true; }
    void hide() { visible_ = false; }
    void toggleVisible() { visible_ = !visible_; }

protected:
    DebugClient& client_;
    GLFWwindow* window_;
    bool visible_ = false;

    // all children + owned
    std::vector<IGeneric*> children_;
    std::vector<std::unique_ptr<IGeneric>> owned_;
};
