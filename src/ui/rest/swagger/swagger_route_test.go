package swagger

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func TestDocsPaths(t *testing.T) {
	app := fiber.New()
	Register(app)

	for _, path := range []string{"/docs", "/docs/", "/docs/openapi.yaml"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			require.NotEqualf(t, 404, resp.StatusCode, "path %s body=%s", path, string(body))
		})
	}
}
