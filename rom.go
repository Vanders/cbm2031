package main

import (
	"io/ioutil"
)

type ROM struct {
	Base Word
	Size Word

	mem []Byte
}

func (r *ROM) GetBase() Word {
	return r.Base
}

func (r *ROM) GetSize() Word {
	return r.Size
}

func (r *ROM) CheckInterrupt() bool {
	return false
}

func (r *ROM) Reset() {
	r.mem = make([]Byte, r.Size)

	for n := Word(0); n < r.Size; n++ {
		r.mem[n] = 0x00
	}
}

func (r *ROM) Read(address Word) Byte {
	return r.mem[address-r.Base]
}

func (r *ROM) Write(Word, Byte) {
	// ROM
}

func (r *ROM) Load(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	if Word(len(data)) > r.Size {
		panic("can't load that there (too big)")
	}
	for n := 0; n < len(data); n++ {
		r.mem[Word(n)] = Byte(data[n])
	}
}

func (r *ROM) ReadVector(address Word) Word {
	// The vector is an absolute JMP followed by the target address
	lo := r.Read(address + 1)
	hi := r.Read(address + 2)
	return Word(hi)<<8 | Word(lo)
}

// Patch the data relative to the given vector with the patch contents
func (r *ROM) PatchVector(vector Word, patch []Byte) {
	addr := r.ReadVector(vector)
	for n, b := range patch {
		r.mem[(addr+Word(n))-r.Base] = b
	}
}
