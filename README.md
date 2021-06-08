# groestl

Grøstl hash function implementation in golang

## Usage

groestl library implements standard `hash.Hash` interface.

You can also run it from command-line:
```
$ ./groestl
Usage:
  ./groestl [options] path/to/file

Options:
  -hash int
    	output hash length (default 256)
```

To compile simply run `go build` in the root directory.
