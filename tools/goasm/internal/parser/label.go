package parser

import (
	"strings"

	"github.com/natoga32/goasm/internal/ast"
	"github.com/natoga32/goasm/internal/token"
)

// parseLabel parses a label definition
func (p *Parser) parseLabel() (*ast.Label, error) {
	tok := p.advance()
	name := tok.Value

	// Remove trailing ':' if present
	if strings.HasSuffix(name, ":") {
		name = name[:len(name)-1]
	}

	// Consume trailing EOL
	p.expect(token.EOL)

	return &ast.Label{
		Name:     name,
		Location: locationFromToken(tok),
	}, nil
}
