// Package groestl provides core groestl functionality. It's based on groestl's
// implementation guide with references in C code.
package groestl

import (
	"encoding/binary"
	"hash"
)

// Toggle verbose output with detailed description of every algorithm's step
const VERBOSE = false

// Struct digest is being used during algorithm execution. Provides easy
// access to all information about current state of data processing.
type digest struct {
	hashbitlen int
	chaining   [16]uint64
	blocks     uint64
	buf        [128]byte
	nbuf       int
	columns    int
	rounds     int
}

// Equivalent to Init from reference implementation. Initiates values
// for digest struct, therefore determines exact type of groestl algorithm.
func (d *digest) Reset() {
	for i, _ := range d.chaining {
		d.chaining[i] = 0
	}

	d.blocks = 0
	d.nbuf = 0

	if d.hashbitlen <= 256 {
		d.columns = 8
		d.rounds = 10
	} else {
		d.columns = 16
		d.rounds = 14
	}

	d.chaining[d.columns-1] = uint64(d.hashbitlen)
}

// Each New...() function creates new hash digest and initiates it
// for according hash size.
func New224() hash.Hash {
	d := new(digest)
	d.hashbitlen = 224
	d.Reset()
	return d
}

func New256() hash.Hash {
	d := new(digest)
	d.hashbitlen = 256
	d.Reset()
	return d
}

func New384() hash.Hash {
	d := new(digest)
	d.hashbitlen = 384
	d.Reset()
	return d
}

func New512() hash.Hash {
	d := new(digest)
	d.hashbitlen = 512
	d.Reset()
	return d
}

// Default function for creating hash digest for 256bit groestl.
func New() hash.Hash {
	return New256()
}

// Return size of digest
func (d *digest) Size() int {
	return d.hashbitlen
}

// Return block size for digest. For hash bigger than 256 bit block
// size is 128, otherwise it's 64.
func (d *digest) BlockSize() int {
	if d.hashbitlen <= 256 {
		return 64
	} else {
		return 128
	}
}

// Equivalent to Update form reference implementation. Performs processing
// on all data except the last block that might need padding.
func (d *digest) Write(p []byte) (n int, err error) {
	n = len(p)
	if d.nbuf > 0 {
		nn := copy(d.buf[d.nbuf:], p)
		d.nbuf += nn
		if d.nbuf == d.BlockSize() {
			err = d.transform(d.buf[:d.BlockSize()])
			if err != nil {
				panic(err)
			}
			d.nbuf = 0
		}
		p = p[nn:]
	}
	if len(p) >= d.BlockSize() {
		nn := len(p) &^ (d.BlockSize() - 1)
		err = d.transform(p[:nn])
		if err != nil {
			panic(err)
		}
		p = p[nn:]
	}
	if len(p) > 0 {
		d.nbuf = copy(d.buf[:], p)
	}
	return
}

func (d *digest) Sum(in []byte) []byte {
	d0 := *d
	hash := d0.checkSum()
	return append(in, hash...)
}

// Equivalent to Final from reference implementation. Creates padding
// for last block of data and performs final output transformation and trumcate.
// Returns hash value.
func (d *digest) checkSum() []byte {
	bs := d.BlockSize()
	var tmp [128]byte
	tmp[0] = 0x80

	if d.nbuf > (bs - 8) {
		d.Write(tmp[:(bs - d.nbuf)])
		d.Write(tmp[8:bs])
	} else {
		d.Write(tmp[0:(bs - d.nbuf - 8)])
	}

	binary.BigEndian.PutUint64(tmp[:], d.blocks+1)
	d.Write(tmp[:8])

	if d.nbuf != 0 {
		panic("padding failed")
	}

	d.finalTransform()

	// store chaining in output byteslice
	hash := make([]byte, d.columns*4)
	for i := 0; i < d.columns/2; i++ {
		binary.BigEndian.PutUint64(hash[(i*8):(i+1)*8], d.chaining[i+(d.columns/2)])
	}
	hash = hash[(len(hash) - d.hashbitlen/8):]
	return hash
}

// Each Sum...() function returns according hash value for provided data.
func Sum224(data []byte) []byte {
	d := New224().(*digest)
	d.Write(data)
	return d.checkSum()
}

func Sum256(data []byte) []byte {
	d := New256().(*digest)
	d.Write(data)
	return d.checkSum()
}

func Sum384(data []byte) []byte {
	d := New384().(*digest)
	d.Write(data)
	return d.checkSum()
}

func Sum512(data []byte) []byte {
	d := New512().(*digest)
	d.Write(data)
	return d.checkSum()
}
