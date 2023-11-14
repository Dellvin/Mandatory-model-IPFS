package download

import (
	"errors"
	"fmt"
	ipfssenc "github.com/jbenet/ipfs-senc"
	"ipfs-senc/ipfs/pkg"
	"ipfs-senc/kuznechik"
)

func Download(link, dstPath, keyRaw, api string, crypto bool) error {
	srcLink := ipfssenc.IPFSLink(link)
	if len(srcLink) < 1 {
		return errors.New("invalid ipfs-link")
	}

	if dstPath == "" {
		return errors.New("requires a destination path")
	}

	// check for Key, get key.
	key, err := pkg.GetSencKey(keyRaw, false)
	if err != nil {
		return err
	}

	fmt.Println("Initializing ipfs node...")
	n := ipfssenc.GetROIPFSNode(api)
	if !n.IsUp() {
		return pkg.ErrNoIPFS
	}

	fmt.Println("Getting", srcLink, "...")
	err = ipfssenc.GetDecryptAndUnbundle(n, srcLink, dstPath, key)
	rCloser, err := ipfssenc.Get(n, srcLink)
	if err != nil {
		return err
	}

	if err := kuznechik.Write(rCloser, key, dstPath, crypto); err != nil {
		return err
	}

	fmt.Println("Unbundled to:", dstPath)
	return nil
}
