package preprocess

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/natoga32/goasm/internal/lexer"
	"github.com/natoga32/goasm/internal/token"
)

// handleInclude processes .include "filename"
func (p *Preprocessor) handleInclude(ts *tokenStream) ([]token.Token, error) {
	incTok := ts.advance() // consume .include

	// Get the filename
	if !ts.check(token.STRING) {
		return nil, locationError(incTok, ".include: expected a file name in quotes")
	}
	filename := ts.advance().Value

	// Strip quotes
	if len(filename) >= 2 && filename[0] == '"' && filename[len(filename)-1] == '"' {
		filename = filename[1 : len(filename)-1]
	}

	// Skip EOL
	if ts.check(token.EOL) {
		ts.advance()
	}

	// Resolve path
	fullPath := filepath.Join(p.baseDir, filename)

	// Check for recursive include
	if p.includes[fullPath] {
		return nil, locationError(incTok, ".include \"%s\": the file includes itself", filename)
	}

	// Mark as included (will be cleared after full processing completes)
	p.includes[fullPath] = true

	// Read and tokenize the file
	file, err := os.Open(fullPath)
	if err != nil {
		delete(p.includes, fullPath)
		return nil, locationError(incTok, "cannot open include file '%s': %v", fullPath, unwrapPathError(err))
	}
	defer file.Close()

	lex, err := lexer.NewFromReaderWithFile(file, fullPath)
	if err != nil {
		delete(p.includes, fullPath)
		return nil, locationError(incTok, "cannot read include file '%s': %v", fullPath, err)
	}

	tokens := lex.Tokenize()
	if err := lex.Err(); err != nil {
		delete(p.includes, fullPath)
		return nil, err
	}

	// Remove EOF from included tokens
	if len(tokens) > 0 && tokens[len(tokens)-1].Type == token.EOF {
		tokens = tokens[:len(tokens)-1]
	}

	// Process the included tokens now (while path is still marked)
	processed, err := p.processTokens(tokens)
	if err != nil {
		delete(p.includes, fullPath)
		return nil, err
	}

	// Now safe to unmark
	delete(p.includes, fullPath)

	// Remove EOF if added by recursive call
	if len(processed) > 0 && processed[len(processed)-1].Type == token.EOF {
		processed = processed[:len(processed)-1]
	}

	return processed, nil
}

// unwrapPathError drops the "open FILE:" prefix of an os error: the
// message already names the file
func unwrapPathError(err error) error {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err
	}
	return err
}
