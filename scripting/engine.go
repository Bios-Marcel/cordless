package scripting

import (
	"io"

	"github.com/Bios-Marcel/discordgo"
)

// Engine describes a type that is capable of handling events from the main
// application and allows mutation of data.
type Engine interface {
	// LoadScripts loads scripts from a directory into the VM
	LoadScripts(string) error
	// SetErrorOutput sets the io.Writer that the errors are piped into.
	SetErrorOutput(errorOutput io.Writer)

	// OnMessageSend handles the client sending new messages, allowing scripts
	// to manipulate what's sent. The order of script execution is undefined
	// and should therefore be expected to be random.
	OnMessageSend(string) string
	// OnMessageReceive gets called every time a message is received, no matter
	// in which channel or guild.
	OnMessageReceive(*discordgo.Message)
	// OnMessageEdit gets called every time a message is edited, no matter in
	// which channel or guild.
	OnMessageEdit(*discordgo.Message)
	// OnMessageDelete gets called every time a message gets deleted, no matter
	// in which channel or guild.
	OnMessageDelete(*discordgo.Message)

	SetTriggerNotificationFunction(func(string, string))
	SetPrintToConsoleFunction(func(string))
	SetPrintLineToConsoleFunction(func(string))

	SetGetCurrentGuildFunction(func() string)
	SetGetCurrentChannelFunction(func() string)
}
