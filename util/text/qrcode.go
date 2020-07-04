package text

import (
	"bytes"

	"github.com/mdp/qrterminal/v3"
	"rsc.io/qr"
)

func GenerateQRCode(text string, redundancyLevel qr.Level) string {
	buffer := bytes.NewBufferString("")
	qrConfig := qrterminal.Config{
		Level:          redundancyLevel,
		Writer:         buffer,
		HalfBlocks:     true,
		BlackChar:      qrterminal.BLACK_BLACK,
		WhiteBlackChar: qrterminal.WHITE_BLACK,
		WhiteChar:      qrterminal.WHITE_WHITE,
		BlackWhiteChar: qrterminal.BLACK_WHITE,
		QuietZone:      2,
	}
	qrterminal.GenerateWithConfig(text, qrConfig)
	return buffer.String()
}
