# Application Profiling

## Step 1: Dependency Gathering

1. Get process information:
   - Executable path
   - Command-line arguments
   - Working directory
   - Environment variables
   - Process owner
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
