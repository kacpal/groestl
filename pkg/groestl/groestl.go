package groestl

import (
	"encoding/binary"
	"hash"
)

const VERBOSE = false

type digest struct {
	hashbitlen int
	chaining   [16]uint64
	blocks     uint64
	buf        [128]byte
	nbuf       int
	columns    int
	rounds     int
}

func (d *digest) Reset() {
	// equivalent to Init from reference implementation

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

func New() hash.Hash {
	return New256()
}

func (d *digest) Size() int {
	return d.hashbitlen
}

func (d *digest) BlockSize() int {
	if d.hashbitlen <= 256 {
		return 64
	} else {
		return 128
	}
}

func (d *digest) Write(p []byte) (n int, err error) {
	// equivalent to Update from reference implementation

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

func (d *digest) checkSum() []byte {
	// equivalent to Final from reference implementation

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
	hash := make([]byte, d.hashbitlen/8)
	for i := 0; i < d.columns/2; i++ {
		binary.BigEndian.PutUint64(hash[(i*8):(i+1)*8], d.chaining[i+(d.columns/2)])
	}

	return hash
}

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
