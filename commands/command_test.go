package commands

import (
	"io"
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "no command",
			input: "",
			want:  nil,
		}, {
			name:  "evil whitespace attack",
			input: "zeige\u200Bmir\u200Bdeins",
			want:  []string{"zeige\u200Bmir\u200Bdeins"},
		}, {
			name:  "kaomojis???",
			input: `¯\_(ツ)_/¯`,
			want:  []string{`¯\_(ツ)_/¯`},
		}, {
			name:  "just poo",
			input: "💩",
			want:  []string{"💩"},
		}, {
			name:  "no command, just whitespace",
			input: "   ",
			want:  nil,
		}, {
			name:  "A command with a string where the quote in the end was forgotten",
			input: "hello \"sunny world",
			want:  []string{"hello", "\"sunny", "world"},
		}, {
			name:  "A command with a backslash in the end",
			input: "hello world\\",
			want:  []string{"hello", "world\\"},
		}, {
			name:  "just quoted whitespace",
			input: "\"   \"",
			want:  []string{"   "},
		}, {
			name:  "just quoted whitespace as an argument",
			input: "echo \"   \"",
			want:  []string{"echo", "   "},
		}, {
			name:  "simple, no arguments",
			input: "command",
			want:  []string{"command"},
		}, {
			name:  "command with simple argument",
			input: "command argument",
			want:  []string{"command", "argument"},
		}, {
			name:  "command with two simple argument",
			input: "command argument argument2",
			want:  []string{"command", "argument", "argument2"},
		}, {
			name:  "command with two simple arguments and more whitespace in between",
			input: "command   argument  argument2  ",
			want:  []string{"command", "argument", "argument2"},
		}, {
			name:  "command with one simple argument and a string",
			input: "command argument \"argument2 is long\"",
			want:  []string{"command", "argument", "argument2 is long"},
		}, {
			name:  "command with one simple argument and a string containing a newline",
			input: "command argument \"argument2 is \nlong\"",
			want:  []string{"command", "argument", "argument2 is \nlong"},
		}, {
			name:  "command with one simple argument and single arguments containing escaped quotes",
			input: "command argument \\\"argument2 argument3\\\"",
			want:  []string{"command", "argument", "\"argument2", "argument3\""},
		}, {
			name:  "command with one simple argument and a string containing an escaped quote",
			input: "command argument \"argument2 is \\\" long\"",
			want:  []string{"command", "argument", "argument2 is \" long"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCommand(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

type cmd struct {
	name    string
	aliases []string
}

func (c cmd) Execute(writer io.Writer, parameters []string) {
	panic("shouldn't be needed for the test")
}

func (c cmd) PrintHelp(writer io.Writer) {
	panic("shouldn't be needed for the test")
}

func (c cmd) Name() string {
	return c.name
}

func (c cmd) Aliases() []string {
	return c.aliases
}

func TestCommandEquals(t *testing.T) {
	type args struct {
		cmd  Command
		text string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "match by name",
			args: args{
				cmd: cmd{
					name:    "match",
					aliases: nil,
				},
				text: "match",
			},
			want: true,
		}, {
			name: "match by name with nonmatching aliases",
			args: args{
				cmd: cmd{
					name:    "match",
					aliases: []string{"hello", "world"},
				},
				text: "match",
			},
			want: true,
		}, {
			name: "match by single alias",
			args: args{
				cmd: cmd{
					name:    "nomatch",
					aliases: []string{"match"},
				},
				text: "match",
			},
			want: true,
		}, {
			name: "match with multiple aliases and match in the middle",
			args: args{
				cmd: cmd{
					name:    "nomatch",
					aliases: []string{"a", "b", "match", "c", "d"},
				},
				text: "match",
			},
			want: true,
		}, {
			name: "match with multiple aliases and match at the end",
			args: args{
				cmd: cmd{
					name:    "nomatch",
					aliases: []string{"a", "b", "c", "d", "match"},
				},
				text: "match",
			},
			want: true,
		}, {
			name: "no match with no aliases",
			args: args{
				cmd: cmd{
					name:    "nomatch",
					aliases: nil,
				},
				text: "match",
			},
			want: false,
		}, {
			name: "no match with single aliases",
			args: args{
				cmd: cmd{
					name:    "nomatch",
					aliases: []string{"a"},
				},
				text: "match",
			},
			want: false,
		}, {
			name: "no match with multiple aliases",
			args: args{
				cmd: cmd{
					name:    "nomatch",
					aliases: []string{"a", "b", "c"},
				},
				text: "match",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommandEquals(tt.args.cmd, tt.args.text); got != tt.want {
				t.Errorf("CommandEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}
