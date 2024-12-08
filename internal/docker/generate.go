package docker

import (
	"fmt"
	"os"
	"strings"

	"application_profiling/internal/process"
)

// GenerateDockerfile generates a Dockerfile based on process information.
func GenerateDockerfile(info *process.ProcessInfo, dockerfilePath, tarFile string) error {
	dockerfile := &strings.Builder{}
	fmt.Fprintln(dockerfile, "FROM ubuntu:latest")
	fmt.Fprintf(dockerfile, "COPY %s /\n", tarFile)
	fmt.Fprintf(dockerfile, "RUN tar --skip-old-files -xvf /%s -C / && rm /%s\n", tarFile, tarFile)

	for _, env := range info.EnvironmentVariables {
		if strings.Contains(env, "=") {
			fmt.Fprintf(dockerfile, "ENV %s\n", env)
		}
	}

	userSpec := info.ProcessOwner
	if userSpec == "" {
		userSpec = "root:root"
	} else if !strings.Contains(userSpec, ":") {
		userSpec = userSpec + ":" + userSpec
	}
	fmt.Fprintf(dockerfile, "USER %s\n", userSpec)

	if info.WorkingDirectory != "" {
		fmt.Fprintf(dockerfile, "WORKDIR %s\n", info.WorkingDirectory)
	}

	for _, port := range info.ListeningTCP {
		fmt.Fprintf(dockerfile, "EXPOSE %d/tcp\n", port)
	}
	for _, port := range info.ListeningUDP {
		fmt.Fprintf(dockerfile, "EXPOSE %d/udp\n", port)
	}

	cmdArgs := strings.Fields(info.ReconstructedCommand)
	if len(cmdArgs) > 0 {
		quoted := make([]string, len(cmdArgs))
		for i, c := range cmdArgs {
			quoted[i] = fmt.Sprintf("\"%s\"", c)
		}
		fmt.Fprintf(dockerfile, "CMD [%s]\n", strings.Join(quoted, ", "))
	}

	return os.WriteFile(dockerfilePath, []byte(dockerfile.String()), 0644)
}
