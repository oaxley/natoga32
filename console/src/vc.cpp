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
 * @brief	Virtual Console - main class implementation
 */

// program-specific includes
#include "vc.h"
#include "core/cpu.h"
#include "core/memory.h"

namespace vc
{

// constructor
Console::Console()
{
    memory_ = std::make_unique<Memory>();
    cpu_ = std::make_unique<CPU>(memory_.get());
}

// destructor
/* virtual */ Console::~Console() = default;

// console reset
void Console::reset()
{
    memory_->reset();
    cpu_->reset();

    total_cycles_ = 0;
    state_ = ConsoleState::Stopped;

    exception_id_ = 0;
    exception_pc_ = 0;
}

// execute one instruction
void Console::tick()
{
    // nothing to be done if console is not running
    if (state_ != ConsoleState::Running) {
        return;
    }

    // execute one instruction
    cpu_->tick();
    total_cycles_++;

    // check CPU state after execution
    CPUState cpu_state = cpu_->getState();
    switch (cpu_state)
    {
        case CPUState::Running:
            state_ = ConsoleState::Running;
            break;

        case CPUState::Idle:
            state_ = ConsoleState::Idle;
            break;

        case CPUState::Halted:
            state_ = ConsoleState::Halted;
            auto exception = cpu_->getLastException();
            exception_id_ = std::get<0>(exception);
            exception_pc_ = std::get<1>(exception);
            break;
    }
}

// run some cycles
void Console::runCycles(u64 cycles)
{
    for (u64 i = 0; i < cycles && state_ == ConsoleState::Running; ++i) {
        tick();
    }
}

// accessors
std::tuple<u32, u32> Console::getLastException() const
{
    return std::make_tuple(exception_id_, exception_pc_);
}

} // namespace vc


