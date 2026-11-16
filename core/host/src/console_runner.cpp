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
 * @brief	Virtual Console - Basic Host Program
 */

// standard library headers
#include <iostream>

// SDL
#include <SDL3/SDL.h>

// program-specific includes
#include "console_runner.h"


//----- constants

// the window shows the console framebuffer, scaled x2 with --superscalar
constexpr int WINDOW_SCALE_NORMAL = 1;
constexpr int WINDOW_SCALE_SUPERSCALAR = 2;


//----- ConsoleRunner

ConsoleRunner::ConsoleRunner(const HostConfig& config)
    : config_(config)
{
}

ConsoleRunner::~ConsoleRunner()
{
    if (console_) {
        vc_destroy(console_);
    }
}

bool ConsoleRunner::boot()
{
    console_ = vc_create();
    if (!console_ || !vc_initialize(console_)) {
        std::cerr << "Failed to initialize the console: "
                  << (console_ ? vc_getLastError(console_) : "out of memory") << "\n";
        return false;
    }

    if (!vc_loadRomFile(console_, config_.rombios.c_str())) {
        std::cerr << "Failed to load ROM: " << config_.rombios << ": "
                  << vc_getLastError(console_) << "\n";
        return false;
    }

    if (!config_.cartridge.empty()) {
        std::cerr << "Cartridge loading is not supported yet: " << config_.cartridge << "\n";
    }

    vc_reset(console_);
    vc_resume(console_);
    return true;
}

int ConsoleRunner::run()
{
    if (!boot()) {
        return EXIT_FAILURE;
    }
    return config_.frames > 0 ? runHeadless() : runWindowed();
}

// no SDL call here, so it works without a display
int ConsoleRunner::runHeadless()
{
    for (int i = 0; i < config_.frames; ++i) {
        vc_runFrame(console_);
    }

    if (vc_getState(console_) == VC_STATE_ERROR) {
        std::cerr << "Console error: " << vc_getLastError(console_) << "\n";
        return EXIT_FAILURE;
    }
    return EXIT_SUCCESS;
}

int ConsoleRunner::runWindowed()
{
    if (!SDL_Init(SDL_INIT_VIDEO)) {
        std::cerr << "SDL_Init failed: " << SDL_GetError() << "\n";
        return EXIT_FAILURE;
    }

    const int scale = config_.superscalar ? WINDOW_SCALE_SUPERSCALAR : WINDOW_SCALE_NORMAL;
    SDL_Window* window = nullptr;
    SDL_Renderer* renderer = nullptr;
    SDL_Texture* texture = nullptr;
    int result = EXIT_FAILURE;

    do {
        window = SDL_CreateWindow("NATOGA32", VC_VIDEO_WIDTH * scale, VC_VIDEO_HEIGHT * scale, 0);
        if (!window) {
            std::cerr << "SDL_CreateWindow failed: " << SDL_GetError() << "\n";
            break;
        }

        renderer = SDL_CreateRenderer(window, nullptr);
        if (!renderer) {
            std::cerr << "SDL_CreateRenderer failed: " << SDL_GetError() << "\n";
            break;
        }
        SDL_SetRenderVSync(renderer, 1);

        texture = SDL_CreateTexture(renderer, SDL_PIXELFORMAT_RGB565,
                                    SDL_TEXTUREACCESS_STREAMING,
                                    VC_VIDEO_WIDTH, VC_VIDEO_HEIGHT);
        if (!texture) {
            std::cerr << "SDL_CreateTexture failed: " << SDL_GetError() << "\n";
            break;
        }
        SDL_SetTextureScaleMode(texture, SDL_SCALEMODE_NEAREST);

        bool running = true;
        while (running) {
            SDL_Event event;
            while (SDL_PollEvent(&event)) {
                if (event.type == SDL_EVENT_QUIT) {
                    running = false;
                } else if (event.type == SDL_EVENT_KEY_DOWN && !event.key.repeat) {
                    if (event.key.key == SDLK_ESCAPE) {
                        running = false;
                    } else if (event.key.key == SDLK_R) {
                        vc_reset(console_);
                        vc_resume(console_);
                    }
                }
            }

            vc_runFrame(console_);

            const uint16_t* fb = vc_getFramebuffer(console_);
            if (fb) {
                SDL_UpdateTexture(texture, nullptr, fb, VC_VIDEO_WIDTH * sizeof(uint16_t));
            }
            SDL_RenderClear(renderer);
            SDL_RenderTexture(renderer, texture, nullptr, nullptr);
            SDL_RenderPresent(renderer);
        }
        result = EXIT_SUCCESS;
    } while (false);

    if (texture) SDL_DestroyTexture(texture);
    if (renderer) SDL_DestroyRenderer(renderer);
    if (window) SDL_DestroyWindow(window);
    SDL_Quit();
    return result;
}
