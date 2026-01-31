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

// program-specific includes
#include "debug_client.h"

//----- Class definition
class IGeneric
{
public:
    IGeneric(DebugClient& client) :
        client_{client}, childs_{}
    { }

    virtual ~IGeneric() { }

    // render the UI element
    virtual void render() = 0;

    // add a new child to this element
    void addChild(std::unique_ptr<IGeneric> child) {
        childs_.push_back(std::move(child));
    }

protected:
    DebugClient& client_;
    std::vector<std::unique_ptr<IGeneric>> childs_;
};
