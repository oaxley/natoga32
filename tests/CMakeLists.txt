cmake_minimum_required(VERSION 3.25)

# CMake fetch Catch2 library
include(FetchContent)
FetchContent_Declare(
    Catch2
    GIT_REPOSITORY https://github.com/catchorg/Catch2.git
    GIT_TAG        v3.12.0
)
FetchContent_MakeAvailable(Catch2)

# create the Virtual Console tests binary
add_executable(vctests
    libvc/test_memmap.cpp
    libvc/test_memory.cpp
    libvc/test_decoder.cpp
    libvc/test_console.cpp
    libvc/test_bootrom.cpp
    libvc/test_crc32.cpp
    libvc/test_cartridge.cpp

    libvc/test_cpu/basic.cpp
    libvc/test_cpu/opcode-memory.cpp
    libvc/test_cpu/opcode-auipc-lui.cpp
    libvc/test_cpu/opcode-misc-mem.cpp
    libvc/test_cpu/branches.cpp
    libvc/test_cpu/opcodes_19.cpp
    libvc/test_cpu/opcodes_51.cpp
    libvc/test_cpu/opcodes-zba-zbb.cpp
    libvc/test_cpu/opcodes-csr.cpp
    libvc/test_cpu/idle.cpp
    libvc/test_cpu/opcodes-thread.cpp
    libvc/test_cpu/private-zone.cpp
    libvc/test_cpu/cartridge.cpp
    libvc/test_cpu/traps.cpp

    host/test_frame_pacer.cpp
)

# link targets
target_link_libraries(vctests PRIVATE vc Catch2::Catch2WithMain)

# include directories
target_include_directories(vctests
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/../core/console/src
        ${CMAKE_CURRENT_SOURCE_DIR}/../core/console/include
        ${CMAKE_CURRENT_SOURCE_DIR}/../core/host/src
        ${CMAKE_CURRENT_SOURCE_DIR}/libvc
)

# boot ROM image (built by the rombios target when Go and Python+Pillow are available)
if(TARGET rombios)
    target_compile_definitions(vctests PRIVATE VC_ROMBIOS_PATH="${VC_ROMBIOS_FILE}")
    add_dependencies(vctests rombios)
endif()

# Catch2 tests auto-discovery
include(CTest)
include(Catch)
catch_discover_tests(vctests)

# cartridge checker: loads a cartridge in the emulator and prints its status
add_executable(vc_cartcheck cartcheck/cartcheck.cpp)
target_link_libraries(vc_cartcheck PRIVATE vc)
target_include_directories(vc_cartcheck
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/../core/console/src
        ${CMAKE_CURRENT_SOURCE_DIR}/../core/console/include
)

# sample cartridge packed with goasm and n32pack, then checked by the emulator
if(TARGET vc_goasm AND TARGET vc_n32pack)
    set(VC_CART_SRC ${CMAKE_CURRENT_SOURCE_DIR}/cartridge)
    set(VC_CART_DIR ${CMAKE_CURRENT_BINARY_DIR}/cartridge)

    # the sources are copied next to the generated files: manifest and .include paths are relative
    foreach(VC_CART_FILE hello.s hello.json level1.tiles)
        configure_file(${VC_CART_SRC}/${VC_CART_FILE} ${VC_CART_DIR}/${VC_CART_FILE} COPYONLY)
    endforeach()

    add_custom_command(
        OUTPUT ${VC_CART_DIR}/assets.inc
        COMMAND ${VC_N32PACK_EXECUTABLE} -inc assets.inc hello.json
        WORKING_DIRECTORY ${VC_CART_DIR}
        DEPENDS vc_n32pack ${VC_N32PACK_EXECUTABLE} ${VC_CART_DIR}/hello.json
        COMMENT "Generating the sample cartridge assets.inc"
        VERBATIM
    )
    add_custom_command(
        OUTPUT ${VC_CART_DIR}/hello.rscv ${VC_CART_DIR}/hello.sym
        COMMAND ${VC_GOASM_EXECUTABLE} hello.s -f rscv -o hello.rscv --sym hello.sym
        WORKING_DIRECTORY ${VC_CART_DIR}
        DEPENDS vc_goasm ${VC_GOASM_EXECUTABLE} ${VC_CART_DIR}/hello.s ${VC_CART_DIR}/assets.inc
        COMMENT "Assembling the sample cartridge"
        VERBATIM
    )
    add_custom_command(
        OUTPUT ${VC_CART_DIR}/hello.n32
        COMMAND ${VC_N32PACK_EXECUTABLE} -o hello.n32 hello.json
        WORKING_DIRECTORY ${VC_CART_DIR}
        DEPENDS vc_n32pack ${VC_N32PACK_EXECUTABLE} ${VC_CART_DIR}/hello.json
                ${VC_CART_DIR}/hello.rscv ${VC_CART_DIR}/hello.sym ${VC_CART_DIR}/level1.tiles
        COMMENT "Packing the sample cartridge"
        VERBATIM
    )
    add_custom_target(sample_cartridge ALL DEPENDS ${VC_CART_DIR}/hello.n32)

    add_test(NAME cartridge_pack_valid
        COMMAND vc_cartcheck ${VC_CART_DIR}/hello.n32)
    set_tests_properties(cartridge_pack_valid PROPERTIES
        PASS_REGULAR_EXPRESSION "cartridge valid, 3 chunks indexed")
    add_test(NAME cartridge_check_rejects_rscv
        COMMAND vc_cartcheck ${VC_CART_DIR}/hello.rscv)
    set_tests_properties(cartridge_check_rejects_rscv PROPERTIES WILL_FAIL TRUE)
    add_test(NAME cartridge_check_rejects_rscv_message
        COMMAND vc_cartcheck ${VC_CART_DIR}/hello.rscv)
    set_tests_properties(cartridge_check_rejects_rscv_message PROPERTIES
        PASS_REGULAR_EXPRESSION "cartridge invalid: error 1")
