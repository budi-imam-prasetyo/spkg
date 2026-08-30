# Changelog

Semua perubahan penting pada `spkg` didokumentasikan di sini. Proyek mengikuti
[Semantic Versioning](https://semver.org/).

## [1.4.0] - 2026-08-30

- **Batch mode**: `spkg ls cat grep node` — deteksi banyak executable sekaligus
  dalam satu perintah; output mengikuti urutan argumen (cache hit & miss bercampur
  tetap terurut), argumen yang tidak ditemukan dilewati diam-diam.
- Fokus dukungan dan artefak rilis pada Linux amd64/arm64.
- Output default menjadi tabel dua kolom Field/Value; output massal
  Executable/Source.
- Perbaikan confidence manual sehingga hasil fallback tetap unknown, bukan
  dinaikkan ke high.
- Runtime manager kini bisa memiliki path meski tidak ada package detector yang
  cocok; command ownership runtime konsisten dengan manager yang dilaporkan.
- Validasi format output dan jumlah argumen executable; dukungan `--version`.
- Scan executable umum dengan concurrency terbatas.

## [1.3.0] - 2026-08-15

### Confidence engine — desain ulang penuh
Sebelumnya field `Confidence` adalah satu integer ajaib yang diisi tidak
konsisten antar detector, mencampur "path cocok" (50) dengan "version-manager
cocok" (80-95), sehingga angkanya bermakna berbeda di konteks berbeda dan skor
rendah tidak punya penjelasan. Diganti dengan **structural confidence engine**:

- **`Score()` per detector** — setiap detector kini mengekspos sinyal struktural
  yang murah dan bebas efek samping: `Score(execPath) -> (band Confidence, reasons []Reason)`.
  Melaporkan seberapa kuat path executable cocok dengan domain manager *tanpa*
  spawn subprocess. Band: `unknown=0`, `low=40` (dir ambigu/bersama),
  `medium=70` (prefix path jelas), `high=90` (bukti DB/runtime), `certain=100`.
- **Engine `CombineConfidence`** — menggabungkan, untuk manager pemenang: band
  struktural (floor) + konfirmasi level command (mis. bukti kepemilikan
  `pacman -Qo`) + sinyal runtime/version-manager-owns-path + bonus
  tanpa-alternatif (maks +10). Output di-clamp ke 0-100 dan selalu membawa
  daftar `reasons` yang menjelaskan *mengapa* skor diberikan.
- **Output transparan** — `pretty`/`table`/`plain` kini menampilkan skor, label
  band-nya (`unknown`/`low`/`medium`/`high`/`certain`, berwarna), dan alasan.
  JSON/YAML menyertakan array `reasons`.
- **Tes** — `confidence_test.go` mencakup threshold band, engine combine
  (DB-backed -> high, cap bonus exclusion, fallback manual), dan `Score()` untuk
  pacman/npm/cargo, dan asersi confidence end-to-end `DetectWithRuntime`.

Hasil: `spkg git` kini melaporkan `confidence: 90% (high)` dengan alasan
`path is under a system install dir` + `pacman -Qo confirms ownership by package git`,
dan `node` kelolaan fnm melaporkan `confidence: 90% (high)` dengan
`a runtime/version manager owns this path` — bukan angka ajaib tanpa penjelasan.

## [1.2.0] - 2026-08-15

### Logika deteksi — analisis & optimasi
- **Resolusi prioritas kini allocation-free di hot path.** Pemenang package-manager
  dulu dipilih dengan membangun ulang `map[Manager]int` pada *setiap* invokasi
  `spkg <bin>` (`priorityOrder` adalah fungsi yang mengalokasikan map setiap
  dipanggil). Diganti `prioritySlice` terurut yang dihitung sekali di `init()`
  dan lookup `priorityOf()`, sehingga `DetectFirst` tidak lagi mengalokasikan
  map per pemanggilan.
- **`spkg all` memakai satu snapshot detector** dan diuntungkan cache persisten:
  cold run membayar biaya probing tiap binary sekali, warm run melayani semua
  entri dari cache disk dalam milidetik.
- **Semantik kepemilikan diperjelas.** Ketika runtime/version manager (fnm, nvm,
  volta, pyenv, rustup, …) memiliki path, ia dilaporkan sebagai manager — bukan
  package manager yang kebetulan cocok — menghilangkan atribusi palsu seperti
  `node` kelolaan fnm yang dilaporkan sebagai `npm`.
- Skala `Confidence` didokumentasikan (0 = fallback manual, 50 = cocok path PM,
  80–95 = cocok version-manager) agar konsumen downstream memperlakukan nilai
  rendah sebagai tebakan terbaik.

### Diwarisi dari perbaikan audit sebelumnya (lihat di bawah)
Rilis ini juga membawa perbaikan kebenaran dan ketahanan dari audit v1.1.x:
cache lintas-invokasi yang berfungsi, pelaporan versi nyata, atribusi manager
akurat, pemeriksaan kepemilikan pip/pipx/uv/gem yang diperketat, timeout command,
exit non-zero saat not-found, output English-only, dan penghapusan dead code.

## [1.1.0] - (sebelumnya) — hardening berbasis audit
- **H1 — Atribusi manager yang benar.** Detector kini resolve berdasarkan prioritas,
  bukan urutan registrasi (nama file); runtime/version manager yang memiliki path
  menang atas package manager yang hanya berbagi pohon direktori.
- **H2 — Cache benar-benar persisten.** Store in-memory sekali pakai diganti
  cache disk berbasis gob di `$XDG_CACHE_HOME/spkg` (atau `~/.cache/spkg`).
  `spkg cache stats`/`clear` kini mencerminkan keadaan nyata lintas invokasi.
- **H3 — Pelaporan versi nyata.** Variabel `main.version` di-set via `-ldflags`;
  `spkg version` / injeksi goreleaser & Makefile kini melaporkan build tag, bukan
  string hardcoded.
- **H4 — Kepemilikan Python akurat.** pip/pipx/uv/gem kini mensyaratkan binary
  berada di direktori yang benar-benar mereka kelola (venv / user scheme / tool
  bin) sebelum mengklaim kepemilikan, mematikan false positive substring nama.

### Medium
- **M1** Melonggarkan `go.mod` dari `go 1.24.4` ke `go 1.24`.
- **M2** Menghapus `sudo pacman` / `sudo apt` hardcoded dari command
  install/upgrade/remove lintas distro (pacman sendiri sah mempertahankan `sudo`).
- **M3** `RunCmd`/`RunCmdStdout` kini memakai timeout konteks 15s agar daemon
  package-manager yang menggantung tidak membekukan CLI.
- **M4** `spkg <missing>` dan `spkg explain <missing>` kini exit non-zero (error
  satu baris yang bersih, tanpa dump usage).
- **M5** Output kini English-only (campuran Indonesia/Inggris dihapus).
- **M6** Menghapus `ManualDetector.Detect`/`IsAvailable` yang mati (hanya
  `NewManualResult` yang dipakai).
- **M7** Error command bawaan dikembalikan (tidak ditelan) agar verbose mode bisa
  menampilkan mengapa deteksi gagal.

### Low
- **L1** docker/podman/distrobox/toolbox tidak lagi cocok hanya dengan substring
  telanjang di PATH; mensyaratkan path berada di lokasi container/runtime nyata.
- **L2** Deteksi AppImage tidak lagi berjalan rekursif di `/opt`; memindai
  direktori aplikasi user terbatas dengan cap jumlah file.
- **L3** `spkg all` menghormati `--no-cache` dan memakai cache bersama.
- **L4** README tidak lagi mengklaim detector `paru`/`yay` yang tidak didukung.
- **L5** `confidence` dihilangkan dari JSON/YAML saat 0 (konsisten dengan output
  list).
- **L6** Field `Confidence` didokumentasikan beserta skalanya.

## [1.0.0] - awal
- Deteksi package-manager inti di 20+ manager dan runtime/version manager.
- Berbagai format output (pretty, json, yaml, table, plain), subcommand `explain`,
  `all`, dan `cache`.
