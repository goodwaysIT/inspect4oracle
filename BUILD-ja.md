# Inspect4Oracle ビルドガイド

このドキュメントでは、`Inspect4Oracle` アプリケーションのクロスプラットフォーム実行可能ファイルをビルドするための具体的なコマンドと手順を説明します。

## 1. 前提条件

ビルドを開始する前に、以下の条件を満たしていることを確認してください：

*   **Go 環境**: Go がインストールされていること（最新の安定版を推奨）。[Go 公式サイト](https://golang.org/dl/) からダウンロードしてインストールできます。
*   **プロジェクトソースコード**: 最新の `Inspect4Oracle` プロジェクトソースコードを入手してください。
*   **(オプション) Git**: Git タグからバージョン番号を自動的に取得したい場合。

## 2. Inspect4Oracle のビルド

以下のコマンドは `Inspect4Oracle` をビルドし、実行可能ファイルをプロジェクトのルートディレクトリにある `build/` フォルダに出力します。

### 2.1. バージョン情報の準備 (オプション)

ビルドコマンドでは、バージョン番号を埋め込むために `-X 'main.AppVersion=${VERSION_STRING}'` を使用します。まず、`VERSION_STRING` 環境変数を定義する必要があります。

*   **Linux / macOS**:
    ```bash
    export VERSION_STRING="0.1.0"
    # または Git タグから取得:
    # export VERSION_STRING=$(git describe --tags --always)
    ```
*   **Windows (コマンドプロンプト)**:
    ```cmd
    set VERSION_STRING=0.1.0
    ```
*   **Windows (PowerShell)**:
    ```powershell
    $env:VERSION_STRING="0.1.0"
    ```
`main.go` ファイル内の `AppVersion` がリンカによって上書きできるよう、**変数**（例: `var AppVersion = "dev"`）であることを確認してください。

### 2.2. 現在のプラットフォーム向けにビルド

*   **Windows (コマンドプロンプト)**:
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

### 2.3. ターゲットプラットフォーム向けのクロスコンパイル

`VERSION_STRING` 環境変数が設定されていることを確認してください（2.1 を参照）。

| ターゲットプラットフォームの説明 | OS (GOOS) | アーキテクチャ (GOARCH) | ビルドコマンド (Linux/macOS または PowerShell で実行)                                                                                                                               |
| :------------------------------- | :-------- | :-------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Windows (64-bit, x86)            | `windows` | `amd64`               | `CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_windows_amd64.exe .`                 |
| Linux (64-bit, x86)              | `linux`   | `amd64`               | `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_linux_amd64 .`                       |
| Linux (64-bit, ARM)              | `linux`   | `arm64`               | `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_linux_arm64 .`                       |
| macOS (64-bit, Intel)            | `darwin`  | `amd64`               | `CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_darwin_amd64 .`                     |
| macOS (64-bit, ARM/M1/M2)        | `darwin`  | `arm64`               | `CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X 'main.AppVersion=${VERSION_STRING}'" -o build/inspect4oracle_darwin_arm64 .`                     |

**Windows コマンドプロンプト (CMD) でクロスコンパイルを実行する場合**:
環境変数を一行ずつ設定する必要があります：
```cmd
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
set OUTPUT_NAME=build\inspect4oracle_linux_amd64
go build -ldflags="-s -w -X 'main.AppVersion=%VERSION_STRING%'" -o %OUTPUT_NAME% .
```
(各ターゲットプラットフォームに対して上記の `set` と `go build` コマンドを繰り返します)

### 2.4. ビルドパラメータの簡単な説明

*   **`CGO_ENABLED=0`**: CGO を無効にします。純粋な Go プロジェクトの場合、これにより静的リンクされた実行可能ファイルが生成され、クロスコンパイル時の潜在的なツールチェーンの問題を回避できます。
*   **`GOOS`**: ターゲットのオペレーティングシステムを指定します（例: `linux`, `windows`, `darwin`）。
*   **`GOARCH`**: ターゲットのプロセッサアーキテクチャを指定します（例: `amd64`, `arm64`, `386`）。
*   **`go build`**: Go 言語のコンパイルコマンド。
*   **`-o <出力パス/ファイル名>`**: コンパイル後の実行可能ファイルの出力パスと名前を指定します。
*   **`-ldflags="..."`**: リンカに渡すフラグ。
    *   **`-s`**: シンボルテーブルを省略し、ファイルサイズを削減します。
    *   **`-w`**: DWARF デバッグ情報を省略し、ファイルサイズを削減します。
    *   **`-X 'package.Variable=value'`**: ビルド時に指定されたパッケージ（ここでは `main` パッケージ）の文字列変数 `Variable`（ここでは `AppVersion`）の値を設定します。
*   **`.` (ドット)**: 現在のディレクトリで `main` パッケージをコンパイルすることを示します。

---

これらの手順に従うことで、`Inspect4Oracle` のプラットフォーム固有の実行可能ファイルをビルドできます。
