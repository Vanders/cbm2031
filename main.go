package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/vanders/pet/mos6502"
)

type (
	Byte = mos6502.Byte
	Word = mos6502.Word
)

func sep() {
	fmt.Println("========================================")
}

func main() {
	var (
		writer io.Writer
	)
	debug := flag.Bool("d", false, "enable CPU dissasembly")
	flag.Parse()

	if *debug {
		writer = os.Stderr
	}

	// Dummy 'A' end connector
	aConnector := &DummyConnector{}
	aConnector.Reset()
	monitor := NewMonitor(aConnector)

	// Create a new CBM2031
	cbm2031 := NewCBM2031(writer)

	// Connect a cable to the "B" end
	bConnector := cbm2031.CreateConnector()

	// Create a cable & connect the 'A' & 'B' ends
	cable := &Cable{
		A: aConnector,
		B: bConnector,
	}
	cbm2031.Cable = cable
	monitor.BVIA = cbm2031.VIA
	monitor.BRAM = cbm2031.RAM

	go cbm2031.Run()
	monitor.Run()
}
