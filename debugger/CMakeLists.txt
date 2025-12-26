cmake_minimum_required(VERSION 3.25)

find_package(SDL2 REQUIRED)

# Dear ImGui as submodule

add_library(imgui STATIC
    #${CMAKE_CURRENT_SOURCE_DIR}/../third_party/imgui/
)

target_include_directories(imgui PUBLIC
    #${CMAKE_CURRENT_SOURCE_DIR}/../third_party/imgui/
)

target_link_libraries(imgui PUBLIC SDL2::SDL2)

set(DEBUGGER_SOURCES
)

add_executable(vcdebug ${DEBUGGER_SOURCES})

target_link_libraries(vcdebug
    PRIVATE
        imgui
        natoga
)
