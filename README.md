# Application Profiling

## Step 1: Dependency Gathering

1. Get process information:
   - Executable path
   - Command-line arguments
   - Working directory
   - Environment variables
   - Process User & Group
   - Sockets
   - User permissions, etc.
2. Save process information to a file.
3. Restart the process.
4. Get new PID and child processes.
5. Monitor the new process and log file access.
6. Filter the trace log based on the PIDs and clean up duplicate logs.

## Step 2: Dockerization

1. Copy files from the trace log to the `profiling` directory (ensure working symlinks).
2. Map the `profiling` directory to the Dockerfile.
3. Map the process info file to the Dockerfile.

# Command-line Parser

### 1. Retrieve the executable path

```bash
/proc/<PID>/exe -> /usr/sbin/nginx
```

### 2. Retrieve the command-line arguments

```bash
/proc/<PID>/cmdline -> "nginx: master process /usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;"
```

### 3. Clean up the command

Remove any values before the executable path:

```bash
/usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g daemon on; master_process on;
```

### 4. Split values into flags and arguments

```bash
[ [--force], [-c /etc/nginx/nginx.conf], [-g daemon on; master_process on;] ]
```

### 5. Handle special characters in arguments

Wrap arguments with special characters (e.g., spaces, tabs, newlines) in quotes.  
For example, `"daemon on; master_process on;"` contains a semicolon and requires wrapping.

**Result:**

```bash
/usr/sbin/nginx --force -c /etc/nginx/nginx.conf -g "daemon on; master_process on;"
```

---

### Notes

- Flags may have multiple arguments. Continue collecting arguments for a flag until encountering the next flag or the end of the list.
- Proper handling of special characters ensures the reconstructed command is shell-compatible.
