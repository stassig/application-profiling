# application-profiling

PHASE 1: Dependency Gathering

Step 1: Get process information (executable path, command-line arguments, working directory, environment variables, process owner, sockets, user permissions, etc.)
Step 2: Save process information to a file
Step 3: Restart the process
Step 4: Get new PID and child processes
Step 5: Monitor the new process and log file access
Step 6: Filter the trace log based on the PIDs & clean up duplicate logs

PHASE 2: Dockerization

Step 1: Copy files from trace log to "profiling" directory (ensure working symlinks)
Step 2: Map the "profiling" directory to the Dockerfile
Step 3: Map process info file to the Dockerfile
