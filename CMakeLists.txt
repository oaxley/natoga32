cmake_minimum_required(VERSION 3.25)
project(natoga32 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

add_subdirectory(console)
add_subdirectory(host)
add_subdirectory(tests)

