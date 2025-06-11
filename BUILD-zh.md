# Inspect4Oracle 构建指南

本文档提供为 `Inspect4Oracle` 应用程序构建跨平台可执行文件的具体命令和步骤。

## 1. 先决条件

在开始构建之前，请确保您已满足以下条件：

*   **Go 环境**: 已安装 Go (推荐最新稳定版本)。您可以从 [Go 官方网站](https://golang.org/dl/) 下载并安装。
*   **项目源码**: 获取最新的 `Inspect4Oracle` 项目源码。
*   **(可选) Git**: 如果您希望从 Git 标签自动获取版本号。

## 2. 构建 Inspect4Oracle

以下命令将构建 `Inspect4Oracle`，并将可执行文件输出到项目根目录下的 `build/` 文件夹中。

### 2.1. 准备版本信息 (可选)

构建命令中使用了 `-X 'main.AppVersion=${VERSION_STRING}'` 来嵌入版本号。您需要先定义 `VERSION_STRING` 环境变量。

*   **Linux / macOS**:
    ```bash
    export VERSION_STRING="0.1.0"
    # 或者从 Git 标签获取:
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
确保 `main.go` 文件中的 `AppVersion` 是一个**变量** (例如 `var AppVersion = "dev"`)，以便可以被链接器覆盖。

### 2.2. 为当前平台构建

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

### 2.3. 交叉编译目标平台

确保已设置 `VERSION_STRING` 环境变量 (参考 2.1)。

| 目标平台描述             | 操作系统 (GOOS) | 架构 (GOARCH) | 构建命令 (在 Linux/macOS 或 PowerShell 中执行)                                                                                                                               |
| :----------------------- | :-------------- | :------------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Windows (64-bit, x86)    | `windows`       | `amd64`       | `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_windows_amd64.exe .`                 |
| Linux (64-bit, x86)      | `linux`         | `amd64`       | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_linux_amd64 .`                       |
| Linux (64-bit, ARM)      | `linux`         | `arm64`       | `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_linux_arm64 .`                       |
| macOS (64-bit, Intel)    | `darwin`        | `amd64`       | `CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_darwin_amd64 .`                     |
| macOS (64-bit, ARM/M1/M2)| `darwin`        | `arm64`       | `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_darwin_arm64 .`                     |

**在 Windows 命令提示符 (CMD) 中执行交叉编译**:
您需要逐行设置环境变量：
```cmd
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
set OUTPUT_NAME=build\inspect4oracle_linux_amd64
go build -ldflags="-s -w -X 'main.AppVersion=%VERSION_STRING%'" -o %OUTPUT_NAME% .
```
(针对每个目标平台重复上述 `set` 和 `go build` 命令)

### 2.4. 构建参数简要说明

*   **`CGO_ENABLED=0`**: 禁用 CGO。对于纯 Go 项目，这可以确保生成静态链接的可执行文件，并避免交叉编译时的潜在工具链问题。
*   **`GOOS`**: 指定目标操作系统 (例如 `linux`, `windows`, `darwin`)。
*   **`GOARCH`**: 指定目标处理器架构 (例如 `amd64`, `arm64`, `386`)。
*   **`go build`**: Go 语言的编译命令。
*   **`-o <输出路径/文件名>`**: 指定编译后可执行文件的输出路径和名称。
*   **`-ldflags="..."`**: 传递给链接器的标志。
    *   **`-s`**: 省略符号表，减小文件大小。
    *   **`-w`**: 省略 DWARF 调试信息，减小文件大小。
    *   **`-X 'package.Variable=value'`**: 在构建时设置指定包中（此处为 `main` 包）的字符串变量 `Variable` (此处为 `AppVersion`) 的值。
*   **`.` (点)**: 表示在当前目录下编译 `main` 包。

---

遵循这些步骤，您可以为 `Inspect4Oracle` 构建特定平台的可执行程序。
