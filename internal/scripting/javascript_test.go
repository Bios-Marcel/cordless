package scripting

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
			"test/simple",
			"Replace me",
			"Replace this",
		},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			e := New()
			if err := e.LoadScripts(tt.dir); err != nil {
				t.Error("LoadScripts failed:", err)
				return
			}
			if gotNewText := e.OnMessage(tt.input); gotNewText != tt.want {
				t.Errorf("JavaScriptEngine.OnMessage() = %v, want %v", gotNewText, tt.want)
			}
		})
	}
}
