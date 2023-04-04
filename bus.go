package main

import (
	"fmt"
	"io"

	"github.com/vanders/pet/mos6502"
)

type Device interface {
	GetBase() Word
	GetSize() Word
	CheckInterrupt() bool
	mos6502.ReadWriter
}

type Bus struct {
	Devices []Device

	Writer io.Writer // io.Writer for log output
}

func (b *Bus) debug(format string, a ...any) (int, error) {
	if b.Writer != nil {
		return fmt.Fprintf(b.Writer, format, a...)
	} else {
		return 0, nil
	}
}

func (b *Bus) Map(device Device) {
	// Insert devices in order of specificity E.g. more specific at the front
	b.Devices = append([]Device{device}, b.Devices...)
}

func (b *Bus) Read(address Word) Byte {
	b.debug("read $%04x\n", address)

	for n, d := range b.Devices {
		base := d.GetBase()
		size := d.GetSize()
		top := base + (size - 1)

		b.debug("device %d at $%04x:$%04x\n", n, base, base+top)

		if address >= base && address <= top {
			b.debug("selected device %d at $%04x\n", n, base)
			return d.Read(address)
		}
	}
	return Byte(0)
}

func (b *Bus) Write(address Word, data Byte) {
	b.debug("write $%04x\n", address)

	for n, d := range b.Devices {
		base := d.GetBase()
		size := d.GetSize()
		top := base + (size - 1)

		b.debug("device %d at $%04x:$%04x\n", n, base, base+top)
		if address >= base && address <= top {
			b.debug("selected device %d at $%04x\n", n, base)
			d.Write(address, data)
		}
	}
}

func (b *Bus) CheckInterrupts() bool {
	// Check devices for interrupts
	for _, d := range b.Devices {
		if d.CheckInterrupt() {
			return true
		}
	}
	return false
}
