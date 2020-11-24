package main

import (
	"fmt"

	"github.com/Bios-Marcel/cordless/tview"
	tcell "github.com/gdamore/tcell/v2"
)

// End shows the final slide.
func End(nextSlide func()) (title string, content tview.Primitive) {
	textView := tview.NewTextView().SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	url := "https://github.com/Bios-Marcel/cordless/tview"
	fmt.Fprint(textView, url)
	return "End", Center(len(url), 1, textView)
}
