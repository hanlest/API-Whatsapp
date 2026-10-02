package storage

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

func TestMiniCloudPutGet(t *testing.T) {
	const bucket = "test-bucket"
	const object = "dev/incoming/2026-01-01/file.txt"
	store := map[string][]byte{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/upload/storage/v1/b/"):
			_, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"name": object, "bucket": bucket})
		case r.Method == http.MethodGet && strings.Contains(r.URL.Query().Get("alt"), "media"):
			if _, ok := store[object]; !ok {
				store[object] = []byte("hello")
			}
			w.Write(store[object])
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	config.MiniCloudBaseURL = srv.URL
	config.MediaStorageDriver = "minicloud"
	config.MiniCloudDefaultBucket = bucket

	b := newMiniCloudBackend()
	ref, err := b.Put(t.Context(), "test-device", bucket, object, []byte("hello"), "text/plain")
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseMC(ref)
	if err != nil {
		t.Fatal(err)
	}
	data, _, err := b.Get(t.Context(), parsed)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("got %q", data)
	}
}
