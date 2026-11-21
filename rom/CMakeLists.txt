cmake_minimum_required(VERSION 3.25)

# Boot ROM: rom/boot.s assembled with n32asm into rombios.sys, next to vchost.
# Needs n32asm (built by tools/) and Python 3 with Pillow (to convert the assets).

find_package(Python3 COMPONENTS Interpreter)

set(VC_ROM_HAS_PILLOW FALSE)
if(Python3_Interpreter_FOUND)
    execute_process(
        COMMAND ${Python3_EXECUTABLE} -c "import PIL"
        RESULT_VARIABLE VC_ROM_PILLOW_RESULT
        OUTPUT_QUIET ERROR_QUIET
    )
    if(VC_ROM_PILLOW_RESULT EQUAL 0)
        set(VC_ROM_HAS_PILLOW TRUE)
    endif()
endif()

if(NOT TARGET vc_n32asm)
    message(WARNING "n32asm not available: the boot ROM (rombios) will not be built")
    return()
endif()
if(NOT VC_ROM_HAS_PILLOW)
    message(WARNING "Python 3 with Pillow not found: the boot ROM (rombios) will not be built")
    return()
endif()

set(VC_ROMBIOS_FILE ${CMAKE_BINARY_DIR}/core/host/rombios.sys)
set(VC_ROMBIOS_FILE ${VC_ROMBIOS_FILE} PARENT_SCOPE)

set(VC_ROM_SOURCE ${CMAKE_CURRENT_BINARY_DIR}/boot.s)
set(VC_ROM_LOGO ${CMAKE_CURRENT_BINARY_DIR}/logo.raw)
set(VC_ROM_FONT ${CMAKE_CURRENT_BINARY_DIR}/font.raw)

# the source is copied next to the assets: .incbin paths are relative to it
add_custom_command(
    OUTPUT ${VC_ROM_SOURCE}
    COMMAND ${CMAKE_COMMAND} -E copy ${CMAKE_CURRENT_SOURCE_DIR}/boot.s ${VC_ROM_SOURCE}
    DEPENDS ${CMAKE_CURRENT_SOURCE_DIR}/boot.s
    VERBATIM
)

# logo and font assets
add_custom_command(
    OUTPUT ${VC_ROM_LOGO} ${VC_ROM_FONT}
    COMMAND ${Python3_EXECUTABLE} ${CMAKE_CURRENT_SOURCE_DIR}/tools/make_assets.py
            ${CMAKE_CURRENT_SOURCE_DIR}/assets/playerone.png ${CMAKE_CURRENT_BINARY_DIR}
    DEPENDS ${CMAKE_CURRENT_SOURCE_DIR}/tools/make_assets.py
            ${CMAKE_CURRENT_SOURCE_DIR}/assets/playerone.png
    COMMENT "Converting the boot ROM assets"
    VERBATIM
)

# the ROM image
add_custom_command(
    OUTPUT ${VC_ROMBIOS_FILE}
    COMMAND ${CMAKE_COMMAND} -E make_directory ${CMAKE_BINARY_DIR}/core/host
    COMMAND ${VC_N32ASM_EXECUTABLE} ${VC_ROM_SOURCE} -f rom -o ${VC_ROMBIOS_FILE}
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    DEPENDS vc_n32asm ${VC_N32ASM_EXECUTABLE} ${VC_ROM_SOURCE} ${VC_ROM_LOGO} ${VC_ROM_FONT}
    COMMENT "Assembling the boot ROM"
    VERBATIM
)

add_custom_target(rombios ALL DEPENDS ${VC_ROMBIOS_FILE})
