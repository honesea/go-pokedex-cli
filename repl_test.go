package main

import "testing"

func TestCleanInput(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input: "hello world",
			expected: []string{
				"hello",
				"world",
			},
		},
		{
			input: "HELLO WORLD",
			expected: []string{
				"hello",
				"world",
			},
		},
	}

	for _, cs := range cases {
		actual := cleanInput(cs.input)
		if len(actual) != len(cs.expected) {
			t.Errorf("The lengths don't match: %v (actual) - %v (expected)", len(actual), len(cs.expected))
			continue
		}

		for i := 0; i < len(cs.expected); i++ {
			if actual[i] != cs.expected[i] {
				t.Errorf("Word mismatch: %v (actual) - %v (expected)", actual[i], cs.expected[i])
				break
			}
		}
	}
}
