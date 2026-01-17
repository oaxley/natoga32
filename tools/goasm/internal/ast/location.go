// SourceLocation tracking for AST nodes
//
// This source file is subject to the MIT License
// that is bundled with this package in the file LICENSE.txt.
package ast

import "fmt"

// SourceLocation represents a location in the source code
type SourceLocation struct {
	Filename string
	Line     int
	Column   int
}

// String returns a formatted location string
func (loc SourceLocation) String() string {
	if loc.Filename == "" {
		return fmt.Sprintf("%d:%d", loc.Line, loc.Column)
	}
	return fmt.Sprintf("%s:%d:%d", loc.Filename, loc.Line, loc.Column)
}

// IsValid returns true if the location has valid line information
func (loc SourceLocation) IsValid() bool {
	return loc.Line > 0
}
