# whatsmeow (fork local)

GoWA usa un `replace` en `src/go.mod` apuntando aquí para soportar **`StatusBroadcastParticipants`** en `SendRequestExtra` (lista explícita de JIDs al publicar en `status@broadcast`, paridad con Baileys `statusJidList`).

El parche está en `send.go`:

- Campo `StatusBroadcastParticipants []types.JID` en `SendRequestExtra`.
- Si el destino es `status@broadcast` y la lista no está vacía, se usa en lugar de la resolución por privacidad/contactos.

Para regenerar esta copia desde el módulo upstream:

```powershell
.\scripts\vendor-whatsmeow-status.ps1
```
