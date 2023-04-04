package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// DummyConnector is a do-nothing stub
type DummyConnector struct {
	In struct {
		DAV  IEEEBool
		EOI  IEEEBool
		NDAC IEEEBool
		NRFD IEEEBool
		SRQ  IEEEBool
		ATN  IEEEBool
		REN  IEEEBool
		IFC  IEEEBool

		DIO Byte
	}
	Out struct {
		DAV  IEEEBool
		EOI  IEEEBool
		NDAC IEEEBool
		NRFD IEEEBool
		SRQ  IEEEBool
		ATN  IEEEBool
		REN  IEEEBool
		IFC  IEEEBool

		DIO Byte
	}
}

func (c *DummyConnector) Reset() {
	c.Out.DAV = FLOATING
	c.Out.EOI = FLOATING
	c.Out.NDAC = FLOATING
	c.Out.NRFD = FLOATING
	c.Out.SRQ = FLOATING
	c.Out.ATN = FLOATING
	c.Out.REN = FLOATING
	c.Out.IFC = FLOATING
}

func (c *DummyConnector) Read() IEEE488 {
	return IEEE488{
		ATN:  c.Out.ATN,
		NRFD: c.Out.NRFD,
		NDAC: c.Out.NDAC,
		EOI:  c.Out.EOI,
		DAV:  c.Out.DAV,
		REN:  c.Out.REN,
		SRQ:  c.Out.SRQ,
		IFC:  c.Out.IFC,
		DIO:  ^c.Out.DIO, // We must invert DIO before writing it
	}
}

func (c *DummyConnector) Write(i IEEE488) {
	c.In.ATN = i.ATN
	c.In.NRFD = i.NRFD
	c.In.NDAC = i.NDAC
	c.In.EOI = i.EOI
	c.In.DAV = i.DAV
	c.In.REN = i.REN
	c.In.SRQ = i.SRQ
	c.In.IFC = i.IFC
	c.In.DIO = ^i.DIO // DIO in comes to us inverted (0v = 1, 5v = 0) so invert it back to a normal byte
}

func (c *DummyConnector) Dump() {
	fmt.Println("IN")
	fmt.Printf("ATN: %s, NRFD: %s, NDAC: %s, EOI: %s, DAV: %s\n",
		c.In.ATN.ToOnOff(),
		c.In.NRFD.ToOnOff(),
		c.In.NDAC.ToOnOff(),
		c.In.EOI.ToOnOff(),
		c.In.DAV.ToOnOff())
	fmt.Printf("DATA: 0x%02x (Inverted from 0x%02x)\n", c.In.DIO, ^c.In.DIO)
	//fmt.Printf("REN: %t, SRQ: %t, IFC: %t\n", c.In.REN, c.In.SRQ, c.In.IFC)
	fmt.Println("")
	fmt.Println("OUT")
	fmt.Printf("ATN: %s, NRFD: %s, NDAC: %s, EOI: %s, DAV: %s\n",
		c.Out.ATN.ToOnOff(),
		c.Out.NRFD.ToOnOff(),
		c.Out.NDAC.ToOnOff(),
		c.Out.EOI.ToOnOff(),
		c.Out.DAV.ToOnOff())
	//fmt.Printf("REN: %t, SRQ: %t, IFC: %t\n", c.Out.REN, c.Out.SRQ, c.Out.IFC)
	fmt.Printf("DATA: 0x%02x (Inverted from 0x%02x)\n", ^c.Out.DIO, c.Out.DIO)
}

type Monitor struct {
	A    *DummyConnector
	BVIA *VIA
	BRAM *RAM
}

func NewMonitor(connector *DummyConnector) *Monitor {
	return &Monitor{
		A: connector,
	}
}

