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
 * @brief	Virtual Console - C API Implementation
 */

// C API header
#include "vc/vc.h"

// C++ headers
#include "core/cpu.h"
#include "core/cpu_thread.h"
#include "core/memory.h"
#include "console.h"

// C++ to C bridge structure
struct vc_console_t {
    vc::Console* console;
};


//----- Console lifecycle functions

vc_console_t* vc_console_create(void)
{
    try {
        vc_console_t* wrapper = new vc_console_t;
        wrapper->console = new vc::Console();
        return wrapper;
    } catch (...) {
        return nullptr;
    }
}

void vc_console_destroy(vc_console_t* console)
{
    if (console) {
        delete console->console;
        delete console;
    }
}

bool vc_console_initialize(vc_console_t* console)
{
    if (!console || !console->console) {
        return false;
    }
    return console->console->initialize();
}

void vc_console_reset(vc_console_t* console)
{
    if (console && console->console) {
        console->console->reset();
    }
}

void vc_console_shutdown(vc_console_t* console)
{
    if (console && console->console) {
        console->console->shutdown();
    }
}


//----- ROM loading functions

bool vc_console_load_rom_file(vc_console_t* console, const char* path)
{
    if (!console || !console->console || !path) {
        return false;
    }
    return console->console->loadROM(std::string(path));
}

bool vc_console_load_rom_data(vc_console_t* console, const uint8_t* data, size_t size)
{
    if (!console || !console->console || !data) {
        return false;
    }
    return console->console->loadROM(data, size);
}


//----- Execution control functions

void vc_console_tick(vc_console_t* console)
{
    if (console && console->console) {
        console->console->tick();
    }
}

void vc_console_run(vc_console_t* console, uint32_t cycles)
{
    if (console && console->console) {
        console->console->run(cycles);
    }
}

void vc_console_run_frame(vc_console_t* console)
{
    if (console && console->console) {
        console->console->runFrame();
    }
}

void vc_console_pause(vc_console_t* console)
{
    if (console && console->console) {
        console->console->pause();
    }
}

void vc_console_resume(vc_console_t* console)
{
    if (console && console->console) {
        console->console->resume();
    }
}


//----- State query functions

vc_console_state_t vc_console_get_state(vc_console_t* console)
{
    if (!console || !console->console) {
        return VC_STATE_UNINITIALIZED;
    }

    auto state = console->console->getState();
    switch (state) {
        case vc::ConsoleState::Uninitialized:
            return VC_STATE_UNINITIALIZED;
        case vc::ConsoleState::Ready:
            return VC_STATE_READY;
        case vc::ConsoleState::Running:
            return VC_STATE_RUNNING;
        case vc::ConsoleState::Paused:
            return VC_STATE_PAUSED;
        case vc::ConsoleState::Error:
            return VC_STATE_ERROR;
        default:
            return VC_STATE_UNINITIALIZED;
    }
}

bool vc_console_is_running(vc_console_t* console)
{
    if (!console || !console->console) {
        return false;
    }
    return console->console->isRunning();
}

bool vc_console_is_paused(vc_console_t* console)
{
    if (!console || !console->console) {
        return false;
    }
    return console->console->isPaused();
}

uint64_t vc_console_get_total_cycles(vc_console_t* console)
{
    if (!console || !console->console) {
        return 0;
    }
    return console->console->getTotalCycles();
}

vc_cpu_state_t vc_console_get_cpu_state(vc_console_t* console)
{
    if (!console || !console->console) {
        return VC_CPU_HALTED;
    }

    auto state = console->console->getCPU().getState();
    switch (state) {
        case vc::CPUState::Running:
            return VC_CPU_RUNNING;
        case vc::CPUState::Idle:
            return VC_CPU_IDLE;
        case vc::CPUState::Halted:
            return VC_CPU_HALTED;
        default:
            return VC_CPU_HALTED;
    }
}

int vc_console_get_current_thread_id(vc_console_t* console)
{
    if (!console || !console->console) {
        return -1;
    }
    return console->console->getCPU().getCurrentThreadID();
}


//----- Event system

void vc_console_trigger_event(vc_console_t* console, uint32_t event)
{
    if (console && console->console) {
        console->console->triggerEvent(event);
    }
}


//----- Error handling

const char* vc_console_get_last_error(vc_console_t* console)
{
    if (!console || !console->console) {
        return "Invalid console instance";
    }
    return console->console->getLastError().c_str();
}

void vc_console_clear_error(vc_console_t* console)
{
    if (console && console->console) {
        console->console->clearError();
    }
}


//----- Memory access functions

uint8_t vc_console_mem_read8(vc_console_t* console, uint32_t addr)
{
    if (!console || !console->console) {
        return 0;
    }
    return console->console->getMemory().read8(addr);
}

uint16_t vc_console_mem_read16(vc_console_t* console, uint32_t addr)
{
    if (!console || !console->console) {
        return 0;
    }
    return console->console->getMemory().read16(addr);
}

uint32_t vc_console_mem_read32(vc_console_t* console, uint32_t addr)
{
    if (!console || !console->console) {
        return 0;
    }
    return console->console->getMemory().read32(addr);
}

void vc_console_mem_write8(vc_console_t* console, uint32_t addr, uint8_t value)
{
    if (console && console->console) {
        console->console->getMemory().write8(addr, value);
    }
}

void vc_console_mem_write16(vc_console_t* console, uint32_t addr, uint16_t value)
{
    if (console && console->console) {
        console->console->getMemory().write16(addr, value);
    }
}

void vc_console_mem_write32(vc_console_t* console, uint32_t addr, uint32_t value)
{
    if (console && console->console) {
        console->console->getMemory().write32(addr, value);
    }
}


//----- CPU/Thread debugging functions

vc_thread_state_t vc_console_get_thread_state(vc_console_t* console, int thread_id)
{
    if (!console || !console->console || thread_id < 0 || thread_id >= 8) {
        return VC_THREAD_FREE;
    }

    auto& thread = console->console->getCPU().getThread(thread_id);
    auto state = thread.getState();

    switch (state) {
        case vc::ThreadState::Free:
            return VC_THREAD_FREE;
        case vc::ThreadState::Ready:
            return VC_THREAD_READY;
        case vc::ThreadState::Running:
            return VC_THREAD_RUNNING;
        case vc::ThreadState::Sleeping:
            return VC_THREAD_SLEEPING;
        case vc::ThreadState::Dead:
            return VC_THREAD_DEAD;
        default:
            return VC_THREAD_FREE;
    }
}

uint32_t vc_console_get_thread_pc(vc_console_t* console, int thread_id)
{
    if (!console || !console->console || thread_id < 0 || thread_id >= 8) {
        return 0;
    }

    auto& thread = console->console->getCPU().getThread(thread_id);
    return thread.getPC();
}

uint32_t vc_console_get_thread_register(vc_console_t* console, int thread_id, int reg_num)
{
    if (!console || !console->console || thread_id < 0 || thread_id >= 8) {
        return 0;
    }

    if (reg_num < 0 || reg_num >= 32) {
        return 0;
    }

    auto& thread = console->console->getCPU().getThread(thread_id);
    return thread.getRegisters()[reg_num];
}
