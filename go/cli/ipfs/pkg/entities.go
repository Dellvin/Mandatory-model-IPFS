package pkg

import "errors"

const (
	GWayGlobal = "https://gateway.ipfs.io"
	GWayLocal  = "http://localhost:8080"
)

var ErrNoIPFS = errors.New("ipfs node error: not online")
