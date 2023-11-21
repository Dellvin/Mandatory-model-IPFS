package security

import (
	"encoding/base64"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/stretchr/testify/assert"
	"ipfs-senc/accumulator"
	"ipfs-senc/stribog"
	"testing"
)

func TestCommon(t *testing.T) {
	keys, err := setupFromFile()
	assert.NoError(t, err)
	wit, err := addUser(keys[0].Level, "alskjdfhiuaf", keys[1].SK)
	assert.NoError(t, err)
	ok, err := checkUser(keys[0].Level, wit, keys[0])
	assert.NoError(t, err)
	assert.True(t, ok)

	elem, err := curves.BLS12381(curves.BLS12381G1().Point).Scalar.SetBytes(hashStr("user_id")[:32])
	assert.NoError(t, err)
	wrongKey, err := new(accumulator.SecretKey).New(curves.BLS12381(curves.BLS12381G1().Point), []byte("asdf"))
	assert.NoError(t, err)
	wrongAcc, err := new(accumulator.Accumulator).WithElements(curves.BLS12381(curves.BLS12381G1().Point), wrongKey, []accumulator.Element{elem})
	assert.NoError(t, err)

	wrongWit, err := new(accumulator.MembershipWitness).New(elem, wrongAcc, wrongKey)
	assert.NoError(t, err)
	wrongWitByte, err := wrongWit.MarshalBinary()
	assert.NoError(t, err)

	ok, err = checkUser(keys[0].Level, wrongWitByte, keys[0])
	assert.NoError(t, err)
	assert.False(t, ok)

	err = deleteUser(keys[0].Level, "alskjdfhiuaf", keys[0].SK)
	assert.NoError(t, err)
	ok, err = checkUser(keys[0].Level, wit, keys[0])
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestMarshaller(t *testing.T) {
	curve := curves.BLS12381(curves.BLS12381G1().Point)
	acc, err := new(accumulator.Accumulator).New(curve)
	assert.NoError(t, err)
	buf, err := acc.MarshalBinary()

	acc2 := new(accumulator.Accumulator)
	err = acc2.UnmarshalBinary(buf)
	assert.NoError(t, err)
	assert.Equal(t, acc2, acc)
}

func TestElemGeneration(t *testing.T) {
	userStr := stribog.New256().Sum([]byte(base64.StdEncoding.EncodeToString([]byte("wrong_user"))))[:32]

	_, err := curves.BLS12381(curves.BLS12381G1().Point).Scalar.SetBytes(userStr)
	assert.NoError(t, err)

	userStr = stribog.New256().Sum([]byte("alskjdfhiuaf"))[:32]

	_, err = curves.BLS12381(curves.BLS12381G1().Point).Scalar.SetBytes(userStr)
	assert.NoError(t, err)
}
