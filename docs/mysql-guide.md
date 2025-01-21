# MySQL Migration Guide

> **Prerequisites**:
>
> - You have elevated privileges (sudo/root access) to run necessary commands.
> - **MySQL and vm2container are installed** as per the [Installation Guide](installation.md).

📌 **Demo:** [MySQL Migration](https://drive.google.com/file/d/1RttL7_uTeGMbr2d-uKmTwogzitQW8PvT/view?usp=sharing)

Follow these steps to profile a MySQL server and generate a Docker container.

---

### 1. Get PIDs for Profiling

To collect all relevant dependencies, retrieve the PIDs of the MySQL console instance and the main MySQL service. The MySQL console is a standalone instance, not a child process of the main service, so it requires a separate profile.

1. **Open a MySQL Console**: Run this in a separate terminal:
   ```bash
   cd /             # Reset CWD
   sudo mysql       # Open MySQL console
   ```
2. **Get the Console PID**:
   ```bash
   pgrep -u root -o -f "sudo mysql"
   ```
3. **Get the Main Service PID**:
   ```bash
   pgrep -o -x mysqld
   ```

---

### 2. Profile the Service

Run the profiling command to collect runtime dependencies. Replace `<console-pid>` and `<mysql-pid>` with the PIDs obtained in step 1:

```bash
sudo vm2container profile <console-pid>,<mysql-pid>
```

This will output the collected data to:

```
./output/<mysql-pid>/profile
```

---

### 3. Dockerize the Profiled Data

Generate a Docker configuration and minimal filesystem:

```bash
sudo vm2container dockerize <mysql-pid>
```

The Docker configuration will be saved to:

```
./output/<mysql-pid>/dockerize
```

---

### 4. Stop the Restarted Process

After profiling, stop the MySQL service to avoid conflicts:

```bash
pgrep -o -x mysqld   # Get the PID
kill <PID>
```

---

### 5. Build the Container Image

Build a Docker image using the generated profile and Dockerfile:

```bash
sudo docker build -t mysql-server ./output/<mysql-pid>/dockerize
```

---

### 6. Run the Container

Start the container and bind it to the host network:

```bash
sudo docker run -d --network=host mysql-server
```

---

### 7. Functionality Test

Verify the container is running and MySQL is functional:

1. List running containers:
   ```bash
   sudo docker ps
   ```
2. Access the container:
   ```bash
   sudo docker exec -it --user=root <container-id> /bin/bash
   ```
3. Inside the container, open the MySQL console:
   ```bash
   mysql
   ```
4. In the MySQL console, verify functionality:
   ```sql
   SHOW DATABASES;
   ```

---

### 8. Performance Test (Optional)

#### Prepare the Test

1. Install `sysbench` on the host system:
   ```bash
   sudo apt install -y sysbench
   ```
2. Inside the container, create a test database:
   ```bash
   sudo mysql
   CREATE DATABASE sbtest;
   ```

#### Run the Benchmark

1. Prepare the test:
   ```bash
   sysbench \
     --db-driver=mysql \
     --mysql-user=root \
     --mysql-password='' \
     --mysql-host=127.0.0.1 \
     --mysql-port=3306 \
     --mysql-db=sbtest \
     oltp_read_write \
     prepare
   ```
2. Run the test:
   ```bash
   sysbench \
     --db-driver=mysql \
     --mysql-user=root \
     --mysql-password='' \
     --mysql-host=127.0.0.1 \
     --mysql-port=3306 \
     --mysql-db=sbtest \
     --time=60 \
     --threads=4 \
     oltp_read_write \
     run
   ```
