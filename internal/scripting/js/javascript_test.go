package js

import (
	"testing"
)

func TestJavaScriptEngine(t *testing.T) {
	tests := []struct {
		dir   string
		input string
		want  string
	}{
		{
			dir:   "test/simple",
			input: "Replace me",
			want:  "Replace this",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			e := New()
			if err := e.LoadScripts(tt.dir); err != nil {
				t.Error("LoadScripts failed:", err)
				return
			}

			if gotNewText := e.OnMessageSend(tt.input); gotNewText != tt.want {
				t.Errorf("JavaScriptEngine.OnMessageSend() = %v, want %v", gotNewText, tt.want)
			}
		})
	}
}
