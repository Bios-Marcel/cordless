package commands

import (
	"io"
	"strings"
)

// Command represents a command that is executable by the user.
type Command interface {
	// Execute runs the command piping its output into the supplied writer.
	Execute(writer io.Writer, parameters []string)

	// Name represents this commands indentifier.
	Name() string

	// PrintHelp prints a static help page for this command
	PrintHelp(writer io.Writer)
}

// ParseCommand takes an arbitrary input string and splits it into parameters.
// The first parameter (index 0) will always be the command itself.
func ParseCommand(input string) []string {
	if len(input) == 0 || len(strings.TrimSpace(input)) == 0 {
		return nil
	}

	if !strings.ContainsRune(input, ' ') {
		return []string{input}
	}

	parameters := make([]string, 0)

	trimmedWhiteSpace := []rune(strings.TrimSpace(input))
	length := len(trimmedWhiteSpace)

	lastArgument := make([]rune, 0)

OUTER_LOOP:
	for index := 0; index < length; index++ {
		char := trimmedWhiteSpace[index]
		if char == ' ' {
			if len(lastArgument) > 0 {
				parameters = append(parameters, string(lastArgument))
				lastArgument = make([]rune, 0)
			}
		} else if char == '\\' {
			if index == length-1 {
				lastArgument = append(lastArgument, char)
			} else if trimmedWhiteSpace[index+1] == '"' {
				lastArgument = append(lastArgument, '"')
				index++
				continue OUTER_LOOP
			}
		} else if char == '"' {
			if index == 0 || trimmedWhiteSpace[index] != '\\' {
				for index2 := index + 1; index2 < length; index2++ {
					nextChar := trimmedWhiteSpace[index2]
					if nextChar == '"' && trimmedWhiteSpace[index2-1] != '\\' {
						lastArgument = trimmedWhiteSpace[index+1 : index2]
						replacedEscapedQuotes := strings.Replace(string(lastArgument), "\\\"", "\"", -1)
						parameters = append(parameters, replacedEscapedQuotes)
						lastArgument = make([]rune, 0)
						index = index2
						continue OUTER_LOOP
					}
				}

				//No quoting end found
				lastArgument = append(lastArgument, char)
			}
		} else {
			lastArgument = append(lastArgument, char)
		}
	}

	if len(lastArgument) > 0 {
		parameters = append(parameters, string(lastArgument))
	}

	return parameters
}
