# third_party

`whatsmeow/` no se versiona (está en `.gitignore`): contiene el fork local con parche de estados (`StatusBroadcastParticipants`).

Generar antes de compilar o en Docker:

```powershell
.\scripts\vendor-whatsmeow-status.ps1
```

```sh
./scripts/vendor-whatsmeow-status.sh
```

Ver `src/go.mod` (`replace go.mau.fi/whatsmeow => ../third_party/whatsmeow`).
