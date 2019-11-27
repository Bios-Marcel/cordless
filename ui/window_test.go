package ui

import (
	"reflect"
	"testing"
)

func Test_emojiSequenceIndexes(t *testing.T) {
	type args struct {
		runes []rune
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "empty",
			args: args{
				runes: []rune(""),
			},
			want: nil,
		}, {
			name: "whitespace",
			args: args{
				runes: []rune("     "),
			},
			want: nil,
		}, {
			name: "simple word",
			args: args{
				runes: []rune("hello"),
			},
			want: nil,
		}, {
			name: "non-closed sequence",
			args: args{
				runes: []rune(":hello"),
			},
			want: nil,
		}, {
			name: "single valid standalone emoji sequence",
			args: args{
				runes: []rune(":test:"),
			},
			want: []int{0, 5},
		}, {
			name: "two valid emoji sequences right next to eachother",
			args: args{
				runes: []rune(":test::lol:"),
			},
			want: []int{6, 10, 0, 5},
		}, {
			name: "two valid emoji sequences with single unnecessary double colons inbetween",
			args: args{
				runes: []rune(":test:::lol:"),
			},
			want: []int{7, 11, 0, 5},
		}, {
			name: "two valid emoji sequences with even amount of double colons inbetween",
			args: args{
				runes: []rune("::test:::lol:"),
			},
			want: []int{8, 12, 1, 6},
		}, {
			name: "two valid emoji sequences with uneven amount of double colons inbetween",
			args: args{
				runes: []rune("::test::::lol:"),
			},
			want: []int{9, 13, 1, 6},
		}, {
			name: "one emoji sequence in the end and the beginning",
			args: args{
				runes: []rune(":test: Hello yeah :lol:"),
			},
			want: []int{18, 22, 0, 5},
		}, {
			name: "emoji sequences with space, therefore invalid",
			args: args{
				runes: []rune(":test :"),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emojiSequenceIndexes(tt.args.runes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("emojiSequenceIndexes() = %v, want %v", got, tt.want)
			}
		})
	}
}
