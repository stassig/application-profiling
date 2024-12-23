package profiler

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseCommandLine reconstructs the command string from the executable path and cmdline data.
// It follows the steps of cleaning, splitting, and reformatting arguments as needed.
func ParseCommandLine(executablePath string, commandLineArguments []string) (string, []FlagArgument) {
	// Step 1: Remove any arguments before the executable path
	filteredArguments := filterArgumentsBeforeExecutable(commandLineArguments, executablePath)

	// Step 2: Parse the filtered arguments into flags and their associated values
	flagsAndArguments := extractFlagsAndArguments(filteredArguments)

	// Step 3: Wrap special arguments containing spaces or special characters in quotes
	quotedArguments := quoteSpecialArguments(flagsAndArguments)

	// Step 4: Construct the final command string
	finalCommand := buildCommandString(executablePath, quotedArguments)

	// Return the reconstructed command and the parsed flags/arguments
	return finalCommand, flagsAndArguments
}

// filterArgumentsBeforeExecutable removes any arguments that occur before the executable path.
func filterArgumentsBeforeExecutable(arguments []string, executablePath string) []string {
	for i, arg := range arguments {
		if arg == executablePath {
			return arguments[i:]
		}
	}
	return arguments
}

// extractFlagsAndArguments parses the arguments into a slice of FlagArgument structs
func extractFlagsAndArguments(arguments []string) []FlagArgument {
	var flagsWithArguments []FlagArgument

	isFlag := func(argument string) bool {
		return strings.HasPrefix(argument, "-")
	}

	i := 1 // Start after the executable path
	for i < len(arguments) {
		if isFlag(arguments[i]) {
			flag := arguments[i]
			var collectedArguments []string

			// Collect all associated arguments for the current flag
			i++
			for i < len(arguments) && !isFlag(arguments[i]) {
				collectedArguments = append(collectedArguments, arguments[i])
				i++
			}

			// Combine collected arguments into a single string
			flagValue := strings.Join(collectedArguments, " ")
			flagsWithArguments = append(flagsWithArguments, FlagArgument{Flag: flag, Value: flagValue})
		} else {
			// Handle standalone arguments (e.g., positional arguments)
			flagsWithArguments = append(flagsWithArguments, FlagArgument{Flag: arguments[i], Value: ""})
			i++
		}
	}
	return flagsWithArguments
}

// quoteSpecialArguments wraps arguments containing special characters or spaces in quotes
func quoteSpecialArguments(flagsAndArguments []FlagArgument) []string {
	var specialCharacterPattern = regexp.MustCompile(`[^\w@%+=:,./-]`)
	var quotedArguments []string

	for _, argument := range flagsAndArguments {
		quotedArguments = append(quotedArguments, argument.Flag)
		if argument.Value != "" {
			if specialCharacterPattern.MatchString(argument.Value) {
				argument.Value = fmt.Sprintf("\"%s\"", argument.Value)
			}
			quotedArguments = append(quotedArguments, argument.Value)
		}
	}
	return quotedArguments
}

// buildCommandString combines the executable path and arguments into a final command string.
func buildCommandString(executablePath string, arguments []string) string {
	return executablePath + " " + strings.Join(arguments, " ")
}
