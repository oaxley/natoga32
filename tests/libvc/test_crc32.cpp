#include <catch2/catch_all.hpp>

#include <array>
#include <cstring>
#include <span>
#include <string_view>

#include "core/crc32.h"

using namespace vc;

static std::span<const u8> bytes_of(std::string_view s) {
    return {reinterpret_cast<const u8*>(s.data()), s.size()};
}

TEST_CASE("CRC-32", "[crc32]") {
    REQUIRE( crc32(bytes_of("123456789")) == 0xCBF43926u );
    REQUIRE( crc32({}) == 0u );

    SECTION("chained over two spans") {
        u32 partial = crc32(bytes_of("1234"));
        REQUIRE( crc32(bytes_of("56789"), partial) == 0xCBF43926u );
    }
}
