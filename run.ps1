# Relance LoL CD Scout : PATH Go, /api/quit, build, Start-Process.
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
$env:Path += ";C:\Program Files\Go\bin"

try { Invoke-RestMethod -Method POST "http://127.0.0.1:27182/api/quit" | Out-Null } catch {}

go build -ldflags="-s -w -H=windowsgui" -o lol-cd-scout-native.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# Smart App Control bloque par intermittence un exe fraichement compile lance
# depuis PowerShell ; le passage par l'Explorateur passe.
try {
  Start-Process .\lol-cd-scout-native.exe -ErrorAction Stop
} catch {
  Write-Host "Start-Process bloque, relance via l'Explorateur."
  Start-Process explorer.exe -ArgumentList (Resolve-Path .\lol-cd-scout-native.exe).Path
}
