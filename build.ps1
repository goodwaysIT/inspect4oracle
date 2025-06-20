param (
    [string]$Version = "0.1.0"
)

Write-Host "Starting build for version $Version..."

# Create build directory if it doesn't exist
if (-not (Test-Path -Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
}

# Define build targets
$targets = @(
    @{ GOOS = "windows"; GOARCH = "amd64"; OutputSuffix = ".exe" },
    @{ GOOS = "linux";   GOARCH = "amd64"; OutputSuffix = "" },
    @{ GOOS = "linux";   GOARCH = "arm64"; OutputSuffix = "" },
    @{ GOOS = "darwin";  GOARCH = "amd64"; OutputSuffix = "" },
    @{ GOOS = "darwin";  GOARCH = "arm64"; OutputSuffix = "" }
)

foreach ($target in $targets) {
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    $outputName = "build/inspect4oracle_${env:GOOS}_${env:GOARCH}${target.OutputSuffix}"
    
    Write-Host "Building for ${env:GOOS}/${env:GOARCH} -> $outputName"
    
    $ldflags = "-s -w -X 'main.AppVersion=$Version'"
    
    go build -ldflags "$ldflags" -o $outputName .
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed for ${env:GOOS}/${env:GOARCH}"
        # Clean up env vars
        Remove-Item env:GOOS
        Remove-Item env:GOARCH
        exit 1
    }
}

# Clean up env vars
if (Test-Path env:GOOS) { Remove-Item env:GOOS }
if (Test-Path env:GOARCH) { Remove-Item env:GOARCH }

Write-Host "Build process completed successfully. Files are in the 'build' directory."
