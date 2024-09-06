package main

import (
	"reflect"
	"testing"
)

func GetURLsFromHTMLTest(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  []string
	}{
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
<html>
	<body>
		<a href="/path/one">
			<span>Boot.dev</span>
		</a>
		<a href="https://other.com/path/one">
			<span>Boot.dev</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:     "relative URLs only",
			inputURL: "https://example.com",
			inputBody: `
<html>
	<body>
		<a href="/relative/path">
			<span>Link</span>
		</a>
		<a href="another/path">
			<span>Another Link</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://example.com/relative/path", "https://example.com/another/path"},
		},
		{
			name:     "absolute URLs only",
			inputURL: "https://example.com",
			inputBody: `
<html>
	<body>
		<a href="https://example.com/absolute/path">
			<span>Absolute Link</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://example.com/absolute/path"},
		},
		{
			name:     "no URLs",
			inputURL: "https://example.com",
			inputBody: `
<html>
	<body>
		<p>No links here!</p>
	</body>
</html>
`,
			expected: []string{},
		},
		{
			name:     "invalid URLs",
			inputURL: "https://example.com",
			inputBody: `
<html>
	<body>
		<a href="not-a-url">
			<span>Invalid URL</span>
		</a>
	</body>
</html>
`,
			expected: []string{"https://example.com/not-a-url"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := getURLsFromHTML(tc.inputBody, tc.inputURL)
			if err != nil {
				t.Errorf("FAIL: unexpected error: %v", err)
				return
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("FAIL: expected URLs: %v, actual: %v", tc.expected, actual)
			}
		})
	}
}
