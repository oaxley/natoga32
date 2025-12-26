#include <SDL2/SDL.h>
#include <iostream>
#include <cstdlib>

const int WIDTH = 800;
const int HEIGHT = 480;

int main(int argc, char* argv[])
{
    // 1. initialize SDL
    if (SDL_Init(SDL_INIT_VIDEO) < 0) {
        std::cerr << "SDL Could not be initialized!" << std::endl;
        std::cerr << "SDL_Error:" << SDL_GetError() << std::endl;
        return EXIT_FAILURE;
    }

    // 2. create a window
    // SDL_Window* window = SDL_CreateWindow("SDL Example",
    //                                       SDL_WINDOWPOS_CENTERED,
    //                                       SDL_WINDOWPOS_CENTERED,
    //                                       WIDTH, HEIGHT,
    //                                       SDL_WINDOW_SHOWN
    // );
    SDL_Window* window = nullptr;
    SDL_Renderer* renderer = nullptr;
    SDL_CreateWindowAndRenderer(WIDTH, HEIGHT, 0, &window, &renderer);
    if (!window || !renderer) {
        std::cerr << "Window/Rendered could not be created!" << std::endl;
        std::cerr << "SDL_Error:" << SDL_GetError() << std::endl;
        SDL_Quit();
        return EXIT_FAILURE;
    }

    // 3. Get the window surface
    // SDL_Surface* screenSurface = SDL_GetWindowSurface(window);
    // if (!screenSurface) {
    //     std::cerr << "Could not obtain an SDL Surface!" << std::endl;
    //     std::cerr << "SDL_Error:" << SDL_GetError() << std::endl;
    //     SDL_Quit();
    //     return EXIT_FAILURE;
    // }

    // 4. fill the surface
    // SDL_FillRect(screenSurface, NULL, SDL_MapRGB(screenSurface->format, 0xAA, 0xAA, 0xAA));
    // SDL_UpdateWindowSurface(window);

    SDL_Texture* texture = SDL_CreateTexture(renderer, SDL_PIXELFORMAT_RGB565,
                    SDL_TEXTUREACCESS_STREAMING, WIDTH, HEIGHT);

    uint16_t* framebuffer = new uint16_t[WIDTH * HEIGHT];
    if (!framebuffer) {
        std::cerr << "Error with memory allocation!" << std::endl;
        SDL_Quit();
        return EXIT_FAILURE;
    }

    // fill everything with gray color
    uint8_t r = 16;
    uint8_t g = 32;
    uint8_t b = 16;
    for(uint32_t i = 0; i < WIDTH*HEIGHT; i++) {
        framebuffer[i] = (((r & 31) << 11) | ((g & 63) << 5) | (b & 31));
    }

    // update the texture
    SDL_UpdateTexture(texture, nullptr, framebuffer, WIDTH*2);
    SDL_RenderClear(renderer);
    SDL_RenderCopy(renderer, texture, nullptr, nullptr);
    SDL_RenderPresent(renderer);

    SDL_Delay(3000);

    SDL_DestroyWindow(window);
    if (renderer) {
        SDL_DestroyRenderer(renderer);
    }

    delete [] framebuffer;
    SDL_Quit();
    return EXIT_SUCCESS;
}
