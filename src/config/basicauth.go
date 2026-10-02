package config

import "strings"

// BasicAuthEnabled reports whether HTTP Basic Auth is active for REST, UI, MCP, and /docs.
func BasicAuthEnabled() bool {
	return len(NormalizeBasicAuthCredentials(AppBasicAuthCredential)) > 0
}

// NormalizeBasicAuthCredentials drops empty entries and trims whitespace.
func NormalizeBasicAuthCredentials(creds []string) []string {
	if len(creds) == 0 {
		return nil
	}
	out := make([]string, 0, len(creds))
	for _, c := range creds {
		c = strings.TrimSpace(c)
		if c != "" {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
