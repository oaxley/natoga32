cmake_minimum_required(VERSION 3.25)

# SDL2 required for imgui
find_package(glfw3 REQUIRED)
find_package(OpenGL REQUIRED)

# CMake fetch argparse library
include(FetchContent)
FetchContent_Declare(
    argparse
    GIT_REPOSITORY https://github.com/p-ranav/argparse.git
)
FetchContent_MakeAvailable(argparse)

# create the Virtual Console debugger binary
set(IMGUI_DIR ${CMAKE_SOURCE_DIR}/third_party/imgui)

add_executable(vcdebug
    src/main.cpp
    src/debug_client.cpp

    ${IMGUI_DIR}/imgui.cpp
    ${IMGUI_DIR}/imgui_draw.cpp
    ${IMGUI_DIR}/imgui_tables.cpp
    ${IMGUI_DIR}/imgui_widgets.cpp
    ${IMGUI_DIR}/backends/imgui_impl_glfw.cpp
    ${IMGUI_DIR}/backends/imgui_impl_opengl3.cpp
)

# include directory
target_include_directories(vcdebug
    PRIVATE
        ${IMGUI_DIR}
        ${IMGUI_DIR}/backends
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# link targets
target_link_libraries(vcdebug
    PUBLIC
        glfw
        OpenGL::GL
    PRIVATE
        argparse
        imgui
        vc
)
