package commandimpls

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/cordless/util/files"
)

const fileSendDocumentation = `[::b]NAME
	file-send - send files from your local machine

[::b]SYNOPSIS
	[::b]file-send [OPTION[]... <FILE_PATH>...

[::b]DESCRIPTION
	The file-send command allows you to send multiple files to your current channel.

	[::b]OPTIONS
	[::b]-b, --bulk
		Zips all files and sends them as a single file.
		Without this option, the folder structure won't be preserved.
	[::b]-r, --recursive
		Allow sending folders as well

[::b]EXAMPLES
	[gray]$ file-send ~/file.txt
	[gray]$ file-send -r ~/folder
	[gray]$ file-send -r -b ~/folder
	[gray]$ file-send ~/file1.txt ~/file2.txt
	[gray]$ file-send "~/file one.txt" ~/file2.txt
	[gray]$ file-send -b "~/file one.txt" ~/file2.txt
	[gray]$ file-send -b -r "~/file one.txt" ~/folder ~/file2.txt`

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
		fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]In order to use this command, you have to be in a channel.")
		return
	}

	if len(parameters) == 0 {
		cmd.PrintHelp(writer)
		return
	}

	var filteredParameters []string
	//Will cause all files in folders to be upload. This counts for subfolders as well.
	var recursive bool
	//Puts all files into one zip.
	var bulk bool

	//Parse flags and sort them out for further processing.
	for _, parameter := range parameters {
		if parameter == "-r" || parameter == "--recursive" {
			recursive = true
		} else if parameter == "-b" || parameter == "--bulk" {
			bulk = true
		} else {
			filteredParameters = append(filteredParameters, parameter)
		}
	}

	//Assume that all leftofer parameters are paths and convert them to absolute paths.
	var consumablePaths []string
	for _, parameter := range filteredParameters {
		resolvedPath, resolveError := files.ToAbsolutePath(parameter)
		if resolveError != nil {
			commands.PrintError(writer, "Error reading file", resolveError.Error())
			return
		}

		consumablePaths = append(consumablePaths, resolvedPath)
	}

	//If folders are not to be included, we error if any folder is found.
	if !recursive {
		for _, path := range consumablePaths {
			stats, statError := os.Stat(path)
			if statError != nil {
				if os.IsNotExist(statError) {
					commands.PrintError(writer, "Invalid input", fmt.Sprintf("'%s' doesn't exist", path))
				} else {
					commands.PrintError(writer, "Invalid input", statError.Error())
				}
				return
			}

			if stats.IsDir() {
				commands.PrintError(writer, "Invalid input", "Directories can only be uploaded if the '-r' flag is set")
				return
			}
		}
	}

	if bulk {
		//We read and write at the same time to save performance and memory.
		zipOutput, zipInput := io.Pipe()

		//While we write, we read in a background thread. We stay in
		//memory, instead of going over the filesystem.
		go func() {
			defer zipOutput.Close()
			_, sendError := cmd.discord.ChannelFileSend(channel.ID, "files.zip", zipOutput)
			if sendError != nil {
				fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error sending file:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", sendError.Error())
			}
		}()

		zipWriter := zip.NewWriter(zipInput)
		defer zipInput.Close()
		defer zipWriter.Close()
		for _, parameter := range consumablePaths {
			zipError := files.AddToZip(zipWriter, parameter)
			if zipError != nil {
				log.Println(zipError.Error())
			}
		}
	} else {
		//We skip directories and flatten the folder structure.
		for _, filePath := range consumablePaths {
			filepath.Walk(filePath, func(file string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}

				data, readError := ioutil.ReadFile(file)
				if readError != nil {
					fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error reading file:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", readError.Error())
					return nil
				}

				dataChannel := bytes.NewReader(data)
				_, sendError := cmd.discord.ChannelFileSend(channel.ID, path.Base(file), dataChannel)
				if sendError != nil {
					fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error sending file:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", sendError.Error())
				}
				return nil
			})
		}
	}
}

// Name represents the main-name of the command.
func (cmd *FileSend) Name() string {
	return "file-send"
}

// Aliases represents all available aliases this command can be called with.
func (cmd *FileSend) Aliases() []string {
	return []string{"filesend", "sendfile", "send-file", "file-upload", "upload-file"}
}

// PrintHelp prints the help for the FileSend command.
func (cmd *FileSend) PrintHelp(writer io.Writer) {
	fmt.Fprint(writer, fileSendDocumentation)
}
