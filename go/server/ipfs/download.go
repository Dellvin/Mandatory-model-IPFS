package ipfs

import (
	"errors"
	"fmt"
	ipfssenc "github.com/jbenet/ipfs-senc"
	"io"
	"server/pkg"
)

func Download(link, api string) ([]byte, error) {
	srcLink := ipfssenc.IPFSLink(link)
	if len(srcLink) < 1 {
		return nil, errors.New("invalid ipfs-link")
	}

	fmt.Println("Initializing ipfs node...")
	n := ipfssenc.GetROIPFSNode(api)
	if !n.IsUp() {
		return nil, fmt.Errorf("failed to IsUp: %w", pkg.ErrNoIPFS)
	}

	fmt.Println("Getting", srcLink, "...")

	rCloser, err := ipfssenc.Get(n, srcLink)
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	defer rCloser.Close()

	b, err := io.ReadAll(rCloser)
	if err != nil {
		return nil, fmt.Errorf("failed to ReadAll: %w", err)
	}
	return b, nil
}
