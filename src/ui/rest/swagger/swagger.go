package swagger

import (
	_ "embed"
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/gofiber/fiber/v3"
	"gopkg.in/yaml.v3"
)

//go:embed openapi.yaml
var embeddedOpenAPI []byte

var (
	specOnce sync.Once
	specRaw  []byte
	specErr  error
)

func loadSpecFile() ([]byte, error) {
	specOnce.Do(func() {
		if len(embeddedOpenAPI) > 0 {
			specRaw = embeddedOpenAPI
			return
		}
		candidates := openAPICandidates()
		for _, p := range candidates {
			b, err := os.ReadFile(p)
			if err == nil && len(b) > 0 {
				specRaw = b
				return
			}
		}
		specErr = fmt.Errorf("openapi spec not found (tried: %s)", strings.Join(candidates, ", "))
	})
	return specRaw, specErr
}

func openAPICandidates() []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(p string) {
		p = filepath.Clean(p)
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	add(filepath.Join("docs", "openapi.yaml"))
	add(filepath.Join("..", "docs", "openapi.yaml"))
	if wd, err := os.Getwd(); err == nil {
		add(filepath.Join(wd, "docs", "openapi.yaml"))
		add(filepath.Join(wd, "..", "docs", "openapi.yaml"))
	}
	return out
}

func stripOperationSecurity(doc map[string]any) {
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return
	}
	for _, item := range paths {
		methods, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for _, op := range methods {
			opMap, ok := op.(map[string]any)
			if !ok {
				continue
			}
			delete(opMap, "security")
		}
	}
}

func stripSecuritySchemes(doc map[string]any) {
	components, ok := doc["components"].(map[string]any)
	if !ok {
		return
	}
	delete(components, "securitySchemes")
}

func stripOpenAPIAuth(doc map[string]any) {
	delete(doc, "security")
	stripOperationSecurity(doc)
	stripSecuritySchemes(doc)
}

func patchServersForRequest(spec []byte, c fiber.Ctx) ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(spec, &doc); err != nil {
		return spec, err
	}
	base := strings.TrimSuffix(config.AppBasePath, "/")
	scheme := c.Protocol()
	if scheme == "" {
		scheme = "http"
	}
	host := c.Host()
	serverURL := fmt.Sprintf("%s://%s%s", scheme, host, base)
	doc["servers"] = []map[string]string{{"url": serverURL}}
	if !config.BasicAuthEnabled() {
		stripOpenAPIAuth(doc)
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return spec, err
	}
	return buf.Bytes(), nil
}

func openAPIHandler(c fiber.Ctx) error {
	raw, err := loadSpecFile()
	if err != nil {
		return c.Status(http.StatusNotFound).SendString(err.Error())
	}
	patched, err := patchServersForRequest(raw, c)
	if err != nil {
		patched = raw
	}
	c.Set("Content-Type", "application/yaml; charset=utf-8")
	c.Set("Cache-Control", "private, max-age=300")
	return c.Send(patched)
}

func docsSpecPath() string {
	base := strings.TrimSuffix(config.AppBasePath, "/")
	return base + "/docs/openapi.yaml"
}

func indexHandler(c fiber.Ctx) error {
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Cache-Control", "private, max-age=3600")
	return c.SendString(swaggerIndexHTML(docsSpecPath(), config.BasicAuthEnabled()))
}

func swaggerIndexHTML(specPath string, basicAuthEnabled bool) string {
	specPath = template.JSEscapeString(specPath)
	authCSS := ""
	persistAuth := "true"
	if !basicAuthEnabled {
		authCSS = " .swagger-ui .auth-wrapper { display: none !important; }"
		persistAuth = "false"
	}
	return `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>GoWA — Documentación API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
  <style>body { margin: 0; } .swagger-ui .topbar { display: none; }` + authCSS + `</style>
</head>
<body>
<div id="swagger-ui"></div>
<noscript><p style="padding:1rem;font-family:sans-serif">Activá JavaScript para ver la documentación.</p></noscript>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
<script>
(function () {
  var specURL = "` + specPath + `";
  function showErr(msg) {
    document.getElementById("swagger-ui").innerHTML =
      "<pre style=\"padding:1rem;font-family:monospace;white-space:pre-wrap\">" + msg + "</pre>";
  }
  if (typeof SwaggerUIBundle === "undefined") {
    showErr("No se pudo cargar Swagger UI desde el CDN. Revisá red o firewall.");
    return;
  }
  window.ui = SwaggerUIBundle({
    url: specURL,
    dom_id: "#swagger-ui",
    deepLinking: true,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    layout: "StandaloneLayout",
    persistAuthorization: ` + persistAuth + `,
    tryItOutEnabled: true,
    onFailure: function (data) {
      showErr("Error al cargar OpenAPI desde " + specURL + "\\n\\n" + (data && data.message ? data.message : ""));
    }
  });
})();
</script>
</body>
</html>`
}

// Register mounts Swagger UI at /docs and the OpenAPI spec at /docs/openapi.yaml.
func Register(router fiber.Router) {
	router.Get("/docs/openapi.yaml", openAPIHandler)
	router.Get("/docs/", indexHandler)
	router.Get("/docs", func(c fiber.Ctx) error {
		base := strings.TrimSuffix(config.AppBasePath, "/")
		return c.Redirect().To(base + "/docs/")
	})
}
