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

// program-specific includes
#include "debug_client.h"

//----- Class definition
class IGeneric
{
public:
    IGeneric(DebugClient& client) :
        client_{client}
    { }

    virtual ~IGeneric() { }

    // render the UI element
    virtual void render() = 0;

protected:
    DebugClient& client_;
};
