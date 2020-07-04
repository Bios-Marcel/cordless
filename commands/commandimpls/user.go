package commandimpls

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Bios-Marcel/discordemojimap"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

const (
	userHelpPage = `[::b]NAME
	user - manipulate and retrieve your user information

[::b]SYNOPSIS
	[::b]user <command>

[::b]DESCRIPTION
	This command allows you to manipulate and retrieve your user information.

	This command is split into multiple subcommands. The default subcommand
	is [::b]user-get[::-] and will be used if no other command was supplied.

[::b]SUBCOMMANDS
	[::b]user-get (default)
		prints the current user information
	[::b]user-set
		updates the current user information`

	userSetHelpPage = `[::b]NAME
	user-set - updates your accounts user information

[::b]SYNOPSIS
	[::b]user-set[::-] [OPTION[]...

[::b]DESCRIPTION
	This command allows you to set all or single values of your user
	information. Every value has a specific parameter and you'll always
	be asked for your password when trying to change any data.

[::b]OPTIONS
	TODO
	[::b]-n, --name
		change your nickname
	[::b]-e, --email
		change the e-mail address associated with your account
	[::b]-a, --avatar
		change your avatar to a new local file of yours
	[::b]-np, --new-password
		changes the password you use to log in to your account

[::b]EXAMPLES
	[gray]$ user-set -n "My new nickname"
	[gray]$ user-set -n NewName
	[gray]$ user-set -n NewName -a /home/pics/avatar.png`

	userGetHelpPage = `[::b]NAME
	user-get - prints your accounts user information

[::b]SYNOPSIS
	[::b]user-get[::-] [OPTION[]...

[::b]DESCRIPTION
	This command prints your accounts user information to the
	commandline in a human readable format. If no options were
	supplied, then "-n", "-e" and "-a" are chosen as the default
	options.

[::b]OPTIONS
	[::b]-n, --name
		Prints nickname and discriminator
	[::b]-e, --email
		Prints your e-mail address
	[::b]-a, --avatar
		Prints the URL of your avatar
	[::b]-t, --tfa
		Prints whether you have two-factor authentication enabled

[::b]EXAMPLES
	[gray]$ user
	Nick: Example#1234
	E-Mail: example@provider.com
	Avatar: https://discordapp.com/XXX/YYY.png

	[gray]$ user -a
	Avatar: https://discordapp.com/XXX/YYY.png`
)

type UserCmd struct {
	userSetCmd *UserSetCmd
	userGetCmd *UserGetCmd
}

type UserSetCmd struct {
	window  *ui.Window
	session *discordgo.Session
}

type UserGetCmd struct {
	window  *ui.Window
	session *discordgo.Session
}

func NewUserCommand(userSetCmd *UserSetCmd, userGetCmd *UserGetCmd) *UserCmd {
	return &UserCmd{userSetCmd, userGetCmd}
}

func NewUserSetCommand(window *ui.Window, session *discordgo.Session) *UserSetCmd {
	return &UserSetCmd{window, session}
}

func NewUserGetCommand(window *ui.Window, session *discordgo.Session) *UserGetCmd {
	return &UserGetCmd{window, session}
}

func (cmd *UserCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) >= 1 {
		if parameters[0] == "set" || parameters[0] == "update" {
			cmd.userSetCmd.Execute(writer, parameters[1:])
		} else if parameters[0] == "get" {
			cmd.userGetCmd.Execute(writer, parameters[1:])
		} else {
			cmd.userGetCmd.Execute(writer, parameters)
		}
	} else {
		cmd.userGetCmd.Execute(writer, parameters)
	}
}

func (cmd *UserGetCmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		//Calling get with defaults
		cmd.Execute(writer, []string{"-n", "-e", "-a"})
	} else {
		userInformation := ""
		for _, param := range parameters {
			switch param {
			case "-n", "--name", "--nick", "-u", "--username":
				userInformation += fmt.Sprintf("Nick: %s#%s\n", cmd.session.State.User.Username, cmd.session.State.User.Discriminator)
			case "-e", "--email", "--e-mail", "--mail":
				userInformation += fmt.Sprintf("E-Mail: %s\n", cmd.session.State.User.Email)
			case "-a", "--avatar", "--profile-picture":
				// FIXME Potential bug if jpeg is uploaded?
				userInformation += fmt.Sprintf("Avatar: https://cdn.discordapp.com/avatars/%s/%s.png\n", cmd.session.State.User.ID, cmd.session.State.User.Avatar)
			case "-m", "--mfa", "--tfa", "--2fa":
				userInformation += fmt.Sprintf("Two-Factor Authentication : %v\n", cmd.session.State.User.MFAEnabled)
			default:
				fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Invalid parameter '%s'\n", param)
				cmd.PrintHelp(writer)
				return
			}
		}

		fmt.Fprint(writer, userInformation)

	}
}

