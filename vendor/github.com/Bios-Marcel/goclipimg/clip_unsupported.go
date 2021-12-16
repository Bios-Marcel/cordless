// +build !linux,!darwin,!windows

package goclipimg

import "errors"

// ErrImagePasteUnsupported means that this system has no implementation for pasting an image.

var ErrImagePasteUnsupported = errors.New("This system doesn't have a paste implementation.")

// getImageFromClipboard always returns ErrImagePasteUnsupported, since the

// compilation target currently doesn't support pasting images.

func getImageFromClipboard() ([]byte, error) {
	return nil, ErrImagePasteUnsupported
}
