# Advice Report

## Introduction

This document outlines the current limitations and areas for improvement in the **Application Profiling** CLI tool. While the tool successfully automates the migration of NGINX and MySQL workloads from Ubuntu servers to Docker containers, certain challenges remain. Key areas covered include process restart consistency, dependency filtering, and expanding validation to more applications. Additionally, the document discusses the need for CI/CD, handling non-deterministic workloads, and security considerations. These points outline the next steps required to improve accuracy and reliability.

## 1. Process Restart

Restarting processes is necessary to capture startup dependencies (e.g., shared libraries loaded via `ld.so`), but doing so on a production system carries risks. A safer approach is to set up a replica of the source VM so profiling can be done without disrupting live services. This way, restarts can be performed as needed without concern for system stability.

## 2. Hardcoded Filters

The current approach uses `strace` to track all file-related syscalls, but this results in excessive noise, including system-wide paths (`/proc`, `/sys`, `/dev`) that are not relevant to migration. Instead of hardcoding filters, it would be better to refine `strace` parameters to capture only key syscalls such as `open`, `execve`, and `readlink`. This would reduce noise and make the approach more adaptable across different Linux environments.

## 3. Standalone Processes

Some applications, like MySQL, launch separate processes that don’t follow a parent-child hierarchy, making them difficult to track using `strace` with `-f` (follow child processes). A better alternative could be `bpftrace`, which allows tracking all instances of a target application by filtering based on process name (`comm == "mysql"`). However, identifying the representative application name for filtering might not always be straightforward. One option could be to use the executable file name from `/proc/<pid>/exe` as input to ensure that all relevant processes are captured.

## 4. Broader Validation

So far, profiling has only been tested on NGINX and MySQL. To refine dependency discovery, it would be useful to test additional single-container applications like Redis, Apache, and PostgreSQL. A logical next step would be to extend the approach to multi-container workloads like Elasticsearch, which requires container orchestration via `docker-compose`. This would help evaluate how well the solution scales beyond single-process applications.

## 5. Linux Distributions

The current method is validated only on Ubuntu, so its compatibility with other Linux distributions is still unknown. Testing on other distributions (CentOS, Debian, Alpine) would ensure compatibility across different package managers (`apt`, `yum`, `apk`) and system layouts. This would also help identify any OS-specific differences that might affect profiling accuracy.

## 6. Automated Testing

A CI pipeline should be set up to automate the validation of profiling results. It should recreate the case studies: deploy an Ubuntu VM, install dependencies, run the profiling process, generate a container, and verify that the application functions correctly. To further improve accuracy, a comparison step could be added to check the expected container filesystem against the actual generated one. This would help identify any gaps in the profiling solution, catch issues early, and ensure the tool remains reliable as changes are made.

## 7. Performance Profiling

Currently, resource usage is captured as a single snapshot, which doesn’t always reflect how an application behaves over time. A more accurate approach would be to use continuous monitoring to track CPU, memory, and disk usage throughout its runtime. This way, resource demands can be observed under different conditions to help create a more reliable and well-optimized container that performs consistently across varying workloads.

## 8. Non-Deterministic Apps

Non-deterministic applications load dependencies only when certain conditions are met. For example, NGINX does not always load HTML files unless specific HTTP requests trigger them, meaning some dependencies might be missed during profiling. In contrast, deterministic applications like databases tend to be more predictable, making them less prone to this issue. A potential solution could be to use machine learning techniques to infer missing dependencies based on historical profiling data.

## 9. Security

Security in the generated container should be at least equivalent to the original VM. User permissions are already addressed in the implementation, so access controls remain consistent post-migration. However, firewall rules (`iptables`, `nftables`) still need profiling to make sure network security settings aren’t lost during migration. A good next step would be integrating vulnerability scans (`trivy`, `grype`) into the existing workflow to automatically check the generated Docker images for known security risks before deployment.
