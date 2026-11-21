// Command line entry point of the NATOGA32 cartridge packer
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package packer

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const usage = "usage: n32pack [-o game.n32] [-inc assets.inc] game.json"

// Run parses the command line arguments and packs; returns the exit code
func Run(args []string, stderr io.Writer) int {
	flags := flag.NewFlagSet("n32pack", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprintln(stderr, usage)
		flags.PrintDefaults()
	}
	out := flags.String("o", "", "write the cartridge to this file")
	inc := flags.String("inc", "", "write the asset ids to this include file")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 1 || (*out == "" && *inc == "") {
		flags.Usage()
		return 2
	}
	path := flags.Arg(0)

	fail := func(err error) int {
		fmt.Fprintf(stderr, "n32pack: %v\n", err)
		return 1
	}
	m, err := loadManifest(path)
	if err != nil {
		return fail(err)
	}

	// build everything in memory first, so that no file is written when packing fails
	var cart []byte
	if *out != "" {
		if cart, err = pack(m, time.Now()); err != nil {
			return fail(err)
		}
	}
	var incFile []byte
	if *inc != "" {
		incFile = assetsInc(m, filepath.Base(path))
	}

	if *out != "" {
		if err := os.WriteFile(*out, cart, 0o644); err != nil {
			return fail(err)
		}
	}
	if *inc != "" {
		if err := os.WriteFile(*inc, incFile, 0o644); err != nil {
			return fail(err)
		}
	}
	return 0
}
