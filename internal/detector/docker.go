package detector

import (
	"path/filepath"
	"strings"
)

type DockerDetector struct{}

func init() { Register(&DockerDetector{}) }

func (d *DockerDetector) Name() Manager { return ManagerDocker }

func (d *DockerDetector) IsAvailable() bool {
	return HasCommand("docker")
}

func (d *DockerDetector) Detect(execPath string) *Result {
	if !strings.Contains(execPath, "docker") {
		return nil
	}

	images, err := RunCmd("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(images, "\n") {
		if strings.Contains(line, filepath.Base(execPath)) {
			return &Result{
				Found:          true,
				ExecutablePath: execPath,
				Manager:        ManagerDocker,
				Package:        line,
				InstallCmd:     "docker pull " + line,
				UpgradeCmd:     "docker pull " + line,
				RemoveCmd:      "docker rmi " + line,
			}
		}
	}

	return nil
}
