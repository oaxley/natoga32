#include <catch2/catch_all.hpp>

#include "core/memory_map.h"


using namespace vc;

TEST_CASE("Memory Segments are 32-bit aligned", "[memory]") {
    REQUIRE( (ROM_BOOT_LOADER % alignof(u32)) == 0 );
}
