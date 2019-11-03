package commandimpls

import (
	"fmt"
	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/cordless/util/text"
	"github.com/Bios-Marcel/discordgo"
	"io"
)

type TFAEnableCmd struct {
	window *ui.Window
	session  *discordgo.Session
}

func NewTFAEnableCommand(window *ui.Window, session *discordgo.Session) *TFAEnableCmd {
	return &TFAEnableCmd{window, session}
}

func (cmd *TFAEnableCmd) Execute(writer io.Writer, parameters []string) {
	if cmd.session.MFA {
		fmt.Fprintln(writer, "TFA is already enabled on this account.")
	} else {
		cmd.window.ShowTFASetup()
	}
}

func (cmd *TFAEnableCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "TODO")
}

func (cmd *TFAEnableCmd) Name() string {
	return "tfa-enable"
}

func (cmd *TFAEnableCmd) Aliases() []string {
	return []string{"mfa-enable", "totp-enable", "2fa-enable", "mfa-activate", "totp-activate", "2fa-activate"}
}

type TFADisableCmd struct {
	session *discordgo.Session
}

func NewTFADisableCommand(session *discordgo.Session) *TFADisableCmd {
	return &TFADisableCmd{session}
}

func (cmd *TFADisableCmd) Execute(writer io.Writer, parameters []string) {
	if cmd.session.State.User.MFAEnabled {
		if len(parameters) != 1 {
			fmt.Fprintln(writer, "Usage: tfa-disable <TFA Token>")
		} else {
			code, parseError := text.ParseTFACode(parameters[0])
			if parseError != nil {
				commands.PrintError(writer, "Error disabling Two-Factor-Authentication", parseError.Error())
			}
			disableError := cmd.session.TwoFactorDisable(code)
			if disableError != nil {
				commands.PrintError(writer, "Error disabling Two-Factor-Authentication", disableError.Error())
			} else {
				config.UpdateCurrentToken(cmd.session.Token)
				configError := config.PersistConfig()
				if configError != nil {
					commands.PrintError(writer, "Error updating access token in configuration. You might have to log in again.", disableError.Error())
				}
			}
		}
	} else {
		fmt.Fprintln(writer, "TFA isn't enabled on this account.")
	}
}

func (cmd *TFADisableCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "TODO")
}

func (cmd *TFADisableCmd) Name() string {
	return "tfa-disable"
}

func (cmd *TFADisableCmd) Aliases() []string {
	return []string{"mfa-disable", "totp-disable", "2fa-disable", "mfa-activate", "totp-activate", "2fa-activate"}
}

type TFABackupGetCmd struct {
	window  *ui.Window
	session *discordgo.Session
}

func NewTFABackupGetCmd(session *discordgo.Session, window *ui.Window) *TFABackupGetCmd {
	return &TFABackupGetCmd{window, session}
}

func (cmd *TFABackupGetCmd) Execute(writer io.Writer, parameters []string) {
	if !cmd.session.MFA {
		fmt.Fprintln(writer, "Two-Factor-Authentication isn't enabled on this account.")
	} else {
		go func() {

			currentPassword := cmd.window.PromptSecretInput("Retrieving TFA backup codes", "Please enter your current password.")
			if currentPassword == "" {
				fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Empty password, aborting.")
			} else {
				codes, codeError := cmd.session.GetTwoFactorBackupCodes(currentPassword)
				if codeError != nil {
					commands.PrintError(writer, "Error retrieving TFA backup codes", codeError.Error())
				} else {
					for _, code := range codes {
						fmt.Fprintf(writer, "Code: %s  | Already used: %v\n", code.Code, code.Consumed)
					}
				}
			}

			cmd.window.ForceRedraw()
		}()
	}
}

func (cmd *TFABackupGetCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "TODO")
}

func (cmd *TFABackupGetCmd) Name() string {
	return "tfa-backup-get"
}

func (cmd *TFABackupGetCmd) Aliases() []string {
	return []string{"mfa-backup-get", "2fa-backup-get", "totp-backup-get"}
}

type TFABackupResetCmd struct {
	window  *ui.Window
	session *discordgo.Session
}

func NewTFABackupResetCmd(session *discordgo.Session, window *ui.Window) *TFABackupResetCmd {
	return &TFABackupResetCmd{window, session}
}

func (cmd *TFABackupResetCmd) Execute(writer io.Writer, parameters []string) {
	if !cmd.session.MFA {
		fmt.Fprintln(writer, "Two-Factor-Authentication isn't enabled on this account.")
	} else {
		go func() {
			currentPassword := cmd.window.PromptSecretInput("Resetting TFA backup codes", "Please enter your current password.")
			if currentPassword == "" {
				fmt.Fprintln(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Empty password, aborting.")
			} else {
				codes, codeError := cmd.session.RegenerateTwoFactorBackupCodes(currentPassword)
				if codeError != nil {
					commands.PrintError(writer, "Error resetting TFA backup codes", codeError.Error())
				} else {
					fmt.Fprintln(writer, "Newly generated codes:")
					for _, code := range codes {
						fmt.Fprintf(writer, "    Code: %s  | Already used: %v\n", code.Code, code.Consumed)
					}
				}
			}

			cmd.window.ForceRedraw()
		}()
	}
}

func (cmd *TFABackupResetCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "TODO")
}

func (cmd *TFABackupResetCmd) Name() string {
	return "tfa-backup-Reset"
}

func (cmd *TFABackupResetCmd) Aliases() []string {
	return []string{"mfa-backup-reset", "2fa-backup-reset", "totp-backup-reset"}
}
