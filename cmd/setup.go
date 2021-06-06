package cmd

import (
	"flag"
	"fmt"
	"groestl/pkg/groestl"
	"io/ioutil"
	"os"
)

func Execute() {
	var sum []byte

	hashlen := flag.Int("hash", 256, "output hash length")
	flag.Parse()

	data, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	switch *hashlen {
	case 224:
		sum = groestl.Sum224(data)
	case 256:
		sum = groestl.Sum256(data)
	case 384:
		sum = groestl.Sum384(data)
	case 512:
		sum = groestl.Sum512(data)
	default:
		fmt.Println("Invalid hash length")
		os.Exit(1)
	}

	fmt.Println(sum)
}
