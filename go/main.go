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
	Key        string
	API        string
	Encrypt    bool
	SecureType int
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
    --h, --help              show usage
    --key <secret-key>       a 256bit secret key, encoded with multibase, ONLY in download required
    --api <ipfs-api-url>     an ipfs node api to use (overrides defaults)
	--crypto                 if true, than it will encrypt your file on upload and decrypt on download
	--secure_type            allowed security levels:
							 	1) 0 - public, available for everybody
								2) 1...n - secret, available for certain group
`

func init() {
	flag.StringVar(&Key, "key", "", "an AES encryption key in hex")
	flag.StringVar(&API, "api", "", "override IPFS node API")
	flag.BoolVar(&Encrypt, "crypto", false, "if true, than it will encrypt your file on upload and decrypt on download")
	flag.IntVar(&SecureType, "secure_type", 0, "security level")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage)
	}
}

func errMain(args []string) error {
	// no command is not an error. it's usage.
	if len(args) == 0 {
		fmt.Println(Usage)
		return nil
	}

	cmd := args[0]
	switch cmd {
	case "download":
		if len(args) < 2 {
			return errors.New("not enough arguments. download requires 2. see -h")
		}
		return download.Download(args[1], args[2], Key, API, Encrypt, SecureType)
	case "share":
		if len(args) < 1 {
			return errors.New("not enough arguments. share requires 1. see -h")
		}
		srcPath := args[0]
		if srcPath == "" {
			return errors.New("requires a source path")
		}
		return upload.Upload(Key, API, srcPath, Encrypt, SecureType)
	default:
		return errors.New("Unknown command: " + cmd)
	}
}

// ipfs daemon
// go go --key D44DHB54VE62PMID4JLG6WYZWTPKUJFO3Q2NJOOTKMUGKLX5B57A==== download /ipfs/Qme4rKqR3iDUa9iEx9iyYRTFhY4X1skXQFGSJdTGFQw9Zx
func main() {
	flag.Parse()
	args := flag.Args()
	if err := errMain(args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(-1)
	}
}
