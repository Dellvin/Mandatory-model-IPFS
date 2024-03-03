package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestABE_OK(t *testing.T) {
	encrypted, depAuth, levelAuth, err := Encrypt(1, 4, []byte("secret msg"))
	require.NoError(t, err)
	require.NotNil(t, encrypted)

	decrypted, err := Decrypt(1, 4, encrypted, depAuth, levelAuth)
	require.NoError(t, err)
	require.Equal(t, []byte("secret msg"), decrypted)

}

func TestABE_OK_withHigherPrivilege(t *testing.T) {
	encrypted, depAuth, levelAuth, err := Encrypt(1, 2, []byte("secret msg"))
	require.NoError(t, err)
	require.NotNil(t, encrypted)

	decrypted, err := Decrypt(1, 4, encrypted, depAuth, levelAuth)
	require.NoError(t, err)
	require.Equal(t, []byte("secret msg"), decrypted)

}

func TestABE_Fail_differentDep(t *testing.T) {
	encrypted, depAuth, levelAuth, err := Encrypt(1, 2, []byte("secret msg"))
	require.NoError(t, err)
	require.NotNil(t, encrypted)

	decrypted, err := Decrypt(2, 2, encrypted, depAuth, levelAuth)
	require.EqualError(t, err, "failed to GenerateAttribKeys: attribute not found in secret key")
	require.Nil(t, decrypted)
}

func TestABE_Fail_lowPrivilege(t *testing.T) {
	encrypted, depAuth, levelAuth, err := Encrypt(1, 2, []byte("secret msg"))
	require.NoError(t, err)
	require.NotNil(t, encrypted)

	decrypted, err := Decrypt(1, 1, encrypted, depAuth, levelAuth)
	require.EqualError(t, err, "failed to GenerateAttribKeys: attribute not found in secret key")
	require.Nil(t, decrypted)
}
