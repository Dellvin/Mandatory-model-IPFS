package security

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/coinbase/kryptology/pkg/core/curves"

	"server/accumulator"
	"server/stribog"
)

type AccumulatorKey struct {
	Level int
	SK    *accumulator.SecretKey
	PK    *accumulator.PublicKey
	Acc   *accumulator.Accumulator
	Path  string
}

const (
	typeLevel = iota
	typeDepartment
)

const levelCount = 5

func Add(level, department int, data []byte) ([]byte, []byte, error) { // data is a pk from abe
	accLevel, err := getOrCreateAccumulatorByType(level, 0, typeLevel)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to getLevelAccumulator: %w", err)
	}
	witLevel, err := accLevel.add(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add to accLevel: %w", err)
	}

	accDepartment, err := getOrCreateAccumulatorByType(0, department, typeDepartment)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to getLevelAccumulator: %w", err)
	}
	witDepartment, err := accDepartment.add(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add to accDepartment: %w", err)
	}
	str := base64.StdEncoding.EncodeToString(witLevel)
	witLevel1, _ := base64.StdEncoding.DecodeString(str)
	fmt.Println(accLevel.check(witLevel1))
	return witLevel, witDepartment, nil
}

func (acc *AccumulatorKey) add(data []byte) ([]byte, error) {
	elem, err := curves.BLS12381(curves.BLS12381G1().Point).Scalar.SetBytes(hashBytes(data))
	if err != nil {
		return nil, fmt.Errorf("failed to SetBytes: %w", err)
	}

	acc.Acc, err = acc.Acc.Add(acc.SK, elem)
	if err != nil {
		return nil, fmt.Errorf("failed to Add: %w", err)
	}

	wit, err := new(accumulator.MembershipWitness).New(elem, acc.Acc, acc.SK)
	if err != nil {
		return nil, fmt.Errorf("failed to New: %w", err)
	}

	body, err := acc.Acc.MarshalBinary()

	if err = rewriteFileByPath(acc.Path+".txt", body); err != nil {
		return nil, fmt.Errorf("failed to rewriteFile: %w", err)
	}

	witBody, err := wit.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to MarshalBinary: %w", err)
	}

	return witBody, nil
}

func Check(level, department int, witLevel, witDepartment []byte) error {
	if level >= levelCount {
		return fmt.Errorf("level is bigger that supported. Max is " + strconv.Itoa(levelCount))
	}
	filePath := strconv.Itoa(level) + "_level"

	acc, err := getAccumulatorByPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to getAccumulatorByPath level: %w", err)
	}

	if ok, err := acc.check(witLevel); err != nil {
		return fmt.Errorf("failed to check level: %w", err)
	} else if !ok {
		return fmt.Errorf("failed to verify wit level")
	}

	filePath = strconv.Itoa(department) + "_department"
	acc, err = getAccumulatorByPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to getAccumulatorByPath dep: %w", err)
	}

	if ok, err := acc.check(witDepartment); err != nil {
		return fmt.Errorf("failed to check dep: %w", err)
	} else if !ok {
		return fmt.Errorf("failed to verify wit dep")
	}

	return nil
}

func (acc *AccumulatorKey) check(witByte []byte) (bool, error) {
	wit := new(accumulator.MembershipWitness)
	if err := wit.UnmarshalBinary(witByte); err != nil {
		return false, fmt.Errorf("failed to UnmarshalBinary: %w", err)
	}

	if err := wit.Verify(acc.PK, acc.Acc); err != nil {
		return false, fmt.Errorf("failed to Verify: %w", err)
	}

	return true, nil
}

func Delete(level, department int, data []byte) error {
	if level >= levelCount {
		return fmt.Errorf("level is bigger that supported. Max is " + strconv.Itoa(levelCount))
	}
	filePath := strconv.Itoa(level) + "_level"

	acc, err := getAccumulatorByPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to getAccumulatorByPath level: %w", err)
	}

	if err = acc.delete(data); err != nil {
		return fmt.Errorf("failed to delete level: %w", err)
	}

	filePath = strconv.Itoa(department) + "_department"

	acc, err = getAccumulatorByPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to getAccumulatorByPath department: %w", err)
	}

	if err = acc.delete(data); err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	return nil
}

func (acc *AccumulatorKey) delete(data []byte) error {
	elem := curves.BLS12381(&curves.PointBls12381G1{}).Scalar.Hash(hashBytes(data))

	a, err := acc.Acc.Remove(acc.SK, elem)
	if err != nil {
		return fmt.Errorf("failed to Remove: %w", err)
	}

	acc.Acc = a

	body, err := acc.Acc.MarshalBinary()

	if err = rewriteFileByPath(acc.Path+".txt", body); err != nil {
		return fmt.Errorf("failed to rewriteFile: %w", err)
	}

	return nil
}

