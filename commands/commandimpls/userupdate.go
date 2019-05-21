package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/discordgo"
)

const userUpdateDocumentation = `[orange]# user-update[white]

This command allows updating your account-information.

All changeable values can be passed via additional parameters. The following parameters are available:

-n, --name <new name>            - Sets the given value as your new name
-e, --email <new e-mail address> - Sets the given value as your accounts new e-mail address
-np, --new-password              - Will cause the application to ask you for a new password 

At least one of the parameters has to be passed, otherwise running this command will fail.

Examples:
  user-update --name Hello
  user-update --name "Hello World"
  user-update --name "Marcel" -e mynewemail@coolprovider.com
  user-update -np
`

// UserUpdate allows changing the users account-information.
type UserUpdate struct {
	window  *ui.Window
	session *discordgo.Session
}

// NewUserUpdateCommand creates a new ready to use UserUpdate command.
func NewUserUpdateCommand(window *ui.Window, session *discordgo.Session) *UserUpdate {
	return &UserUpdate{window, session}
}

// Execute runs the UserUpdate command.
func (nick *UserUpdate) Execute(writer io.Writer, parameters []string) {
	var newName, newEmail string
	var askForNewPassword bool
	for index, param := range parameters {
		switch param {
		case "-n", "--name", "--nick", "-u", "--username":
			if index != len(parameters)-1 {
				newName = parameters[index+1]
			} else {
				fmt.Fprintln(writer, "[red]Error, you didn't supply a new name.")
			}
		case "-e", "--email", "--e-mail", "--mail":
			if index != len(parameters)-1 {
				newEmail = parameters[index+1]
			} else {
				fmt.Fprintln(writer, "[red]Error, you didn't supply a new e-mail address.")
			}
		case "--new-password", "-np":
			askForNewPassword = true
		}
	}

	if newName == "" && !askForNewPassword && newEmail == "" {
		fmt.Fprintln(writer, "[red]No valid parameters were supplied. See `help user-update` for more information.")
		return
	}

	if newName != "" {
		newName = strings.TrimSpace(newName)
	}

	if newEmail != "" {
		newEmail = strings.TrimSpace(newEmail)
	}

	var newPassword string

	go func() {
		if askForNewPassword {
			newPassword = nick.window.PromptSecretInput("Updating your user information", "Please enter your new password.")
			newPasswordConfirmation := nick.window.PromptSecretInput("Updating your user information", "Please enter your new password again, to make sure it is correct.")

			if newPassword != newPasswordConfirmation {
				fmt.Fprintln(writer, "[red]Error, new passwords differ from each other, please try again.")
				nick.window.ForceRedraw()
				return
			}
		}

		currentPassword := nick.window.PromptSecretInput("Updating your user information", "Please enter your current password.")
		if currentPassword != "" {
			fmt.Fprintln(writer, "[red]Empty password, aborting.")
		} else {
			_, err := nick.session.UserUpdate(newEmail, currentPassword, newName, nick.session.State.User.Avatar, newPassword)
			if err != nil {
				fmt.Fprintln(writer, err)
			} else {
				fmt.Fprintln(writer, "Your user has been updated.")
			}
		}

		nick.window.ForceRedraw()
	}()
}

// Name returns the string that this command can be called by.
func (nick *UserUpdate) Name() string {
	return "user-update"
}

// PrintHelp prints the general help page for this command
func (nick *UserUpdate) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, userUpdateDocumentation)
}
