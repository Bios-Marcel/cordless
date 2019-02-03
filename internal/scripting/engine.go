package scripting

// Engine describes a type that is capable of handling events from the main
// application and allows mutation of data.
type Engine interface {
	LoadScripts(string) error // loads scripts from a directory into the VM
	OnMessage(string) string  // OnMessage handles new messages
}
