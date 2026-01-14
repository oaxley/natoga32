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
struct VC_Console_t {
    vc::Console* ptr;
};


//----- Console lifecycle functions

VC_Console_t* vc_create(void)
{
    try {
        VC_Console_t* wrapper = new VC_Console_t;
        wrapper->ptr = new vc::Console();
        return wrapper;
    } catch (...) {
        return nullptr;
    }
}

void vc_destroy(VC_Console_t* console)
{
    if (console) {
        delete console->ptr;
        delete console;
    }
}

bool vc_initialize(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return false;
    }
    return console->ptr->initialize();
}

void vc_reset(VC_Console_t* console)
{
    if (console && console->ptr) {
        console->ptr->reset();
    }
}

void vc_shutdown(VC_Console_t* console)
{
    if (console && console->ptr) {
        console->ptr->shutdown();
    }
}


//----- ROM loading functions

bool vc_loadRomFile(VC_Console_t* console, const char* path)
{
    if (!console || !console->ptr || !path) {
        return false;
    }
    return console->ptr->loadROM(std::string(path));
}

bool vc_loadRomData(VC_Console_t* console, const uint8_t* data, size_t size)
{
    if (!console || !console->ptr || !data) {
        return false;
    }
    return console->ptr->loadROM(data, size);
}


//----- Execution control functions

void vc_tick(VC_Console_t* console)
{
    if (console && console->ptr) {
        console->ptr->tick();
    }
}

void vc_run(VC_Console_t* console, uint32_t cycles)
{
    if (console && console->ptr) {
        console->ptr->run(cycles);
    }
}

void vc_runFrame(VC_Console_t* console)
{
    if (console && console->ptr) {
        console->ptr->runFrame();
    }
}

void vc_pause(VC_Console_t* console)
{
    if (console && console->ptr) {
        console->ptr->pause();
    }
}

void vc_resume(VC_Console_t* console)
{
    if (console && console->ptr) {
        console->ptr->resume();
    }
}


//----- State query functions

VC_ConsoleState_t vc_getState(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return VC_STATE_UNINITIALIZED;
    }

    auto state = console->ptr->getState();
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

bool vc_isRunning(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return false;
    }
    return console->ptr->isRunning();
}

bool vc_isPaused(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return false;
    }
    return console->ptr->isPaused();
}

uint64_t vc_getTotalCycles(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return 0;
    }
    return console->ptr->getTotalCycles();
}

VC_CPUState_t vc_getCpuState(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return VC_CPU_HALTED;
    }

    auto state = console->ptr->getCPU().getState();
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

int vc_getCurrentThreadId(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return -1;
    }
    return console->ptr->getCPU().getCurrentThreadID();
}


//----- Event system

void vc_triggerEvent(VC_Console_t* console, uint32_t event)
{
    if (console && console->ptr) {
        console->ptr->triggerEvent(event);
    }
}


//----- Error handling

const char* vc_getLastError(VC_Console_t* console)
{
    if (!console || !console->ptr) {
        return "Invalid console instance";
    }
    return console->ptr->getLastError().c_str();
}

void vc_clearError(VC_Console_t* console)
{
    if (console && console->ptr) {
        console->ptr->clearError();
    }
}


//----- Memory access functions

uint8_t vc_memRead8(VC_Console_t* console, uint32_t addr)
{
    if (!console || !console->ptr) {
        return 0;
    }
    return console->ptr->getMemory().read8(addr);
}

uint16_t vc_memRead16(VC_Console_t* console, uint32_t addr)
{
    if (!console || !console->ptr) {
        return 0;
    }
    return console->ptr->getMemory().read16(addr);
}

uint32_t vc_memRead32(VC_Console_t* console, uint32_t addr)
{
    if (!console || !console->ptr) {
        return 0;
    }
    return console->ptr->getMemory().read32(addr);
}

void vc_memWrite8(VC_Console_t* console, uint32_t addr, uint8_t value)
{
    if (console && console->ptr) {
        console->ptr->getMemory().write8(addr, value);
    }
}

void vc_memWrite16(VC_Console_t* console, uint32_t addr, uint16_t value)
{
    if (console && console->ptr) {
        console->ptr->getMemory().write16(addr, value);
    }
}

void vc_memWrite32(VC_Console_t* console, uint32_t addr, uint32_t value)
{
    if (console && console->ptr) {
        console->ptr->getMemory().write32(addr, value);
    }
}


//----- CPU/Thread debugging functions

VC_ThreadState_t vc_getThreadState(VC_Console_t* console, int thread_id)
{
    if (!console || !console->ptr || thread_id < 0 || thread_id >= 8) {
        return VC_THREAD_FREE;
    }

    auto& thread = console->ptr->getCPU().getThread(thread_id);
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

uint32_t vc_getThreadPc(VC_Console_t* console, int thread_id)
{
    if (!console || !console->ptr || thread_id < 0 || thread_id >= 8) {
        return 0;
    }

    auto& thread = console->ptr->getCPU().getThread(thread_id);
    return thread.getPC();
}

uint32_t vc_getThreadRegister(VC_Console_t* console, int thread_id, int reg_num)
{
    if (!console || !console->ptr || thread_id < 0 || thread_id >= 8) {
        return 0;
    }

    if (reg_num < 0 || reg_num >= 32) {
        return 0;
    }

    auto& thread = console->ptr->getCPU().getThread(thread_id);
    return thread.getRegisters()[reg_num];
}
