# Windows Build Guide

This guide describes how to build `cg.exe` on a new Windows x64 development machine.

## Scope

The build target is a Windows executable. Running the application also requires a compatible game client, but the game client is not required to compile the project.

## Required tools

Install the following before building:

1. Git, to clone the repository.
2. Go 1.21.x. The project declares `go 1.21.1` in `go.mod`; use that release line until a newer Go version is explicitly verified.
3. MSYS2 with the MinGW-w64 x64 C/C++ toolchain. Fyne requires a C compiler for its Windows graphics integration.

The following must be available in the Windows `PATH`:

- `go`
- `gcc` (normally supplied by `C:\msys64\mingw64\bin`)

## Install the C toolchain

Install MSYS2 from its official distribution. Open **MSYS2 MinGW x64** and run:

```sh
pacman -Syu
pacman -S --needed mingw-w64-x86_64-toolchain
```

Then add the following directory to the Windows user or system `PATH` and open a new PowerShell window:

```text
C:\msys64\mingw64\bin
```

Verify the compiler from PowerShell:

```powershell
gcc --version
```

## Obtain the source and verify prerequisites

```powershell
git clone https://github.com/g70245/cg.git
Set-Location cg

go version
gcc --version
```

`go version` must report Go 1.21.x. This project has not yet been verified against newer Go releases.

## Download dependencies and build

Run from the repository root:

```powershell
go mod download
go build .
```

The successful build output is `cg.exe` in the repository root. To compile and start the GUI for a local smoke test:

```powershell
go run .
```

The first module download requires access to the configured Go module proxy and the project dependencies.

### Repeatable build command

For the standard developer build, use the repository script instead of entering the individual commands:

```powershell
.\scripts\build.ps1
```

It verifies that Go and GCC are available, checks the tracked `go.sum` dependencies, downloads modules, and writes the executable to `dist\cg.exe`. To use already downloaded modules, run:

```powershell
.\scripts\build.ps1 -SkipDependencyDownload
```

### Windows CI

GitHub Actions runs the verified build path on `windows-2022` for pushes and pull requests targeting `dev` or `main`, and it also supports manual runs. The workflow sets up Go `1.21.x`, enables CGO with GCC, runs `scripts/build.ps1`, verifies `dist\cg.exe`, and then runs `go test ./...` and `go vet ./...`.

The CI workflow is defined in `.github/workflows/windows-ci.yml`. It validates the build, automated tests, and static analysis; it does not launch the GUI, interact with a game client, or create a Fyne release package.

## Package a Windows release with Fyne

The project pins the packaging CLI at `fyne.io/tools v1.7.2`. The packaging script runs it through Go, so a globally installed `fyne` command is not required. The first packaging run requires access to the configured Go module proxy unless that module is already cached locally.

Use the repository-owned `app.png` icon and run:

```powershell
.\scripts\package.ps1
```

To embed a semantic application version in the Windows package:

```powershell
.\scripts\package.ps1 -AppVersion 1.0.0
```

The release package is written to `dist\CG.exe`. The script uses Fyne release mode and fails if the expected executable is not created.

For manual Fyne CLI use, install the same pinned version and ensure `%USERPROFILE%\go\bin` is on `PATH`:

```powershell
go install fyne.io/tools/cmd/fyne@v1.7.2
```

## Current limitations

- Automated tests cover selected package logic and filesystem boundaries, but they do not exercise a live game window, process memory, audio device, real game logs, or GUI workflows. Those integrations still require supervised manual testing.
- The program itself is Windows-specific because it invokes Win32 APIs.

## Verified environment and packaging findings

The following was verified on the development machine on 2026-07-15:

- Go `1.21.13` for `windows/amd64`.
- MSYS2 MinGW-w64 GCC `16.1.0`.
- `scripts/build.ps1` completed successfully and produced `dist\cg.exe`.
- Fyne CLI `v1.7.2` successfully produced a Windows release executable when invoked with an explicit app ID.

The verified direct Fyne command is:

```powershell
fyne package --target windows --src . --name CG --icon app.png --app-id com.github.g70245.cg --release
```

### Important output-location distinction

The direct `fyne package` command writes `CG.exe` to the repository root. In contrast, `scripts/build.ps1` writes the ordinary Go build to `dist\cg.exe`.

These are different files. If Explorer displays the icon for `dist\cg.exe` after only running the direct Fyne command, it is displaying the ordinary Go build rather than the newly packaged executable. Move the packaged root `CG.exe` to `dist\CG.exe` to replace it; Windows treats `CG.exe` and `cg.exe` as the same file name.

## Troubleshooting information to collect

When reporting a build failure, include:

```powershell
go version
go env GOOS GOARCH CGO_ENABLED CC GOPROXY
gcc --version
Get-Command go,gcc,fyne -ErrorAction SilentlyContinue
```

Also include the exact failing command and complete error output. Do not include game account data or personal filesystem information in the report.
