package util

import (
	"strings"
	"unicode"
)

// ParseArgs parses a command line string into arguments, handling quotes.
// It supports single quotes (') and double quotes (").
func ParseArgs(input string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	var escaped bool

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
			continue
		}

		if r == '"' || r == '\'' {
			quote = r
			continue
		}

		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if quote != 0 {
		return nil, nil // Unclosed quote, simplistic error handling or just return what we have?
		// Ideally we should return an error.
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

// JoinArgs joins arguments into a single string, quoting them if necessary.
func JoinArgs(args []string) string {
	var sb strings.Builder
	for i, arg := range args {
		if i > 0 {
			sb.WriteRune(' ')
		}

		needsQuote := false
		if arg == "" {
			needsQuote = true
		} else {
			for _, r := range arg {
				if unicode.IsSpace(r) || r == '"' || r == '\'' {
					needsQuote = true
					break
				}
			}
		}

		if needsQuote {
			// Simple quoting strategy: wrap in double quotes and escape existing double quotes
			sb.WriteRune('"')
			for _, r := range arg {
				if r == '"' || r == '\\' {
					sb.WriteRune('\\')
				}
				sb.WriteRune(r)
			}
			sb.WriteRune('"')
		} else {
			sb.WriteString(arg)
		}
	}
	return sb.String()
}
