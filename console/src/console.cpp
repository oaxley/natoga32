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
 * @brief	Virtual Console - Console Implementation
 */

// standard library headers
#include <fstream>
#include <vector>
#include <cstring>

// program-specific headers
#include "core/memory.h"
#include "core/cpu.h"
#include "core/constants.h"
#include "console.h"


namespace vc
{

//----- Constructor / Destructor

Console::Console() :
    state_(ConsoleState::Uninitialized),
    is_initialized_(false)
{
}

Console::~Console()
{
    shutdown();
}


//----- Lifecycle methods

bool Console::initialize()
{
    if (is_initialized_) {
        return true;
    }

    try {
        // Create memory subsystem
        memory_ = std::make_unique<Memory>(MMAP_RAM_SIZE);

        // Create CPU subsystem
        cpu_ = std::make_unique<CPU>(*memory_);

        // Set state to ready
        state_ = ConsoleState::Ready;
        is_initialized_ = true;

        return true;

    } catch (const std::exception& e) {
        setError(std::string("Initialization failed: ") + e.what());
        return false;
    }
}

void Console::reset()
{
    if (!is_initialized_) {
        return;
    }

    // Reset memory
    memory_->reset();

    // Reset CPU
    cpu_->reset();

    // Clear error state
    clearError();

    // Set state to ready
    state_ = ConsoleState::Ready;
}

void Console::shutdown()
{
    if (!is_initialized_) {
        return;
    }

    // Release CPU
    cpu_.reset();

    // Release Memory
    memory_.reset();

    // Update state
    state_ = ConsoleState::Uninitialized;
    is_initialized_ = false;
}


//----- ROM/Cartridge loading

bool Console::loadROM(const std::string& path)
{
    if (!is_initialized_) {
        setError("Console not initialized");
        return false;
    }

    // Open file
    std::ifstream file(path, std::ios::binary | std::ios::ate);
    if (!file.is_open()) {
        setError("Failed to open ROM file: " + path);
        return false;
    }

    // Get file size
    std::streamsize size = file.tellg();
    file.seekg(0, std::ios::beg);

    // Read file content
    std::vector<u8> buffer(size);
    if (!file.read(reinterpret_cast<char*>(buffer.data()), size)) {
        setError("Failed to read ROM file: " + path);
        return false;
    }

    return loadROM(buffer.data(), buffer.size());
}

bool Console::loadROM(const u8* data, size_t size)
{
    if (!is_initialized_) {
        setError("Console not initialized");
        return false;
    }

    if (!validateROM(data, size)) {
        return false;
    }

    return loadData(data, size, MMAP_ROM_BOOT_LOADER, MMAP_ROM_SIZE);
}

bool Console::loadCartridge(const std::string& path)
{
    // TO BE DONE
    return true;
}

bool Console::loadCartridge(const u8* data, size_t size)
{
    // TO BE DONE
    return true;
}


//----- Execution control

void Console::tick()
{
    if (!is_initialized_) {
        return;
    }

    // Only tick if running
    if (state_ != ConsoleState::Running) {
        return;
    }

    // Execute one CPU cycle
    cpu_->tick();

    // Check if CPU encountered an error
    if (cpu_->getState() == CPUState::Halted) {
        setError("CPU halted due to exception");
        state_ = ConsoleState::Error;
    }
}

void Console::run(u32 cycles)
{
    if (!is_initialized_) {
        return;
    }

    // Only run if in running state
    if (state_ != ConsoleState::Running) {
        return;
    }

    // Execute N cycles
    for (u32 i = 0; i < cycles; ++i) {
        tick();

        // Stop if we hit an error or paused
        if (state_ != ConsoleState::Running) {
            break;
        }
    }
}

void Console::runFrame()
{
    if (!is_initialized_) {
        return;
    }

    // For now, we'll define a frame as a fixed number of cycles
    // Future: This should be based on vsync signal from video subsystem
    // Assuming 60 FPS and a hypothetical CPU speed
    constexpr u32 CYCLES_PER_FRAME = 100000;  // placeholder value

    run(CYCLES_PER_FRAME);
}

void Console::pause()
{
    if (state_ == ConsoleState::Running) {
        state_ = ConsoleState::Paused;
    }
}

void Console::resume()
{
    if (state_ == ConsoleState::Paused) {
        state_ = ConsoleState::Running;
    }
}


//----- State management

u64 Console::getTotalCycles() const
{
    if (!is_initialized_) {
        return 0;
    }

    // Sum all thread cycles
    u64 total = 0;
    for (int i = 0; i < THREADS_COUNT; ++i) {
        total += cpu_->getThread(i).getTotalCycles();
    }

    return total;
}


//----- Subsystem access

CPU& Console::getCPU()
{
    return *cpu_;
}

const CPU& Console::getCPU() const
{
    return *cpu_;
}

Memory& Console::getMemory()
{
    return *memory_;
}

const Memory& Console::getMemory() const
{
    return *memory_;
}


//----- Event system

void Console::triggerEvent(u32 event)
{
    if (!is_initialized_) {
        return;
    }

    // Wake any threads waiting for this event
    cpu_->wakeThreadOnEvent(event);
}


//----- Error handling

void Console::clearError()
{
    last_error_.clear();

    if (state_ == ConsoleState::Error) {
        state_ = ConsoleState::Ready;
    }
}


//----- Private methods

void Console::setError(const std::string& error)
{
    last_error_ = error;
    state_ = ConsoleState::Error;
}

bool Console::validateROM(const u8* data, size_t size)
{
    if (data == nullptr) {
        setError("ROM data is null");
        return false;
    }

    if (size == 0) {
        setError("ROM size is zero");
        return false;
    }

    if (size > MMAP_ROM_SIZE) {
        setError("ROM size exceeds maximum allowed size");
        return false;
    }

    // Future: Add more validation (magic numbers, checksums, etc.)

    return true;
}

bool Console::validateCartridge(const u8* data, size_t size)
{
    if (data == nullptr) {
        setError("Cartridge data is null");
        return false;
    }

    if (size == 0) {
        setError("Cartridge size is zero");
        return false;
    }

    return true;
}

bool Console::loadData(const u8* data, size_t size, u32 base_addr, size_t max_size)
{
    if (size > max_size) {
        setError("Data size exceeds allowed region size");
        return false;
    }

    // Copy data into memory byte by byte
    auto ram = memory_->memview<u8>(base_addr, max_size);
    for (size_t i = 0; i < size; ++i) {
        ram[i] = data[i];
    }

    return true;
}


} // namespace vc
