package main

import (
	"encoding/base32"
	"errors"
	"flag"
	"fmt"
	senc "github.com/jbenet/go-simple-encrypt"
	"ipfs-senc/kuznechik"
	"ipfs-senc/stribog"
	"os"
	"strings"

	ipfssenc "github.com/jbenet/ipfs-senc"
	mb "github.com/multiformats/go-multibase"
)

// flags
var (
	Key       string
	API       string
	RandomKey bool
	DirWrap   bool
)

// errors
var (
	ErrNoIPFS = errors.New("ipfs node error: not online")
)

const (
	gwayGlobal = "https://gateway.ipfs.io"
	gwayLocal  = "http://localhost:8080"
)

var Usage = `ENCRYPT AND SEND
    # will ask for a key
    go share <local-source-path>

    # encrypt with a known key. (256 bits please)
    go --key <secret-key> share <local-source-path>

    # encrypt with a randomly generated key. will be printed out.
    go --random-key share <local-source-path>


GET AND DECRYPT
    # will ask for key
    go download <ipfs-link> <local-destination-path>

    # decrypt with given key.
    go --key <secret-key> download <ipfs-link> <local-destination-path>

OPTIONS
    --h, --help              show usage
    --key <secret-key>       a 256bit secret key, encoded with multibase (no key = random key)
    --api <ipfs-api-url>     an ipfs node api to use (overrides defaults)
    -w, --wrap               if adding a directory, wrap it first to preserve dir

EXAMPLES
    > go share my_secret_dir
    Enter a 256 bit AES key in multibase:
`

func init() {
	flag.BoolVar(&RandomKey, "random-key", false, "use a randomly generated key (deprecated opt)")
	flag.StringVar(&Key, "key", "", "an AES encryption key in hex")
	flag.StringVar(&API, "api", "", "override IPFS node API")
	flag.BoolVar(&DirWrap, "w", false, "if adding a directory, wrap it first to preserve dir")
	flag.BoolVar(&DirWrap, "wrap", false, "if adding a directory, wrap it first to preserve dir")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage)
	}
}

func decodeKey(k string) ([]byte, error) {
	_, b, err := mb.Decode(k)
	if err != nil {
		return nil, fmt.Errorf("multibase decoding error: %v", err)
	}
	if len(b) != 32 {
		return nil, fmt.Errorf("key must be exactly 256 bits. Was: %d", len(b))
	}
	return b, nil
}

func getSencKey(randomIfNone bool) (ipfssenc.Key, error) {
	NilKey := ipfssenc.Key(nil)

	var k []byte
	var err error
	if Key != "" {
		k, err = base32.StdEncoding.DecodeString(Key)
	} else if randomIfNone { // random key
		randBuf, err := senc.RandomKey()
		if err != nil {
			return nil, fmt.Errorf("failed to RandomKey: %w", err)
		}
		sb := stribog.New256()
		_, err = sb.Write(randBuf)
		k = sb.Sum([]byte{})
	} else {
		err = errors.New("Please enter a key with --key")
	}
	if err != nil {
		return NilKey, err
	}

	return ipfssenc.Key(k), nil
}

func cmdDownload(args []string) error {
	if RandomKey {
		return errors.New("cannot use --random-key with download")
	}
	if len(args) < 2 {
		return errors.New("not enough arguments. download requires 2. see -h")
	}

	srcLink := ipfssenc.IPFSLink(args[0])
	if len(srcLink) < 1 {
		return errors.New("invalid ipfs-link")
	}

	dstPath := args[1]
	if dstPath == "" {
		return errors.New("requires a destination path")
	}

	// check for Key, get key.
	key, err := getSencKey(false)
	if err != nil {
		return err
	}

	// fmt.Println("Initializing ipfs node...")
	n := ipfssenc.GetROIPFSNode(API)
	if !n.IsUp() {
		return ErrNoIPFS
	}

	// fmt.Println("Getting", srcLink, "...")
	err = ipfssenc.GetDecryptAndUnbundle(n, srcLink, dstPath, key)
	rCloser, err := ipfssenc.Get(n, srcLink)
	if err != nil {
		return err
	}

	if err := kuznechik.FileDecode(rCloser, key, dstPath); err != nil {
		return err
	}

	fmt.Println("SUCCESS!!!!!!!!!")
	fmt.Println("Unbundled to:", dstPath)
	return nil
}

func cmdShare(args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments. share requires 1. see -h")
	}
	srcPath := args[0]
	if srcPath == "" {
		return errors.New("requires a source path")
	}

	// check for Key, get key.
	key, err := getSencKey(true)
	if err != nil {
		return err
	}
	//API = "ipfs.io"
	// fmt.Println("Initializing ipfs node...")
	n, err := ipfssenc.GetRWIPFSNode(API)
	if err != nil {
		return err
	}
	if !n.IsUp() {
		return ErrNoIPFS
	}

	// fmt.Println("Sharing", srcPath, "...")

	fmt.Println(key)

	reader, err := kuznechik.FileEncode(key, srcPath)
	if err != nil {
		return fmt.Errorf("failed to cmdShare: %w", err)
	}
	link, err := ipfssenc.Put(n, reader)
	//link, err := ipfssenc.BundleEncryptAndPut(n, srcPath, key, DirWrap)
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
	fmt.Println("Ciphertext on local gateway: ", gwayGlobal, l)
	fmt.Println("Ciphertext on global gateway: ", gwayLocal, l)
	fmt.Println("")
	fmt.Println("Get, Decrypt, and Unbundle with:")
	fmt.Println("    go --key", keyStr, "download", l, "dstPath")
	fmt.Println("")
	fmt.Printf("View on the web: https://ipfs.io/ipns/ipfs-senc.net/#%s:%s\n", key, l)
	return nil
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
		return cmdDownload(args[1:])
	case "share":
		return cmdShare(args[1:])
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