func rewriteFileByPath(path string, body []byte) error {
	if err := os.Truncate(path, 0); err != nil {
		return fmt.Errorf("failed to Truncate: %w", err)
	}

	if err := os.WriteFile(path, body, os.ModeAppend); err != nil {
		return fmt.Errorf("failed to Write: %w", err)
	}

	return nil
}

func getOrCreateAccumulatorByType(level, department int, which int) (AccumulatorKey, error) {
	var filePath string
	switch which {
	case typeLevel:
		if level >= levelCount {
			return AccumulatorKey{}, fmt.Errorf("level is bigger that supported. Max is " + strconv.Itoa(levelCount))
		}
		filePath = strconv.Itoa(level) + "_level"
	case typeDepartment:
		filePath = strconv.Itoa(department) + "_department"
	default:
		return AccumulatorKey{}, fmt.Errorf("unknown type of accumulator")
	}

	if _, err := os.Stat(filePath + ".txt"); !errors.Is(err, os.ErrNotExist) {
		acc, err := getAccumulatorByPath(filePath)
		if err != nil {
			return AccumulatorKey{}, fmt.Errorf("failed to getAccumulatorByPath: %w", err)
		}

		return acc, err
	}

	// file not exists
	curve := curves.BLS12381(curves.BLS12381G1().Point)
	f, err := os.OpenFile(filePath+".txt", os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to setupFromFile: %w", err)
	}

	defer f.Close()
	acc, err := new(accumulator.Accumulator).New(curve)
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to New: %w", err)
	}
	buf, err := acc.MarshalBinary()
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to MarshalBinary: %w", err)
	}

	if _, err := f.Write(buf); err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to Write: %w", err)
	}
	sk, err := new(accumulator.SecretKey).New(curve, hashStr(""))
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to create secret key: %w", err)
	}
	pk, err := sk.GetPublicKey(curve)
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to GetPublicKey: %w", err)
	}

	pbFile, err := os.Create(filePath + ".pk")
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to Create pb: %w", err)
	}

	defer pbFile.Close()
	pkRaw, err := pk.MarshalBinary()
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to Marshal pk: %w", err)
	}
	if _, err = pbFile.Write(pkRaw); err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to Write: %w", err)
	}

	skFile, err := os.Create(filePath + ".sk")
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to Create sk: %w", err)
	}
	defer skFile.Close()
	skRaw, err := sk.MarshalBinary()
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to Marshal sk: %w", err)
	}
	if _, err = skFile.Write(skRaw); err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to Write: %w", err)
	}

	return AccumulatorKey{PK: pk, SK: sk, Level: level, Acc: acc, Path: filePath}, nil
}

func getAccumulatorByPath(filePath string) (AccumulatorKey, error) {
	var accKey = AccumulatorKey{Path: filePath}
	buf, err := os.ReadFile(filePath + ".txt")
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to ReadFile acc: %w", err)
	}

	acc := new(accumulator.Accumulator)
	if err = acc.UnmarshalBinary(buf); err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to UnmarshalBinary: %w", err)
	}

	accKey.Acc = acc

	buf, err = os.ReadFile(filePath + ".pk")
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to ReadFile pk: %w", err)
	}

	var pk accumulator.PublicKey
	if err = pk.UnmarshalBinary(buf); err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to UnmarshalBinary pk: %w", err)
	}

	accKey.PK = &pk

	buf, err = os.ReadFile(filePath + ".sk")
	if err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to ReadFile sk: %w", err)
	}

	var sk accumulator.SecretKey
	if err = sk.UnmarshalBinary(buf); err != nil {
		return AccumulatorKey{}, fmt.Errorf("failed to UnmarshalBinary sk: %w", err)
	}

	accKey.SK = &sk

	return accKey, nil
}

func hashStr(str string) []byte {
	return stribog.New256().Sum([]byte(base64.StdEncoding.EncodeToString([]byte(str))))[:32]
}

func hashBytes(str []byte) []byte {
	return stribog.New256().Sum([]byte(base64.StdEncoding.EncodeToString(str)))[:32]
}