func (m *Monitor) Run() {
	// Read input
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')

		args := strings.Split(strings.Trim(input, "\r\n"), " ")
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "exit":
			os.Exit(0)
		case "dump":
			m.A.Dump()
		case "via":
			switch args[1] {
			case "a":
				fmt.Printf("Port A: $%02x\n", m.BVIA.PeekRegister(PORT_A))
			case "adir":
				fmt.Printf("Port A Dir: %02x\n", m.BVIA.PeekRegister(PORT_A_DIR))
			case "b":
				fmt.Printf("Port B: $%02x\n", m.BVIA.PeekRegister(PORT_B))
			case "bdir":
				fmt.Printf("Port B Dir: %02x\n", m.BVIA.PeekRegister(PORT_B_DIR))
			case "t1lo", "t1low":
				fmt.Printf("Timer 1 Low: %02x\n", m.BVIA.PeekRegister(TIMER_1_LOW))
			case "t1hi", "t1high":
				fmt.Printf("Timer 1 High: %02x\n", m.BVIA.PeekRegister(TIMER_1_HIGH))
			case "ifr":
				fmt.Printf("IFR: %02x\n", m.BVIA.PeekRegister(INT_FLAGS))
			case "ie":
				fmt.Printf("IE: %02x\n", m.BVIA.PeekRegister(INT_ENABLE))
			case "acr":
				fmt.Printf("ACR: %02x\n", m.BVIA.PeekRegister(AUXILLERY_CTRL))
			case "pcr":
				fmt.Printf("PCR: %02x\n", m.BVIA.PeekRegister(PERIPHERAL_CTRL))
			case "irq":
				fmt.Printf("IRQ: %t\n", m.BVIA.CheckInterrupt())
			}
		case "peek":
			if len(args) != 2 {
				fmt.Println("peek addr")
				break
			}
			addr, err := strconv.ParseInt(args[1], 16, 17)
			if err != nil {
				fmt.Printf("invalid addr: %w", err)
				break
			}
			data := m.BRAM.Read(Word(addr))
			fmt.Printf("$%02x: $%02x\n", addr, data)
		case "poke":
			if len(args) != 3 {
				fmt.Println("poke addr data")
				break
			}
			addr, err := strconv.ParseInt(args[1], 16, 17)
			if err != nil {
				fmt.Printf("invalid addr: %w", err)
				break
			}
			data, err := strconv.ParseInt(args[2], 16, 9)
			if err != nil {
				fmt.Printf("invalid data: %w", err)
				break
			}
			m.BRAM.Write(Word(addr), Byte(data))
			fmt.Printf("$%02x: $%02x\n", addr, data)
		case "open":
			err := m.open(args[1:])
			if err != nil {
				fmt.Println(err)
			}
		case "input":
			err := m.open(args[1:])
			if err != nil {
				fmt.Println(err)
			}
			data, err := m.input()
			if err != nil {
				fmt.Println(err)
			}
			m.untalk()

			for _, ch := range data {
				fmt.Printf("0x%02x ", ch)
			}
			fmt.Printf("\n")
			for _, ch := range data {
				fmt.Printf("%s", string(ch))
			}
			fmt.Printf("\n")
		default:
			fmt.Println("?")
		}
	}
}

func (m *Monitor) open(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: open primary_addr [secondary_addr]")
	}

	var (
		primary, secondary int64
		err                error
	)

	primary, err = strconv.ParseInt(args[0], 16, 9)
	if err != nil {
		return fmt.Errorf("invalid primary_addr: %w", err)
	}
	// Mark the primary address TALK
	primary = primary + 0x40

	if len(args) == 2 {
		secondary, err = strconv.ParseInt(args[1], 16, 9)
		if err != nil {
			return fmt.Errorf("invalid secondary_addr: %w", err)
		}
		// Mark the secondary address SECOND
		secondary = secondary + 0x60
	} else {
		secondary = 0xff
	}

	return m.cmd(Byte(primary), Byte(secondary))
}

