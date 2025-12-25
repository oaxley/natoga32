add_library(natoga STATIC
    src/main.cpp
)

target_include_directories(natoga
    PUBLIC
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)

target_compile_features(natoga PUBLIC cxx_std_20)
