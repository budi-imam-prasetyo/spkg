package detector

import (
	"strings"
)

type PodmanDetector struct{}

func init() { Register(&PodmanDetector{}) }

func (d *PodmanDetector) Name() Manager { return ManagerPodman }

func (d *PodmanDetector) IsAvailable() bool {
	return HasCommand("podman")
}

func (d *PodmanDetector) Detect(execPath string) *Result {
	if !strings.Contains(execPath, "podman") {
		return nil
	}

	images, err := RunCmd("podman", "images", "--format", "{{.Repository}}:{{.Tag}}")
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(images, "\n") {
		if strings.Contains(line, execPath) {
			return &Result{
				Found:          true,
				ExecutablePath: execPath,
				Manager:        ManagerPodman,
				Package:        line,
				InstallCmd:     "podman pull " + line,
				UpgradeCmd:     "podman pull " + line,
				RemoveCmd:      "podman rmi " + line,
			}
		}
	}

	return nil
}
