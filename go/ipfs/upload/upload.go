package upload

import (
	"bytes"
	"encoding/base32"
	"fmt"
	ipfssenc "github.com/jbenet/ipfs-senc"
	"ipfs-senc/abe"
	"ipfs-senc/ipfs/pkg"
	"ipfs-senc/kuznechik"
	"strings"
)

func Upload(keyRaw, api, srcPath string, crypto bool, department, securityLevel int) error {
	// check for Key, get key.
	key, err := pkg.GetSencKey(keyRaw, true)
	if err != nil {
		return err
	}

	//API = "ipfs.io"
	fmt.Println("Initializing ipfs node...")
	n, err := ipfssenc.GetRWIPFSNode(api)
	if err != nil {
		return err
	}
	if !n.IsUp() {
		return pkg.ErrNoIPFS
	}

	fmt.Println("Sharing", srcPath, "...")

	fmt.Println(key)

	buf, err := kuznechik.Read(key, srcPath, crypto)
	if err != nil {
		return fmt.Errorf("failed to Upload: %w", err)
	}

	SecretFile, err := abe.EncryptFile(department, securityLevel, buf)
	if err != nil {
		return fmt.Errorf("failed to EncryptFile: %w", err)
	}

	link, err := ipfssenc.Put(n, bytes.NewReader(SecretFile.Cipher))
	if err != nil {
		return fmt.Errorf("failed to Put: %w", err)
	}

	l := string(link)
	if !strings.HasPrefix(l, "/ipfs/") {
		l = "/ipfs/" + l
	}
	keyStr := base32.StdEncoding.EncodeToString(key)

	if err != nil {
		return err
	}
	fmt.Println("Shared as: ", l)
	fmt.Println("Key: '" + string(keyStr) + "'")
	fmt.Println("Ciphertext on global gateway: ", pkg.GWayGlobal, l)
	fmt.Println("Ciphertext on local gateway: ", pkg.GWayLocal, l)
	fmt.Println("")
	fmt.Println("Get, Decrypt, and Unbundle with:")
	fmt.Println("    go --key", keyStr, "download", l, "dstPath")
	fmt.Println("")
	fmt.Printf("View on the web: https://ipfs.io/ipns/ipfs-senc.net/#%s:%s\n", key, l)
	fmt.Printf("Your department: '%d'\n", department)
	fmt.Printf("Your securityLevel: '%d'\n", securityLevel)
	fmt.Println("Your crypto:", crypto)
	//fmt.Printf("Your public key: '%s'\n", string(SecretFile.PubKey))
	//fmt.Printf("Your private key: '%s'\n", string(SecretFile.SecKey))
	return nil
}
