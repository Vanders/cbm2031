package main

import (
	"fmt"

	"github.com/vanders/pet/mos6502"
)

type (
	VIARegister  = Byte
	VIAInterrupt = Byte
	VIAControl   = Byte
)

type VIA struct {
	Base Word

	portB    Byte
	portBOut Byte
	portBDir Byte
	portA    Byte
	portAOut Byte
	portADir Byte

	timer1Low       Byte
	timer1High      Byte
	timer1LatchLow  Byte
	timer1LatchHigh Byte
	timer1Counter   Word

	acr Byte // Auxillery Control Register
	pcr Byte // Peripheral Control Register

	ifr Byte // Interrupt Flags Register
	ie  Byte // Interrupt Enable Register

	ca1 bool // Current state of the CA1 line
	ca2 bool // Current state of the CA2 line
	cb1 bool // Current state of the CB1 line
	cb2 bool // Current state of the CB2 line
}

func (v *VIA) GetBase() Word {
	return v.Base
}

func (v *VIA) GetSize() Word {
	return Word(16)
}

func (v *VIA) CheckInterrupt() bool {
	return (v.ifr & 0x80) != 0
}

func (v *VIA) Clock() {
	t1start := v.timer1Counter

	if v.timer1Counter > 0 {
		v.timer1Counter = v.timer1Counter - 1
	}
	//fmt.Printf("$%04x t1=0x%02x, acr=0x%02x, ie=0x%02x\r", v.Base, v.timer1Counter, v.acr, v.ie)

	// Did T1 decrement to 0? If so raise an interrupt
	if t1start > 0 && v.timer1Counter == 0 {
		v.setInterrupt(INT_T1)

		// If T1 is in continuous mode, reload the counter
		if v.acr&mos6502.BIT_6 != 0 {
			//fmt.Println("loading T1")
			v.setTimer1Counter()
		}
	}
}

// Interrupt bits
const (
	INT_CA2 VIAInterrupt = 1 << 0
	INT_CA1 VIAInterrupt = 1 << 1

	INT_CB2 VIAInterrupt = 1 << 3
	INT_CB1 VIAInterrupt = 1 << 4

	INT_T2 VIAInterrupt = 1 << 5
	INT_T1 VIAInterrupt = 1 << 6

	INT_IRQ VIAInterrupt = 1 << 7
)

func (v *VIA) setOrClearIRQ() {
	var irq Byte

	irq = (v.ifr & 0x7f) & (v.ie & 0x7f)
	if irq == 0 {
		v.ifr = v.ifr & 0x7f
	} else {
		v.ifr = v.ifr | INT_IRQ
	}
}

func (v *VIA) setInterrupt(i VIAInterrupt) {
	v.ifr = v.ifr | i
	v.setOrClearIRQ()
}

func (v *VIA) clearInterrupt(i VIAInterrupt) {
	v.ifr = v.ifr & ^i
	v.setOrClearIRQ()
}

// VIA Registers
const (
	PORT_B             VIARegister = 0x0
	PORT_A             VIARegister = 0x1
	PORT_B_DIR         VIARegister = 0x2
	PORT_A_DIR         VIARegister = 0x3
	TIMER_1_LOW        VIARegister = 0x4
	TIMER_1_HIGH       VIARegister = 0x5
	TIMER_1_LATCH_LOW  VIARegister = 0x6
	TIMER_1_LATCH_HIGH VIARegister = 0x7
	TIMER_2_LOW        VIARegister = 0x8
	TIMER_2_HIGH       VIARegister = 0x9
	AUXILLERY_CTRL     VIARegister = 0xb // Auxillery control
	PERIPHERAL_CTRL    VIARegister = 0xc // Peripheral control
	INT_FLAGS          VIARegister = 0xd // Interrupt flag register (IFR)
	INT_ENABLE         VIARegister = 0xe // Interrupt enable register
)

func (v *VIA) setTimer1Counter() {
	v.timer1Counter = Word(v.timer1LatchHigh)<<8 | Word(v.timer1LatchLow)
}

func (v *VIA) ReadRegister(r VIARegister) Byte {
	switch r {
	case PORT_B:
		v.clearInterrupt(INT_CB1)
		if v.pcr&PCR_CB2_1 == 0 {
			v.clearInterrupt(INT_CB2)
		}
		return v.portB
	case PORT_A:
		v.clearInterrupt(INT_CA1)
		if v.pcr&PCR_CA2_1 == 0 {
			v.clearInterrupt(INT_CA2)
		}
		return v.portA
	case PORT_B_DIR:
		return v.portBDir
	case PORT_A_DIR:
		return v.portADir
	case TIMER_1_LOW:
		v.clearInterrupt(INT_T1)
		return v.timer1Low
	case TIMER_1_HIGH:
		return v.timer1High
	case TIMER_1_LATCH_LOW:
		return v.timer1LatchLow
	case TIMER_1_LATCH_HIGH:
		return v.timer1LatchHigh
	case AUXILLERY_CTRL:
		return v.acr
	case PERIPHERAL_CTRL:
		return v.pcr
	case INT_FLAGS:
		return v.ifr
	case INT_ENABLE:
		return v.ie
	}

	return Byte(0)
}