endif()

# vchost headless runs (no SDL call when --frames is given)
if(TARGET vchost)
    add_test(NAME vchost_missing_rom
        COMMAND vchost --rombios /nonexistent/rom.sys --frames 1)
    set_tests_properties(vchost_missing_rom PROPERTIES WILL_FAIL TRUE)
    add_test(NAME vchost_missing_rom_message
        COMMAND vchost --rombios /nonexistent/rom.sys --frames 1)
    set_tests_properties(vchost_missing_rom_message PROPERTIES
        PASS_REGULAR_EXPRESSION "Failed to load ROM")
    add_test(NAME vchost_negative_frames
        COMMAND vchost --frames -1)
    set_tests_properties(vchost_negative_frames PROPERTIES WILL_FAIL TRUE)
    add_test(NAME vchost_negative_frames_message
        COMMAND vchost --frames -1)
    set_tests_properties(vchost_negative_frames_message PROPERTIES
        PASS_REGULAR_EXPRESSION "--frames must be >= 0")

    # a zero-filled ROM traps at 0x100 without a handler: the console stops in error
    set(VC_BAD_ROM_CMD "head -c 260 /dev/zero > $0 && exec $<TARGET_FILE:vchost> --rombios $0 --frames 1")
    add_test(NAME vchost_headless_console_error
        COMMAND /bin/sh -c "${VC_BAD_ROM_CMD}" trap_rom_exit.sys)
    set_tests_properties(vchost_headless_console_error PROPERTIES WILL_FAIL TRUE)
    add_test(NAME vchost_headless_console_error_message
        COMMAND /bin/sh -c "${VC_BAD_ROM_CMD}" trap_rom_message.sys)
    set_tests_properties(vchost_headless_console_error_message PROPERTIES
        PASS_REGULAR_EXPRESSION "Console error: CPU halted due to exception")

    if(TARGET rombios)
        add_test(NAME vchost_headless
            COMMAND vchost --rombios ${VC_ROMBIOS_FILE} --frames 5)
        add_test(NAME vchost_headless_superscalar
            COMMAND vchost --rombios ${VC_ROMBIOS_FILE} --superscalar --frames 5)

        add_test(NAME vchost_missing_cartridge
            COMMAND vchost --rombios ${VC_ROMBIOS_FILE} /nonexistent/game.n32 --frames 1)
        set_tests_properties(vchost_missing_cartridge PROPERTIES WILL_FAIL TRUE)
        add_test(NAME vchost_missing_cartridge_message
            COMMAND vchost --rombios ${VC_ROMBIOS_FILE} /nonexistent/game.n32 --frames 1)
        set_tests_properties(vchost_missing_cartridge_message PROPERTIES
            PASS_REGULAR_EXPRESSION "Failed to load cartridge")

        # an invalid cartridge does not stop the console: the ROM reads CART_STATUS
        set(VC_BAD_CART_CMD "head -c 64 /dev/zero > $0 && exec $<TARGET_FILE:vchost> --rombios ${VC_ROMBIOS_FILE} $0 --frames 5")
        add_test(NAME vchost_invalid_cartridge_boots
            COMMAND /bin/sh -c "${VC_BAD_CART_CMD}" bad_cart.n32)

        # run by name through PATH from another directory: the default ROM
        # is still found next to the executable
        get_filename_component(VC_VCHOST_DIR ${VC_ROMBIOS_FILE} DIRECTORY)
        add_test(NAME vchost_default_rom_via_path
            COMMAND /bin/sh -c "exec vchost --frames 1")
        set_tests_properties(vchost_default_rom_via_path PROPERTIES
            WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
            ENVIRONMENT "PATH=${VC_VCHOST_DIR}:/usr/bin:/bin")
    endif()
endif()
