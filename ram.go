package main

type RAM struct {
	Base Word // Base address
	Size Word // Size

	mem []Byte
}

func (r *RAM) GetBase() Word {
	return r.Base
}

func (r *RAM) GetSize() Word {
	return r.Size
}

func (r *RAM) CheckInterrupt() bool {
	return false
}

func (r *RAM) Reset() {
	r.mem = make([]Byte, r.Size)

	for n := Word(0); n < r.Size; n++ {
		r.mem[n] = 0x00
	}
}

func (r *RAM) Read(address Word) Byte {
	return r.mem[address-r.Base]
}

func (r *RAM) Write(address Word, data Byte) {
	r.mem[address-r.Base] = data
}
