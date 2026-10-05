package whatsapp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatusRecipientJIDs(t *testing.T) {
	jids, err := ParseStatusRecipientJIDs([]string{
		"56911111111@s.whatsapp.net",
		"56911111111@s.whatsapp.net",
		"56922222222@s.whatsapp.net",
	})
	require.NoError(t, err)
	require.Len(t, jids, 2)
	assert.Equal(t, "56911111111@s.whatsapp.net", jids[0].String())
}

func TestParseStatusRecipientJIDs_DedupesByNonAD(t *testing.T) {
	jids, err := ParseStatusRecipientJIDs([]string{
		"56911111111:1@s.whatsapp.net",
		"56911111111@s.whatsapp.net",
	})
	require.NoError(t, err)
	assert.Len(t, jids, 1)
}
