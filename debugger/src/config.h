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
 * @brief	Configuration structure
 */
#pragma once

// standard library headers
#include <optional>
#include <cstdlib>

namespace Config
{

//----- structs definition
struct ConnectionParams
{
    std::string host;
    int port;
};

struct WindowParams
{
    int x;
    int y;
    int width;
    int height;
};

// configuration main object
struct Config
{
    WindowParams window;
    ConnectionParams connection;
};

//----- templates
// simple result structure for configuration
template<typename T>
struct Result
{
    std::optional<T> value;
    std::string error;

    bool ok() const { return value.has_value(); }
    explicit operator bool() const { return ok(); }
    T& operator*() { return *value; }
    const T& operator*() const { return *value; }
};

//----- forward definitions

// load the configuration
Result<Config> loadConfig(std::optional<std::string> user_path = std::nullopt);

// save the configuration
bool saveConfig(const Config& config, std::optional<std::string> user_path = std::nullopt);

} // namespace Config


