package abe

import (
	"github.com/fentec-project/gofe/abe"
)

func TEstMarshalFameCipher(cipher *abe.FAMECipher) string {
	raw, err := json.Marshal(cipher)

}
