package kuznechik

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const batchSize = 16

func FileEncode(key []byte, filePath string) (io.Reader, error) {
	if buf, err := readFile(filePath); err == nil {
		if encoded, err := bigTextEncode(key, buf); err == nil {
			return bytes.NewReader(encoded), nil
		} else {
			return nil, fmt.Errorf("failed to bigTextEncode: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed to FileEncode: %w", err)
	}
}

func readFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path) // just pass the file name
	if err != nil {
		return nil, fmt.Errorf("failed to readFile: %w", err)
	}

	return b, nil
}

func FileDecode(closer io.ReadCloser, key []byte, path string) error {
	var raw bytes.Buffer
	tee := io.TeeReader(closer, &raw)

	newBuf, err := ioutil.ReadAll(tee)
	if err != nil {
		return fmt.Errorf("failed to ReadAll: %w", err)
	}

	decoded, err := bigTextDecode(key, newBuf)
	if err != nil {
		return fmt.Errorf("failed to bigTextDecode: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to Create: %w", err)
	}
	if _, err := f.Write(decoded); err != nil {
		return fmt.Errorf("failed to Write: %w", err)
	}

	return nil
}

func bigTextEncode(key, body []byte) ([]byte, error) {
	batchCount := len(body) / batchSize
	fmt.Println("KEY: ", key)
	her, err := NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to NewCipher: %w", err)
	}

	ecrypted := make([]byte, 0, len(body))

	for j, i := 0, 1; i <= batchCount; i++ {
		srcBatch := body[j*batchSize : i*batchSize]
		dstBatch := make([]byte, batchSize)
		her.Encrypt(dstBatch, srcBatch)
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

func bigTextDecode(key, body []byte) ([]byte, error) {
	batchCount := len(body) / batchSize
	her, err := NewCipher(key)
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
