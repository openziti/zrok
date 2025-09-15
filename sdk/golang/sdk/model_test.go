package sdk

import (
	"testing"
)

func TestParseNamespaceSelection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected NamespaceSelection
		hasError bool
	}{
		{
			name:  "namespace token only",
			input: "token123",
			expected: NamespaceSelection{
				NamespaceToken: "token123",
				Name:           "",
			},
			hasError: false,
		},
		{
			name:  "namespace token with name",
			input: "token123:myname",
			expected: NamespaceSelection{
				NamespaceToken: "token123",
				Name:           "myname",
			},
			hasError: false,
		},
		{
			name:  "namespace token with empty name",
			input: "token123:",
			expected: NamespaceSelection{
				NamespaceToken: "token123",
				Name:           "",
			},
			hasError: false,
		},
		{
			name:  "name with colon in it",
			input: "token123:name:with:colons",
			expected: NamespaceSelection{
				NamespaceToken: "token123",
				Name:           "name:with:colons",
			},
			hasError: false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: NamespaceSelection{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseNamespaceSelection(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.NamespaceToken != tt.expected.NamespaceToken {
				t.Errorf("expected NamespaceToken %q, got %q", tt.expected.NamespaceToken, result.NamespaceToken)
			}

			if result.Name != tt.expected.Name {
				t.Errorf("expected Name %q, got %q", tt.expected.Name, result.Name)
			}
		})
	}
}