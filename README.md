# Application Profiling

This command-line tool automates the migration of Linux-based VM applications to containers using a process-level profiling approach. It identifies runtime dependencies to generate lean and optimized Docker containers.

## Workflow Overview

### Phase 1: Profile

Analyze the underlying Unix processes of an application to collect its runtime dependencies required for containerization. ([profile.go](./cmd/commands/profile.go))

1. **Collect Metadata**: Extract process information from `/proc`, including environment variables and ports, and save it to a YAML file. ([info.go](./internal/profiler/info.go))
2. **Restart with Tracing**: Restart the application process with `strace` to monitor runtime behavior . ([restart.go](./internal/profiler/restart.go))
3. **Capture Dependencies**: Log file-related system calls to identify transient dependencies.
4. **Filter Logs**: Clean up trace logs to retain only relevant file paths. ([filter.go](./internal/profiler/filter.go))

---

### Phase 2: Dockerize

Generate a minimal Docker container configuration from the profiled dependencies. ([dockerize.go](./cmd/commands/dockerize.go))

1. **Load Data**: Import process metadata and file paths from the profiling phase.
2. **Build Filesystem**: Copy identified files and directories into a minimal profile filesystem. ([filesystem.go](./internal/dockerizer/filesystem.go))
3. **Archive and Package**: Create a compressed tarball of the profile filesystem.
4. **Generate Dockerfile**: Create a tailored Dockerfile based on the collected dependencies and runtime configuration. ([generate.go](./internal/dockerizer/generate.go))

---

## Installation Guide

### Prerequisites

Before proceeding with installation, ensure the following:

- **Operating System**: Ubuntu 24.04 server or a compatible Ubuntu distribution
- **System Resources**:
  - 2 CPU cores
  - 4GB RAM
  - 20GB disk space
- **Root/Sudo Access**: Required for installing software and managing services

---

### 1. Golang

Download and install Golang from the [official website](https://golang.org/dl/), or use:

```bash
sudo apt update
sudo apt install -y golang
```

Verify the installation:

```bash
go version
```

### 2. Docker

Follow the [official Docker installation guide](https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository).

```bash
docker --version
```

### 3. Strace

Install the tracing tool for capturing runtime dependencies:

```bash
sudo apt install -y strace
```

### 4. MySQL

Set up MySQL for testing the tool:

```bash
sudo apt install -y mysql-server
sudo systemctl start mysql
sudo systemctl enable mysql
```

Verify MySQL status:

```bash
sudo systemctl status mysql
```

### 5. NGINX

Set up NGINX for testing the tool:

```bash
sudo apt install -y nginx
sudo systemctl start nginx
sudo systemctl enable nginx
```

Verify NGINX status:

```bash
sudo systemctl status nginx
```

### 6. Build Tool

Clone the repository and build the CLI tool:

```bash
git clone git@github.com:stassig/application-profiling.git
cd application-profiling
go build -o vm2container cmd/main.go
```

Verify the installation:

```bash
./vm2container --help
```

---

## Usage Instuctions

### Case 1: NGINX

> **Note**: This guide assumes that NGINX is installed and running as per the [Installation Guide](#installation--setup).

Follow these steps to profile an NGINX web server and generate a Docker container.

#### 1. Get the Main PID of the NGINX Service

```bash
pgrep -o -x nginx
```

#### 2. Profile the NGINX Service

Run the profiling command to collect runtime dependencies. Replace <PID> with the PID obtained from step 1:

```bash
./vm2container profile <PID>
```

This will output the collected data to:

```bash
./output/<PID>/profile
```

#### 3. Dockerize the Profiled Data

Generate a Docker configuration and minimal filesystem for the application:

```bash
./vm2container dockerize <PID>
```

The Docker configuration will be saved to:

```bash
./output/<PID>/dockerize
```

####

#### 4. Stop the Restarted NGINX Process

After profiling, stop the NGINX process to avoid conflicts:

```bash
pgrep -o -x nginx    # Get the PID
kill <PID>           # Stop the NGINX process
```

#### 5. Build the Container Image

Build a Docker container image using the generated Dockerfile and profile:

```bash
docker build -t nginx-server ./output/<PID>/dockerize
```

#### 6. Run the Container

Start the container and bind it to the host network:

```bash
docker run -d --network=host nginx-server
```

#### 7. Functionality Test

Verify the container is running and serving content:

```bash
curl http://localhost
```

#### 8. Performance Test

(Optional) Measure the performance of the containerized NGINX server.

Install the wrk benchmarking tool:

```bash
sudo apt install -y wrk
```

Run a performance test for 30 seconds with 2 threads and 100 concurrent connections:

```bash
wrk -t2 -c100 -d30s http://localhost
```
