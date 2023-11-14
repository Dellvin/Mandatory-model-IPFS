package pkg

import (
	"encoding/base32"
	"errors"
	"fmt"
	senc "github.com/jbenet/go-simple-encrypt"
	ipfssenc "github.com/jbenet/ipfs-senc"
	"ipfs-senc/stribog"
)

func GetSencKey(key string, randomIfNone bool) (ipfssenc.Key, error) {
	NilKey := ipfssenc.Key(nil)

	var k []byte
	var err error
	if key != "" {
		k, err = base32.StdEncoding.DecodeString(key)
	} else if randomIfNone { // random key
		randBuf, err := senc.RandomKey()
		if err != nil {
			return nil, fmt.Errorf("failed to RandomKey: %w", err)
		}
		sb := stribog.New256()
		_, err = sb.Write(randBuf)
		k = sb.Sum([]byte{})
	} else {
		err = errors.New("Please enter a key with --key")
	}
	if err != nil {
		return NilKey, err
	}

	return ipfssenc.Key(k), nil
}
