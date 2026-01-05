package html_utils

import (
	"reflect"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	// http://, https:// наверное неважно.
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "No links",
			input:    "<html><h1>There is no links!</h1></html>",
			expected: []string{},
		},
		{
			name:     "One link",
			input:    `<html><a href="https://google.com">A</a></html>`,
			expected: []string{"https://google.com"},
		},
		{
			name: "Many links",
			input: `<html>
				<a href="https://google.com">A</a>
				<a href="ozon.ru">B</a>
			</html>`,
			expected: []string{"https://google.com", "ozon.ru"},
		},
		{
			name: "Relative links (Should be discarded, or append to full addr)",
			input: `<html>
				<a href="/cart">A</a>
			</html>`,
			expected: []string{"example.com/cart"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := ParseHtml(tt.input)
			if err != nil {
				t.Errorf("Failed! Error: %e", err)
			}

			result, err := ExtractLinks(input, "example.com")
			if err != nil {
				t.Errorf("Failed! Error: %e", err)
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Failed! Found: %v, Expected: %v", result, tt.expected)
			}
		})
	}

}
