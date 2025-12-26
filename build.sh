#!/bin/bash
set -e

# by default build the debug one
if [[ -z $1 ]]; then
    BUILD_TYPE="Debug"
else
    BUILD_TYPE="Release"
fi


# build the library and host
cmake -B build -DCMAKE_BUILD_TYPE=${BUILD_TYPE} "$@"
cmake --build build -j$(nproc)

# build tools
# to be done

# end
echo "Build complete."
