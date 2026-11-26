# spkg

Cari package manager atau runtime manager yang kemungkinan mengelola executable di Linux.

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-4EAA25?logo=linux&logoColor=white)](.github/workflows/ci.yml)

```bash
spkg git
spkg node npm python3
spkg explain node -o pretty
```

`spkg` mencari executable melalui `PATH` atau path yang diberikan, mengikuti
symlink, lalu menggabungkan informasi path dan hasil query tool yang tersedia.
Hasilnya dapat mencakup manager, paket, versi, metadata runtime, alasan deteksi,
serta saran perintah install, upgrade, dan remove. Kelengkapan informasi
bergantung pada detektor yang cocok.

Repositori: [budi-imam-prasetyo/spkg](https://github.com/budi-imam-prasetyo/spkg).

## Instalasi

### Build dari source

Memerlukan Linux, Git, Go **1.24 atau lebih baru**, dan Make untuk perintah `make`.

```bash
git clone https://github.com/budi-imam-prasetyo/spkg.git
cd spkg
make build
```

Pasang binary hasil build untuk pengguna saat ini:

```bash
install -Dm755 ./spkg "$HOME/.local/bin/spkg"
```

Pastikan `~/.local/bin` masuk dalam `PATH`. Untuk instalasi sistem ke
`/usr/local/bin`, jalankan `make install` **setelah** `make build`;
target install menggunakan `sudo` dan hanya menyalin binary yang sudah ada.

Tanpa Make, binary pengembangan dapat dibangun dengan:

```bash
go build -o spkg .
```

### Binary dari GitHub Releases

[Rilis GitHub](https://github.com/budi-imam-prasetyo/spkg/releases) menyediakan
binary Linux untuk `amd64` dan `arm64`. Aset yang diterbitkan workflow saat ini
bernama `spkg-linux-amd64` dan `spkg-linux-arm64`, tanpa arsip `.tar.gz`.

Untuk Linux amd64 / x86-64:

```bash
curl -fL https://github.com/budi-imam-prasetyo/spkg/releases/latest/download/spkg-linux-amd64 -o spkg-linux-amd64
install -Dm755 ./spkg-linux-amd64 "$HOME/.local/bin/spkg"
```

Untuk Linux arm64 / AArch64:

```bash
curl -fL https://github.com/budi-imam-prasetyo/spkg/releases/latest/download/spkg-linux-arm64 -o spkg-linux-arm64
install -Dm755 ./spkg-linux-arm64 "$HOME/.local/bin/spkg"
```

### Memeriksa versi

```bash
spkg --version
spkg version
```

Versi berasal dari nilai yang disisipkan saat build. `make build` mengambilnya
dari Git; `go build` biasa memakai `dev`. Workflow CI saat ini juga membangun
binary rilis tanpa menyisipkan versi, sehingga binary unduhan dapat melaporkan
`dev` meskipun berada pada halaman rilis bertag.

## Penggunaan

### Satu executable

```bash
spkg git
spkg /usr/bin/python3
spkg node -o json
spkg npm --no-cache -v
```

Output default berupa tabel `Field` / `Value`. Pilihan `pretty` menampilkan
rincian per bagian, `plain` menampilkan sejumlah field sebagai teks sederhana,
dan `json` / `yaml` menyediakan hasil terstruktur.

Perintah install, upgrade, dan remove pada hasil merupakan saran yang dicetak;
`spkg` tidak menjalankannya. Detektor tetap menjalankan subprocess untuk membaca
informasi package manager.

### Beberapa executable

```bash
spkg ls cat grep
spkg node npm python3 -o json
```

Tabel batch berisi kolom `Executable` / `Source`, mengikuti urutan argumen.
Executable yang tidak ditemukan dilewati; proses gagal jika tidak ada satu pun
argumen yang ditemukan. Query satu executable yang tidak ditemukan juga
menghasilkan exit code non-zero.

JSON/YAML batch berupa daftar ringkas dengan field `executable`, `manager`,
serta `package`, `runtime`, dan `runtime_version` jika tersedia. Untuk rincian
seperti versi paket, confidence, dan saran perintah, gunakan query satu executable
berformat JSON/YAML.

### Memeriksa executable umum

```bash
spkg all
spkg all -o json > packages.json
```

`all` memeriksa daftar nama executable umum yang ditentukan di kode, seperti
`git`, `node`, `python3`, dan `cargo`. Perintah ini tidak menginventarisasi seluruh
paket terpasang atau seluruh isi `PATH`. Hasilnya memakai format ringkas batch.

### Penjelasan lokasi dan runtime

```bash
spkg explain node -o pretty
spkg explain python3 -o json
```

`explain` menjalankan deteksi langsung dan menambahkan penjelasan lokasi,
symlink, runtime, serta manager. Gunakan `pretty`, `json`, atau `yaml` untuk
melihat penjelasan; renderer `table` dan `plain` saat ini tidak menampilkan
field tersebut.

### Opsi

| Opsi | Perilaku |
|---|---|
| `-o, --output` | `table` (default), `pretty`, `plain`, `json`, atau `yaml` |
| `-v, --verbose` | Diagnostik pencarian dan command yang gagal ke stderr pada query executable utama |
| `--no-cache` | Lewati pembacaan dan penulisan hasil deteksi persisten pada query utama dan `all` |
| `--cache-ttl` | TTL dalam detik untuk entri baru; default `600`. Gunakan nilai positif; nilai nol/negatif tidak menonaktifkan cache |
| `--version` | Tampilkan versi build |
| `-h, --help` | Tampilkan bantuan |

```bash
spkg --help
spkg explain --help
spkg cache --help
```

### Cache

```bash
spkg cache stats
spkg cache clear
```

Hasil deteksi disimpan dalam format gob di
`$XDG_CACHE_HOME/spkg/cache-v2.gob`, atau `~/.cache/spkg/cache-v2.gob` jika
`XDG_CACHE_HOME` tidak diatur. Key mencakup nama query, path awal, dan target
symlink. Batch menyimpan hasil baru sekaligus; file diperbarui secara atomik
dengan penguncian antarproses.

TTL default adalah 10 menit. Perubahan paket di luar `spkg` dapat membuat hasil
cache tertinggal sampai kedaluwarsa; gunakan `--no-cache` atau `cache clear`
untuk memeriksa ulang. Opsi `--cache-ttl` tidak mengubah kedaluwarsa entri yang
sudah tersimpan.

Query global tertentu juga dimemoisasi selama proses berjalan. `--no-cache`
tidak menonaktifkan memoization tersebut. Kecepatan bergantung pada jumlah
executable, tool yang tersedia, dan waktu respons masing-masing tool.

## Cakupan deteksi

| Kelompok | Detektor yang tersedia |
|---|---|
| Paket sistem | pacman |
| JavaScript | Bun, npm, pnpm, Yarn |
| Python | pip, pipx, uv |
| Bahasa lain | Cargo, Go, RubyGems, Composer |
| Aplikasi dan lingkungan | Flatpak, Snap, Homebrew, Nix, AppImage |
| Container | Docker, Podman, Distrobox, Toolbox; pemeriksaan path/nama terbatas |

Runtime detector mencakup fnm, nvm, Volta, Bun, rustup, pyenv, conda, mamba,
micromamba, pixi, rye, Poetry, rbenv, RVM, SDKMAN!, jenv, gvm, goenv, phpenv,
asdf, mise, dan kerl. Pengenalan terutama bergantung pada struktur path dan
ketersediaan tool atau direktori yang diperiksa oleh masing-masing detektor.

Untuk tooling bawaan runtime, fnm, nvm, Volta, rustup, dan Bun dapat menjadi
`manager` utama. Runtime lain dapat dipetakan ke `pip`, `gem`, `go`, atau
`manual`, sementara identitas runtime tetap tersedia pada metadata. Paket pihak
ketiga yang berhasil dikenali tetap dapat dimiliki package manager, dengan
runtime sebagai informasi tambahan.

### Paket foreign dan helper AUR

Detektor pacman memakai `pacman -Qm` untuk mengenali paket foreign, kemudian
memilih helper yang tersedia dengan urutan `paru`, `yay`, `pikaur`, lalu `pakku`.
Jika ditemukan, helper tersebut digunakan pada field manager dan saran perintah.

Ini tidak membuktikan helper mana yang dahulu menginstal paket. Paket foreign
juga dapat berasal dari build lokal atau sumber lain; label `aur` pada hasil
merupakan asumsi detektor, bukan verifikasi terhadap AUR. `paru` dan `yay`
bukan detektor kepemilikan yang terpisah dari pacman.

### Batasan

- Belum ada detektor kepemilikan paket apt/dpkg, RPM/dnf, atau APK. Pada distro
  tersebut, executable sistem dapat jatuh ke fallback `manual`.
- `manual` berarti tidak ada detektor yang cocok. Hasil itu tidak membuktikan
  bahwa executable dipasang secara manual.
- Beberapa detektor memakai heuristik path atau nama. Nama paket, versi,
  repository, dan saran perintah tidak selalu lengkap atau benar untuk setiap
  layout instalasi.
- Query subprocess menggunakan context timeout 15 detik per command. Satu
  deteksi dapat menjalankan beberapa command, sehingga total waktunya dapat
  lebih dari 15 detik.

## Confidence

Confidence adalah skor heuristik 0–100, bukan probabilitas terkalibrasi atau
jaminan bahwa atribusi benar. Field `Reasons` menjelaskan sinyal yang dipakai.

| Label | Rentang skor |
|---|---|
| `unknown` | 0 |
| `low` | 1–49 |
| `medium` | 50–89 |
| `high` | 90–99 |
| `certain` | 100 |

Skor awal berasal dari `Score()`, umumnya 0, 40, atau 70. Bukti konfirmasi yang
dilampirkan detektor dapat menaikkannya ke setidaknya 90. Kepemilikan runtime
memberi dasar setidaknya 70; pengecualian manager lain dapat menambah hingga
10 poin ketika skor dasar masih di bawah 90. Konfirmasi command tidak otomatis
menghasilkan 100, dan label `certain` tidak berarti bukti kriptografis.

Fallback manual memiliki skor 0. Pada JSON/YAML satu hasil, field `confidence`
yang bernilai nol dihilangkan oleh serializer; alasannya tetap dapat tersedia.

## Pengembangan

```bash
make build     # build binary dengan metadata versi Git
make test      # go test -v -race -coverprofile=coverage.out ./...
make lint      # golangci-lint run ./...; membutuhkan golangci-lint terpasang
make clean     # hapus binary dan coverage.out
```

Kode utama berada di:

- [`internal/cli`](internal/cli): command, opsi, dan cache hasil query.
- [`internal/detector`](internal/detector): resolusi path, detektor paket/runtime,
  prioritas pemilihan, memoization command, dan perhitungan confidence.
- [`internal/cache`](internal/cache): penyimpanan hasil deteksi di memori dan disk.
- [`internal/output`](internal/output): renderer tabel, pretty, plain, JSON, dan YAML.

Detektor paket mengimplementasikan `Name()`, `Detect()`, `IsAvailable()`, dan
`Score()`, lalu didaftarkan melalui `Register()`. Penambahan manager baru juga
perlu meninjau konstanta manager, daftar manager, prioritas, dan tes; membuat
satu file saja belum tentu cukup. Detektor runtime memakai interface dan
registrasi terpisah. Batch menjalankan paling banyak delapan worker, dibatasi
juga oleh `GOMAXPROCS`.

CI menguji dengan race detector dan membangun Linux amd64/arm64. Lihat
[workflow CI](.github/workflows/ci.yml) untuk proses yang benar-benar dijalankan;
konfigurasi [GoReleaser](.goreleaser.yml) tersedia tetapi tidak dipanggil oleh
workflow rilis saat ini.

## Lisensi

MIT — hak cipta Budi Imam Prasetyo. Lihat [LICENSE](LICENSE),
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md), dan [CHANGELOG.md](CHANGELOG.md).