//// admin only
//func setupFromFile() ([]AccumulatorKey, error) {
//	var keys = []AccumulatorKey{}
//
//	for i := 0; i < levelCount; i++ {
//		err := func() error {
//			curve := curves.BLS12381(curves.BLS12381G1().Point)
//			if f, err := os.OpenFile(strconv.Itoa(i)+".txt", os.O_RDWR|os.O_CREATE, 0777); err != nil {
//				return fmt.Errorf("failed to setupFromFile: %w", err)
//			} else {
//				defer func() {
//					if err := f.Close(); err != nil {
//						fmt.Println(err)
//						return
//					}
//
//				}()
//				acc, err := new(accumulator.Accumulator).New(curve)
//				if err != nil {
//					return fmt.Errorf("failed to New: %w", err)
//				}
//				buf, err := acc.MarshalBinary()
//				if err != nil {
//					return fmt.Errorf("failed to MarshalBinary: %w", err)
//				}
//
//				if _, err := f.Write(buf); err != nil {
//					return fmt.Errorf("failed to Write: %w", err)
//				}
//				sk, err := new(accumulator.SecretKey).New(curve, hashStr(""))
//				if err != nil {
//					return fmt.Errorf("failed to create secret key: %w", err)
//				}
//				pk, err := sk.GetPublicKey(curve)
//				if err != nil {
//					return fmt.Errorf("failed to GetPublicKey: %w", err)
//				}
//				keys = append(keys, AccumulatorKey{Level: i, SK: sk, PK: pk})
//			}
//			return nil
//		}()
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	return keys, nil
//}
//
//func addUser(level int, user_id string, sk *accumulator.SecretKey) ([]byte, error) {
//	acc, err := getAccumulator(level)
//	if err != nil {
//		return nil, fmt.Errorf("failed to getAccumulator: %w", err)
//	}
//
//	elem, err := curves.BLS12381(curves.BLS12381G1().Point).Scalar.SetBytes(hashStr(user_id))
//	if err != nil {
//		return nil, fmt.Errorf("failed to SetBytes: %w", err)
//	}
//
//	acc, err = acc.Add(sk, elem)
//	if err != nil {
//		return nil, fmt.Errorf("failed to Add: %w", err)
//	}
//
//	wit, err := new(accumulator.MembershipWitness).New(elem, acc, sk)
//	if err != nil {
//		return nil, fmt.Errorf("failed to New: %w", err)
//	}
//
//	body, err := acc.MarshalBinary()
//
//	if err = rewriteFile(level, body); err != nil {
//		return nil, fmt.Errorf("failed to rewriteFile: %w", err)
//	}
//
//	witBody, err := wit.MarshalBinary()
//	if err != nil {
//		return nil, fmt.Errorf("failed to MarshalBinary: %w", err)
//	}
//
//	return witBody, nil
//}
//
//func checkUser(level int, witByte []byte, keys AccumulatorKey) (bool, error) {
//	acc, err := getAccumulator(level)
//	if err != nil {
//		return false, fmt.Errorf("failed to getAccumulator: %w", err)
//	}
//
//	wit := new(accumulator.MembershipWitness)
//	if err := wit.UnmarshalBinary(witByte); err != nil {
//		return false, fmt.Errorf("failed to UnmarshalBinary: %w", err)
//	}
//
//	if err := wit.Verify(keys.PK, acc); err != nil {
//		return false, nil
//	}
//
//	return true, nil
//}
//
//func deleteUser(level int, user_id string, key *accumulator.SecretKey) error {
//	acc, err := getAccumulator(level)
//	if err != nil {
//		return fmt.Errorf("failed to getAccumulator: %w", err)
//	}
//
//	elem := curves.BLS12381(&curves.PointBls12381G1{}).Scalar.Hash(hashStr(user_id))
//	if err != nil {
//		return fmt.Errorf("failed to SetBytes: %w", err)
//	}
//
//	acc, err = acc.Remove(key, elem)
//	if err != nil {
//		return fmt.Errorf("failed to Remove: %w", err)
//	}
//
//	body, err := acc.MarshalBinary()
//
//	if err = rewriteFile(level, body); err != nil {
//		return fmt.Errorf("failed to rewriteFile: %w", err)
//	}
//
//	return nil
//}
//
//func getAccumulator(level int) (*accumulator.Accumulator, error) {
//	if level >= levelCount {
//		return nil, fmt.Errorf("level is bigger that supported. Max is " + strconv.Itoa(levelCount))
//	}
//
//	buf, err := os.ReadFile(strconv.Itoa(level) + ".txt")
//	if err != nil {
//		return nil, fmt.Errorf("failed to ReadFile: %w", err)
//	}
//
//	acc := new(accumulator.Accumulator)
//	if err1 := acc.UnmarshalBinary(buf); err1 != nil {
//		return nil, fmt.Errorf("failed to UnmarshalBinary: %w", err1)
//	}
//
//	return acc, nil
//}
//
//func rewriteFile(level int, body []byte) error {
//	if err := os.Truncate(strconv.Itoa(level)+".txt", 0); err != nil {
//		return fmt.Errorf("failed to Truncate: %w", err)
//	}
//
//	f, err := os.Open(strconv.Itoa(level) + ".txt")
//	if err != nil {
//		return fmt.Errorf("failed to Open: %w", err)
//	}
//	defer f.Close()
//
//	if err = os.WriteFile(strconv.Itoa(level)+".txt", body, os.ModeAppend); err != nil {
//		return fmt.Errorf("failed to Write: %w", err)
//	}
//
//	return nil
//}
//
