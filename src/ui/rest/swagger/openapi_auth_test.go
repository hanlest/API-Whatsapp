package swagger

import (
	"testing"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPatchOpenAPIStripsAuthWhenDisabled(t *testing.T) {
	prev := config.AppBasicAuthCredential
	t.Cleanup(func() { config.AppBasicAuthCredential = prev })
	config.AppBasicAuthCredential = nil

	raw := []byte(`
openapi: "3.0.0"
security:
  - basicAuth: []
paths:
  /ping:
    get:
      security:
        - basicAuth: []
components:
  securitySchemes:
    basicAuth:
      type: http
      scheme: basic
`)
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(raw, &doc))
	stripOpenAPIAuth(doc)
	_, hasGlobal := doc["security"]
	require.False(t, hasGlobal)
	components := doc["components"].(map[string]any)
	_, hasSchemes := components["securitySchemes"]
	require.False(t, hasSchemes)
	paths := doc["paths"].(map[string]any)
	get := paths["/ping"].(map[string]any)["get"].(map[string]any)
	_, hasOpSec := get["security"]
	require.False(t, hasOpSec)
}
