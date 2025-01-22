# NGINX Migration Guide

> **Prerequisites**:
>
> - You have elevated privileges (sudo/root access) to run necessary commands.
> - **NGINX and vm2container are installed** as per the [Installation Guide](installation.md).

📌 **Demo:** [NGINX Migration](https://drive.google.com/file/d/1Kn98CszPfvt1ucZSWKFq31j6Hj7Mc3ow/view?usp=sharing)

Follow these steps to profile a web server and generate a Docker container.

---

### 1. Get the Main PID of the Service

```bash
pgrep -o -x nginx
```

---

### 2. Profile the Service

Run the profiling command to collect runtime dependencies. Replace <PID> with the PID obtained from step 1:

```bash
sudo vm2container profile <PID>
```

This will output the collected data to:

```bash
./output/<PID>/profile
```

---

### 3. Dockerize the Profiled Data

Generate a Docker configuration and minimal filesystem for the application:

```bash
sudo vm2container dockerize <PID>
```

The Docker configuration will be saved to:

```bash
./output/<PID>/dockerize
```

---

### 4. Stop the Restarted Process

After profiling, stop the NGINX process to avoid conflicts:

```bash
pgrep -o -x nginx    # Get the PID
kill <PID>           # Stop the NGINX process
```

---

### 5. Build the Container Image

Build a Docker container image using the generated Dockerfile and profile:

```bash
sudo docker build -t nginx-server ./output/<PID>/dockerize
```

---

### 6. Run the Container

Start the container and bind it to the host network:

```bash
sudo docker run -d --network=host nginx-server
```

---

### 7. Functionality Test

Verify the container is running and serving content:

```bash
curl http://localhost
```

---

### 8. Performance Test (Optional)

Measure the performance of the containerized NGINX server.

Install the wrk benchmarking tool:

```bash
sudo apt install -y wrk
```

Run a performance test for 30 seconds with 2 threads and 100 concurrent connections:

```bash
wrk -t2 -c100 -d30s http://localhost
```
