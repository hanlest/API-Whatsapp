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
$c = Get-Content $sendGo -Raw
if ($c -notmatch "StatusBroadcastParticipants") {
  $c = $c -replace '(AdditionalNodes \*\[]waBinary\.Node\r?\n\})', "AdditionalNodes *[]waBinary.Node`r`n`t// StatusBroadcastParticipants overrides status privacy resolution when sending to status@broadcast.`r`n`tStatusBroadcastParticipants []types.JID`r`n}"
  $old = @'
		} else {
			groupParticipants, err = cli.getBroadcastListParticipants(ctx, to)
			if err != nil {
				err = fmt.Errorf("failed to get broadcast list members: %w", err)
				return
			}
		}
'@
  $new = @'
		} else if len(req.StatusBroadcastParticipants) > 0 && to == types.StatusBroadcastJID {
			groupParticipants = append([]types.JID(nil), req.StatusBroadcastParticipants...)
			ownIDNonAD := cli.getOwnID().ToNonAD()
			if ownIDNonAD.IsEmpty() {
				err = ErrNotLoggedIn
				return
			}
			selfIndex := -1
			for i, participant := range groupParticipants {
				if participant.User == ownIDNonAD.User {
					selfIndex = i
					break
				}
			}
			if selfIndex < 0 {
				groupParticipants = append(groupParticipants, ownIDNonAD)
			}
		} else {
			groupParticipants, err = cli.getBroadcastListParticipants(ctx, to)
			if err != nil {
				err = fmt.Errorf("failed to get broadcast list members: %w", err)
				return
			}
		}
'@
  $c2 = $c.Replace($old, $new)
  if ($c2 -eq $c) { throw "Failed to patch send.go (upstream layout changed?)" }
  Set-Content -Path $sendGo -Value $c2 -NoNewline
}
Write-Host "whatsmeow vendored and patched at $Dest"
