package main

import "fmt"

type IEEEBool int

/*
IEEE bus logic is inverted: 5V is 0 (false) and 0V is 1 (true). To make
that make sense here we'll use the literal voltage values, rather than the
logical values.
*/
const (
	TRUE     IEEEBool = 0
	FALSE    IEEEBool = 5
	FLOATING IEEEBool = -1 // Neither true nor false, a floating input
)

func (i IEEEBool) ToBool() bool {
	if i == TRUE {
		return true
	}
	return false
}

/*
Convert an IEEE signal to TTL:

0V = TRUE (IEEE) = 0 (TTL Low)
5V = FALSE (IEEE) = 1 (TTL High)
*/
func (i IEEEBool) ToTTL() bool {
	if i == FALSE {
		return true
	}
	return false
}

func (i IEEEBool) String() string {
	switch i {
	case TRUE:
		return "true"
	case FALSE:
		return "false"
	case FLOATING:
		return "floating"
	}
	return "???"
}

func (i IEEEBool) ToOnOff() string {
	switch i {
	case TRUE:
		return "ON"
	case FALSE:
		return "OFF"
	case FLOATING:
		return "FLOATING"
	}
	return "???"
}

func IEEEBoolFromBool(b bool) IEEEBool {
	if b {
		return TRUE
	}
	return FALSE
}

/* IEEE488 models a cable that contains the signals for an IEEE488 interface */
type IEEE488 struct {
	DAV  IEEEBool /* Data Valid */
	EOI  IEEEBool /* End or Identity */
	NDAC IEEEBool /* No Data Accepted */
	NRFD IEEEBool /* Not Ready For Data */
	SRQ  IEEEBool /* Service Request */
	ATN  IEEEBool /* Attention */
	REN  IEEEBool /* Remote Enable */
	IFC  IEEEBool /* Interface Clear */

	DIO Byte /* Data */
}

func (i IEEE488) Dump() {
	fmt.Printf("ATN: %s, NDAC: %s, NRFD: %s, DAV: %s, EOI: %s\n",
		i.ATN.ToOnOff(),
		i.NDAC.ToOnOff(),
		i.NRFD.ToOnOff(),
		i.DAV.ToOnOff(),
		i.EOI.ToOnOff())
	fmt.Printf("ATN: %s (%s/%dv)\n", i.ATN.ToOnOff(), i.ATN, i.ATN)
	fmt.Printf("NDAC: %s (%s/%dv)\n", i.NDAC.ToOnOff(), i.NDAC, i.NDAC)
	fmt.Printf("NRFD: %s (%s/%dv)\n", i.NRFD.ToOnOff(), i.NRFD, i.NRFD)
	fmt.Printf("DAV: %s (%s/%dv)\n", i.DAV.ToOnOff(), i.DAV, i.DAV)
}

/*
When all participants write 0 (false), the line will read back 0, but if any
device writes 1 (true), the bus will read back as 1.

0V = TRUE/1 (IEEE) = 0 (TTL Low)
5V = FALSE/0 (IEEE) = 1 (TTL High)
*/
func (i IEEE488) Or(o IEEE488) IEEE488 {
	var out IEEE488

	if i.DAV == TRUE || o.DAV == TRUE {
		out.DAV = TRUE
	} else {
		out.DAV = FALSE
	}
	if i.EOI == TRUE || o.EOI == TRUE {
		out.EOI = TRUE
	} else {
		out.EOI = FALSE
	}
	if i.NDAC == TRUE || o.NDAC == TRUE {
		out.NDAC = TRUE
	} else {
		out.NDAC = FALSE
	}
	if i.NRFD == TRUE || o.NRFD == TRUE {
		out.NRFD = TRUE
	} else {
		out.NRFD = FALSE
	}
	if i.SRQ == TRUE || o.SRQ == TRUE {
		out.SRQ = TRUE
	} else {
		out.SRQ = FALSE
	}
	if i.ATN == TRUE || o.ATN == TRUE {
		out.ATN = TRUE
	} else {
		out.ATN = FALSE
	}
	if i.REN == TRUE || o.REN == TRUE {
		out.REN = TRUE
	} else {
		out.REN = FALSE
	}
	if i.IFC == TRUE || o.IFC == TRUE {
		out.IFC = TRUE
	} else {
		out.IFC = FALSE
	}

	out.DIO = i.DIO | o.DIO

	return out
}

func MakeIEEE488() IEEE488 {
	return IEEE488{
		DAV:  FALSE,
		EOI:  FALSE,
		NDAC: FALSE,
		NRFD: FALSE,
		SRQ:  FALSE,
		ATN:  FALSE,
		REN:  FALSE,
		IFC:  FALSE,
	}
}

/*
Connectors provide methods to map the cable signals to physical devices on
either end
*/
type Connector interface {
	Read() IEEE488
	Write(IEEE488)
}

/*
Cables connect two Connectors at the A & B end
*/
type Cable struct {
	A Connector
	B Connector
}

func (c *Cable) Sync() {
	aEnd := c.A.Read()
	bEnd := c.B.Read()

	// OR the states together
	combined := aEnd.Or(bEnd)

	// Write the combined state to both ends
	c.A.Write(combined)
	c.B.Write(combined)
}
