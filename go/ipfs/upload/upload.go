package upload

import (
	"encoding/base32"
	"fmt"
	ipfssenc "github.com/jbenet/ipfs-senc"
	"ipfs-senc/ipfs/pkg"
	"ipfs-senc/kuznechik"
	"strings"
)

func Upload(keyRaw, api, srcPath string, crypto bool) error {
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

	reader, err := kuznechik.Read(key, srcPath, crypto)
	if err != nil {
		return fmt.Errorf("failed to Upload: %w", err)
	}
	link, err := ipfssenc.Put(n, reader)

	if err != nil {
		return err
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
	return nil
}
