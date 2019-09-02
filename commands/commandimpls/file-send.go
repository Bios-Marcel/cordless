package commandimpls

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/user"
	"path"
	"path/filepath"
	"strings"

	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/discordgo"
)

const fileSendDocumentation = `[::b]NAME
	file-send - send files from your local machine

[::b]SYNOPSIS
	[::b]file-send <FILE_PATH>...

[::b]DESCRIPTION
	The file-send command allows you to send multiple files to your current channel.

[::b]EXAMPLES
	[gray]$ file-send ~/file.txt
	[gray]$ file-send ~/file1.txt ~/file2.txt
	[gray]$ file-send "~/file one.txt" ~/file2.txt`

// FileSend represents the command used to send multiple files to a channel.
type FileSend struct {
	discord *discordgo.Session
	window  *ui.Window
}

// NewFileSendCommand creates a ready to use FileSend instance.
func NewFileSendCommand(discord *discordgo.Session, window *ui.Window) *FileSend {
	return &FileSend{
		discord: discord,
		window:  window,
	}
}

// Execute runs the command piping its output into the supplied writer.
func (cmd *FileSend) Execute(writer io.Writer, parameters []string) {
	channel := cmd.window.GetSelectedChannel()
	if channel == nil {
		fmt.Fprintln(writer, "[red]In order to use this command, you have to be in a channel.")
		return
	}

	if len(parameters) == 0 {
		cmd.PrintHelp(writer)
		return
	}

	for _, parameter := range parameters {
		var resolvedPath string
		if strings.HasPrefix(parameter, "~") {
			currentUser, userResolveError := user.Current()
			if userResolveError != nil {
				fmt.Fprintf(writer, "[red]Error resolving path:\n\t[red]%s\n", userResolveError.Error())
				continue
			}

			resolvedPath = filepath.Join(currentUser.HomeDir, strings.TrimPrefix(parameter, "~"))
		} else {
			resolvedPath = parameter
		}

		isAbs := filepath.IsAbs(resolvedPath)
		if !isAbs {
			fmt.Fprintln(writer, "[red]Error reading file:\n\t[red]the path is not absolute")
			continue
		}

		data, readError := ioutil.ReadFile(resolvedPath)
		if readError != nil {
			fmt.Fprintf(writer, "[red]Error reading file:\n\t[red]%s\n", readError.Error())
			continue
		}

		dataChannel := bytes.NewReader(data)
		_, sendError := cmd.discord.ChannelFileSend(channel.ID, path.Base(resolvedPath), dataChannel)
		if sendError != nil {
			fmt.Fprintf(writer, "[red]Error sending file:\n\t[red]%s\n", sendError.Error())
		}
	}
}

func (cmd *FileSend) Name() string {
	return "file-send"
}

func (cmd *FileSend) Aliases() []string {
	return []string{"filesend"}
}

// PrintHelp prints the help for the FileSend command.
func (cmd *FileSend) PrintHelp(writer io.Writer) {
	fmt.Fprint(writer, fileSendDocumentation)
}
