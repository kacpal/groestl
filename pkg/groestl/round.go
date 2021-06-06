package groestl

import (
	"encoding/binary"
	"fmt"
)

func buildColumns(data []byte, cols chan uint64) {
	for i, l := 8, len(data); i <= l; i += 8 {
		cols <- binary.BigEndian.Uint64(data[i-8:i])
	}
	close(cols)
}

func (d *digest) transform(data []byte) error {
	if VERBOSE {
		fmt.Println("Call to transform:", data)
	}

	if (len(data) % d.BlockSize()) != 0 {
		return fmt.Errorf("data len in transform is not a multiple of BlockSize")
	}

	cols := make(chan uint64)
	go buildColumns(data, cols)

	eb := d.blocks + uint64(len(data) / d.BlockSize())
	for d.blocks < eb {
		m := make([]uint64, d.columns)
		hxm := make([]uint64, d.columns)

		for i := 0; i < d.columns; i++ {
			m[i] = <- cols
			hxm[i] = d.chaining[i] ^ m[i]
		}

		if VERBOSE {
			fmt.Printf("Block: %d\n", d.blocks)
			fmt.Println("M:  ", m)
			printUintSlice(m)
			fmt.Println("HxM:", hxm)
			printUintSlice(hxm)
		}

		round(d, hxm, 'P')
		round(d, m, 'Q')

		if VERBOSE {
			fmt.Println("after round transformations...")
			fmt.Println("M:  ", m)
			printUintSlice(m)
			fmt.Println("HxM:", hxm)
			printUintSlice(hxm)
		}

		for i := 0; i < d.columns; i++ {
			d.chaining[i] ^= hxm[i] ^ m[i]
		}

		d.blocks += 1
		
		if VERBOSE {
			fmt.Println(d)
		}
	}

	return nil
}

func round(d *digest, x []uint64, variant rune) {
	if d.BlockSize() == 64 {
		// for smaller blocksize change variant to lowercase letter
		variant += 0x20
	}

	for i := 0; i < d.rounds; i++ {
		x = addRoundConstant(x, i, variant)
		x = subBytes(x)
		x = shiftBytes(x, variant)
		x = mixBytes(x)
	}
}

func addRoundConstant(x []uint64, r int, variant rune) []uint64 {
	switch variant {
	case 'P', 'p':
		for i, l := 0, len(x); i < l; i++ {
			// byte from row 0: ((col >> (8*7)) & 0xFF)
			// we want to xor the byte below with row 0
			// therefore we have to shift it by 8*7 bits
			x[i] ^= uint64((i<<4)^r) << (8*7)
		}
	case 'Q', 'q':
		for i, l := 0, len(x); i < l; i++ {
			x[i] ^= ^uint64(0) ^ uint64((i<<4)^r)
		}
	}
	return x
}

func subBytes(x []uint64) []uint64 {
	// TODO
	return nil
}

func shiftBytes(x []uint64, variant rune) []uint64 {
	// TODO
	return nil
}

func mixBytes(x []uint64) []uint64 {
	// TODO
	return nil
}

