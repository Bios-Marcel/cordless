package commandimpls

import (
	"fmt"
	"io"
	"strings"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
)

const accountDocumentation = `[orange][::u]# account command[white]

The account command allows you to manage multiple discord accounts within
cordless.

Subcommands:
  * add         - Adds a new account
  * delete      - Deletes the given account
  * switch      - Allows you to switch your active account
  * list        - Lists all available accounts
  * current     - Displays the currently used account
  * add-current - Add the token currently in use as a new account
`

// Account manages the users account
type Account struct {
	window  *ui.Window
	runNext chan bool
}

// NewAccount creates a ready-to-use Account command.
func NewAccount(runNext chan bool, window *ui.Window) *Account {
	return &Account{window: window, runNext: runNext}
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
				account.addAcount(writer, parameters[1:])
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
		}
	}
}

func (account *Account) printAccountAddHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account add <Name> <Token>")
}

func (account *Account) addAcount(writer io.Writer, parameters []string) {
	newName := strings.ToLower(parameters[0])
	for _, acc := range config.GetConfig().Accounts {
		if acc.Name == newName {
			fmt.Fprintf(writer, "[red]The name '%s' is already in use.\n", acc.Name)
			return
		}
	}

	newAccount := &config.Account{
		Name:  newName,
		Token: parameters[1],
	}
	config.GetConfig().Accounts = append(config.GetConfig().Accounts, newAccount)
	config.PersistConfig()

	fmt.Fprintf(writer, "The account '%s' has been created successfully.\n", newName)
}

func (account *Account) printAccountDeleteHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage `account delete <Name>`")
}

func deleteAccount(writer io.Writer, account string) {
	var deletionSuccessful bool

	newAccounts := make([]*config.Account, 0)
	for _, acc := range config.GetConfig().Accounts {
		if acc.Name != account {
			newAccounts = append(newAccounts, acc)
		} else {
			deletionSuccessful = true
		}
	}

	if deletionSuccessful {
		fmt.Fprintf(writer, "Account '%s' has been deleted.\n", account)
		config.GetConfig().Accounts = newAccounts
		config.PersistConfig()
	} else {
		fmt.Fprintf(writer, "[red]Account '%s' could not be found.\n", account)
	}

}

func (account *Account) printAccountSwitchHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account switch <Name>")
}

func (account *Account) switchAccount(writer io.Writer, accountName string) {
	var newToken string
	for _, acc := range config.GetConfig().Accounts {
		if acc.Name == accountName {
			newToken = acc.Token
			break
		}
	}

	config.GetConfig().Token = newToken
	persistError := config.PersistConfig()
	if persistError != nil {
		fmt.Fprintf(writer, "[red]Error switching accounts '%s'.\n", persistError.Error())
	} else {
		//Using a go routine, so this instance doesn't stay alive and pollutes the memory.
		account.runNext <- true
		account.window.Shutdown()
	}

}

func (account *Account) printAccountListHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account list")
}

func (account *Account) listAccounts(writer io.Writer) {
	fmt.Fprintln(writer, "Available accounts:")
	for _, acc := range config.GetConfig().Accounts {
		fmt.Fprintln(writer, "  * "+acc.Name)
	}
}

func (account *Account) printAccountCurrentHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account current")
}

func (account *Account) currentAccount(writer io.Writer) {
	var currentAccount *config.Account
	for _, acc := range config.GetConfig().Accounts {
		if acc.Token == config.GetConfig().Token {
			currentAccount = acc
			break
		}
	}

	if currentAccount != nil {
		fmt.Fprintf(writer, "Current account '%s'.\n", currentAccount.Name)
	} else {
		fmt.Fprintln(writer, "Current token belongs to no saved account.")
	}
}

func (account *Account) printAccountAddCurrentHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account add-current <Name>")
}

func (account *Account) addCurrentAccount(writer io.Writer, name string) {
	account.addAcount(writer, []string{name, config.GetConfig().Token})
}

// Name represents this commands indentifier.
func (account *Account) Name() string {
	return "account"
}

// PrintHelp prints a static help page for this command
func (account *Account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, accountDocumentation)
}
