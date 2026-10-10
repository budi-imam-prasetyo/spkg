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
	// Only claim ownership if the executable lives under a Docker container
	// storage root (the layered image trees) or is the docker CLI itself.
	// A plain substring check on the PATH (e.g. ".../docker-scripts/foo") would
	// produce false positives.
	if !dockerOwnsPath(execPath) {
		return nil
	}

	images, err := memoizedRunCmd("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	if err != nil {
		return nil
	}

	base := filepath.Base(execPath)
	for _, line := range strings.Split(images, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Exact match on repository name (before :tag, after last /):
		// "myapp" must not match "myapp-helper".
		repo := line
		if idx := strings.LastIndex(repo, ":"); idx != -1 {
			repo = repo[:idx]
		}
		if idx := strings.LastIndex(repo, "/"); idx != -1 {
			repo = repo[idx+1:]
		}
		if repo != base {
			continue
		}
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerDocker,
			Package:        line,
			InstallCmd:     "docker pull " + line,
			UpgradeCmd:     "docker pull " + line,
			RemoveCmd:      "docker rmi " + line,
			ConfirmReasons: []Reason{{
				Signal: "manager_db",
				Detail: "docker images confirms repository " + line,
				Weight: int(ConfidenceHigh),
			}},
		}
	}

	return nil
}

// dockerOwnsPath reports whether execPath is a Docker-managed binary (inside a
// container image layer or the docker CLI).
func dockerOwnsPath(execPath string) bool {
	if filepath.Base(execPath) == "docker" {
		return true
	}
	return strings.Contains(execPath, "/var/lib/docker/overlay2/") ||
		strings.Contains(execPath, "/docker/overlay2/")
}

func (d *DockerDetector) Score(execPath string) (Confidence, []Reason) {
	if filepath.Base(execPath) == "docker" {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "executable is the docker CLI", Weight: int(ConfidenceMedium)}}
	}
	if strings.Contains(execPath, "/var/lib/docker/overlay2/") || strings.Contains(execPath, "/docker/overlay2/") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is inside a docker image layer", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
