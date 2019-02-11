package commandimpls

import (
	"fmt"
	"io"
	"github.com/Bios-Marcel/cordless/internal/config"
)

const logoutDocumentation = `[orange]# logout[white]

Logs you out from cordless. You'll be able to log in with the same or another account after re-launching cordless.
`

// Logout is command which nulls discord token in config.json
type Logout struct{}

// NewLogoutCommand constructs a new usable logout command for the user.
func NewLogoutCommand() *Logout {
	return &Logout{}
}

func (logout *Logout) Execute(writer io.Writer, parameters []string) {
	config.GetConfig().Token = ""
	config.PersistConfig()
}

// Name represents this commands indentifier.
func (logout *Logout) Name() string {
	return "logout"
}

func (logout *Logout) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, logoutDocumentation)
}