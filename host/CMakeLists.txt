cmake_minimum_required(VERSION 3.25)

# find SDL2
find_package(SDL2 REQUIRED)

# library / exexcutable name
set(VCLIB_NAME "natoga")
set(VCRE_NAME "nare")

# source files
set(HOST_SOURCES
    src/main.cpp
)

add_executable(${VCRE_NAME} ${HOST_SOURCES})

target_link_libraries(${VCRE_NAME}
    PRIVATE
        ${VCLIB_NAME}
        SDL2::SDL2
)

target_include_directories(${VCRE_NAME}
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# Install
install(TARGETS ${VCRE_NAME}
    RUNTIME DESTINATION bin
)
