package groestl

import (
	"encoding/hex"
	"fmt"
)

func PrintHash(hash []byte) {
	fmt.Println(hex.EncodeToString(hash))
}

func pickRow(col uint64, i int) byte {
	return byte((col >> (8 * (7 - i))) & 0xFF)
}

func printUintSlice(x []uint64) {
	l := len(x)
	for i := 0; i < 8; i++ {
		for j := 0; j < l; j++ {
			fmt.Printf("%02x ", pickRow(x[j], i))
		}
		fmt.Println()
	}
}
