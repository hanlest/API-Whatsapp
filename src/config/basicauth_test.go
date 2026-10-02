package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeBasicAuthCredentials(t *testing.T) {
	require.Nil(t, NormalizeBasicAuthCredentials(nil))
	require.Nil(t, NormalizeBasicAuthCredentials([]string{"", "  "}))
	require.Equal(t, []string{"a:b"}, NormalizeBasicAuthCredentials([]string{" a:b ", ""}))
}

func TestBasicAuthEnabled(t *testing.T) {
	prev := AppBasicAuthCredential
	t.Cleanup(func() { AppBasicAuthCredential = prev })

	AppBasicAuthCredential = nil
	require.False(t, BasicAuthEnabled())

	AppBasicAuthCredential = []string{"user:pass"}
	require.True(t, BasicAuthEnabled())

	AppBasicAuthCredential = []string{"  "}
	require.False(t, BasicAuthEnabled())
}
