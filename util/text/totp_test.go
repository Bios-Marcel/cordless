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

// TestGenerateBase32Key tests whether 100 unique keys can be generated
// generated in a row. This is simple in order to make sure that the keys
// aren't always the same. On top of that, keys are checked against the
// valid base 32 characters and a length of 16.
func TestGenerateBase32Key(t *testing.T) {
	availableCharacters := [...]rune{
		'2', '3', '4', '5', '6', '7',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	}

	iterations := 50
	generated := make(map[string]bool, iterations)
	for i := 0; i < iterations; i++ {
		newKey, keyError := GenerateBase32Key()
		if keyError != nil {
			t.Errorf("Error generating key: %s", keyError)
		}
		_, ok := generated[newKey]
		if ok {
			t.Errorf("Duplicated key: %s", newKey)
		}
		generated[newKey] = true

		if len(newKey) != 16 {
			t.Errorf("Keylength is invalid: %d", len(newKey))
		}

	OUTER_LOOP:
		for _, char := range []rune(newKey) {
			for _, validChar := range availableCharacters {
				if validChar == char {
					continue OUTER_LOOP
				}
			}
			t.Errorf("Generated key contains invalid character: %c", char)
		}
	}
}