func (v *VIA) PeekRegister(r VIARegister) Byte {
	switch r {
	case PORT_B:
		return v.portB
	case PORT_B_DIR:
		return v.portBDir
	case PORT_A:
		return v.portA
	case PORT_A_DIR:
		return v.portADir
	case TIMER_1_LOW:
		return v.timer1Low
	case TIMER_1_HIGH:
		return v.timer1High
	case AUXILLERY_CTRL:
		return v.acr
	case PERIPHERAL_CTRL:
		return v.pcr
	case INT_FLAGS:
		return v.ifr
	case INT_ENABLE:
		return v.ie
	}

	return Byte(0)
}

func (v *VIA) WriteRegister(r VIARegister, data Byte) {
	switch r {
	case PORT_B:
		v.portB = data
		v.clearInterrupt(INT_CB1)
		if v.pcr&PCR_CB2_1 == 0 {
			v.clearInterrupt(INT_CB2)
		}
		v.portBOut = v.portB & v.portBDir
	case PORT_B_DIR:
		v.portBDir = data
	case PORT_A:
		v.portA = data
		v.clearInterrupt(INT_CA1)
		if v.pcr&PCR_CA2_1 == 0 {
			v.clearInterrupt(INT_CA2)
		}
		v.portAOut = v.portA & v.portADir
	case PORT_A_DIR:
		v.portADir = data
	case TIMER_1_LOW:
		v.timer1Low = data
		v.timer1LatchLow = data
	case TIMER_1_HIGH:
		v.timer1High = data
		v.timer1LatchHigh = data
		v.timer1Low = v.timer1LatchLow

		v.clearInterrupt(INT_T1)
		v.setTimer1Counter()
	case TIMER_1_LATCH_LOW:
		v.timer1LatchLow = data
	case TIMER_1_LATCH_HIGH:
		v.timer1LatchHigh = data
		v.clearInterrupt(INT_T1)
	case AUXILLERY_CTRL:
		v.acr = data
	case PERIPHERAL_CTRL:
		v.pcr = data
	case INT_FLAGS:
		if data&INT_CA2 != 0 {
			v.clearInterrupt(INT_CA2)
		}
		if data&INT_CA1 != 0 {
			v.clearInterrupt(INT_CA1)
		}
		if data&INT_CB2 != 0 {
			v.clearInterrupt(INT_CB2)
		}
		if data&INT_CB1 != 0 {
			v.clearInterrupt(INT_CB1)
		}
	case INT_ENABLE:
		if data&0x80 == 0 {
			// Clear
			v.ie = (^v.ie | ^data) & 0x7f
		} else {
			// Set
			v.ie = (v.ie | data) & 0x7f
		}
	}
}

// Read & Write are CPU-bus side methods that take an address that decodes into
// the register
func (v *VIA) Read(addr Word) Byte {
	r := Byte(addr - v.Base)
	return v.ReadRegister(r)
}

func (v *VIA) Write(addr Word, data Byte) {
	r := Byte(addr - v.Base)
	v.WriteRegister(r, data)
}

/*
Each port has a Data Direction Register (DDRA, DDRB) for specifying whether
the peripheral pins are to act as inputs or outputs. A 0 in a bit of the Data
Direction Register causes the corresponding peripheral pin to act as an input.
A 1 causes the pin to act as an output.
*/

// In & Out are the GPIO side methods that take a register
func (v *VIA) Out(r VIARegister) Byte {
	var data Byte
	switch r {
	case PORT_A:
		data = v.portAOut
	case PORT_B:
		data = v.portBOut
	}
	return data
}

func (v *VIA) In(r VIARegister, data Byte) {
	switch r {
	case PORT_A:
		dir := v.ReadRegister(PORT_A_DIR)

		in := data & ^dir
		portA := v.portA & dir

		v.portA = in | portA
	case PORT_B:
		dir := v.ReadRegister(PORT_B_DIR)

		in := data & ^dir
		portB := v.portB & dir

		v.portB = in | portB
	}
}

// Control lines
const (
	CTRL_CA1 VIAControl = 0
	CTRL_CA2 VIAControl = 1
	CTRL_CB1 VIAControl = 2
	CTRL_CB2 VIAControl = 3
)

