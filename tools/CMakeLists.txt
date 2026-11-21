cmake_minimum_required(VERSION 3.25)

# Go tools: the n32asm assembler and the n32pack cartridge packer, built next to
# the top-level build directory. Both are rebuilt whenever one of their Go sources changes.

find_program(VC_GO_EXECUTABLE go)

if(NOT VC_GO_EXECUTABLE)
    message(WARNING "Go not found: n32asm and n32pack will not be built")
    return()
endif()

set(VC_N32ASM_DIR ${CMAKE_CURRENT_SOURCE_DIR}/n32asm)
set(VC_N32ASM_EXECUTABLE ${CMAKE_BINARY_DIR}/n32asm)
set(VC_N32PACK_DIR ${CMAKE_CURRENT_SOURCE_DIR}/n32pack)
set(VC_N32PACK_EXECUTABLE ${CMAKE_BINARY_DIR}/n32pack)

# n32asm
file(GLOB_RECURSE VC_N32ASM_SOURCES CONFIGURE_DEPENDS
    ${VC_N32ASM_DIR}/*.go
    ${VC_N32ASM_DIR}/go.mod
    ${VC_N32ASM_DIR}/go.sum
)
add_custom_command(
    OUTPUT ${VC_N32ASM_EXECUTABLE}
    COMMAND ${VC_GO_EXECUTABLE} build -o ${VC_N32ASM_EXECUTABLE} .
    WORKING_DIRECTORY ${VC_N32ASM_DIR}
    DEPENDS ${VC_N32ASM_SOURCES}
    COMMENT "Building n32asm"
    VERBATIM
)
add_custom_target(vc_n32asm ALL DEPENDS ${VC_N32ASM_EXECUTABLE})

# n32pack
file(GLOB_RECURSE VC_N32PACK_SOURCES CONFIGURE_DEPENDS
    ${VC_N32PACK_DIR}/*.go
    ${VC_N32PACK_DIR}/go.mod
    ${VC_N32PACK_DIR}/go.sum
)
add_custom_command(
    OUTPUT ${VC_N32PACK_EXECUTABLE}
    COMMAND ${VC_GO_EXECUTABLE} build -o ${VC_N32PACK_EXECUTABLE} .
    WORKING_DIRECTORY ${VC_N32PACK_DIR}
    DEPENDS ${VC_N32PACK_SOURCES}
    COMMENT "Building n32pack"
    VERBATIM
)
add_custom_target(vc_n32pack ALL DEPENDS ${VC_N32PACK_EXECUTABLE})

set(VC_N32ASM_EXECUTABLE ${VC_N32ASM_EXECUTABLE} PARENT_SCOPE)
set(VC_N32PACK_EXECUTABLE ${VC_N32PACK_EXECUTABLE} PARENT_SCOPE)
