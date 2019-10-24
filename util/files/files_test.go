//+build !windows

package files

import (
	"testing"
)

func TestToAbsolutePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "absolute path",
			input:   "/test.txt",
			want:    "/test.txt",
			wantErr: false,
		}, {
			name:    "absolute path; file uri",
			input:   "file:///te%20st.txt",
			want:    "/te st.txt",
			wantErr: false,
		}, {
			name:    "absolute path; contains Escaped characters",
			input:   "/te%20st.txt",
			want:    "/te st.txt",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToAbsolutePath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToAbsolutePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToAbsolutePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
