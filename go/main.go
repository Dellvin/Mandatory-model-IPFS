package main

import (
	"errors"
	"flag"
	"fmt"
	"ipfs-senc/ipfs/download"
	"ipfs-senc/ipfs/upload"
	"os"
)

// flags
var (
	Use        = flag.String("use", "", "share or download")
	Link       = flag.String("link", "", "link to file in IPFS")
	Path       = flag.String("path", "", "path to file")
	Key        = flag.String("key", "", "an AES encryption key in hex")
	API        = flag.String("api", "", "override IPFS node API")
	Encrypt    = flag.Bool("crypto", false, "if true, than it will encrypt your file on upload and decrypt on download")
	SecureType = flag.Int("secure_type", 0, "security level")
	Department = flag.Int("department", 0, "your department")
)

var Usage = `ENCRYPT AND SEND
    # encrypt with a randomly generated key. will be printed out.
    go share <local-source-path>


GET AND DECRYPT
    # will ask for key
    go download <ipfs-link> <local-destination-path>

    # decrypt with given key.
    go --key <secret-key> download <ipfs-link> <local-destination-path>

OPTIONS
	--use					 share or download
    --link					 link to file in IPFS
	--path         			 path to file
    --h, --help              show usage
    --key <secret-key>       a 256bit secret key, encoded with multibase, ONLY in download required
    --api <ipfs-api-url>     an ipfs node api to use (overrides defaults)
	--crypto                 if true, than it will encrypt your file on upload and decrypt on download
	--secure_type            allowed security levels:
							 	1) 0 - spacial importance
								2) 1 - absolutely secretly
								3) 2 - secretly
	--department			number of your department
`

func errMain() error {

	switch *Use {
	case "download":
		return download.Download(*Link, *Path, *Key, *API, *Encrypt, *SecureType)
	case "share":
		return upload.Upload(*Key, *API, *Path, *Encrypt, *Department, *SecureType)
	default:
		return errors.New("Unknown command: " + *Use)
	}
}

// ipfs daemon
// go --key D44DHB54VE62PMID4JLG6WYZWTPKUJFO3Q2NJOOTKMUGKLX5B57A==== download /ipfs/Qme4rKqR3iDUa9iEx9iyYRTFhY4X1skXQFGSJdTGFQw9Zx
func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage)
	}

	flag.Parse()
	if err := errMain(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(-1)
	}
}
