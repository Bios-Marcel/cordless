package tviewutil

import (
	"testing"

	"github.com/gdamore/tcell"
)

func TestColorToHex(t *testing.T) {

	tests := []struct {
		name  string
		color tcell.Color
		want  string
	}{
		{
			name:  "black",
			color: tcell.ColorBlack,
			want:  "#000000",
		}, {
			name:  "white",
			color: tcell.ColorWhite,
			want:  "#ffffff",
		}, {
			name:  "red",
			color: tcell.ColorRed,
			want:  "#ff0000",
		}, {
			name:  "green",
			color: tcell.ColorRed,
			want:  "#00ff00",
		}, {
			name:  "blue",
			color: tcell.ColorRed,
			want:  "#0000ff",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ColorToHex(tt.color); got != tt.want {
				t.Errorf("ColorToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}
