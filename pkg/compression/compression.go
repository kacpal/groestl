package compression

import (	
	"groestl/pkg/permutation"
)

func Compress(hash, msg []byte) []byte {
	a := permutation.P(Xor(hash, msg))
	b := permutation.Q(msg)
	return Xor(Xor(a, b), hash)
}

func Xor(a, b []byte) []byte {
	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] ^ b[i]
	}
	return c
}