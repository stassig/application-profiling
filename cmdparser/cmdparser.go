// Step 1: Extract the executable path and command-line for the process from /proc/<pid>/exe and /proc/<pid>/cmdline
// Step 2: Remove any values in the cmdline that are before the executable path
// Step 3: Split the values after the executable path into flags
// Step 4: Wrap flag arguments in quotes, if they contain special shell characters or whitespace
// Flags: starting with - or -- (e.g. -f, --force)
// Arguments: values that follow the flags (e.g. -f value, --force value)

// Example Flow:

// PID: 1234
// 0. /proc/1234/exe -> /usr/sbin/nginx
// 1. /proc/1234/cmdline -> "nginx: master process /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;"
// 2. Remove any values before the executable path -> "/usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;""
// 3. Split the values after the executable path into flags combined with their respective argument(s):

// [[--force ] [-c /etc/nginx/nginx.conf] [-g daemon on; master_process on;]]

// 4. Wrap any arguments containing a special shell character/whitespace/tabs/newlines in quotes
// -> only {argument: daemon on; master_process on; } should be wraped in quotes in this case (contains semi-colons)

// 5. Final command: /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g "daemon on; master_process on;"

// IMPORTANT: Each flag can have more than one argument, so we need to collect all arguments for a flag until the next flag or end of list

package cmdparser

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseCmdline parses the cmdline data and returns the reconstructed command string
func ParseCmdline(exePath string, cmdlineData []byte) string {

	// Step 1: Parse cmdlineData into arguments
	cmdlineArgs := parseCmdline(cmdlineData)

	// Step 2: Remove any values before the executable path
	cmdlineArgs = removeBeforeExe(cmdlineArgs, exePath)

	// Step 3: Split the values after the executable path into flags and arguments
	flagsWithArgs := parseFlagsAndArgs(cmdlineArgs)

	// Step 4: Wrap arguments in quotes if they contain special characters or whitespaces
	finalArgs := wrapSpecialArgs(flagsWithArgs)

	// Step 5: Construct the final command
	finalCmd := constructCommand(exePath, finalArgs)

	return finalCmd
}

// parseCmdline splits the cmdline data into arguments
func parseCmdline(data []byte) []string {
	// Replace null bytes with spaces
	processedString := strings.ReplaceAll(string(data), "\x00", " ")

	// Split the resulting string by whitespace
	args := strings.Fields(processedString)

	return args
}

// removeBeforeExe removes any values before the executable path
func removeBeforeExe(args []string, exePath string) []string {
	for i, arg := range args {
		if arg == exePath {
			return args[i:]
		}
	}
	return args
}

// parseFlagsAndArgs splits the values into flags and their arguments
func parseFlagsAndArgs(args []string) [][2]string {
    fmt.Printf("Args: %v\n", args)
    var result [][2]string
    isFlag := func(s string) bool {
        return strings.HasPrefix(s, "-")
    }

    i := 1 // Start after the executable path
    for i < len(args) {
        if isFlag(args[i]) {
            flag := args[i]
            var argParts []string

            // Collect all arguments for the flag until the next flag or end of list
            i++
            for i < len(args) && !isFlag(args[i]) {
                argParts = append(argParts, args[i])
                i++
            }

            // Combine collected arguments into a single string
            arg := strings.Join(argParts, " ")
            result = append(result, [2]string{flag, arg})
        } else {
            // Handle standalone arguments if any (unlikely in this context)
            result = append(result, [2]string{args[i], ""})
            i++
        }
    }
	fmt.Printf("Result: %v\n", result)
    return result
}


// wrapSpecialArgs wraps arguments containing special characters in quotes
func wrapSpecialArgs(flagsWithArgs [][2]string) []string {
	var specialChars = regexp.MustCompile(`[^\w@%+=:,./-]`)
	var finalArgs []string
	for _, pair := range flagsWithArgs {
		flag, arg := pair[0], pair[1]
		finalArgs = append(finalArgs, flag)
		if arg != "" {
			if specialChars.MatchString(arg) {
				arg = fmt.Sprintf(`"%s"`, arg)
			}
			finalArgs = append(finalArgs, arg)
		}
	}
	return finalArgs
}

// constructCommand builds the final command string
func constructCommand(exePath string, args []string) string {
	return exePath + " " + strings.Join(args, " ")
}