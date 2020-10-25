package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/cordless/commands"
	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
)

const accountDocumentation = `[orange][::u]# account command[white]

The account command allows you to manage multiple discord accounts within
cordless.

Subcommands:
  * add         - Adds a new account
  * delete      - Deletes the given account
  * switch      - Allows you to switch accounts
  * list        - Lists all available accounts
  * current     - Displays the current account
  * add-current - Adds the currently logged in token as a new account
  * logout      - Logs out of the current account logged into cordless
`

// Account manages the users account
type Account struct {
	window        *ui.Window
	accountLogout *AccountLogout
}

// AccountLogout allows logging out of cordless. This clears the token saved
// in the users folder.
type AccountLogout struct {
	window  *ui.Window
	restart func()
}

// NewAccount creates a ready-to-use Account command.
func NewAccount(accountLogout *AccountLogout, window *ui.Window) *Account {
	return &Account{window: window, accountLogout: accountLogout}
}

// NewAccountLogout creates a ready-to-use Logout command.
func NewAccountLogout(restart func(), window *ui.Window) *AccountLogout {
	return &AccountLogout{window: window, restart: restart}
}

// Execute runs the command piping its output into the supplied writer.
func (account *Account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		account.PrintHelp(writer)
	} else {
		switch parameters[0] {
		case "add", "create", "new":
			if len(parameters) != 3 {
				account.printAccountAddHelp(writer)
			} else {
				account.addAccount(writer, parameters[1:])
			}
		case "delete", "remove":
			if len(parameters) != 2 {
				account.printAccountDeleteHelp(writer)
			} else {
				deleteAccount(writer, parameters[1])
			}
		case "switch", "change":
			if len(parameters) != 2 {
				account.printAccountSwitchHelp(writer)
			} else {
				account.switchAccount(writer, parameters[1])
			}
		case "list":
			if len(parameters) != 1 {
				account.printAccountListHelp(writer)
			} else {
				account.listAccounts(writer)
			}
		case "current":
			if len(parameters) != 1 {
				account.printAccountCurrentHelp(writer)
			} else {
				account.currentAccount(writer)
			}
		case "add-current":
			if len(parameters) != 2 {
				account.printAccountAddCurrentHelp(writer)
			} else {
				account.addCurrentAccount(writer, parameters[1])
			}
		case "logout", "sign-out", "signout", "logoff":
			account.accountLogout.Execute(writer, parameters[1:])
		default:
			account.PrintHelp(writer)
		}

	}
}

func (account *Account) printAccountAddHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account add <Name> <Token>")
}

func (account *Account) addAccount(writer io.Writer, parameters []string) {
	newName := strings.ToLower(parameters[0])
	for _, acc := range config.Current.Accounts {
		if acc.Name == newName {
			fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]The name '%s' is already in use.\n", acc.Name)
			return
		}
	}

	newAccount := &config.Account{
		Name:  newName,
		Token: parameters[1],
	}
	config.Current.Accounts = append(config.Current.Accounts, newAccount)
	config.PersistConfig()

	fmt.Fprintf(writer, "The account '%s' has been created successfully.\n", newName)
}

func (account *Account) printAccountDeleteHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage `account delete <Name>`")
}

func deleteAccount(writer io.Writer, account string) {
	var deletionSuccessful bool

	newAccounts := make([]*config.Account, 0)
	for _, acc := range config.Current.Accounts {
		if acc.Name != account {
			newAccounts = append(newAccounts, acc)
		} else {
			deletionSuccessful = true
		}
	}

	if deletionSuccessful {
		fmt.Fprintf(writer, "Account '%s' has been deleted.\n", account)
		config.Current.Accounts = newAccounts
		config.PersistConfig()
	} else {
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Account '%s' could not be found.\n", account)
	}

}

func (account *Account) printAccountSwitchHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account switch <Name>")
}

func (account *Account) switchAccount(writer io.Writer, accountName string) {
	for _, acc := range config.Current.Accounts {
		if acc.Name == accountName {
			oldToken := config.Current.Token
			config.Current.Token = acc.Token
			persistError := account.accountLogout.saveAndRestart(writer)
			if persistError != nil {
				config.Current.Token = oldToken
				commands.PrintError(writer, "Error switching accounts", persistError.Error())
			}
			return
		}
	}

	commands.PrintError(writer, "Error switching accounts", fmt.Sprintf("No account named '%s' was found", accountName))
}

func (account *Account) printAccountListHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account list")
}

func (account *Account) listAccounts(writer io.Writer) {
	fmt.Fprintln(writer, "Available accounts:")
	for _, acc := range config.Current.Accounts {
		if acc.Token == config.Current.Token {
			fmt.Fprintln(writer, "  ["+tviewutil.ColorToHex(config.GetTheme().AttentionColor)+"]> "+acc.Name+"["+tviewutil.ColorToHex(config.GetTheme().PrimaryTextColor)+"]")
		} else {
			fmt.Fprintln(writer, "  * "+acc.Name)
		}
	}
}

func (account *Account) printAccountCurrentHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account current")
}

func (account *Account) currentAccount(writer io.Writer) {
	var currentAccount *config.Account
	for _, acc := range config.Current.Accounts {
		if acc.Token == config.Current.Token {
			currentAccount = acc
			break
		}
	}

	if currentAccount != nil {
		fmt.Fprintf(writer, "Current account '%s'.\n", currentAccount.Name)
	} else {
		fmt.Fprintln(writer, "You have not saved an account with this token.")
	}
}

func (account *Account) printAccountAddCurrentHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account add-current <Name>")
}

func (account *Account) addCurrentAccount(writer io.Writer, name string) {
	account.addAccount(writer, []string{name, config.Current.Token})
}

func (account *Account) Name() string {
	return "account"
}

func (account *Account) Aliases() []string {
	return []string{"profile"}
}

// PrintHelp prints a static help page for this command
func (account *Account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, accountDocumentation)
}

// Execute runs the command piping its output into the supplied writer.
func (accountLogout *AccountLogout) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		accountLogout.PrintHelp(writer)
	} else {
		accountLogout.logout(writer)
	}
}

func (accountLogout *AccountLogout) logout(writer io.Writer) {
	oldToken := config.Current.Token
	config.Current.Token = ""
	err := accountLogout.saveAndRestart(writer)
	if err != nil {
		config.Current.Token = oldToken
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error logging you out '%s'.\n", err.Error())
	}
}

func (accountLogout *AccountLogout) saveAndRestart(writer io.Writer) error {
	persistError := config.PersistConfig()
	if persistError != nil {
		return persistError
	}

	//Using a go routine, so this instance doesn't stay alive and pollutes the memory.
	accountLogout.window.Shutdown()
	accountLogout.restart()

	return nil
}

// PrintHelp prints a static help page for this command
func (accountLogout *AccountLogout) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account logout")
}

// Name returns the primary name for this command. This name will also be
// used for listing the command in the commandlist.
func (accountLogout *AccountLogout) Name() string {
	return "account-logout"
}

// Aliases are a list of aliases for this command. There might be none.
func (accountLogout *AccountLogout) Aliases() []string {
	return []string{"logout"}
}
