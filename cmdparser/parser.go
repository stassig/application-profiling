// TO DO: Refactor extractFlagsAndArguments

package cmdparser

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseCommandLine reconstructs the command string from the executable path and cmdline data.
// It follows the steps of cleaning, splitting, and reformatting arguments as needed.
func ParseCommandLine(executablePath string, commandLineData []byte) string {
	// Step 1: Split the raw cmdline data into arguments
	commandLineArguments := splitCommandLineData(commandLineData)

	// Step 2: Remove any arguments before the executable path
	filteredArguments := filterArgumentsBeforeExecutable(commandLineArguments, executablePath)

	// Step 3: Parse the filtered arguments into flags and their associated values
	flagsAndArguments := extractFlagsAndArguments(filteredArguments)

	// Step 4: Wrap special arguments containing spaces or special characters in quotes
	quotedArguments := quoteSpecialArguments(flagsAndArguments)

	// Step 5: Construct the final command string
	finalCommand := buildCommandString(executablePath, quotedArguments)

	return finalCommand
}

// splitCommandLineData splits the raw cmdline data into individual arguments.
func splitCommandLineData(data []byte) []string {
	// Replace null bytes with spaces and split by whitespace
	processedString := strings.ReplaceAll(string(data), "\x00", " ")
	return strings.Fields(processedString)
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

// extractFlagsAndArguments parses the arguments into a slice of flags and their corresponding values.
func extractFlagsAndArguments(arguments []string) [][2]string {
	fmt.Printf("Arguments: %v\n", arguments)
	var flagsWithArguments [][2]string

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
			flagArguments := strings.Join(collectedArguments, " ")
			flagsWithArguments = append(flagsWithArguments, [2]string{flag, flagArguments})
		} else {
			// Handle standalone arguments (if any, though rare in this case)
			flagsWithArguments = append(flagsWithArguments, [2]string{arguments[i], ""})
			i++
		}
	}
	fmt.Printf("Flags with arguments: %v\n", flagsWithArguments)
	return flagsWithArguments
}

// quoteSpecialArguments wraps arguments containing special characters or spaces in quotes.
func quoteSpecialArguments(flagsAndArguments [][2]string) []string {
	var specialCharacterPattern = regexp.MustCompile(`[^\w@%+=:,./-]`)
	var quotedArguments []string

	for _, pair := range flagsAndArguments {
		flag, argument := pair[0], pair[1]
		quotedArguments = append(quotedArguments, flag)
		if argument != "" {
			if specialCharacterPattern.MatchString(argument) {
				argument = fmt.Sprintf(`"%s"`, argument)
			}
			quotedArguments = append(quotedArguments, argument)
		}
	}
	return quotedArguments
}

// buildCommandString combines the executable path and arguments into a final command string.
func buildCommandString(executablePath string, arguments []string) string {
	return executablePath + " " + strings.Join(arguments, " ")
}
