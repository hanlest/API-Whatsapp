package storage

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

func TestPutAutoCreatesBucketWithDeviceStorageID(t *testing.T) {
	const bucket = "new-bucket"
	const deviceID = "Estar Siempre"
	const storageID = "Estar-Siempre"
	var created bool
	var createdStorage bool
	var patched bool

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/upload/storage/v1/b/"):
			if !patched && !createdStorage {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":{"code":"BUCKET_NOT_CONFIGURED","message":"Bucket sin almacenamientos"}}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"name": "obj", "bucket": bucket})
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/storage/buckets"):
			created = true
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/storage/storages"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"storages": []map[string]any{{
					"id": "seed", "provider": "local",
					"config": map[string]string{"rootPath": "/data/storage", "maxBytes": "100GB"},
				}},
			})
		case r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/storage/storages/"+storageID):
			patched = true
			w.WriteHeader(http.StatusNotFound)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/storage/storages"):
			createdStorage = true
			patched = true
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	config.MiniCloudBaseURL = srv.URL
	config.MiniCloudAutoCreateBucket = true
	config.MiniCloudDefaultStorageID = ""

	b := newMiniCloudBackend()
	_, err := b.Put(t.Context(), deviceID, bucket, "dev/incoming/a.txt", []byte("x"), "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	if !created || !createdStorage {
		t.Fatalf("created=%v createdStorage=%v patched=%v", created, createdStorage, patched)
	}
}

func TestResolveMiniCloudStorageIDUsesDeviceName(t *testing.T) {
	config.MiniCloudDefaultStorageID = ""
	if got := resolveMiniCloudStorageID("Mi Device"); got != "Mi-Device" {
		t.Fatalf("got %q", got)
	}
}
