package ui

import (
	"testing"

	_ "github.com/Bios-Marcel/cordless/syntax"
)

func TestParseBoldAndUnderline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple bold",
			input: "**Hallo Welt**",
			want:  "[::b]Hallo Welt[::-]",
		},
		{
			name:  "Useless bold",
			input: "****Hallo Welt",
			want:  "****Hallo Welt",
		},
		{
			name:  "Non closed bold",
			input: "**Hallo Welt",
			want:  "**Hallo Welt",
		},
		{
			name:  "Bold newline",
			input: "**Hallo\nWelt**",
			want:  "[::b]Hallo\n[::b]Welt[::-]",
		},
		{
			name:  "Bold newline2",
			input: "Hallo**\nWelt**",
			want:  "Hallo[::b]\n[::b]Welt[::-]",
		},
		{
			name:  "Bold newline3",
			input: "Hal**lo\nWelt**",
			want:  "Hal[::b]lo\n[::b]Welt[::-]",
		},
		{
			name:  "Simple underline",
			input: "__Hallo Welt__",
			want:  "[::u]Hallo Welt[::-]",
		},
		{
			name:  "Useless underline",
			input: "____Hallo Welt",
			want:  "____Hallo Welt",
		},
		{
			name:  "Non closed underline",
			input: "__Hallo Welt",
			want:  "__Hallo Welt",
		},
		{
			name:  "Underline newline",
			input: "__Hallo\nWelt__",
			want:  "[::u]Hallo\n[::u]Welt[::-]",
		},
		{
			name:  "Underline newline2",
			input: "Hallo__\nWelt__",
			want:  "Hallo[::u]\n[::u]Welt[::-]",
		},
		{
			name:  "Underline newline3",
			input: "Hal__lo\nWelt__",
			want:  "Hal[::u]lo\n[::u]Welt[::-]",
		},
		{
			name:  "Underline and bold",
			input: "**__Hallo Welt__**",
			want:  "[::b][::bu]Hallo Welt[::b][::-]",
		},
		{
			name:  "Underline and bold2",
			input: "** OwO__Hallo Welt__**",
			want:  "[::b] OwO[::bu]Hallo Welt[::b][::-]",
		},
		{
			name:  "Underline and bold3",
			input: "** OwO__Hallo Welt__** What",
			want:  "[::b] OwO[::bu]Hallo Welt[::b][::-] What",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBoldAndUnderline(tt.input); got != tt.want {
				t.Errorf("ParseBoldAndUnderline() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func Test_removeLeadingWhitespaceInCode(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "no whitespace; single line",
			code: "1",
			want: "1",
		}, {
			name: "no whitespace; two lines",
			code: "1\n2",
			want: "1\n2",
		}, {
			name: "single space on single line",
			code: " 1",
			want: "1",
		}, {
			name: "two spaces on single line",
			code: "  1",
			want: "1",
		}, {
			name: "multiple spaces on single line",
			code: "    1",
			want: "1",
		}, {
			name: "multiple spaces on multiple lines with different amounts",
			code: "    1\n  2\n 3",
			want: "   1\n 2\n3",
		}, {
			name: "single tab on single line",
			code: "	1",
			want: "1",
		}, {
			name: "single tab on multiple lines",
			code: "	1\n	2",
			want: "1\n2",
		}, {
			name: "multiple tabs on multiple lines",
			code: "		1\n		2",
			want: "1\n2",
		}, {
			name: "multiple tabs on multiple lines; each line different tab amount",
			code: "				1\n		2",
			want: "		1\n2",
		}, {
			name: "only one line with a tab",
			code: "	1\n2",
			want: "	1\n2",
		}, {
			name: "multiple tabs on single line",
			code: "			1",
			want: "1",
		}, {
			name: "mixed tabs and spaces of multiple lines",
			code: " 	1\n  	2\n 	3",
			want: "	1\n 	2\n	3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeLeadingWhitespaceInCode(tt.code); got != tt.want {
				t.Errorf("removeLeadingWhitespaceInCode() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
