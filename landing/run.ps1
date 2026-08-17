$env:Path += ";C:\Program Files\Go\bin"
Set-Location $PSScriptRoot
Write-Host "Landing CD Scout → http://127.0.0.1:27183"
go run ./tools/serve
