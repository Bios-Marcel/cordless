package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/discordgo"
)

const serverDocumentation = `[orange][::u]# server[white]

The server command allows you to manage the servers you are in. This command
doesn't allow administrating a server though.

Available subcommands
  * join - takes an invite and accepts it
`

// Server is the command that allows managing the UserGuilds.
type Server struct {
	session *discordgo.Session
	window  *ui.Window
}

// NewServerCommand makes a new ready to use Server instance.
func NewServerCommand(session *discordgo.Session, window *ui.Window) *Server {
	return &Server{
		session: session,
		window:  window,
	}
}

// Execute runs the command logic.
func (s *Server) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		s.PrintHelp(writer)
		return
	}

	switch parameters[0] {
	case "join", "accept", "enter":
		if len(parameters) != 2 {
			s.printServerJoinHelp(writer)
		} else {
			s.joinServer(writer, parameters[1])
		}
	default:
		s.PrintHelp(writer)
	}
}

func (s *Server) printServerJoinHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: server join <InviteCode|InviteURL>")
}

func (s *Server) joinServer(writer io.Writer, input string) {
	lastSlash := strings.LastIndex(input, "/")
	var inviteID string
	if lastSlash == -1 {
		inviteID = input
	} else {
		inviteID = input[lastSlash+1:]
	}

	invite, err := s.session.InviteAccept(inviteID)
	if err != nil {
		fmt.Fprintf(writer, "[red]Error accepting invite with ID '%s':\n\t[red]%s\n", inviteID, err)
	} else {
		fmt.Fprintf(writer, "Joined server '%s'\n", invite.Guild.Name)
	}
}

// Name identifies this command, allowing you to call it by using the returned
// value.
func (s *Server) Name() string {
	return "server"
}

// PrintHelp prints the general help text for the Server command.
func (s *Server) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, serverDocumentation)
}
