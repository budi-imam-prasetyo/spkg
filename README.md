# spkg

<div align="center">

**Kenali siapa yang menginstall executable di sistem Linux Anda.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-4EAA25?logo=linux&logoColor=white)](https://www.linux.org)

</div>

<br>

`spkg` menjawab satu pertanyaan sederhana tapi sering membingungkan:

> *"Executable ini dipasang oleh apa? pacman? paru? npm? cargo? fnm?"*

Cukup satu perintah, dan `spkg` memberi tahu **siapa pemiliknya**, **versi berapa**,
**di repository mana**, dan **command apa yang harus dipakai** untuk install, upgrade,
atau hapus — lengkap dengan skor confidence berbasis bukti.

---

## ✨ Fitur

| | |
|---|---|
| 🔍 **23 package manager** | pacman, paru, yay, bun, npm, pnpm, yarn, pip, pipx, uv, cargo, go, gem, composer, flatpak, snap, homebrew, nix, docker, podman, distrobox, toolbox, appimage + manual fallback |
| ⚡ **Runtime detection** | fnm, nvm, volta, rustup diakui sebagai *owner* utama executable; asdf, mise, pyenv, conda, mamba, pixi, poetry, rbenv, RVM, SDKMAN!, jenv, gvm, goenv, phpenv, kerl dideteksi sebagai runtime dan diatribusikan ke package manager di bawahnya (pip/gem/go) |
| 🎯 **AUR helper aware** | Mendeteksi apakah paket AUR dipasang via `paru`, `yay`, `pikaur`, atau `pakku` — command install/upgrade otomatis disesuaikan |
| 📊 **Confidence scoring** | Skor 0–100 berbasis bukti (path domain, konfirmasi DB, runtime ownership, exclusion bonus) + alasan eksplisit, bukan angka misterius |
| 🖼️ **Output berborder** | Tabel ASCII dengan rounded corners, header berwarna, dan padding rapi |
| 📦 **Batch mode** | `spkg ls cat grep` — cek banyak executable sekaligus dalam satu perintah |
| 💾 **Cache persisten** | Gob serialization ke disk — query ulang <1ms, cold scan ~20s |
| 🧵 **Timeout & safety** | Setiap subprocess dibatasi 15 detik; tidak ada command injection |
| 🧪 **Testable engine** | Interface `System` yang bisa di-mock untuk test deterministik tanpa tooling host |

---

## 🚀 Instalasi

### Dari source (disarankan)

```bash
git clone https://github.com/ryoukaii/spkg.git
cd spkg
make install
```

### Dari GitHub Releases

```bash
# Linux amd64
curl -sL https://github.com/ryoukaii/spkg/releases/latest/download/spkg_linux_amd64.tar.gz \
  | tar xz -C ~/.local/bin spkg

# Linux arm64
curl -sL https://github.com/ryoukaii/spkg/releases/latest/download/spkg_linux_arm64.tar.gz \
  | tar xz -C ~/.local/bin spkg
```

### Verifikasi

```bash
spkg --version
# spkg version v0.3.0
```

---

## 📖 Penggunaan

### Query satu executable

```bash
spkg python3
```

```
╭────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│   Field    │                                                        Value                                                        │
├────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Executable │ /bin/python3                                                                                                        │
│ Manager    │ pacman                                                                                                              │
│ Package    │ python                                                                                                              │
│ Version    │ 3.14.6-3                                                                                                            │
│ Confidence │ 90% (high)                                                                                                          │
│ Reasons    │ path is under a system install dir (pacman domain); pacman -Qo confirms ownership by package python                  │
│ Install    │ sudo pacman -S python                                                                                               │
│ Upgrade    │ sudo pacman -Syu python                                                                                             │
│ Remove     │ sudo pacman -Rns python                                                                                             │
╰────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Query banyak executable sekaligus

```bash
spkg ls cat grep awk sed
```

```
╭────────────┬────────╮
│ Executable │ Source │
├────────────┼────────┤
│ ls         │ pacman │
│ cat        │ pacman │
│ grep       │ pacman │
│ awk        │ pacman │
│ sed        │ pacman │
╰────────────┴────────╯
```

### Scan executable umum di sistem

```bash
spkg all
spkg all -o json > packages.json
```

### Runtime manager (fnm, nvm, pyenv, dll)

```bash
spkg node
```

```
╭─────────────────┬─────────────────────────────────────────────────────────────────────────────────────────────────────╮
│     Field       │                                                        Value                                                        │
├─────────────────┼─────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Executable      │ /home/ryoukaii/.local/share/fnm/node-versions/v24.20.0/installation/bin/node                                    │
│ Manager         │ fnm                                                                                                              │
│ Runtime         │ fnm                                                                                                              │
│ Runtime version │ Node.js 24.20.0                                                                                                  │
│ Package         │ node                                                                                                             │
│ Confidence      │ 80% (medium)                                                                                                     │
│ Install         │ fnm install 24.20.0                                                                                              │
│ Upgrade         │ fnm install --latest                                                                                             │
│ Remove          │ fnm uninstall 24.20.0                                                                                            │
╰─────────────────┴─────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

### Paket AUR via helper

```bash
spkg code
```

```
╭────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│   Field    │                                                        Value                                                        │
├────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ Executable │ /bin/code                                                                                                           │
│ Manager    │ pacman                                                                                                              │
│ Package    │ visual-studio-code-bin                                                                                              │
│ Repository │ aur                                                                                                                 │
│ AUR        │ yes                                                                                                                 │
│ Version    │ 1.134.0-1                                                                                                           │
│ Install    │ paru -S visual-studio-code-bin                                                                                      │
│ Upgrade    │ paru -Syu visual-studio-code-bin                                                                                    │
│ Remove     │ paru -Rns visual-studio-code-bin                                                                                    │
╰────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

Perhatikan: `spkg` tahu VS Code dipasang via **`paru`**, bukan `pacman` langsung —
jadi command-nya sudah benar.

---

## 🎛️ Opsi

```
spkg [executable...] [flags]

Flags:
  -o, --output string   Output format: table, json, yaml, plain, pretty (default "table")
  -v, --verbose         Show verbose detection output
      --no-cache        Skip cache
      --cache-ttl int   Cache TTL in seconds (default: 600)
  -h, --help            Help for spkg
      --version         Version for spkg
```

### Subcommand

| Command | Deskripsi |
|---------|-----------|
| `spkg all` | Scan executable umum di sistem |
| `spkg cache stats` | Lihat jumlah item cache |
| `spkg cache clear` | Bersihkan cache |
| `spkg cache items` | Tampilkan isi cache |

---

## 🧠 Bagaimana Confidence Scoring Bekerja

`spkg` tidak menebak. Setiap skor 0–100 dibangun dari bukti berlapis:

| Level | Skor | Artinya |
|-------|------|---------|
| `unknown` | 0 | Tidak ada bukti — fallback manual |
| `low` | 40 | Indikasi lemah (prefix path ambigu) |
| `medium` | 70 | Bukti struktural kuat (path di domain manager) |
| `high` | 90 | Konfirmasi command-level (`pacman -Qo`) |
| `certain` | 100 | Ownership DB-backed |

Komponen yang dikombinasikan oleh `CombineConfidence`:

1. **Structural band** — sinyal dari `Score()` tanpa subprocess
2. **Command confirmation** — menaikkan ke `high` jika DB konfirmasi
3. **Runtime ownership** — runtime manager meng-own path tersebut
4. **Exclusion bonus** — semakin sedikit manager lain yang match, semakin yakin (maksimal +10)

Setiap alasan dicetak eksplisit di field `Reasons`, jadi Anda selalu tahu **mengapa**
`spkg` menyimpulkan sesuatu.

---

## 🏗️ Arsitektur Engine

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI (cobra)                           │
│   runDetect → single | batch → output (table/json/yaml)     │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                   Registry (registry.go)                     │
│   DetectWithRuntime: priority order + runtime overlay       │
│   DetectBatch: bounded worker pool (max 8)                  │
└──────┬───────────────────────────────────────────┬──────────┘
       │                                           │
┌──────▼──────────────┐                  ┌─────────▼──────────┐
│  Detector interface │                  │ RuntimeDetector    │
│  • Name()           │                  │ • Name()           │
│  • Detect(path)     │                  │ • Detect(path)     │
│  • IsAvailable()    │                  │ • IsAvailable()    │
│  • Score(path)      │                  └────────────────────┘
└──────┬──────────────┘
       │
┌──────▼──────────────────────────────────────────────────────┐
│                Confidence Engine (confidence.go)              │
│   CombineConfidence → structural + confirm + runtime + bonus │
└──────┬──────────────────────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────────────────────┐
│              System interface (mockable)                     │
│   LookPath() • CmdContext()  ← RWMutex-guarded              │
└──────────────────────────────────────────────────────────────┘
```

**Keputusan desain penting:**

- **`Detect()` vs `Score()`** — pemisahan operasi mahal (subprocess) dari sinyal murah (string match), sehingga `Score()` bisa dijalankan paralel tanpa side effect.
- **Priority order eksplisit** — `pacman(1) < paru(2) < fnm(8) < npm(11) < manual(99)`. Bukan urutan registrasi, tapi urutan bisnis.
- **Runtime menang jika meng-own path** — `node` di bawah `~/.local/share/fnm/` dimiliki `fnm`, bukan `npm`, meskipun `npm` juga ada di sana.
- **System interface di-mock** — `SetSystem(fake)` memungkinkan test deterministik tanpa tooling host, dan dilindungi `RWMutex` untuk aman concurrent.

---

## 🧪 Development

```bash
make test      # go test -v -race -coverprofile
make lint      # golangci-lint run ./...
make build     # go build dengan ldflags version
make clean     # hapus artefak
```

Menambah detector baru hanya butuh satu file:

```go
// internal/detector/mytool.go
type MyToolDetector struct{}

func (d *MyToolDetector) Name() Manager { return ManagerMyTool }
func (d *MyToolDetector) Detect(execPath string) *Result { /* ... */ }
func (d *MyToolDetector) IsAvailable() bool { return HasCommand("mytool") }
func (d *MyToolDetector) Score(execPath string) (Confidence, []Reason) { /* ... */ }

func init() { Register(&MyToolDetector{}) }
```

---

## 📄 Lisensi

MIT — lihat [LICENSE](LICENSE).