// Peripheral Control Register bits
const (
	PCR_CA1_CTRL = 1 << 0
	PCR_CA2_1    = 1 << 1
	PCR_CA2_2    = 1 << 2
	PCR_CA2_3    = 1 << 3
	PCR_CA2_CTRL = 0x0e

	PCR_CB1_CTRL = 1 << 4
	PCR_CB2_1    = 1 << 5
	PCR_CB2_2    = 1 << 6
	PCR_CB2_3    = 1 << 7
	PCR_CB2_CTRL = 0xe0
)

func (v *VIA) ctrlInSet(c VIAControl) {
	switch c {
	case CTRL_CA1:
		v.setInterrupt(INT_CA1)
	case CTRL_CA2:
		v.setInterrupt(INT_CA2)
	case CTRL_CB1:
		v.setInterrupt(INT_CB1)
	case CTRL_CB2:
		v.setInterrupt(INT_CB2)
	}
}

func (v *VIA) CtrlIn(c VIAControl, ttl bool) {
	var currentTTL bool

	switch c {
	case CTRL_CA1:
		currentTTL = v.ca1
	case CTRL_CA2:
		ca2_ctrl := v.PeekRegister(PERIPHERAL_CTRL) & PCR_CA2_CTRL
		if ca2_ctrl&PCR_CA2_3 != 0 {
			// CA2 is in output mode
			return
		}
		currentTTL = v.ca2
	case CTRL_CB1:
		currentTTL = v.cb1
	case CTRL_CB2:
		cb2_ctrl := v.PeekRegister(PERIPHERAL_CTRL) & PCR_CB2_CTRL
		if cb2_ctrl&PCR_CB2_3 != 0 {
			// CB2 is in output mode
			return
		}
		currentTTL = v.cb2
	}

	// If the current & input TTL are the same, no transition has taken place
	if ttl == currentTTL {
		return
	}

	// Was the transition from low->high, or high->low?
	const (
		LO_TO_HI = 0
		HI_TO_LO = 1
	)

	var transition int
	if ttl == true {
		// Low to high
		transition = LO_TO_HI
	} else {
		// High to low
		transition = HI_TO_LO
	}

	switch c {
	case CTRL_CA1:
		if v.pcr&PCR_CA1_CTRL == 0 && transition == HI_TO_LO {
			// Negative transition (high to low)
			fmt.Println("CA1 negative transition (high to low)")
			v.ctrlInSet(c)
		}
		if v.pcr&PCR_CA1_CTRL == 1 && transition == LO_TO_HI {
			// Positive transition (low to high)
			fmt.Println("CA1 positive transition (low to high)")
			v.ctrlInSet(c)
		}
		v.ca1 = ttl
	case CTRL_CA2:
		if v.pcr&PCR_CA2_3 == 0 { // Is CA2 in Input mode?
			// Is it in positive or negative mode?
			if v.pcr&PCR_CA2_2 == 0 && transition == HI_TO_LO {
				// Negative transition (high to low)
				v.ctrlInSet(c)
			}
			if v.pcr&PCR_CA2_2 == 1 && transition == LO_TO_HI {
				// Positive transition (low to high)
				v.ctrlInSet(c)
			}
			v.ca2 = ttl
		}
	case CTRL_CB1:
		if v.pcr&PCR_CB1_CTRL == 0 && transition == HI_TO_LO {
			// Negative transition (high to low)
			v.ctrlInSet(c)
		}
		if v.pcr&PCR_CB1_CTRL == 1 && transition == LO_TO_HI {
			// Positive transition (low to high)
			v.ctrlInSet(c)
		}
		v.cb1 = ttl
	case CTRL_CB2:
		if v.pcr&PCR_CB2_3 == 0 { // Is CB2 in Input mode?
			// Is it in positive or negative mode?
			if v.pcr&PCR_CB2_2 == 0 && transition == HI_TO_LO {
				// Negative transition (high to low)
				v.ctrlInSet(c)
			}
			if v.pcr&PCR_CB2_2 == 1 && transition == LO_TO_HI {
				// Positive transtion (low to high)
				v.ctrlInSet(c)
			}
			v.cb2 = ttl
		}
	}
}

func (v *VIA) CtrlOut(c VIAControl) bool {
	switch c {
	case CTRL_CA2:
		if v.pcr&PCR_CA2_3 != 0 {
			return (v.pcr & PCR_CA2_1) != 0
		}
	case CTRL_CB2:
		if v.pcr&PCR_CB2_3 != 0 {
			return (v.pcr & PCR_CB2_1) != 0
		}
	}
	return false
}

func (v *VIA) CtrlPeek(c VIAControl) bool {
	var currentTTL bool

	switch c {
	case CTRL_CA1:
		currentTTL = v.ca1
	case CTRL_CA2:
		currentTTL = v.ca2
	case CTRL_CB1:
		currentTTL = v.cb1
	case CTRL_CB2:
		currentTTL = v.cb2
	}
	return currentTTL
}
