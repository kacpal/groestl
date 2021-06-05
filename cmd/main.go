package main

import (
	"fmt"
	"groestl/pkg/compression"
	"groestl/pkg/transformation"
	"io/ioutil"
)

func main() {
	dat, err := ioutil.ReadFile("data")
	if err != nil {
		fmt.Println(err)
	}

	hash := make([]byte, len(dat))

	for i := 512; i < len(dat); i += 512 {
		copy(hash[i-512:i], compression.Compress(hash[i-512:i], dat[i:i+512]))
	}

	transformation.Transform(hash)
}
