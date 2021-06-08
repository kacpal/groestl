package groestl

import (
	"encoding/binary"
	"fmt"
)

func buildColumns(data []byte, cols chan uint64) {
	for i, l := 8, len(data); i <= l; i += 8 {
		cols <- binary.BigEndian.Uint64(data[i-8 : i])
	}
	close(cols)
}

// Performs compression function. Returns nil on success, error otherwise.
func (d *digest) transform(data []byte) error {
	if (len(data) % d.BlockSize()) != 0 {
		return fmt.Errorf("data len in transform is not a multiple of BlockSize")
	}

	cols := make(chan uint64)
	go buildColumns(data, cols)

	eb := d.blocks + uint64(len(data)/d.BlockSize())
	for d.blocks < eb {
		m := make([]uint64, d.columns)
		hxm := make([]uint64, d.columns)

		for i := 0; i < d.columns; i++ {
			m[i] = <-cols
			hxm[i] = d.chaining[i] ^ m[i]
		}

		if VERBOSE {
			fmt.Println("\n========================================\n")
			fmt.Println("Block Contents:")
			printUintSlice(m)
			fmt.Println()
		}

		hxm = round(d, hxm, 'P')
		m = round(d, m, 'Q')

		for i := 0; i < d.columns; i++ {
			d.chaining[i] ^= hxm[i] ^ m[i]
		}

		d.blocks += 1

		if VERBOSE {
			fmt.Println("P(h+m) + Q(m) + h =")
			printUintSlice(d.chaining[:d.columns])
			fmt.Println()
		}
	}

	return nil
}

// Performs last compression. After this function, data
// is ready for truncation.
func (d *digest) finalTransform() {
	h := make([]uint64, d.columns)

	for i := 0; i < d.columns; i++ {
		h[i] = d.chaining[i]
	}

	if VERBOSE {
		fmt.Println("\n========================================\n")
		fmt.Println("Output transformation:\n")
	}

	h = round(d, h, 'P')

	for i := 0; i < d.columns; i++ {
		d.chaining[i] ^= h[i]
	}

	d.blocks += 1

	if VERBOSE {
		fmt.Println("P(h) + h =")
		printUintSlice(d.chaining[:d.columns])
		fmt.Println("\n---------------------------------------\n")
	}
}

// Performs whole set of rounds on data provided in x. Variant denotes type
// of permutation being performed. P and Q are for groestl-512
// and lowercase are for groestl-256
func round(d *digest, x []uint64, variant rune) []uint64 {
	if VERBOSE {
		fmt.Println(":: BEGIN " + string(variant))
		defer fmt.Println(":: END " + string(variant) + "\n")
		fmt.Println("Input:")
		printUintSlice(x)
	}

	if d.BlockSize() == 64 {
		// for smaller blocksize change variant to lowercase letter
		variant += 0x20
	}

	for i := 0; i < d.rounds; i++ {
		x = addRoundConstant(x, i, variant)
		if VERBOSE {
			fmt.Printf("t=%d (AddRoundConstant):\n", i)
			printUintSlice(x)
		}
		x = subBytes(x)
		if VERBOSE {
			fmt.Printf("t=%d (SubBytes):\n", i)
			printUintSlice(x)
		}
		x = shiftBytes(x, variant)
		if VERBOSE {
			fmt.Printf("t=%d (ShiftBytes):\n", i)
			printUintSlice(x)
		}
		x = mixBytes(x)
		if VERBOSE {
			fmt.Printf("t=%d (MixBytes):\n", i)
			printUintSlice(x)
		}
	}

	return x
}

// AddRoundConstant transformation for data provided in x. Variant denotes type
// of permutation being performed. P and Q are for groestl-512
// and lowercase are for groestl-256
func addRoundConstant(x []uint64, r int, variant rune) []uint64 {
	switch variant {
	case 'P', 'p':
		for i, l := 0, len(x); i < l; i++ {
			// byte from row 0: ((col >> (8*7)) & 0xFF)
			// we want to xor the byte below with row 0
			// therefore we have to shift it by 8*7 bits
			x[i] ^= uint64((i<<4)^r) << (8 * 7)
		}
	case 'Q', 'q':
		for i, l := 0, len(x); i < l; i++ {
			x[i] ^= ^uint64(0) ^ uint64((i<<4)^r)
		}
	}
	return x
}

// SubBytes transformation for data provided in x.
func subBytes(x []uint64) []uint64 {
	var newCol [8]byte
	for i, l := 0, len(x); i < l; i++ {
		for j := 0; j < 8; j++ {
			newCol[j] = sbox[pickRow(x[i], j)]
		}
		x[i] = binary.BigEndian.Uint64(newCol[:])
	}
	return x
}

// ShiftBytes transformation for data provided in x. Variant denotes type
// of permutation being performed. P and Q are for groestl-512
// and lowercase are for groestl-256
func shiftBytes(x []uint64, variant rune) []uint64 {
	var shiftVector [8]int
	switch variant {
	case 'p':
		shiftVector = [8]int{0, 1, 2, 3, 4, 5, 6, 7}
	case 'P':
		shiftVector = [8]int{0, 1, 2, 3, 4, 5, 6, 11}
	case 'q':
		shiftVector = [8]int{1, 3, 5, 7, 0, 2, 4, 6}
	case 'Q':
		shiftVector = [8]int{1, 3, 5, 11, 0, 2, 4, 6}
	}
	l := len(x)
	ret := make([]uint64, l)
	for i := 0; i < l; i++ {
		ret[i] = uint64(pickRow(x[(i+shiftVector[0])%l], 0))
		for j := 1; j < 8; j++ {
			ret[i] <<= 8
			ret[i] ^= uint64(pickRow(x[(i+shiftVector[j])%l], j))
		}
	}
	return ret
}

// MixBytes transformation for data provided in x.
func mixBytes(x []uint64) []uint64 {
	// this part is tricky
	// so here comes yet another rough translation straight from reference implementation

	mul2 := func(b uint8) uint8 { return uint8((b << 1) ^ (0x1B * ((b >> 7) & 1))) }
	mul3 := func(b uint8) uint8 { return (mul2(b) ^ (b)) }
	mul4 := func(b uint8) uint8 { return mul2(mul2(b)) }
	mul5 := func(b uint8) uint8 { return (mul4(b) ^ (b)) }
	mul7 := func(b uint8) uint8 { return (mul4(b) ^ mul2(b) ^ (b)) }

	var temp [8]uint8
	for i, l := 0, len(x); i < l; i++ {
		for j := 0; j < 8; j++ {
			temp[j] =
				mul2(pickRow(x[i], (j+0)%8)) ^
					mul2(pickRow(x[i], (j+1)%8)) ^
					mul3(pickRow(x[i], (j+2)%8)) ^
					mul4(pickRow(x[i], (j+3)%8)) ^
					mul5(pickRow(x[i], (j+4)%8)) ^
					mul3(pickRow(x[i], (j+5)%8)) ^
					mul5(pickRow(x[i], (j+6)%8)) ^
					mul7(pickRow(x[i], (j+7)%8))
		}
		x[i] = binary.BigEndian.Uint64(temp[:])
	}
	return x
}
