# spkg

Detect which package manager installed an executable on your system.

## Features

- Detects 20+ package managers: pacman, paru, yay, bun, npm, pnpm, yarn, pip, pipx, uv, cargo, go, gem, composer, flatpak, snap, homebrew, nix, docker, podman
- Beautiful terminal output with lipgloss
- Multiple output formats: pretty, JSON, YAML, table, plain
- Verbose mode for debugging detection
- Explain mode for understanding PATH resolution
- Caching for fast repeated queries
- `--all` to list all detectable executables
- Cross-platform: Linux, macOS, Windows

## Installation

### Binary releases

Download from [GitHub Releases](https://github.com/ryoukaii/spkg/releases).

### Build from source

```bash
git clone https://github.com/ryoukaii/spkg.git
cd spkg
go build -o spkg .
sudo cp spkg /usr/local/bin/
```

### AUR (Arch Linux)

```bash
paru -S spkg
# or
yay -S spkg
```

## Usage

### Basic detection

```bash
spkg git
```

Output:
```
✓ Package ditemukan

Executable:
  /usr/bin/git

Package manager:
  pacman

Package:
  git

Repository:
  extra

Installed version:
  2.51.0

Install command:
  sudo pacman -S git

Upgrade command:
  sudo pacman -Syu git

Remove command:
  sudo pacman -Rns git
```

### Output formats

```bash
spkg git --output json
spkg git --output yaml
spkg git --output table
spkg git --output plain
```

### Verbose mode

```bash
spkg git -v
```

Shows the detection process:
```
Searching executable "git"...
Found /usr/bin/git
Running detectors...
✓ Package ditemukan
...
```

### Explain mode

```bash
spkg explain git
```

Explains why an executable is in your PATH:
```
✓ Package ditemukan

Executable:
  /usr/bin/git

Package manager:
  pacman

Package:
  git

Explanation:
  git is located at:
    /usr/bin/git

  That directory is owned by the system.
  pacman installed this as a system package.
```

### List all detectable executables

```bash
spkg all
```

Output:
```
Executable    Manager
────────────  ──────────────
git           pacman
node          bun
python        pacman
docker        pacman
code          pacman
lazygit       pacman
```

### Cache management

```bash
spkg cache stats    # Show cache statistics
spkg cache clear    # Clear cache
```

## Adding a new detector

1. Create `internal/detector/yourmanager.go`:

```go
package detector

type YourManagerDetector struct{}

func init() { Register(&YourManagerDetector{}) }

func (d *YourManagerDetector) Name() Manager { return ManagerYourManager }

func (d *YourManagerDetector) IsAvailable() bool {
    return HasCommand("your-manager")
}

func (d *YourManagerDetector) Detect(execPath string) *Result {
    // Implement detection logic
    return &Result{
        Found:          true,
        ExecutablePath: execPath,
        Manager:        ManagerYourManager,
        Package:        "package-name",
        Version:        "1.0.0",
        InstallCmd:     "your-manager install package-name",
        UpgradeCmd:     "your-manager upgrade package-name",
        RemoveCmd:      "your-manager remove package-name",
    }
}
```

2. Add the manager constant to `internal/detector/detector.go`:

```go
const (
    ManagerYourManager Manager = "your-manager"
)
```

3. Add priority in `internal/detector/registry.go`:

```go
func priorityOrder(m Manager) int {
    // Add your manager's priority
    ManagerYourManager: 15,
}
```

## Supported package managers

| Manager | Detection method |
|---------|-----------------|
| pacman | `pacman -Qo` |
| paru/yay | `pacman -Qm` (AUR check) |
| bun | `~/.bun/bin` prefix check |
| npm | `npm prefix -g` + `npm root -g` |
| pnpm | `pnpm prefix -g` |
| yarn | `yarn global dir` |
| pip | `pip show` |
| pipx | `pipx list` |
| uv | `uv tool list` |
| cargo | `cargo install --list` |
| go | `go env GOPATH` |
| gem | `gem list -d` |
| composer | `composer global config home` |
| flatpak | `flatpak list --app` |
| snap | `snap list` |
| homebrew | `brew --prefix` |
| nix | `nix profile list` |
| docker | `docker images` |
| podman | `podman images` |
| appimage | `.AppImage` suffix check |
| manual | Fallback (no package manager found) |

## Troubleshooting

### "command not found"

The executable is not in your PATH. Check with `which <command>`.

### Wrong package manager detected

Some executables may be installed by multiple package managers. The detector uses a priority system:

1. System package managers (pacman, apt, dnf)
2. Application stores (flatpak, snap)
3. Language-specific managers (npm, pip, cargo)
4. Manual installation (fallback)

### Slow detection

First run may be slow due to package manager queries. Subsequent runs use cache (default TTL: 10 minutes).

Clear cache with:
```bash
spkg cache clear
```

## License

MIT
