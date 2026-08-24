package meetings

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptTokenRoundTrip(t *testing.T) {
	secret := "test-secret-value"
	plain := "refresh-token-abc123"
	enc, err := EncryptToken(secret, plain)
	require.NoError(t, err)
	require.NotEmpty(t, enc)
	require.NotEqual(t, plain, enc)

	out, err := DecryptToken(secret, enc)
	require.NoError(t, err)
	require.Equal(t, plain, out)
}

func TestDecryptTokenWrongSecretFails(t *testing.T) {
	enc, err := EncryptToken("secret-a", "token")
	require.NoError(t, err)
	_, err = DecryptToken("secret-b", enc)
	require.Error(t, err)
}
