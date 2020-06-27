package commandimpls

import (
	"fmt"
	"github.com/Bios-Marcel/cordless/commands"
	"io"
	"strings"

	"github.com/Bios-Marcel/cordless/config"
	"github.com/Bios-Marcel/cordless/ui"
	"github.com/Bios-Marcel/cordless/ui/tviewutil"
	"github.com/Bios-Marcel/cordless/util/fuzzy"
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
	window  *ui.Window
	runNext chan bool
	subcommands map[string]commands.Command
}

type add_account struct {}
type add_current_account struct {
	account *Account
}
type current_account struct {}
type delete_account struct {}
type list_account struct {}
type logout_account struct {
	account *Account
}
type switch_account struct {
	account *Account
}

// NewAccount creates a ready-to-use Account command.
func NewAccount(runNext chan bool, window *ui.Window) *Account {
	a := &Account{window: window, runNext: runNext}
	a.subcommands = make(map[string]commands.Command)
	a.subcommands["add"] = &add_account{}
	a.subcommands["add-current"] = &add_current_account{account: a}
	a.subcommands["current"] = &current_account{}
	a.subcommands["delete"] = &delete_account{}
	a.subcommands["list"] = &list_account{}
	a.subcommands["logout"] = &logout_account{account: a}
	a.subcommands["switch"] = &switch_account{account: a}

	for _, cmd := range a.subcommands {
		for _, alias := range cmd.Aliases() {
			a.subcommands[alias] = a.subcommands[cmd.Name()]
		}
	}

	return a
}

// Execute runs the command piping its output into the supplied writer.
func (account *Account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) == 0 {
		account.PrintHelp(writer)
	} else {
		subcommand := account.subcommands[parameters[0]]
		if subcommand != nil {
			if len(parameters) > 1 {
				parameters = parameters[1:]
			} else {
				parameters = []string{}
			}
			subcommand.Execute(writer, parameters)
		} else {
			account.PrintHelp(writer)
		}
	}
}

func (_ add_account) Name() string{
	return "add"
}

func (_ add_account) Aliases() []string{
	return []string{"create", "new"}
}

func (_ add_account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account add <Name> <Token>")
}

func (cmd *add_account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 2 {
		cmd.PrintHelp(writer)
		return
	}
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

func (_ add_account) Complete(cmdline []string) []string {
	return []string{}
}

func (_ delete_account) Name() string{
	return "delete"
}

func (_ delete_account) Aliases() []string {
	return []string{"delete", "remove"}
}

func (_ delete_account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage `account delete <Name>`")
}

func (cmd delete_account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		cmd.PrintHelp(writer)
		return
	}

	var deletionSuccessful bool
	account := parameters[0]

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

func (_ delete_account) Complete(args []string) []string {
	var accountNames []string
	for _, acc := range config.Current.Accounts {
		accountNames = append(accountNames, acc.Name)
	}
	if len(args) == 0 {
		return accountNames
	} else if len(args) > 1 {
		return []string{}
	}
	results := fuzzy.ScoreSearch(args[0], accountNames)
	return fuzzy.RankMap(results)
}

func (_ switch_account) Name() string{
	return "switch"
}

func (_ switch_account) Aliases() []string{
	return []string{"change"}
}

func (_ switch_account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account switch <Name>")
}

func (cmd *switch_account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		cmd.PrintHelp(writer)
		return
	}
	accountName := parameters[0]
	for _, acc := range config.Current.Accounts {
		if acc.Name == accountName {
			oldToken := config.Current.Token
			config.Current.Token = acc.Token
			persistError := cmd.account.saveAndRestart(writer)
			if persistError != nil {
				config.Current.Token = oldToken
				commands.PrintError(writer, "Error switching accounts", persistError.Error())
			}
			return
		}
	}

	commands.PrintError(writer, "Error switching accounts", fmt.Sprintf("No account named '%s' was found", accountName))
}

func (cmd switch_account) Complete(args []string) []string {
	var accountNames []string
	for _, acc := range config.Current.Accounts {
		if acc.Token != config.Current.Token {
			accountNames = append(accountNames, acc.Name)
		}
	}
	if len(args) == 0 {
		return accountNames
	} else if len(args) > 1 {
		return []string{}
	}
	results := fuzzy.ScoreSearch(args[0], accountNames)
	return fuzzy.RankMap(results)
}


func (_ logout_account) Name() string {
	return "logout"
}

func (_ logout_account) Aliases() []string {
	return []string{"logout", "sign-out", "signout", "logoff"}
}

func (_ logout_account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account logout")
}

func (cmd *logout_account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 0 {
		cmd.PrintHelp(writer)
		return
	}
	oldToken := config.Current.Token
	config.Current.Token = ""
	err := cmd.account.saveAndRestart(writer)
	if err != nil {
		config.Current.Token = oldToken
		fmt.Fprintf(writer, "["+tviewutil.ColorToHex(config.GetTheme().ErrorColor)+"]Error logging you out '%s'.\n", err.Error())
	}
}

func (_ logout_account) Complet(_ []string) []string {
	return []string{}
}

func (account *Account) saveAndRestart(writer io.Writer) error {
	persistError := config.PersistConfig()
	if persistError != nil {
		return persistError
	}

	//Using a go routine, so this instance doesn't stay alive and pollutes the memory.
	account.runNext <- true
	account.window.Shutdown()

	return nil
}

func (_ list_account) Name() string {
	return "list"
}

func (_ list_account) Aliases() []string {
	return []string{}
}

func (_ list_account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account list")
}

func (cmd list_account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) > 0 {
		cmd.PrintHelp(writer)
		return
	}

	fmt.Fprintln(writer, "Available accounts:")
	for _, acc := range config.Current.Accounts {
		fmt.Fprintln(writer, "  * "+acc.Name)
	}
}

func (_ list_account) Complete(_ []string) []string {
	return []string{}
}

func (_ current_account) Name() string {
	return "current"
}

func (_ current_account) Aliases() []string {
	return []string{}
}

func (_ current_account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account current")
}

func (cmd current_account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 0 {
		cmd.PrintHelp(writer)
		return
	}
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

func (_ current_account) Complete(_ []string) []string {
	return []string{}
}

func (_ add_current_account) Name() string {
	return "add-current"
}

func (_ add_current_account) Aliases() []string {
	return []string{}
}

func (_ add_current_account) PrintHelp(writer io.Writer) {
	fmt.Fprintln(writer, "Usage: account add-current <Name>")
}

func (cmd *add_current_account) Execute(writer io.Writer, parameters []string) {
	if len(parameters) != 1 {
		cmd.PrintHelp(writer)
		return
	}
	cmd.account.subcommands["add"].Execute(writer, []string{parameters[0], config.Current.Token})
}

func (_ add_current_account) Complete(_ []string) []string {
	return []string{}
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

func (account *Account) Complete(args []string) []string {
	var subcommands []string
	for name, _ := range account.subcommands {
		subcommands = append(subcommands, name)
	}

	if len(args) == 0 {
		return subcommands
	}

	name := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = []string{}
	}

	cmd := account.subcommands[name]
	if cmd != nil {
		// TODO remove after implementing completion for all commands
		cmd, ok := cmd.(commands.Completable)
		if ! ok {
			return []string{}
		}

		var results []string
		for _, c := range cmd.Complete(args) {
			results = append(results, name + " " + c)
		}
		return results
	} else {
		results := fuzzy.ScoreSearch(name, subcommands)
		return fuzzy.RankMap(results)
	}
}