func (cmd *UserSetCmd) Execute(writer io.Writer, parameters []string) {
	if cmd.session.State.User.Bot {
		fmt.Fprintln(writer, "[red]This command can't be used by bots due to Discord API restrictions.")
		return
	}

	var newName, newEmail string
	newAvatar := cmd.session.State.User.Avatar
	var askForNewPassword bool
	for index, param := range parameters {
		switch param {
		case "-n", "--name", "--nick", "-u", "--username":
			if index != len(parameters)-1 {
				newName = parameters[index+1]
			} else {
				fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error, you didn't supply a new name.")
			}
		case "-e", "--email", "--e-mail", "--mail":
			if index != len(parameters)-1 {
				newEmail = parameters[index+1]
			} else {
				fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error, you didn't supply a new e-mail address.")
			}
		case "-a", "--avatar", "--profile-picture":
			if index != len(parameters)-1 && !strings.HasPrefix(parameters[index+1], "-") {
				newAvatar = parameters[index+1]
			} else {
				newAvatar = ""
			}
		case "--new-password", "-np":
			askForNewPassword = true
		}
	}

	if newName == "" && !askForNewPassword && newEmail == "" && newAvatar == cmd.session.State.User.Avatar {
		fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]No valid parameters were supplied.")
		cmd.PrintHelp(writer)
		return
	}

	if newName != "" {
		newName = discordemojimap.Replace(strings.TrimSpace(newName))
	}

	if newEmail != "" {
		newEmail = strings.TrimSpace(newEmail)
	}

	if newAvatar != "" && newAvatar != cmd.session.State.User.Avatar {
		newAvatar = strings.TrimSpace(newAvatar)
		var resolvedPath string
		if strings.HasPrefix(newAvatar, "~") {
			currentUser, userResolveError := user.Current()
			if userResolveError != nil {
				fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error resolving path:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", userResolveError.Error())
				return
			}

			resolvedPath = filepath.Join(currentUser.HomeDir, strings.TrimPrefix(newAvatar, "~"))
		} else {
			resolvedPath = newAvatar
		}

		resolvedPath, resolveError := filepath.EvalSymlinks(resolvedPath)
		if resolveError != nil {
			fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error resolving path:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", resolveError.Error())
			return
		}

		isAbs := filepath.IsAbs(resolvedPath)
		if !isAbs {
			fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error reading file:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]the path is not absolute")
			return
		}

		data, readError := ioutil.ReadFile(resolvedPath)
		if readError != nil {
			fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error reading file:\n\t["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]%s\n", readError.Error())
			return
		}

		contentType := http.DetectContentType(data)
		newAvatar = base64.StdEncoding.EncodeToString(data)
		if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/gif" {
			fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error updating avatar:\n\r["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]content type '%s' not supported", contentType)
			return
		}
		newAvatar = fmt.Sprintf("data:%s;base64,%s", contentType, newAvatar)
	}

	var newPassword string

	go func() {
		if askForNewPassword {
			newPassword = cmd.window.PromptSecretInput("Updating your user information", "Please enter your new password.")
			newPasswordConfirmation := cmd.window.PromptSecretInput("Updating your user information", "Please enter your new password again, to make sure it is correct.")

			if newPassword != newPasswordConfirmation {
				fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error, new passwords differ from each other, please try again.")
				cmd.window.ForceRedraw()
				return
			}
		}

		currentPassword := cmd.window.PromptSecretInput("Updating your user information", "Please enter your current password.")
		if currentPassword == "" {
			fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Empty password, aborting.")
		} else {
			_, err := cmd.session.UserUpdate(newEmail, currentPassword, newName, newAvatar, newPassword)
			if err == nil {
				fmt.Fprintln(writer, "Your user has been updated.")
			} else {
				fmt.Fprintln(writer, err)
			}
		}

		cmd.window.ForceRedraw()
	}()
}

func (cmd *UserCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, userHelpPage)
}

func (cmd *UserSetCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, userSetHelpPage)
}

func (cmd *UserGetCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, userGetHelpPage)
}

func (cmd *UserCmd) Name() string {
	return "user"
}

func (cmd *UserSetCmd) Name() string {
	return "user-set"
}

func (cmd *UserGetCmd) Name() string {
	return "user-get"
}

func (cmd *UserCmd) Aliases() []string {
	return nil
}

func (cmd *UserSetCmd) Aliases() []string {
	return []string{"user-update"}
}

func (cmd *UserGetCmd) Aliases() []string {
	return nil
}
