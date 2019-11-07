package text

import "testing"

func TestParseTFACode(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    string
		wantErr bool
	}{
		{
			name:    "empty",
			text:    "",
			want:    "",
			wantErr: true,
		}, {
			name:    "empty, but spaces",
			text:    "     ",
			want:    "",
			wantErr: true,
		}, {
			name:    "negative number of length 7 including the minus",
			text:    "-100000",
			want:    "",
			wantErr: true,
		}, {
			name:    "negative number of length 6 including the minus",
			text:    "-10000",
			want:    "",
			wantErr: true,
		}, {
			name:    "negative 0",
			text:    "-0",
			want:    "000000",
			wantErr: false,
		}, {
			name:    "0",
			text:    "0",
			want:    "000000",
			wantErr: false,
		}, {
			name:    "upper limit",
			text:    "999999",
			want:    "999999",
			wantErr: false,
		}, {
			name:    "above upper limit",
			text:    "1000000",
			want:    "",
			wantErr: true,
		}, {
			name:    "non numeric",
			text:    "javascript is good",
			want:    "",
			wantErr: true,
		}, {
			name:    "correct with spaces",
			text:    "  123456  ",
			want:    "123456",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTFACode(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTFACode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseTFACode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
