# Inspect4Oracle Build Guide

This document provides the specific commands and steps to build cross-platform executables for the `Inspect4Oracle` application.

## 1. Prerequisites

Before you start building, please ensure you have met the following conditions:

*   **Go Environment**: Go installed (latest stable version recommended). You can download and install it from the [official Go website](https://golang.org/dl/).
*   **Project Source Code**: Get the latest `Inspect4Oracle` project source code.
*   **(Optional) Git**: If you want to automatically get the version number from Git tags.

## 2. Building Inspect4Oracle

The following commands will build `Inspect4Oracle` and output the executable to the `build/` folder in the project root directory.

### 2.1. Prepare Version Information (Optional)

The build command uses `-X 'main.AppVersion=${VERSION_STRING}'` to embed the version number. You need to define the `VERSION_STRING` environment variable first.

*   **Linux / macOS**:
    ```bash
    export VERSION_STRING="0.1.0"
    # Or get it from Git tags:
    # export VERSION_STRING=$(git describe --tags --always)
    ```
*   **Windows (Command Prompt)**:
    ```cmd
    set VERSION_STRING=0.1.0
    ```
*   **Windows (PowerShell)**:
    ```powershell
    $env:VERSION_STRING="0.1.0"
    ```
Ensure that `AppVersion` in the `main.go` file is a **variable** (e.g., `var AppVersion = "dev"`) so that it can be overwritten by the linker.

### 2.2. Build for the Current Platform

*   **Windows (Command Prompt)**:
    ```cmd
    md build 2>nul
    go build -ldflags="-s -w -X 'main.AppVersion=%VERSION_STRING%'" -o build\inspect4oracle.exe .
    ```
*   **Windows (PowerShell)**:
    ```powershell
    New-Item -ItemType Directory -Force -Path "build" | Out-Null
    go build -ldflags="-s -w -X 'main.AppVersion=$env:VERSION_STRING'" -o build/inspect4oracle.exe .
    ```
*   **Linux / macOS**:
    ```bash
    mkdir -p build
    go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle .
    ```

### 2.3. Cross-Compile for Target Platforms

Ensure the `VERSION_STRING` environment variable is set (see 2.1).

| Target Platform Description | OS (GOOS) | Architecture (GOARCH) | Build Command (execute in Linux/macOS or PowerShell)                                                                                                                               |
| :-------------------------- | :-------- | :-------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Windows (64-bit, x86)       | `windows` | `amd64`               | `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_windows_amd64.exe .`                 |
| Linux (64-bit, x86)         | `linux`   | `amd64`               | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_linux_amd64 .`                       |
| Linux (64-bit, ARM)         | `linux`   | `arm64`               | `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_linux_arm64 .`                       |
| macOS (64-bit, Intel)       | `darwin`  | `amd64`               | `CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_darwin_amd64 .`                     |
| macOS (64-bit, ARM/M1/M2)   | `darwin`  | `arm64`               | `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_darwin_arm64 .`                     |

**To cross-compile in Windows Command Prompt (CMD)**:
You need to set the environment variables line by line:
```cmd
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
set OUTPUT_NAME=build\inspect4oracle_linux_amd64
go build -ldflags="-s -w -X 'main.AppVersion=%VERSION_STRING%'" -o %OUTPUT_NAME% .
```
(Repeat the `set` and `go build` commands for each target platform)

### 2.4. Brief Explanation of Build Parameters

*   **`CGO_ENABLED=0`**: Disables CGO. For pure Go projects, this ensures a statically linked executable is generated and avoids potential toolchain issues during cross-compilation.
*   **`GOOS`**: Specifies the target operating system (e.g., `linux`, `windows`, `darwin`).
*   **`GOARCH`**: Specifies the target processor architecture (e.g., `amd64`, `arm64`, `386`).
*   **`go build`**: The Go language compile command.
*   **`-o <output_path/filename>`**: Specifies the output path and name for the compiled executable.
*   **`-ldflags="..."`**: Flags to pass to the linker.
    *   **`-s`**: Omits the symbol table, reducing file size.
    *   **`-w`**: Omits the DWARF debugging information, reducing file size.
    *   **`-X 'package.Variable=value'`**: Sets the value of a string variable `Variable` (here, `AppVersion`) in the specified package (here, `main`) at build time.
*   **`.` (dot)**: Indicates compiling the `main` package in the current directory.

---

By following these steps, you can build platform-specific executables for `Inspect4Oracle`.
