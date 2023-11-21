package security

import (
	"encoding/base64"
	"fmt"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"ipfs-senc/accumulator"
	"ipfs-senc/stribog"
	"os"
	"strconv"
)

type AccumulatorKey struct {
	Level int
	SK    *accumulator.SecretKey
	PK    *accumulator.PublicKey
}

const levelCount = 3

// admin only
func setupFromFile() ([]AccumulatorKey, error) {
	var keys = []AccumulatorKey{}

	for i := 0; i < levelCount; i++ {
		err := func() error {
			curve := curves.BLS12381(curves.BLS12381G1().Point)
			if f, err := os.OpenFile(strconv.Itoa(i)+".txt", os.O_RDWR|os.O_CREATE, 0777); err != nil {
				return fmt.Errorf("failed to setupFromFile: %w", err)
			} else {
				defer func() {
					if err := f.Close(); err != nil {
						fmt.Println(err)
						return
					}

				}()
				acc, err := new(accumulator.Accumulator).New(curve)
				if err != nil {
					return fmt.Errorf("failed to New: %w", err)
				}
				buf, err := acc.MarshalBinary()
				if err != nil {
					return fmt.Errorf("failed to MarshalBinary: %w", err)
				}

				if _, err := f.Write(buf); err != nil {
					return fmt.Errorf("failed to Write: %w", err)
				}
				sk, err := new(accumulator.SecretKey).New(curve, hashStr(""))
				if err != nil {
					return fmt.Errorf("failed to create secret key: %w", err)
				}
				pk, err := sk.GetPublicKey(curve)
				if err != nil {
					return fmt.Errorf("failed to GetPublicKey: %w", err)
				}
				keys = append(keys, AccumulatorKey{i, sk, pk})
			}
			return nil
		}()
		if err != nil {
			return nil, err
		}
	}

	return keys, nil
}

func addUser(level int, user_id string, sk *accumulator.SecretKey) ([]byte, error) {
	acc, err := getAccumulator(level)
	if err != nil {
		return nil, fmt.Errorf("failed to getAccumulator: %w", err)
	}

	elem, err := curves.BLS12381(curves.BLS12381G1().Point).Scalar.SetBytes(hashStr(user_id)[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to SetBytes: %w", err)
	}

	acc, err = acc.Add(sk, elem)
	if err != nil {
		return nil, fmt.Errorf("failed to Add: %w", err)
	}

	wit, err := new(accumulator.MembershipWitness).New(elem, acc, sk)
	if err != nil {
		return nil, fmt.Errorf("failed to New: %w", err)
	}

	body, err := acc.MarshalBinary()

	if err = rewriteFile(level, body); err != nil {
		return nil, fmt.Errorf("failed to rewriteFile: %w", err)
	}

	witBody, err := wit.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to MarshalBinary: %w", err)
	}

	return witBody, nil
}

func checkUser(level int, witByte []byte, keys AccumulatorKey) (bool, error) {
	acc, err := getAccumulator(level)
	if err != nil {
		return false, fmt.Errorf("failed to getAccumulator: %w", err)
	}

	wit := new(accumulator.MembershipWitness)
	if err := wit.UnmarshalBinary(witByte); err != nil {
		return false, fmt.Errorf("failed to UnmarshalBinary: %w", err)
	}

	if err := wit.Verify(keys.PK, acc); err != nil {
		return false, nil
	}

	return true, nil
}

func deleteUser(level int, user_id string, key *accumulator.SecretKey) error {
	acc, err := getAccumulator(level)
	if err != nil {
		return fmt.Errorf("failed to getAccumulator: %w", err)
	}

	elem := curves.BLS12381(&curves.PointBls12381G1{}).Scalar.Hash(hashStr(user_id))
	if err != nil {
		return fmt.Errorf("failed to SetBytes: %w", err)
	}

	acc, err = acc.Remove(key, elem)
	if err != nil {
		return fmt.Errorf("failed to Remove: %w", err)
	}

	body, err := acc.MarshalBinary()

	if err = rewriteFile(level, body); err != nil {
		return fmt.Errorf("failed to rewriteFile: %w", err)
	}

	return nil
}

func getAccumulator(level int) (*accumulator.Accumulator, error) {
	if level >= levelCount {
		return nil, fmt.Errorf("level is bigger that supported. Max is " + strconv.Itoa(levelCount))
	}

	buf, err := os.ReadFile(strconv.Itoa(level) + ".txt")
	if err != nil {
		return nil, fmt.Errorf("failed to ReadFile: %w", err)
	}

	acc := new(accumulator.Accumulator)
	if err1 := acc.UnmarshalBinary(buf); err1 != nil {
		return nil, fmt.Errorf("failed to UnmarshalBinary: %w", err1)
	}

	return acc, nil
}

func rewriteFile(level int, body []byte) error {
	if err := os.Truncate(strconv.Itoa(level)+".txt", 0); err != nil {
		return fmt.Errorf("failed to Truncate: %w", err)
	}

	f, err := os.Open(strconv.Itoa(level) + ".txt")
	if err != nil {
		return fmt.Errorf("failed to Open: %w", err)
	}
	defer f.Close()

	if err = os.WriteFile(strconv.Itoa(level)+".txt", body, os.ModeAppend); err != nil {
		return fmt.Errorf("failed to Write: %w", err)
	}

	return nil
}

func hashStr(str string) []byte {
	return stribog.New256().Sum([]byte(base64.StdEncoding.EncodeToString([]byte(str))))[:32]
}
