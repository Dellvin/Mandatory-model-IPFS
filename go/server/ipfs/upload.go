package ipfs

import (
	"bytes"
	"fmt"
	ipfssenc "github.com/jbenet/ipfs-senc"

	"server/pkg"
	"strings"
)

func Upload(api string, file []byte) (string, error) {
	//API = "ipfs.io"
	fmt.Println("Initializing ipfs node...")
	n, err := ipfssenc.GetRWIPFSNode(api)
	if err != nil {
		return "", err
	}
	if !n.IsUp() {
		return "", pkg.ErrNoIPFS
	}

	link, err := ipfssenc.Put(n, bytes.NewReader(file))
	if err != nil {
		return "", fmt.Errorf("failed to Put: %w", err)
	}

	l := string(link)
	if !strings.HasPrefix(l, "/ipfs/") {
		l = "/ipfs/" + l
	}
	return string(link), nil
}
