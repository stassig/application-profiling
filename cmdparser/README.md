// Example Flow:
//
// Process ID (PID): 1234
//
// 1. Retrieve the executable path:
//    /proc/1234/exe -> /usr/sbin/nginx
//
// 2. Retrieve the command-line arguments:
//    /proc/1234/cmdline -> "nginx: master process /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;"
//
// 3. Remove any values before the executable path:
//    "/usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;"
//
// 4. Split values after the executable path into flags and their arguments:
//    [
//      [--force ],
//      [-c /etc/nginx/nginx.conf],
//      [-g daemon on; master_process on;]
//    ]
//
// 5. Wrap arguments with special characters (e.g., spaces, tabs, newlines) in quotes:
//    For example, "daemon on; master_process on;" contains a semicolon and requires wrapping.
//
//    Result:
//    /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g "daemon on; master_process on;"
//
// Notes:
// - Each flag can have multiple arguments. Continue collecting arguments for a flag until the next flag or the end of the list.
// - Proper handling of special characters ensures the reconstructed command is shell-compatible.
