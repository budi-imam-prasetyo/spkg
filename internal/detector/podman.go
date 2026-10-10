package detector

import (
	"path/filepath"
	"strings"
)

type PodmanDetector struct{}

func init() { Register(&PodmanDetector{}) }

func (d *PodmanDetector) Name() Manager { return ManagerPodman }

func (d *PodmanDetector) IsAvailable() bool {
	return HasCommand("podman")
}

func (d *PodmanDetector) Detect(execPath string) *Result {
	// Only claim ownership if the executable lives under a Podman container
	// storage root (the layered image trees) or is the podman CLI itself.
	if !podmanOwnsPath(execPath) {
		return nil
	}

	images, err := memoizedRunCmd("podman", "images", "--format", "{{.Repository}}:{{.Tag}}")
	if err != nil {
		return nil
	}

	base := filepath.Base(execPath)
	for _, line := range strings.Split(images, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		repo := line
		if idx := strings.LastIndex(repo, ":"); idx != -1 {
			repo = repo[:idx]
		}
		if idx := strings.LastIndex(repo, "/"); idx != -1 {
			repo = repo[idx+1:]
		}
		if repo != base && !strings.Contains(line, execPath) {
			continue
		}
		return &Result{
			Found:          true,
			ExecutablePath: execPath,
			Manager:        ManagerPodman,
			Package:        line,
			InstallCmd:     "podman pull " + line,
			UpgradeCmd:     "podman pull " + line,
			RemoveCmd:      "podman rmi " + line,
			ConfirmReasons: []Reason{{
				Signal: "manager_db",
				Detail: "podman images confirms repository " + line,
				Weight: int(ConfidenceHigh),
			}},
		}
	}

	return nil
}

// podmanOwnsPath reports whether execPath is a Podman-managed binary (inside a
// container image layer or the podman CLI).
func podmanOwnsPath(execPath string) bool {
	if filepath.Base(execPath) == "podman" {
		return true
	}
	return strings.Contains(execPath, "/var/lib/containers/storage/overlay/") ||
		strings.Contains(execPath, "/.local/share/containers/storage/overlay/")
}

func (d *PodmanDetector) Score(execPath string) (Confidence, []Reason) {
	if filepath.Base(execPath) == "podman" {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "executable is the podman CLI", Weight: int(ConfidenceMedium)}}
	}
	if strings.Contains(execPath, "/var/lib/containers/storage/overlay/") || strings.Contains(execPath, "/.local/share/containers/storage/overlay/") {
		return ConfidenceMedium, []Reason{{Signal: "path_match", Detail: "path is inside a podman image layer", Weight: int(ConfidenceMedium)}}
	}
	return ConfidenceUnknown, nil
}
