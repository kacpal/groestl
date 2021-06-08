package cmd

import (
	"flag"
	"fmt"
	"groestl/pkg/groestl"
	"io/ioutil"
	"os"
)

const Usage = `Usage:
%s <args> <path>
`

func Execute() {
	var sum []byte

	hashlen := flag.Int("hash", 256, "output hash length")
	flag.Parse()

	if flag.Arg(0) == "" {
		fmt.Printf(Usage, os.Args[0])
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	switch *hashlen {
	case 256:
		sum = groestl.Sum256(data)
	case 512:
		sum = groestl.Sum512(data)
	default:
		fmt.Println("Invalid hash length")
		os.Exit(1)
	}

	groestl.PrintHash(sum)
}
