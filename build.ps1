# build.ps1 - 一键构建 Tea Cloud
# 用法: .\build.ps1

$ErrorActionPreference = "Stop"
$root = $PSScriptRoot

Write-Host "=== Building Tea Cloud ===" -ForegroundColor Cyan

# 1. 构建前端
Write-Host "[1/2] Building frontend..." -ForegroundColor Yellow
Push-Location "$root\frontend"
npm run build
if ($LASTEXITCODE -ne 0) { Write-Host "Frontend build failed!" -ForegroundColor Red; exit 1 }
Pop-Location
Write-Host "Frontend built -> backend/static/" -ForegroundColor Green

# 2. 构建后端（嵌入前端资源）
Write-Host "[2/2] Building backend..." -ForegroundColor Yellow
Push-Location "$root\backend"

# 嵌入 Windows 版本信息和 manifest（需要 goversioninfo，首次自动安装）
if (-not (Get-Command goversioninfo -ErrorAction SilentlyContinue)) {
    Write-Host "  Installing goversioninfo..." -ForegroundColor Gray
    go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest
}
Write-Host "  Generating resource file..." -ForegroundColor Gray
goversioninfo -64 -o resource.syso versioninfo.json
if ($LASTEXITCODE -ne 0) { Write-Host "goversioninfo failed!" -ForegroundColor Red; exit 1 }

go build -ldflags="-s -w" -o "$root\tea-cloud.exe" .
if ($LASTEXITCODE -ne 0) { Write-Host "Backend build failed!" -ForegroundColor Red; exit 1 }

# 清理临时资源文件
Remove-Item -Force resource.syso -ErrorAction SilentlyContinue
Pop-Location
Write-Host "Backend built -> tea-cloud.exe" -ForegroundColor Green

$size = [math]::Round((Get-Item "$root\tea-cloud.exe").Length / 1MB, 1)
Write-Host ""
Write-Host "=== Build successful! ($size MB) ===" -ForegroundColor Cyan
Write-Host "Run: .\tea-cloud.exe" -ForegroundColor White
Write-Host "Then open: http://localhost:8080" -ForegroundColor White
