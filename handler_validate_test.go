package main

import (
	"testing"
)

func TestReplaceBadWords(t *testing.T) {
	cases := []struct {
		input struct {
			body  string
			words []string
		}
		expected string
	}{
		{
			input: struct {
				body  string
				words []string
			}{
				body:  "Test to see if it find bad words and replaces them, here is one Lonnie",
				words: []string{"lonnie"},
			},
			expected: "Test to see if it find bad words and replaces them, here is one ****",
		},
		{
			input: struct {
				body  string
				words []string
			}{
				body:  "A bunch of bad words, Lonnie Enzo Odie",
				words: []string{"lonnie", "enzo", "odie"},
			},
			expected: "A bunch of bad words, **** **** ****",
		},
	}

	for _, c := range cases {
		got := replaceBadWords(c.input.body, c.input.words)
		if got != c.expected {
			t.Fatalf("expected %q, got %q", c.expected, got)
		}
	}
}
