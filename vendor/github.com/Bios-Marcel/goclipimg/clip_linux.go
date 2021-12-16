// +build linux

package goclipimg

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
)

// ErrImagePasteUnsupported means that xclip(X11) or wl-clipboard(Wayland)
// can't be found or isn't installed.
var ErrImagePasteUnsupported = errors.New("xclip/wl-clipboard is not available on this system")

func isCommandAvailable(name string) bool {
	_, fileError := exec.LookPath(name)
	return fileError == nil
}

func getImageFromWayland(buffer *bytes.Buffer) error {
	//The wl-clipboard package has the commands wl-copy and wl-paste
	if !isCommandAvailable("wl-paste") {
		return ErrImagePasteUnsupported
	}

	wlPaste := exec.Command("wl-paste", "-t", "image/png")
	wlPaste.Stdout = buffer
	return wlPaste.Run()
}

func getImageFromXclip(buffer *bytes.Buffer) error {
	if !isCommandAvailable("xclip") {
		return ErrImagePasteUnsupported
	}

	xclip := exec.Command("xclip", "-sel", "clipboard", "-t", "image/png", "-o")
	xclip.Stdout = buffer
	return xclip.Run()
}

func getImageFromClipboard() ([]byte, error) {
	var buffer bytes.Buffer
	//500KB
	buffer.Grow(500000)
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	var clipError error
	if sessionType == "wayland" {
		clipError = getImageFromWayland(&buffer)
	} else {
		//For everything not wayland, we default to x11
		clipError = getImageFromXclip(&buffer)
	}

	if clipError != nil {
		return nil, ErrNoImageInClipboard
	}

	image := buffer.Bytes()
	if len(image) == 0 {
		return nil, ErrNoImageInClipboard
	}

	return image, nil
}
