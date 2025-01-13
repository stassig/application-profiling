# Workflow Overview

## Phase 1: Profile

Analyze the underlying Unix processes of an application to collect its runtime dependencies required for containerization. ([profile.go](./cmd/commands/profile.go))

1. **Collect Metadata**: Extract process information from `/proc`, including environment variables and ports, and save it to a YAML file. ([info.go](./internal/profiler/info.go))
2. **Restart with Tracing**: Restart the application process with `strace` to monitor runtime behavior . ([restart.go](./internal/profiler/restart.go))
3. **Capture Dependencies**: Log file-related system calls to identify transient dependencies.
4. **Filter Logs**: Clean up trace logs to retain only relevant file paths. ([filter.go](./internal/profiler/filter.go))

## Phase 2: Dockerize

Generate a minimal Docker container configuration from the profiled dependencies. ([dockerize.go](./cmd/commands/dockerize.go))

1. **Load Data**: Import process metadata and file paths from the profiling phase.
2. **Build Filesystem**: Copy identified files and directories into a minimal profile filesystem. ([filesystem.go](./internal/dockerizer/filesystem.go))
3. **Archive and Package**: Create a compressed tarball of the profile filesystem.
4. **Generate Dockerfile**: Create a tailored Dockerfile based on the collected dependencies and runtime configuration. ([generate.go](./internal/dockerizer/generate.go))
