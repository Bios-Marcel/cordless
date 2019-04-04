package scripting

import "io"

// Engine describes a type that is capable of handling events from the main
// application and allows mutation of data.
type Engine interface {
	// LoadScripts loads scripts from a directory into the VM
	LoadScripts(string) error
	// OnMessageSend handles the client sending new messages
	OnMessageSend(string) string
	// SetErrorOutput sets the io.Writer that the errors are piped into.
	SetErrorOutput(errorOutput io.Writer)
}