// Send a command (Pull ATN, set primary && secondary address)
func (m *Monitor) cmd(primary, secondary Byte) error {
	var ack, ready, busy, accepted bool

	// Pull ATN
	defer func() {
		/* Command sequence finished */
		m.A.Out.ATN = FALSE
	}()
	m.A.Reset()
	m.A.Out.ATN = TRUE

	/* The drive will pull both NDAC & NRFD when acknowledging ATN: the
	standard says that only NDAC is significant */
	ack = false
	for n := 0; n < 500 && !ack; n++ {
		if m.A.In.NDAC == TRUE {
			ack = true
		} else {
			fmt.Printf("NDAC=%s, NRFD=%s\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
			time.Sleep(1 * time.Millisecond)
		}
	}
	if ack {
		fmt.Printf("remote acknowledged ATN (NDAC=%s, NRFD=%s)\n",
			m.A.In.NDAC.ToOnOff(),
			m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not acknowledge ATN (NDAC=%s, NRFD=%s)\n",
			m.A.In.NDAC.ToOnOff(),
			m.A.In.NRFD.ToOnOff())
	}

	// We have its attention
	m.A.Out.DIO = primary
	fmt.Printf("DIO set to 0x%02x (Inverted to 0x%02x)\n", m.A.Out.DIO, ^m.A.Out.DIO)

	/* It should then release NRFD */
	ready = false
	for n := 0; n < 500 && !ready; n++ {
		if m.A.In.NDAC == TRUE && m.A.In.NRFD == FALSE {
			ready = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if ready {
		fmt.Printf("remote became ready (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not become ready (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	}

	/* Data is valid */
	m.A.Out.DAV = TRUE

	/* Wait for remote to pull NRFD to indicate it is busy */
	busy = false
	for n := 0; n < 500 && !busy; n++ {
		if m.A.In.NRFD == TRUE {
			busy = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if busy {
		fmt.Printf("remote became busy (NRFD=%s)\n", m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not become busy (NRFD=%s)\n", m.A.In.NRFD.ToOnOff())
	}

	/* Wait for remote to release NDAC to indicate it has accepted the byte */
	accepted = false
	for n := 0; n < 500 && !accepted; n++ {
		if m.A.In.NDAC == FALSE {
			accepted = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if accepted {
		fmt.Printf("remote accepted byte (NDAC=%s)\n", m.A.In.NDAC.ToOnOff())
	} else {
		return fmt.Errorf("remote did not accept byte (NDAC=%s)\n", m.A.In.NDAC.ToOnOff())
	}

	/* First byte sent, data is no longer valid */
	m.A.Out.DAV = FALSE
	m.A.Out.DIO = Byte(0)

	idle := false
	for n := 0; n < 500 && !idle; n++ {
		if m.A.In.NDAC == TRUE {
			idle = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if idle {
		fmt.Printf("remote became idle (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not became idle (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	}

	// Send the secondary address if there is one
	if secondary != 0xff {
		fmt.Printf("secondary_addr=%x\n", secondary)

		m.A.Out.DIO = secondary
		fmt.Printf("DIO set to 0x%02x (Inverted to 0x%02x)\n", m.A.Out.DIO, ^m.A.Out.DIO)

		/* Ensure NRFD is high */
		ready = false
		for n := 0; n < 500 && !ready; n++ {
			if m.A.In.NDAC == TRUE && m.A.In.NRFD == FALSE {
				ready = true
			} else {
				time.Sleep(1 * time.Millisecond)
			}
		}
		if ready {
			fmt.Printf("remote became ready (NDAC=%s, NRFD=%s)\n",
				m.A.In.NDAC.ToOnOff(),
				m.A.In.NRFD.ToOnOff())
		} else {
			return fmt.Errorf("remote did not become ready (NDAC=%s, NRFD=%s)\n",
				m.A.In.NDAC.ToOnOff(),
				m.A.In.NRFD.ToOnOff())
		}

		/* Data is valid */
		m.A.Out.DAV = TRUE

		/* Wait for remote to pull NRFD to indicate it is busy */
		busy = false
		for n := 0; n < 500 && !busy; n++ {
			if m.A.In.NRFD == TRUE {
				busy = true
			} else {
				time.Sleep(1 * time.Millisecond)
			}
		}
		if busy {
			fmt.Printf("remote became busy (NRFD=%s)\n", m.A.In.NRFD.ToOnOff())
		} else {
			return fmt.Errorf("remote did not become busy (NRFD=%s)\n", m.A.In.NRFD.ToOnOff())
		}

		/* Wait for remote to release NDAC to indicate it has accepted the byte */
		accepted = false
		for n := 0; n < 500 && !accepted; n++ {
			if m.A.In.NDAC == FALSE {
				accepted = true
			} else {
				time.Sleep(1 * time.Millisecond)
			}
		}
		if accepted {
			fmt.Printf("remote accepted byte (NDAC=%s)\n", m.A.In.NDAC.ToOnOff())
		} else {
			return fmt.Errorf("remote did not accept byte (NDAC=%s)\n", m.A.In.NDAC.ToOnOff())
		}

		/* Second byte sent, data is no longer valid */
		m.A.Out.DAV = FALSE
		m.A.Out.DIO = Byte(0xff)

		idle := false
		for n := 0; n < 500 && !idle; n++ {
			if m.A.In.NDAC == TRUE {
				idle = true
			} else {
				time.Sleep(1 * time.Millisecond)
			}
		}
		if idle {
			fmt.Printf("remote became idle (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
		} else {
			return fmt.Errorf("remote did not became idle (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
		}
	}

	/* Command sequence finished */
	m.A.Out.ATN = FALSE
	m.A.Out.NRFD = TRUE
	m.A.Out.NDAC = TRUE

	return nil
}

// Input from the bus
func (m *Monitor) input() ([]Byte, error) {
	var data []Byte

	// Max. data is 80 characters, or until we read CR
	for len(data) < 80 {
		fmt.Printf("portA=0x%02x, portAdir=0x%02x\n", m.BVIA.PeekRegister(PORT_A), m.BVIA.PeekRegister(PORT_A_DIR))
		ch, err := m.readByte()
		if err != nil {
			return data, fmt.Errorf("failed to read from bus: %w", err)
		}
		data = append(data, ch)

		// Did we read CR?
		if ch == 0x13 {
			break
		}
	}
	fmt.Printf("read %d characters\n", len(data))

	return data, nil
}

// Reads a single byte from the bus
func (m *Monitor) readByte() (Byte, error) {
	var (
		data            Byte
		valid, finished bool
	)

	// We are ready for data
	defer func() {
		m.A.Out.NRFD = TRUE
		//m.A.Out.NDAC = FALSE
	}()
	m.A.Out.DAV = FLOATING
	m.A.Out.NRFD = FALSE
	m.A.Out.NDAC = TRUE

	fmt.Println("Waiting for data...")

	// Wait for the remote end to assert DAV
	valid = false
	for n := 0; n < 500 && !valid; n++ {
		if m.A.In.DAV == TRUE {
			valid = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if valid {
		fmt.Printf("data valid (DAV=%s)\n", m.A.In.DAV.ToOnOff())
	} else {
		return data, fmt.Errorf("no valid data (DAV=%s)\n", m.A.In.DAV.ToOnOff())
	}
	m.A.Out.NRFD = TRUE
	time.Sleep(1 * time.Millisecond)

	// Read DIO
	data = m.A.In.DIO
	fmt.Printf("DIO set to 0x%02x (Inverted from 0x%02x)\n", m.A.In.DIO, ^m.A.In.DIO)

	// Was EOI asserted?
	if m.A.In.EOI == TRUE {
		fmt.Println("EOI asserted")
		data = 0x13 // Force data to CR
	}

	// Pull NDAC to indicate we have read the byte
	m.A.Out.NDAC = FALSE

	// Wait for the remote to raise DAV
	finished = false
	for n := 0; n < 500 && !finished; n++ {
		if m.A.In.DAV == FALSE {
			finished = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if finished {
		fmt.Printf("read accepted (DAV=%s)\n", m.A.In.DAV.ToOnOff())
	} else {
		return data, fmt.Errorf("read not accepted (DAV=%s)\n", m.A.In.DAV.ToOnOff())
	}

	// We are not ready for the next byte
	m.A.Out.NDAC = TRUE
	m.A.Out.NRFD = TRUE

	return data, nil
}

func (m *Monitor) untalk() error {
	var ack, ready, busy, accepted bool

	// Pull ATN
	defer func() {
		/* Command sequence finished */
		m.A.Out.ATN = FALSE
	}()
	m.A.Reset()
	m.A.Out.ATN = TRUE

	/* The drive will pull both NDAC & NRFD when acknowledging ATN: the
	standard says that only NDAC is significant */
	ack = false
	for n := 0; n < 500 && !ack; n++ {
		if m.A.In.NDAC == TRUE {
			ack = true
		} else {
			fmt.Printf("NDAC=%s, NRFD=%s\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
			time.Sleep(1 * time.Millisecond)
		}
	}
	if ack {
		fmt.Printf("remote acknowledged ATN (NDAC=%s, NRFD=%s)\n",
			m.A.In.NDAC.ToOnOff(),
			m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not acknowledge ATN (NDAC=%s, NRFD=%s)\n",
			m.A.In.NDAC.ToOnOff(),
			m.A.In.NRFD.ToOnOff())
	}

	// We have its attention; send UNTALK
	m.A.Out.DIO = 0x5f
	fmt.Printf("DIO set to 0x%02x (Inverted to 0x%02x)\n", m.A.Out.DIO, ^m.A.Out.DIO)

	/* It should then release NRFD */
	ready = false
	for n := 0; n < 500 && !ready; n++ {
		if m.A.In.NDAC == TRUE && m.A.In.NRFD == FALSE {
			ready = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if ready {
		fmt.Printf("remote became ready (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not become ready (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	}

	/* Data is valid */
	m.A.Out.DAV = TRUE

	/* Wait for remote to pull NRFD to indicate it is busy */
	busy = false
	for n := 0; n < 500 && !busy; n++ {
		if m.A.In.NRFD == TRUE {
			busy = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if busy {
		fmt.Printf("remote became busy (NRFD=%s)\n", m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not become busy (NRFD=%s)\n", m.A.In.NRFD.ToOnOff())
	}

	/* Wait for remote to release NDAC to indicate it has accepted the byte */
	accepted = false
	for n := 0; n < 500 && !accepted; n++ {
		if m.A.In.NDAC == FALSE {
			accepted = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if accepted {
		fmt.Printf("remote accepted byte (NDAC=%s)\n", m.A.In.NDAC.ToOnOff())
	} else {
		return fmt.Errorf("remote did not accept byte (NDAC=%s)\n", m.A.In.NDAC.ToOnOff())
	}

	/* First byte sent, data is no longer valid */
	m.A.Out.DAV = FALSE
	m.A.Out.DIO = Byte(0)

	idle := false
	for n := 0; n < 500 && !idle; n++ {
		if m.A.In.NDAC == TRUE {
			idle = true
		} else {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if idle {
		fmt.Printf("remote became idle (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	} else {
		return fmt.Errorf("remote did not became idle (NDAC=%s, NRFD=%s)\n", m.A.In.NDAC.ToOnOff(), m.A.In.NRFD.ToOnOff())
	}

	return nil
}
