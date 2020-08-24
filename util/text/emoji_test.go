package text

import (
	"reflect"
	"testing"
)

func TestFindEmojiIndices(t *testing.T) {
	type testcase struct {
		name     string
		input    string
		expected []int
	}
	tests := []*testcase{
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "spaces",
			input:    "         ",
			expected: nil,
		},
		{
			name:     "text with outer spaces",
			input:    "     Numma Numma yay    ",
			expected: nil,
		},
		{
			name:     "just single emoji, but with spaces, therefore invaid",
			input:    ":hel lo:",
			expected: nil,
		},
		{
			name:     "unclosed sequence",
			input:    ":hello",
			expected: nil,
		},
		{
			name:     "unopened sequence",
			input:    "hello:",
			expected: nil,
		},
		{
			name:     "tab inbetween sequence",
			input:    ":hel\tlo:",
			expected: nil,
		},
		{
			name:     "newline inbetween sequence",
			input:    ":hel\nlo:",
			expected: nil,
		},
		{
			name:     "just single emoji",
			input:    ":hello:",
			expected: []int{0, 6},
		},
		{
			name:     "two emoji without anything inbetween",
			input:    ":hello::world:",
			expected: []int{7, 13, 0, 6},
		},
		{
			name:     "two emoji withspaces inbetween",
			input:    ":hello:  :world:",
			expected: []int{9, 15, 0, 6},
		},
		{
			name:     "two emoji with text inbetween",
			input:    ":hello: Lorem Ipsum whatever :world:",
			expected: []int{29, 35, 0, 6},
		},
		{
			name:     "two valid emoji sequences with single unnecessary double colons in between",
			input:    ":test:::lol:",
			expected: []int{7, 11, 0, 5},
		},
		{
			name:     "two valid emoji sequences with even amount of double colons in between",
			input:    "::test:::lol:",
			expected: []int{8, 12, 1, 6},
		},
		{
			name:     "two valid emoji sequences with uneven amount of double colons in between",
			input:    "::test::::lol:",
			expected: []int{9, 13, 1, 6},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t2 *testing.T) {
			result := FindEmojiIndices([]rune(test.input))

			//We don't care whether it's empty or nil
			if len(result) == 0 && len(test.expected) == 0 {
				return
			}

			if !reflect.DeepEqual(result, test.expected) {
				t2.Errorf("Got %v, but expected %v", result, test.expected)
			}
		})
	}
}
