// Section size checks for the NATOGA32 assembler
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package assembler

import (
	"fmt"

	"github.com/natoga32/n32asm/internal/section"
)

// sectionLimits gives the size of each section window, in memory map order
var sectionLimits = []struct {
	name  string
	limit int64
}{
	{".rom", MMAP_ROM_SIZE},
	{".bss", MMAP_BSS_SIZE},
	{".rodata", MMAP_RODATA_SIZE},
	{".data", MMAP_DATA_SIZE},
	{".text", MMAP_TEXT_SIZE},
}

// checkSectionSizes returns an error when a section grows past its
// memory map window, so it never spills into the next section
func checkSectionSizes(sections *section.Manager) error {
	for _, l := range sectionLimits {
		sec := sections.Get(l.name)
		if sec == nil || sec.Size <= l.limit {
			continue
		}
		return fmt.Errorf("section %s is %d bytes, exceeds its %dMiB window (0x%08X-0x%08X)",
			l.name, sec.Size, l.limit/(1024*1024), sec.BaseAddr, sec.BaseAddr+l.limit-1)
	}
	return nil
}
