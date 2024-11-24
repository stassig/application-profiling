// Step 1: Extract the executable path and command-line for the process from /proc/<pid>/exe and /proc/<pid>/cmdline
// Step 2: Remove trailing empty element if any (\x00)
// Step 3: Remove any values in the cmdline that are before the executable path
// Step 4: Split the values after the executable path into flags and arguments
// Step 5: Wrap any arguments in quotes
// Flags: starting with - or -- (e.g. -f, --force)
// Arguments: valid characters between flags or at the end  (e.g. daemon on; master_process on; , /etc/nginx/nginx.conf)

// Example Flow:

// PID: 1234
// 0. /proc/1234/exe -> /usr/sbin/nginx
// 1. /proc/1234/cmdline -> "nginx: master process /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;""
// 2. Remove any values before the executable path -> "/usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;""
// 3. Split the values after the executable path into flags and arguments -> [--force, -c, /etc/nginx/nginx.conf, -g, daemon on; master_process on;] -> [1, 2, 3, 4, 5] (arguments are any valid characters between flags or at the end)
// 4. Wrap any arguments containing a special shell character/whitespace/tabs/newlines in quotes -> [1, 2, 3, 4, "5"]
// 5. Final command: /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g "daemon on; master_process on;"

package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
    // Example usage with sample data
    cmdline := []byte("nginx: master process /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;")
    exePath := "/usr/sbin/nginx"

    command := ConvertCmdlineToShellCommand(cmdline, exePath)
    fmt.Println(command)
}

// Converts the cmdline and exePath into a working shell command.
func ConvertCmdlineToShellCommand(cmdline []byte, exePath string) string {
    args := strings.Split(string(cmdline), " ")
    if len(args) > 0 && args[len(args)-1] == "" {
        args = args[:len(args)-1]
    }

    exeIndex := -1
    for i, arg := range args {
        if arg == exePath {
            exeIndex = i
            break
        }
    }

    if exeIndex == -1 {
        // Try to find the first argument that starts with '/'
        for i, arg := range args {
            if strings.HasPrefix(arg, "/") {
                exeIndex = i
                break
            }
        }
    }

    command := exePath
    if exeIndex != -1 {
        args = args[exeIndex+1:]
    } else {
        if len(args) > 0 && args[0] == filepath.Base(exePath) {
            args = args[1:]
        }
    }

    // Process the arguments
    for _, arg := range args {
        if needsQuoting(arg) {
            arg = quoteArgument(arg)
        }
        command += " " + arg
    }

    return command
}

// Checks if the argument contains special shell characters or whitespace.
func needsQuoting(s string) bool {
    specialChars := " \t\n\"'`$&*()[]{}<>|\\;?!#~%"
    for _, c := range s {
        if unicode.IsSpace(c) || strings.ContainsRune(specialChars, c) {
            return true
        }
    }
    return false
}

// Wraps the argument in double quotes and escapes inner quotes and backslashes.
func quoteArgument(s string) string {
    var buf bytes.Buffer
    buf.WriteByte('"')
    for _, c := range s {
        if c == '\\' || c == '"' {
            buf.WriteByte('\\')
        }
        buf.WriteRune(c)
    }
    buf.WriteByte('"')
    return buf.String()
}

