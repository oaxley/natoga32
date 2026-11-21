package ast

import "testing"

func TestSourceLocationString(t *testing.T) {
	tests := []struct {
		name     string
		loc      SourceLocation
		expected string
	}{
		{
			name:     "with filename",
			loc:      SourceLocation{Filename: "test.asm", Line: 10, Column: 5},
			expected: "test.asm:10:5",
		},
		{
			name:     "without filename",
			loc:      SourceLocation{Filename: "", Line: 10, Column: 5},
			expected: "10:5",
		},
		{
			name:     "with path",
			loc:      SourceLocation{Filename: "/path/to/file.asm", Line: 42, Column: 18},
			expected: "/path/to/file.asm:42:18",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.loc.String()
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSourceLocationIsValid(t *testing.T) {
	tests := []struct {
		name     string
		loc      SourceLocation
		expected bool
	}{
		{
			name:     "valid with line > 0",
			loc:      SourceLocation{Line: 1, Column: 1},
			expected: true,
		},
		{
			name:     "invalid with line = 0",
			loc:      SourceLocation{Line: 0, Column: 1},
			expected: false,
		},
		{
			name:     "invalid with negative line",
			loc:      SourceLocation{Line: -1, Column: 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.loc.IsValid()
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
