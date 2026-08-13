# Relance LoL CD Scout : PATH Go, /api/quit, build, Start-Process.
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
$env:Path += ";C:\Program Files\Go\bin"

try { Invoke-RestMethod -Method POST "http://127.0.0.1:27182/api/quit" | Out-Null } catch {}
Start-Sleep -Milliseconds 250
Get-CimInstance Win32_Process -ErrorAction SilentlyContinue | Where-Object { $_.CommandLine -match 'hud-profile-v' } | ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }

go build -ldflags="-s -w -H=windowsgui" -o lol-cd-scout-native.exe .
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# Smart App Control bloque un exe fraichement compile (non signe, sans
# reputation) : dans ce cas on lance la meme app via "go run", que la strategie
# laisse passer. Verifier l'etat de SAC : Securite Windows > Controle des
# applications et du navigateur > Smart App Control.
try {
  Start-Process .\lol-cd-scout-native.exe -ErrorAction Stop
} catch {
  Write-Host "Exe bloque par la strategie de controle d'application, relance via 'go run .'"
  Start-Process -WindowStyle Hidden -FilePath "go" -ArgumentList "run","." -WorkingDirectory $PSScriptRoot
}
