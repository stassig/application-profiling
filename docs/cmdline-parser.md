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
