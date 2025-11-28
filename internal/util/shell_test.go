package util

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"simple command", []string{"simple", "command"}},
		{"echo \"hello world\"", []string{"echo", "hello world"}},
		{"echo 'hello world'", []string{"echo", "hello world"}},
		{"git commit -m \"fix: bug\"", []string{"git", "commit", "-m", "fix: bug"}},
		{"escaped \\\"quote\\\"", []string{"escaped", "\"quote\""}},
		{"mixed 'quotes' \"here\"", []string{"mixed", "quotes", "here"}},
		{"   leading and trailing spaces   ", []string{"leading", "and", "trailing", "spaces"}},
		{"", nil},
	}

	for _, test := range tests {
		got, err := ParseArgs(test.input)
		if err != nil {
			t.Errorf("ParseArgs(%q) returned error: %v", test.input, err)
			continue
		}
		if !reflect.DeepEqual(got, test.expected) {
			// Handle nil vs empty slice for comparison if needed, but deepEqual handles it well usually if both are nil or both empty.
			// My implementation returns nil for empty.
			if len(got) == 0 && len(test.expected) == 0 {
				continue
			}
			t.Errorf("ParseArgs(%q) = %v, want %v", test.input, got, test.expected)
		}
	}
}

func TestJoinArgs(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"simple", "command"}, "simple command"},
		{[]string{"echo", "hello world"}, "echo \"hello world\""},
		{[]string{"git", "commit", "-m", "fix: bug"}, "git commit -m \"fix: bug\""},
		{[]string{"empty", ""}, "empty \"\""},
		{[]string{"has", "\"quote\""}, "has \"\\\"quote\\\"\""},
	}

	for _, test := range tests {
		got := JoinArgs(test.input)
		if got != test.expected {
			t.Errorf("JoinArgs(%v) = %q, want %q", test.input, got, test.expected)
		}
	}
}
