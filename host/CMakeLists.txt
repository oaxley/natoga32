add_executable(nare
    src/main.cpp
)

target_include_directories(nare
    PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)

target_link_libraries(nare PRIVATE natoga)

target_compile_features(nare PRIVATE cxx_std_20)
