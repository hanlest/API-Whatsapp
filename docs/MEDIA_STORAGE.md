# Media storage (MiniCloud)

GoWA can store downloaded WhatsApp media on local disk (`statics/media`) or in [MiniCloud](https://github.com/) via its GCS-compatible HTTP API.

## Global environment

| Variable | Description |
|----------|-------------|
| `MEDIA_STORAGE_DRIVER` | `local` (default) or `minicloud` |
| `MINICLOUD_BASE_URL` | MiniCloud root URL, e.g. `http://localhost:8082` |
| `MINICLOUD_DEFAULT_BUCKET` | Bucket when a device has no per-device bucket |
| `MINICLOUD_AUTO_CREATE_BUCKET` | When true, first upload registers the bucket and binds a primary storage named after the device |
| `MINICLOUD_DEFAULT_STORAGE_ID` | Optional: same storage id for every device; when empty, uses the device slot name (e.g. `Estar-Siempre`) |
| `MINICLOUD_LOCAL_STORAGE_QUOTA` | Quota when auto-creating a new per-device local storage (clones `rootPath` from an existing local backend) |
| `MEDIA_PROXY_ENABLED` | When true, `mc:` refs are exposed via authenticated proxy URLs |

## Per-device configuration

Each device can use its own bucket (or share one). Objects are stored under a folder named after the device slot (sanitized), e.g. `Estar-Siempre/incoming/2026-09-14/...`.

```bash
curl -u user:pass -X PUT "http://localhost:3000/devices/Estar%20Siempre/media-storage" \
  -H "Content-Type: application/json" \
  -d '{"bucket":"cliente-a","folder_name":"","enabled":true}'
```

List all configs:

```bash
curl -u user:pass http://localhost:3000/media-storage/configs
```

## MiniCloud prerequisites

1. Ensure MiniCloud has at least one **local** storage backend (template for `rootPath` when auto-creating per-device storages).
2. Buckets can be created manually in admin, or left to GoWA when `MINICLOUD_AUTO_CREATE_BUCKET=true` (registers bucket + assigns primary storage on first failed upload).
3. Upload API: `POST {MINICLOUD_BASE_URL}/storage/upload/storage/v1/b/{bucket}/o?uploadType=multipart`

Without auto-create, uploading to an unknown bucket returns **HTTP 400** (`BUCKET_NOT_CONFIGURED`: bucket without assigned storages).

## References in webhooks

Auto-downloaded media paths use the `mc:{bucket}/{objectKey}` form when the driver is `minicloud`. Public download URLs use:

`GET {APP_BASE_PATH}/devices/{device_id}/media/{objectKey...}`

(requires the same Basic Auth as the REST API).

## Local driver layout

With `MEDIA_STORAGE_DRIVER=local`, files are written under:

`statics/media/{deviceFolder}/incoming/{date}/...`

Per-device bucket settings are ignored; folder overrides still apply.
