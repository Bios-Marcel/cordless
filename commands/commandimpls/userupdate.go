package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/discordgo"
)

type UserUpdate struct {
	window  *ui.Window
	session *discordgo.Session
}

func NewUserUpdateCommand(window *ui.Window, session *discordgo.Session) *UserUpdate {
	return &UserUpdate{window, session}
}

func (nick *UserUpdate) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		fmt.Println()
		return
	}

	var newName, newPassword, newEmail string
	for index, param := range parameters {
		switch param {
		case "-n", "--name", "--nick", "-u", "--username":
			if index != len(parameters)-1 {
				newName = parameters[index+1]
			}
		case "-e", "--email", "--e-mail", "--mail":
			if index != len(parameters)-1 {
				newEmail = parameters[index+1]
			}
		case "--new-password", "-np":
			if index != len(parameters)-1 {
				newPassword = parameters[index+1]
			}
		}
	}

	go func() {
		currentPassword := nick.window.PromptSecretInput("Updating your user information", "Please enter your current password.")
		if currentPassword != "" {
			//TODO Error
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

func (nick *UserUpdate) Name() string {
	return "user-update"
}

func (nick *UserUpdate) PrintHelp(writer io.Writer) {
	panic("not implemented")
}
