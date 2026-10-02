package swagger

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func TestOpenAPISpecLoadsFromRepoDocs(t *testing.T) {
	moduleRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	repoDocs := filepath.Join(moduleRoot, "..", "docs", "openapi.yaml")
	if _, err := os.Stat(repoDocs); err != nil {
		t.Skip("repo docs/openapi.yaml not found")
	}
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(moduleRoot))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	specOnce = sync.Once{}
	specRaw = nil
	specErr = nil

	raw, err := loadSpecFile()
	require.NoError(t, err)
	require.Contains(t, string(raw), "openapi:")
}

func TestRegisterServesIndex(t *testing.T) {
	app := fiber.New()
	Register(app)
	req := httptest.NewRequest("GET", "/docs/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

	req2 := httptest.NewRequest("GET", "/docs", nil)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	require.Contains(t, []int{200, 303}, resp2.StatusCode)
}
