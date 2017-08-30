package utils

import (
	"bytes"
	"io"
)

type DirectiveFunc func(reader io.Reader) []byte

var (
	directives = []DirectiveFunc{}
)

func AddDirective(df DirectiveFunc) {
	directives = append(directives, df)
}

func ApplyDirectives(templ []byte) string {
	if len(directives) == 0 {
		return string(templ)
	}

	result := templ
	for _, directive := range directives {
		result = directive(bytes.NewReader(result))
	}
	return string(result)
}

// instead of stripping the CR, we maintain it..
func CustomScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
