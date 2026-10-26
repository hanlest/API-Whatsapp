$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Dest = Join-Path $Root "third_party\whatsmeow"
$Version = (Select-String -Path (Join-Path $Root "src\go.mod") -Pattern "go.mau.fi/whatsmeow" | Select-Object -First 1).Line
if ($Version -notmatch "v0\.0\.0-[0-9]+-[a-f0-9]+") { throw "Could not parse whatsmeow version from go.mod" }
$Ver = $Matches[0]
$Src = Join-Path $env:GOMODCACHE "go.mau.fi\whatsmeow@$Ver"
if (-not (Test-Path $Src)) { go mod download "go.mau.fi/whatsmeow@$Ver" }
if (Test-Path $Dest) { Remove-Item -Recurse -Force $Dest }
Copy-Item -Recurse $Src $Dest
attrib -R "$Dest\*" /S /D 2>$null

$sendGo = Join-Path $Dest "send.go"
$py = Join-Path $Root "scripts\patch_whatsmeow_status_send.py"
if (-not (Get-Command python -ErrorAction SilentlyContinue)) {
  throw "python required to patch send.go (run patch_whatsmeow_status_send.py)"
}
& python $py $sendGo
if ($LASTEXITCODE -ne 0) { throw "Failed to patch send.go (exit $LASTEXITCODE)" }
Write-Host "whatsmeow vendored and patched at $Dest"
