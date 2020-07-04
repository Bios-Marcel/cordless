package commandimpls

import (
	"fmt"
	"io"

	"github.com/Bios-Marcel/discordgo"

	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/cordless/util/text"
)

const tfaHelpPage = `[::b]NAME
	tfa - allows you to manage two-factor-authentication on your account

[::b]SYNOPSIS
	[::b]tfa[::-] <subcommand [args[]>

[::b]DESCRIPTION
	The tfa command allows you to enable / disable TFA and retrieve or
	reset your TFA.backup codes. Some actions require you to either input a
	valid TFA code or your current password.

[::b]SUBCOMMANDS
	[::b]tfa-enable
		enables two-factor-authentication on your account
	[::b]tfa-disable
		disables two-factor-authentication on your account
	[::b]tfa-backup-get
		retrieves your TFA backup codes from discord
	[::b]tfa-backup-reset
		resets and retrieves your TFA backup codes from discord`

const tfaEnableHelpPage = `[::b]NAME
	tfa-enable - enables two-factor-authentication on your account

[::b]SYNOPSIS
	[::b]tfa-enable[::-]

[::b]DESCRIPTION
	This command will open a view that shows a QR-code and instructions
    on how to proceed in order to enable TFA on your discord account.`

const tfaDisableHelpPage = `[::b]NAME
	tfa-disable - disables two-factor-authentication on your account

[::b]SYNOPSIS
	[::b]tfa-disable[::-] <TFA Code>

[::b]DESCRIPTION
	This command will disable TFA on your discord account. In order to disable
	TFA, you need to pass a valid TFA code for confirming that you actually own
	the currently registered TFA secret.

[::b]EXAMPLES
	[gray]$ tfa-disable 123456
	[gray]$ tfa-disable "123 456"`

const tfaBackupGetHelpPage = `[::b]NAME
	tfa-backup-get - retrieves your TFA backup codes from discord

[::b]SYNOPSIS
	[::b]tfa-backup-get[::-]

[::b]DESCRIPTION
	This command will retrieve your TFA backup codes from discord. Those can
	be used in order to recover your account in case you've	lost your active
	TFA device. In order to retrieve the codes, you need to	supply your
	current password.`

const tfaBackupResetHelpPage = `[::b]NAME
	tfa-backup-reset - resets and retrieves your TFA backup codes from discord

[::b]SYNOPSIS
	[::b]tfa-backup-reset[::-]

[::b]DESCRIPTION
	This command will reset and retrieve your TFA backup codes from discord.
	Those can be used in order to recover your account in case you've lost
	your active TFA device. In order to retrieve the codes, you need to
	supply your current password. If you still have unused backup codes
	lying around, those will be invalidated and only the newly returned ones
	can be used.`

type TFACmd struct {
	tfaEnable      *TFAEnableCmd
	tfaDisable     *TFADisableCmd
	tfaBackupGet   *TFABackupGetCmd
	tfaBackupReset *TFABackupResetCmd
}

func NewTFACommand(tfaEnable *TFAEnableCmd, tfaDisable *TFADisableCmd, tfaBackupGet *TFABackupGetCmd, tfaBackupReset *TFABackupResetCmd) *TFACmd {
	return &TFACmd{tfaEnable, tfaDisable, tfaBackupGet, tfaBackupReset}
}

func (cmd *TFACmd) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		cmd.PrintHelp(writer)
	} else {
		combinedName := cmd.Name() + "-" + parameters[0]
		if commands.CommandEquals(cmd.tfaEnable, combinedName) {
			cmd.tfaEnable.Execute(writer, parameters[1:])
		} else if commands.CommandEquals(cmd.tfaDisable, combinedName) {
			cmd.tfaDisable.Execute(writer, parameters[1:])
		} else if commands.CommandEquals(cmd.tfaBackupGet, combinedName) {
			cmd.tfaBackupGet.Execute(writer, parameters[1:])
		} else if commands.CommandEquals(cmd.tfaBackupReset, combinedName) {
			cmd.tfaBackupReset.Execute(writer, parameters[1:])
		} else {
			cmd.PrintHelp(writer)
		}
	}
}

func (cmd *TFACmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, tfaHelpPage)
}

func (cmd *TFACmd) Name() string {
	return "tfa"
}

func (cmd *TFACmd) Aliases() []string {
	return []string{"mfa", "2fa", "totp"}
}

type TFAEnableCmd struct {
	window  *ui.Window
	session *discordgo.Session
}

func NewTFAEnableCommand(window *ui.Window, session *discordgo.Session) *TFAEnableCmd {
	return &TFAEnableCmd{window, session}
}

func (cmd *TFAEnableCmd) Execute(writer io.Writer, parameters []string) {
	if cmd.session.MFA {
		fmt.Fprintln(writer, "TFA is already enabled on this account.")
	} else {
		tfaError := cmd.window.ShowTFASetup()
		if tfaError != nil {
			commands.PrintError(writer, "error showing tfa gui", tfaError.Error())
		}
	}
}

func (cmd *TFAEnableCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, tfaEnableHelpPage)
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
	if cmd.session.MFA {
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
					commands.PrintError(writer, "Error updating access token in configuration. You might have to log in again.", configError.Error())
				}
			}
		}
	} else {
		fmt.Fprintln(writer, "TFA isn't enabled on this account.")
	}
}

func (cmd *TFADisableCmd) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, tfaDisableHelpPage)
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
	fmt.Fprintln(writer, tfaBackupGetHelpPage)
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
	fmt.Fprintln(writer, tfaBackupResetHelpPage)
}

func (cmd *TFABackupResetCmd) Name() string {
	return "tfa-backup-Reset"
}

func (cmd *TFABackupResetCmd) Aliases() []string {
	return []string{"mfa-backup-reset", "2fa-backup-reset", "totp-backup-reset"}
}
