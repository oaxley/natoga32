/* -*- coding: utf-8 -*-
 * vim: filetype=c
 *
 * This source file is subject to the MIT License
 * that is bundled with this package in the file LICENSE.txt.
 * It is also available through the Internet at this address:
 * https://opensource.org/license/mit
 *
 * @author	Sebastien LEGRAND
 * @license	MIT License
 *
 * @brief	Virtual Console - C API Header
 */
#pragma once

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>
#include <stddef.h>

//----- Opaque handle for console instance
typedef struct vc_console_t vc_console_t;

//----- Console states
typedef enum {
    VC_STATE_UNINITIALIZED = 0,
    VC_STATE_READY = 1,
    VC_STATE_RUNNING = 2,
    VC_STATE_PAUSED = 3,
    VC_STATE_ERROR = 4
} vc_console_state_t;

//----- CPU states
typedef enum {
    VC_CPU_RUNNING = 0,
    VC_CPU_IDLE = 1,
    VC_CPU_HALTED = 2
} vc_cpu_state_t;

//----- Thread states
typedef enum {
    VC_THREAD_FREE = 0,
    VC_THREAD_READY = 1,
    VC_THREAD_RUNNING = 2,
    VC_THREAD_SLEEPING = 4,
    VC_THREAD_DEAD = 8
} vc_thread_state_t;


//----- Console lifecycle functions

/**
 * Create a new console instance
 * @return Pointer to console instance, or NULL on failure
 */
vc_console_t* vc_console_create(void);

/**
 * Destroy a console instance and free all resources
 * @param console Console instance to destroy
 */
void vc_console_destroy(vc_console_t* console);

/**
 * Initialize the console (allocate memory, create CPU)
 * @param console Console instance
 * @return true on success, false on failure
 */
bool vc_console_initialize(vc_console_t* console);

/**
 * Reset the console to initial state
 * @param console Console instance
 */
void vc_console_reset(vc_console_t* console);

/**
 * Shutdown the console and free resources
 * @param console Console instance
 */
void vc_console_shutdown(vc_console_t* console);


//----- ROM loading functions

/**
 * Load ROM from file
 * @param console Console instance
 * @param path Path to ROM file
 * @return true on success, false on failure
 */
bool vc_console_load_rom_file(vc_console_t* console, const char* path);

/**
 * Load ROM from memory buffer
 * @param console Console instance
 * @param data Pointer to ROM data
 * @param size Size of ROM data in bytes
 * @return true on success, false on failure
 */
bool vc_console_load_rom_data(vc_console_t* console, const uint8_t* data, size_t size);


//----- Execution control functions

/**
 * Execute one CPU cycle
 * @param console Console instance
 */
void vc_console_tick(vc_console_t* console);

/**
 * Execute N CPU cycles
 * @param console Console instance
 * @param cycles Number of cycles to execute
 */
void vc_console_run(vc_console_t* console, uint32_t cycles);

/**
 * Execute one frame worth of cycles
 * @param console Console instance
 */
void vc_console_run_frame(vc_console_t* console);

/**
 * Pause console execution
 * @param console Console instance
 */
void vc_console_pause(vc_console_t* console);

/**
 * Resume console execution
 * @param console Console instance
 */
void vc_console_resume(vc_console_t* console);


//----- State query functions

/**
 * Get current console state
 * @param console Console instance
 * @return Current console state
 */
vc_console_state_t vc_console_get_state(vc_console_t* console);

/**
 * Check if console is running
 * @param console Console instance
 * @return true if running, false otherwise
 */
bool vc_console_is_running(vc_console_t* console);

/**
 * Check if console is paused
 * @param console Console instance
 * @return true if paused, false otherwise
 */
bool vc_console_is_paused(vc_console_t* console);

/**
 * Get total cycles executed
 * @param console Console instance
 * @return Total cycles executed
 */
uint64_t vc_console_get_total_cycles(vc_console_t* console);

/**
 * Get CPU state
 * @param console Console instance
 * @return Current CPU state
 */
vc_cpu_state_t vc_console_get_cpu_state(vc_console_t* console);

/**
 * Get current thread ID
 * @param console Console instance
 * @return Current thread ID (0-7)
 */
int vc_console_get_current_thread_id(vc_console_t* console);


//----- Event system

/**
 * Trigger a hardware event
 * @param console Console instance
 * @param event Event code to trigger
 */
void vc_console_trigger_event(vc_console_t* console, uint32_t event);


//----- Error handling

/**
 * Get last error message
 * @param console Console instance
 * @return Error message string (valid until next operation)
 */
const char* vc_console_get_last_error(vc_console_t* console);

/**
 * Clear error state
 * @param console Console instance
 */
void vc_console_clear_error(vc_console_t* console);


//----- Memory access functions (for debugging)

/**
 * Read 8-bit value from memory
 * @param console Console instance
 * @param addr Memory address
 * @return 8-bit value
 */
uint8_t vc_console_mem_read8(vc_console_t* console, uint32_t addr);

/**
 * Read 16-bit value from memory
 * @param console Console instance
 * @param addr Memory address
 * @return 16-bit value
 */
uint16_t vc_console_mem_read16(vc_console_t* console, uint32_t addr);

/**
 * Read 32-bit value from memory
 * @param console Console instance
 * @param addr Memory address
 * @return 32-bit value
 */
uint32_t vc_console_mem_read32(vc_console_t* console, uint32_t addr);

/**
 * Write 8-bit value to memory
 * @param console Console instance
 * @param addr Memory address
 * @param value Value to write
 */
void vc_console_mem_write8(vc_console_t* console, uint32_t addr, uint8_t value);

/**
 * Write 16-bit value to memory
 * @param console Console instance
 * @param addr Memory address
 * @param value Value to write
 */
void vc_console_mem_write16(vc_console_t* console, uint32_t addr, uint16_t value);

/**
 * Write 32-bit value to memory
 * @param console Console instance
 * @param addr Memory address
 * @param value Value to write
 */
void vc_console_mem_write32(vc_console_t* console, uint32_t addr, uint32_t value);


//----- CPU/Thread debugging functions

/**
 * Get thread state
 * @param console Console instance
 * @param thread_id Thread ID (0-7)
 * @return Thread state
 */
vc_thread_state_t vc_console_get_thread_state(vc_console_t* console, int thread_id);

/**
 * Get thread program counter
 * @param console Console instance
 * @param thread_id Thread ID (0-7)
 * @return Program counter value
 */
uint32_t vc_console_get_thread_pc(vc_console_t* console, int thread_id);

/**
 * Get thread register value
 * @param console Console instance
 * @param thread_id Thread ID (0-7)
 * @param reg_num Register number (0-31)
 * @return Register value
 */
uint32_t vc_console_get_thread_register(vc_console_t* console, int thread_id, int reg_num);


#ifdef __cplusplus
}
#endif
