# Makefile

#----- constants
BUILD_DIR := build
MAKEFLAGS += --silent

#----- rules
.PHONY: debug release test host tools

debug:
	@cmake -B $(BUILD_DIR) -DCMAKE_BUILD_TYPE=debug
	@cmake --build $(BUILD_DIR) -j $(nproc)

release:
	@cmake -B $(BUILD_DIR) -DCMAKE_BUILD_TYPE=release
	@cmake --build $(BUILD_DIR) -j $(nproc)

test:
	@cd $(BUILD_DIR) && ctest --output-on-failure

host:
	cd $(BUILD_DIR)/host && ./vchost

clean:
	@rm -rf $(BUILD_DIR)

tools:
	@cd tools/goasm && go build -o ../../build/vcasm ./cmd/goasm/
