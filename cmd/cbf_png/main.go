package main

import (
	"flag"
	"fmt"

	"cbf2go/internal/cbf"
)

func main() {
	var fin, fout string
	var verbose int
	flag.StringVar(&fin, "fin", "", "CBF file")
	flag.StringVar(&fout, "fout", "", "output file")
	flag.IntVar(&verbose, "verbose", 0, "verbose level")
	flag.Parse()

	if fin == "" || fout == "" {
		panic("No input or output file name is provided")
	}

	pixels, w, h, err := cbf.ReadCBF(fin, verbose)
	if err != nil {
		panic(err)
	}

	err = cbf.WritePNG(pixels, w, h, fout)

	if err != nil {
		panic(err)
	}

	fmt.Println("created:", fout)
}
