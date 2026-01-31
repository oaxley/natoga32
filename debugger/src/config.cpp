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
 * @brief	Configuration
 */
// standard library headers
#include <filesystem>
#include <iostream>
#include <fstream>
#include <toml++/toml.hpp>

// program-specific includes
#include "config.h"

namespace Config
{
namespace fs = std::filesystem;

// constants
constexpr const char* CONFIG_FILENAME = "vcdebug.toml";
constexpr const char* CONFIG_DIR = "vcdebug";

// default initialization of the configuration
Config initializeConfig()
{
    return Config{
        .connection = { .host = "localhost", .port = 2600}
    };
}

// parse the configuration from a file
Config parseConfig(const fs::path& path)
{
    auto data = toml::parse_file(path.string());

    Config cfg;
    cfg.connection.host = data["connection"]["host"].value_or(std::string("localhost"));
    cfg.connection.port = data["connection"]["port"].value_or(2600);

    return cfg;
}

// retrieve the value of XDF_CONFIG_HOME
fs::path getConfigHome()
{
    // look for XDG_CONFIG_HOME
    if (const char* xdg = std::getenv("XDG_CONFIG_HOME"); xdg && *xdg) {
        return fs::path(xdg);
    }

    // build it from HOME/.config
    if (const char* home = std::getenv("HOME"); home && *home) {
        return fs::path(home) / ".config";
    }

    // nothing found
    return {};
}

// save the configuration
bool saveConfig(const fs::path& path, const Config& config)
{
    fs::create_directories(path.parent_path());

    toml::table cfg;

    // connection parameters
    cfg.insert("connection", toml::table{
        {"host", config.connection.host},
        {"port", config.connection.port}
    });

    // save the file
    std::ofstream file(path);
    if (!file)
        return false;

    file << cfg;
    return true;
}

// load the configuration
Result<Config> loadConfig(std::optional<std::string> user_path)
{
    // user specified path
    if (user_path.has_value()) {
        fs::path path(*user_path);
        if (!fs::exists(path)) {
            return {
                .value = std::nullopt, .error = "Config file not found: " + path.string()
            };
        }

        return {
            .value = parseConfig(path), .error = { }
        };
    }

    // local directory
    fs::path local_path = fs::current_path() / CONFIG_FILENAME;
    if (fs::exists(local_path)) {
        return {
            .value = parseConfig(local_path), .error = { }
        };
    }

    // check XDG_CONFIG_HOME
    fs::path config_home = getConfigHome();
    if (!config_home.empty()) {
        fs::path xdg_path = config_home / CONFIG_DIR / CONFIG_FILENAME;
        if (fs::exists(xdg_path)) {
            return {
                .value = parseConfig(xdg_path), .error = { }
            };
        }
    }

    // nothing found - first launch -> create the config
    Config cfg = initializeConfig();

    if (!config_home.empty()) {
        fs::path save_path = config_home / CONFIG_DIR / CONFIG_FILENAME;
        if (!saveConfig(save_path, cfg)) {
            std::cerr << "Warning: could not save configuration to " << save_path << "\n";
        }
    }

    // return the configuration
    return {
        .value = cfg, .error = { }
    };
}



} // namespace Config
