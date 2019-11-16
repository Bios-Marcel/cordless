package text

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ParseTFACodes takes an arbitrary string and checks whether it's a valid 6
// digit number for usage as a tfa code.
func ParseTFACode(text string) (string, error) {
	var mfaToken int64
	mfaTokenText := strings.ReplaceAll(text, " ", "")
	if mfaTokenText != "" {
		var parseError error
		mfaToken, parseError = strconv.ParseInt(mfaTokenText, 10, 32)
		if parseError != nil {
			return "", errors.New("token has to be a 6 digit number between 000000 and 999999")
		}

		if mfaToken > 999999 || mfaToken < 0 {
			return "", errors.New("token has to be a 6 digit number between 000000 and 999999")
		}

		return fmt.Sprintf("%06d", mfaToken), nil
	}

	return "", errors.New("tfa code must not be empty")
}

// GenerateBase32Key generates a 16 character key containing 2-7 and A-Z.
// The implementation uses crypto/rand.
func GenerateBase32Key() (string, error) {
	randomBytes := make([]byte, 16, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	availableCharacters := [...]rune{
		'2', '3', '4', '5', '6', '7',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
	}
	tfaSecretRaw := make([]rune, 16, 16)
	for index, randomByte := range randomBytes {
		tfaSecretRaw[index] = availableCharacters[rune(int(randomByte)%len(availableCharacters))]
	}

	return string(tfaSecretRaw), nil
}
