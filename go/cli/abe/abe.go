package abe

import (
	"encoding/json"
	"fmt"
	"github.com/fentec-project/gofe/abe"
	"math"
)

const (
	securityLevelSecretly = iota + math.MaxInt - 3
	securityLevelAbsolutelySecretly
	securityLevelSpacialImportance
)

type SecretFile struct {
	Cipher []byte
	PubKey []byte
	SecKey []byte
}

func EncryptFile(department, securityLevel int, file []byte) (SecretFile, error) {
	policy, err := createPolicy(department, securityLevel)
	if err != nil {
		return SecretFile{}, fmt.Errorf("failed to createPolicy: %w", err)
	}
	a := abe.NewFAME()
	pubKey, secKey, err := a.GenerateMasterKeys()
	if err != nil {
		return SecretFile{}, fmt.Errorf("failed to GenerateMasterKeys: %w", err)
	}
	msp, err := abe.BooleanToMSP(policy, false)
	if err != nil {
		return SecretFile{}, fmt.Errorf("failet to BooleanToMSP: %w", err)
	}

	cipher, err := a.Encrypt(string(file), msp, pubKey)
	if err != nil {
		return SecretFile{}, fmt.Errorf("failed to Encrypt: %w", err)
	}

	byteCipher, err := json.Marshal(cipher)
	if err != nil {
		return SecretFile{}, fmt.Errorf("failed to Marshal byteCipher: %w", err)
	}

	bytePubKey, err := json.Marshal(pubKey)
	if err != nil {
		return SecretFile{}, fmt.Errorf("failed to Marshal bytePubKey: %w", err)
	}

	byteSecKey, err := json.Marshal(secKey)
	if err != nil {
		return SecretFile{}, fmt.Errorf("failed to Marshal bytePubKey: %w", err)
	}

	return SecretFile{
		Cipher: byteCipher,
		PubKey: bytePubKey,
		SecKey: byteSecKey,
	}, nil
}

func createPolicy(department, securityLevel int) (string, error) {
	policy := fmt.Sprintf("%d AND (", department)
	switch securityLevel {
	case 0:
		policy = fmt.Sprintf("%s%d OR %d OR %d", policy, securityLevelAbsolutelySecretly, securityLevelSecretly, securityLevelSpacialImportance)
	case 1:
		policy = fmt.Sprintf("%s%d OR %d", policy, securityLevelAbsolutelySecretly, securityLevelSecretly)
	case 2:
		policy = fmt.Sprintf("%s%d", policy, securityLevelSecretly)
	default:
		return "", fmt.Errorf("unknown policy")
	}

	return policy + ")", nil
}
