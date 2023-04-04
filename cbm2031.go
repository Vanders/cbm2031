package main

import (
	"fmt"
	"io"
	"os"

	"github.com/vanders/pet/mos6502"
)

type CBM2031Connector struct {
	Via *VIA
}

func (c CBM2031Connector) Read() IEEE488 {
	var (
		portA, portB, portADir, portBDir Byte
		ca1                              bool
		out                              IEEE488
	)

	out = MakeIEEE488()
	portA = c.Via.Out(PORT_A)
	portADir = c.Via.PeekRegister(PORT_A_DIR)
	ca1 = c.Via.CtrlPeek(CTRL_CA1)
	portB = c.Via.Out(PORT_B)
	portBDir = c.Via.PeekRegister(PORT_B_DIR)

	var atna, ndac, nrfd bool
	if portB&mos6502.BIT_0 != 0 {
		atna = true // TTL high
	}
	if portB&mos6502.BIT_1 != 0 {
		nrfd = true // TTL high
	}
	if portB&mos6502.BIT_2 != 0 {
		ndac = true // TTL high
	}

	/*
		ATNA & CA1 (which is the inverse of ATN) are XOR'd together. If the output
		is true it is then inverted to drive NDAC & NRFD
	*/
	if atna != ca1 {
		nrfd = false // TTL low, inverted by a NOT gate
		ndac = false // TTL low, inverted by a NOT gate
	}

	/* Convert TTL to IEEE488 */
	if portBDir&mos6502.BIT_1 == 0 {
		// NRFD is an input
		out.NRFD = FLOATING
	} else {
		if nrfd {
			out.NRFD = FALSE
		} else {
			out.NRFD = TRUE
		}
	}
	if portBDir&mos6502.BIT_2 == 0 {
		// NDAC is an input
		out.NRFD = FLOATING
	} else {
		if ndac {
			out.NDAC = FALSE
		} else {
			out.NDAC = TRUE
		}
	}

	if portBDir&mos6502.BIT_3 == 0 {
		// EOI is an input
		out.EOI = FLOATING
	} else {
		if portB&mos6502.BIT_3 != 0 {
			out.EOI = FALSE
		} else {
			out.EOI = TRUE
		}
	}
	if portBDir&mos6502.BIT_6 == 0 {
		// DAV is an input
		out.DAV = FLOATING
	} else {
		if portB&mos6502.BIT_6 != 0 {
			out.DAV = FALSE
		} else {
			out.DAV = TRUE
		}
	}

	out.DIO = portA & portADir

	return out
}

func (c CBM2031Connector) Write(i IEEE488) {
	var (
		TTL   bool
		portB Byte
	)

	// ATN. The signal is inverted (NOR in the circuit) & connected to PB7/CA1
	TTL = !i.ATN.ToTTL()
	//fmt.Printf("ATN is %s. Seting PB7/CA1 to inverted ATN (%t)\n", i.ATN.ToOnOff(), TTL)
	if TTL {
		portB |= mos6502.BIT_7
	}
	c.Via.CtrlIn(CTRL_CA1, TTL)

	// NRFD is connected to PB1/CA2
	TTL = i.NRFD.ToTTL()
	//fmt.Printf("NRFD is %s. Setting PB1/CA2 to NRFD (%t)\n", i.NRFD.ToOnOff(), TTL)
	if TTL {
		portB |= mos6502.BIT_1
	}
	c.Via.CtrlIn(CTRL_CA2, TTL)

	// NDAC is connected to PB2
	TTL = i.NDAC.ToTTL()
	//fmt.Printf("NDAC is %s. Setting PB2 to NDAC (%t)\n", i.NDAC.ToOnOff(), TTL)
	if TTL {
		portB |= mos6502.BIT_2
	}

	// EOI is connected to PB3
	TTL = i.EOI.ToTTL()
	//fmt.Printf("EOI is %s. Setting PB3 to EOI (%t)\n", i.EOI.ToOnOff(), TTL)
	if TTL {
		portB |= mos6502.BIT_3
	}

	// DAV is connected to PB6
	TTL = i.DAV.ToTTL()
	//fmt.Printf("DAV is %s. Setting PB6 to DAV (%t)\n", i.DAV.ToOnOff(), TTL)
	if TTL {
		portB |= mos6502.BIT_6
	}

	c.Via.In(PORT_A, i.DIO)
	c.Via.In(PORT_B, portB) // Port B
}

type CBM2031 struct {
	Cable *Cable
	VIA   *VIA
	RAM   *RAM

	cpu   *mos6502.CPU
	bus   *Bus
	via1  *VIA
	via2  *VIA
	ram   *RAM
	hiRom *ROM
	loRom *ROM
}

func NewCBM2031(writer io.Writer) *CBM2031 {
	// Create a new memory bus
	bus := &Bus{}

	// Initialise memory

	// Main memory
	ram := &RAM{
		Base: 0x0000,
		Size: Word(8 * 1024), // 8k
	}
	ram.Reset()
	bus.Map(ram)

	// Load ROMs
	loRom := &ROM{
		Base: 0xc000,
		Size: 0x2000, // 8k
	}
	loRom.Reset()
	loRom.Load("roms/901484-03.bin")
	bus.Map(loRom)

	hiRom := &ROM{
		Base: 0xe000,
		Size: 0x2000, // 8k
	}
	hiRom.Reset()
	hiRom.Load("roms/901484-05.bin")
	bus.Map(hiRom)

	// VIA1
	via1 := &VIA{
		Base: 0x1800,
	}
	bus.Map(via1)

	// VIA2
	via2 := &VIA{
		Base: 0x1c00,
	}
	bus.Map(via2)

	cpu := mos6502.NewCPU(bus.Read, bus.Write, nil, writer)
	cpu.Reset()

	return &CBM2031{
		cpu:   cpu,
		bus:   bus,
		via1:  via1,
		VIA:   via1,
		via2:  via2,
		ram:   ram,
		RAM:   ram,
		hiRom: hiRom,
		loRom: loRom,
	}
}

// Create a new IEEE488 connector
func (c *CBM2031) CreateConnector() *CBM2031Connector {
	return &CBM2031Connector{
		Via: c.via1,
	}
}

func (c *CBM2031) Run() {
	// Run the CPU
	for {
		// Execute a single instruction
		err := c.cpu.Step()
		if err != nil {
			dumpAndExit(c.cpu, c.ram, fmt.Errorf("\nexecution stopped: %s", err))
		}

		c.via1.Clock()
		c.via2.Clock()

		// Sync data on the IEEE488 interface
		c.Cable.Sync()

		// Check devices for interrupts
		if c.bus.CheckInterrupts() {
			c.cpu.Interrupt()
		}
	}
}

func dumpAndExit(cpu *mos6502.CPU, ram *RAM, err error) {
	fmt.Println(err)
	dump(cpu, ram)
	os.Exit(1)
}

func dump(cpu *mos6502.CPU, ram *RAM) {
	cpu.Dump()

	if cpu.Writer != nil {
		// Dump the Zero Page
		for n := 0; n < 256; n = n + 4 {
			fmt.Fprintf(cpu.Writer,
				"0x%04x: 0x%02x,\t0x%04x: 0x%02x,\t0x%04x: 0x%02x,\t0x%04x: 0x%02x\n",
				Word(n),
				ram.Read(Word(n)),
				Word(n+1),
				ram.Read(Word(n+1)),
				Word(n+2),
				ram.Read(Word(n+2)),
				Word(n+3),
				ram.Read(Word(n+3)))
		}
	}
}
