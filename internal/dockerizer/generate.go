package dockerizer

import (
	"application_profiling/internal/profiler"
	"fmt"
	"os"
	"strings"
	"text/template"
)

const dockerfileTemplateContent = `# Use the official Ubuntu image as the base
FROM ubuntu:latest

# Copy the profile archive
COPY {{.TarFile}} /

# Extract the profile and clean up the archive
RUN tar --skip-old-files -xvf /{{.TarFile}} -C / && rm /{{.TarFile}}

# Set environment variables
{{- range .EnvironmentVariables }}
ENV {{.}}
{{- end }}

# Set the user and group
USER {{.UserAndGroup}}

# Set the working directory
WORKDIR {{.WorkingDirectory}}

# Expose ports
{{- range .TCPPorts }}
EXPOSE {{.}}/tcp
{{- end }}
{{- range .UDPPorts }}
EXPOSE {{.}}/udp
{{- end }}

# Set the entry point
CMD [{{.Command}}]
`

// DockerfileData holds the data needed for the Dockerfile template.
type DockerfileData struct {
	TarFile              string
	EnvironmentVariables []string
	UserAndGroup         string
	WorkingDirectory     string
	TCPPorts             []int
	UDPPorts             []int
	Command              string
}

// GenerateDockerfile generates a Dockerfile from thegiven ProcessInfo.
func GenerateDockerfile(info *profiler.ProcessInfo, dockerfilePath string, tarFile string) error {
	commandLine := buildCommandLine(info)
	userAndGroup := fmt.Sprintf("%s:%s", info.ProcessUser, info.ProcessGroup)

	dockerfileData := DockerfileData{
		TarFile:              tarFile,
		EnvironmentVariables: info.EnvironmentVariables,
		UserAndGroup:         userAndGroup,
		WorkingDirectory:     info.WorkingDirectory,
		TCPPorts:             info.ListeningTCP,
		UDPPorts:             info.ListeningUDP,
		Command:              commandLine,
	}

	return writeDockerfile(dockerfileData, dockerfilePath)
}

// buildCommandLine constructs the CMD array from the executable path
// and the associated command-line arguments. It produces something like:
// ["/usr/sbin/nginx", "-g", "daemon on; master_process on;"]
func buildCommandLine(processInformation *profiler.ProcessInfo) string {
	var commandSegments []string

	// Add the executable path (quoted for Docker CMD array).
	commandSegments = append(commandSegments, fmt.Sprintf("\"%s\"", processInformation.ExecutablePath))

	// Append flags and optional values (quoted).
	for _, argument := range processInformation.CommandLineArguments {
		commandSegments = append(commandSegments, fmt.Sprintf("\"%s\"", argument.Flag))
		if argument.Value != "" {
			commandSegments = append(commandSegments, fmt.Sprintf("\"%s\"", argument.Value))
		}
	}

	// Join them with commas to form a valid Docker CMD array, e.g.:
	// CMD ["/usr/sbin/nginx", "-g", "daemon on; master_process on;"]
	return strings.Join(commandSegments, ", ")
}

// writeDockerfile writes the Dockerfile to the specified path using the provided data.
func writeDockerfile(data DockerfileData, dockerfilePath string) error {
	dockerfileTemplate, parseErr := template.New("Dockerfile").Parse(dockerfileTemplateContent)
	if parseErr != nil {
		return parseErr
	}

	dockerfileHandle, createErr := os.Create(dockerfilePath)
	if createErr != nil {
		return createErr
	}
	defer dockerfileHandle.Close()

	return dockerfileTemplate.Execute(dockerfileHandle, data)
}
