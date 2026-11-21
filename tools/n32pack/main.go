// n32pack - NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package main

import (
	"os"

	"github.com/natoga32/n32pack/internal/packer"
)

func main() {
	os.Exit(packer.Run(os.Args[1:], os.Stderr))
}
