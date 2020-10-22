package tviewutil

import (
	"testing"

	tcell "github.com/gdamore/tcell/v2"
)

func TestColorToHex(t *testing.T) {

	tests := []struct {
		name  string
		color tcell.Color
		want  string
	}{
		{
			name:  "black",
			color: tcell.NewRGBColor(0, 0, 0),
			want:  "#000000",
		}, {
			name:  "white",
			color: tcell.NewRGBColor(255, 255, 255),
			want:  "#ffffff",
		}, {
			name:  "red",
			color: tcell.NewRGBColor(255, 0, 0),
			want:  "#ff0000",
		}, {
			name:  "green",
			color: tcell.NewRGBColor(0, 255, 0),
			want:  "#00ff00",
		}, {
			name:  "blue",
			color: tcell.NewRGBColor(0, 0, 255),
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
