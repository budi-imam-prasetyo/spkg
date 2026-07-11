package detector

type ManualDetector struct{}

func (d *ManualDetector) Name() Manager { return ManagerManual }

func (d *ManualDetector) IsAvailable() bool { return true }

func (d *ManualDetector) Detect(execPath string) *Result {
	return &Result{
		Found:          true,
		ExecutablePath: execPath,
		Manager:        ManagerManual,
		Package:        "",
		InstallCmd:     "# Manual installation",
		UpgradeCmd:     "# Manual upgrade required",
		RemoveCmd:      "rm " + execPath,
	}
}

func NewManualResult(execPath, name string) *Result {
	return &Result{
		Found:          true,
		Executable:     name,
		ExecutablePath: execPath,
		Manager:        ManagerManual,
		Package:        name,
		InstallCmd:     "# Manual installation",
		UpgradeCmd:     "# Manual upgrade required",
		RemoveCmd:      "rm " + execPath,
	}
}
