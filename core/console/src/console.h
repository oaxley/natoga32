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
 * @brief	Virtual Console - Main Console class
 */
#pragma once

// standard library headers
#include <string>
#include <memory>
#include <functional>

// program-specific headers
#include "vc/types.h"

namespace vc
{
// forward declarations
class Memory;
class CPU;


//----- enums
enum class ConsoleState
{
    Uninitialized,      // Console created but not initialized
    Ready,              // Console initialized and ready
    Running,            // Console is actively running
    Paused,             // Console paused by user
    Error               // Console in error state
};


//----- Console class
class Console
{
public:
    Console();
    ~Console();

    // prevent copy/move
    Console(const Console&) = delete;
    Console& operator=(const Console&) = delete;
    Console(Console&&) = delete;
    Console& operator=(Console&&) = delete;

    //----- Lifecycle methods

    // Initialize the console (allocate memory, create CPU, etc.)
    bool initialize();

    // Reset the console to initial state
    void reset();

    // Shutdown the console and free resources
    void shutdown();

    //----- ROM/Cartridge loading

    // Load ROM data from file
    bool loadROM(const std::string& path);

    // Load ROM data from memory buffer
    bool loadROM(const u8* data, size_t size);

    // Load cartridge from file
    bool loadCartridge(const std::string& path);

    // Load cartridge from memory buffer
    bool loadCartridge(const u8* data, size_t size);

    //----- Execution control

    // Execute one CPU cycle
    void tick();

    // Execute N CPU cycles
    void run(u32 cycles);

    // Run until a frame is complete (future: based on vsync)
    void runFrame();

    // Pause execution
    void pause();

    // Resume execution
    void resume();

    //----- State management

    // Get current console state
    ConsoleState getState() const { return state_; }

    // Check if console is running
    bool isRunning() const { return state_ == ConsoleState::Running; }

    // Check if console is paused
    bool isPaused() const { return state_ == ConsoleState::Paused; }

    // Get total cycles executed
    u64 getTotalCycles() const;

    //----- Subsystem access (for host/debugger)

    // Get CPU instance
    CPU& getCPU();
    const CPU& getCPU() const;

    // Get Memory instance
    Memory& getMemory();
    const Memory& getMemory() const;

    //----- Event system (for future hardware events)

    // Trigger a hardware event (keyboard, gamepad, etc.)
    void triggerEvent(u32 event);

    //----- Error handling

    // Get last error message
    const std::string& getLastError() const { return last_error_; }

    // Clear error state
    void clearError();

private:
    //----- Private members

    std::unique_ptr<Memory> memory_;
    std::unique_ptr<CPU> cpu_;

    ConsoleState state_;
    std::string last_error_;

    bool is_initialized_;

    //----- Private methods

    // Set error state with message
    void setError(const std::string& error);

    // Validate ROM data
    bool validateROM(const u8* data, size_t size);

    // Validate cartridge data
    bool validateCartridge(const u8* data, size_t size);

    // Load data into memory region
    bool loadData(const u8* data, size_t size, u32 base_addr, size_t max_size);
};


} // namespace vc
