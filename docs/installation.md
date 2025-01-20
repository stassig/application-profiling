# Installation Guide

### Prerequisites

Before proceeding with installation, ensure the following:

- **OS**: Ubuntu 24.04 server (or compatible)
- **Resources**: 2 CPU cores, 4GB RAM, 20GB disk space
- **Access**: Root or sudo privileges

---

### 1. Install Golang

```bash
sudo apt update
sudo apt install -y golang
go version  # Verify installation
```

### 2. Install Docker

Follow the [official Docker installation guide](https://docs.docker.com/engine/install/ubuntu/#install-using-the-repository).

### 3. Install Strace

```bash
sudo apt install -y strace
strace --version # Verify installation
```

### 4. MySQL

Set up MySQL for testing the tool:

```bash
sudo apt install -y mysql-server
sudo systemctl start mysql
sudo systemctl enable mysql
sudo systemctl status mysql # Verify installation
```

### 5. NGINX

Set up NGINX for testing the tool:

```bash
sudo apt install -y nginx
sudo systemctl start nginx
sudo systemctl enable nginx
sudo systemctl status nginx # Verify installation
```

### 6. Build Tool

Clone the repository and build the CLI tool:

```bash
git clone https://github.com/stassig/application-profiling.git
cd application-profiling
go build -o /usr/local/bin/vm2container cmd/main.go
chmod +x /usr/local/bin/vm2container
vm2container --help  # Verify the build
```
