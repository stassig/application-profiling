package docker

import (
	"fmt"
	"os"
	"strings"

	"application_profiling/internal/profiler"
)

// writeBlock writes a block to the Dockerfile with an optional comment and blank line for spacing.
func writeBlock(builder *strings.Builder, comment, content string) {
	if comment != "" {
		fmt.Fprintf(builder, "# %s\n", comment)
	}
	fmt.Fprintln(builder, content)
	fmt.Fprintln(builder) // Add a blank line for spacing
}

// GenerateDockerfile generates a Dockerfile based on process information
func GenerateDockerfile(info *profiler.ProcessInfo, dockerfilePath, tarFile string) error {
	dockerfile := &strings.Builder{}

	writeBlock(dockerfile, "", "FROM ubuntu:latest")
	writeBlock(dockerfile, "Copy the profile archive", fmt.Sprintf("COPY %s /", tarFile))
	writeBlock(dockerfile, "Extract the profile and clean up the archive",
		fmt.Sprintf("RUN tar --skip-old-files -xvf /%s -C / && rm /%s", tarFile, tarFile))

	// Add environment variables
	if len(info.EnvironmentVariables) > 0 {
		for _, env := range info.EnvironmentVariables {
			if strings.Contains(env, "=") {
				writeBlock(dockerfile, "", fmt.Sprintf("ENV %s", env))
			}
		}
	}

	// Set user and group
	userSpec := info.ProcessOwner
	if userSpec == "" {
		userSpec = "root:root"
	} else if !strings.Contains(userSpec, ":") {
		userSpec = userSpec + ":" + userSpec
	}
	writeBlock(dockerfile, "Set the user and group", fmt.Sprintf("USER %s", userSpec))

	// Set working directory
	if info.WorkingDirectory != "" {
		writeBlock(dockerfile, "Set the working directory", fmt.Sprintf("WORKDIR %s", info.WorkingDirectory))
	}

	// Expose ports
	if len(info.ListeningTCP) > 0 || len(info.ListeningUDP) > 0 {
		for _, port := range info.ListeningTCP {
			writeBlock(dockerfile, "Set TCP ports", fmt.Sprintf("EXPOSE %d/tcp", port))
		}
		for _, port := range info.ListeningUDP {
			writeBlock(dockerfile, "Set UDP ports", fmt.Sprintf("EXPOSE %d/udp", port))
		}
	}

	// Add CMD
	var cmdComponents []string

	// Add the executable path as the first argument
	cmdComponents = append(cmdComponents, fmt.Sprintf("\"%s\"", info.ExecutablePath))

	// Add the command-line arguments
	for _, argument := range info.CommandLineArguments {
		cmdComponents = append(cmdComponents, fmt.Sprintf("\"%s\"", argument.Flag))
		if argument.Value != "" {
			cmdComponents = append(cmdComponents, fmt.Sprintf("\"%s\"", argument.Value))
		}
	}
	writeBlock(dockerfile, "Entry point", fmt.Sprintf("CMD [%s]", strings.Join(cmdComponents, ", ")))

	return os.WriteFile(dockerfilePath, []byte(dockerfile.String()), 0644)
}
