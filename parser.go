// TO DO: Look for a library to format text to shell command automatically

// Step 1: Extract the executable path and command-line for the process from /proc/<pid>/exe and /proc/<pid>/cmdline
// Step 2: Remove trailing empty element if any (\x00)
// Step 3: Remove any values in the cmdline that are before the executable path
// Step 4: Split the values after the executable path into flags
// Step 5: Wrap flag arguments in quotes, if they contain special shell characters or whitespace
// Flags: starting with - or -- (e.g. -f, --force)
// Arguments: values that follow the flags (e.g. -f value, --force value)

// Example Flow:

// PID: 1234
// 0. /proc/1234/exe -> /usr/sbin/nginx
// 1. /proc/1234/cmdline -> "nginx: master process /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;"
// 2. Remove any values before the executable path -> "/usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;""
// 3. Split the values after the executable path into flags with their respective values:

// { flag: "-c", argument: "/etc/nginx/nginx.conf" }
// { flag: "--force", argument: null }
// { flag: "-g", argument: "daemon on; master_process on;" }

// 4. Wrap any arguments containing a special shell character/whitespace/tabs/newlines in quotes
// -> only {argument: daemon on; master_process on; } should be wraped in quotes in this case (contains semi-colons)

// 5. Final command: /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g "daemon on; master_process on;"

// IMPORTANT: Flags should act as breakpoints. Any values after a flag should be considered arguments (unless it's another flag). Flag arguments should be treated togather as shown in the example above. Do not split them into separate arguments.

package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	// Hardcoded values for testing
	exePath := "/usr/sbin/nginx"
	cmdlineData := []byte("nginx: master process /usr/sbin/nginx\x00--force\x00-c\x00/etc/nginx/nginx.conf\x00-g\x00daemon on; master_process on;\x00")

	// Step 1: Parse cmdlineData into arguments
	cmdlineArgs := parseCmdline(cmdlineData)

	// Step 2: Remove trailing empty element if any
	cmdlineArgs = removeTrailingEmpty(cmdlineArgs)

	// Step 3: Remove any values before the executable path
	cmdlineArgs = removeBeforeExe(cmdlineArgs, exePath)

	// Step 4: Split the values after the executable path into flags and arguments
	flagsWithArgs := parseFlagsAndArgs(cmdlineArgs)
    fmt.Printf("Flags with arguments: %v\n", flagsWithArgs)

	// Step 5: Wrap arguments in quotes if they contain special characters or whitespace
	finalArgs := wrapSpecialArgs(flagsWithArgs)

	// Construct the final command
	finalCmd := constructCommand(exePath, finalArgs)

	fmt.Println("Final command:", finalCmd)

	// Uncomment the following lines to execute the command
	// cmd := exec.Command(exePath, finalArgs...)
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// err := cmd.Run()
	// if err != nil {
	//     fmt.Println("Error executing command:", err)
	// }
}

// parseCmdline splits the cmdline data into arguments
func parseCmdline(data []byte) []string {
	// Split by null byte
	args := strings.Split(string(data), "\x00")
	return args
}

// removeTrailingEmpty removes the trailing empty element if any
func removeTrailingEmpty(args []string) []string {
	if len(args) > 0 && args[len(args)-1] == "" {
		return args[:len(args)-1]
	}
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
	var result [][2]string
	isFlag := func(s string) bool {
		return strings.HasPrefix(s, "-")
	}

	i := 1 // Start after the executable path
	for i < len(args) {
		if isFlag(args[i]) {
			flag := args[i]
			arg := ""
			// Collect argument(s) for the flag
			if i+1 < len(args) && !isFlag(args[i+1]) {
				arg = args[i+1]
				i += 2
			} else {
				i++
			}
			result = append(result, [2]string{flag, arg})
		} else {
			// Handle standalone arguments if any
			result = append(result, [2]string{args[i], ""})
			i++
		}
	}
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