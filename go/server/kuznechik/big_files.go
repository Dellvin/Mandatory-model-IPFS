package kuznechik

import (
	"fmt"
)

const batchSize = 16

func Encode(key [32]uint8, body string) ([]byte, error) {
	batchCount := len(body) / batchSize
	fmt.Println("KEY: ", key)
	her, err := NewCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to NewCipher: %w", err)
	}

	ecrypted := make([]byte, 0, len(body))

	for j, i := 0, 1; i <= batchCount; i++ {
		srcBatch := body[j*batchSize : i*batchSize]
		dstBatch := make([]byte, batchSize)
		her.Encrypt(dstBatch, []byte(srcBatch))
		ecrypted = append(ecrypted, dstBatch...)
		j = i
	}

	if len(body) != batchCount*batchSize {
		//add ostatok
		zeroCount := batchSize - (len(body) - (batchCount * batchSize))
		srcBatch := make([]byte, 0, batchSize)
		dstBatch := make([]byte, batchSize)
		for i := 0; i < zeroCount; i++ {
			srcBatch = append(srcBatch, 0)
		}
		srcBatch = append(srcBatch, body[batchCount*batchSize:]...)
		her.Encrypt(dstBatch, srcBatch)
		ecrypted = append(ecrypted, dstBatch...)
	}

	return ecrypted, nil
}

func Decode(key [32]uint8, body []byte) ([]byte, error) {
	batchCount := len(body) / batchSize
	her, err := NewCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to NewCipher: %w", err)
	}

	decrypted := make([]byte, 0, len(body))

	for j, i := 0, 1; i <= batchCount; i++ {
		srcBatch := body[j*batchSize : i*batchSize]
		dstBatch := make([]byte, batchSize)
		her.Decrypt(dstBatch, srcBatch)
		if i == batchCount {
			tmpBuf := make([]byte, 0, batchSize)
			var skipNull bool
			for k := 0; k < 16; k++ {
				if dstBatch[k] == 0 && !skipNull {
					continue
				} else if !skipNull {
					skipNull = true
				}
				tmpBuf = append(tmpBuf, dstBatch[k])
			}
			dstBatch = tmpBuf
		}
		decrypted = append(decrypted, dstBatch...)
		j = i
	}

	return decrypted, nil
}
