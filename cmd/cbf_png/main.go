package main

import (
	"flag"
	"fmt"

	"cbf2go/internal/cbf"
)

func main() {
	fin := flag.String("fin", "", "CBF file")
	fout := flag.String("fout", "", "output file")
	flag.Parse()

	if *fin == "" || *fout == "" {
		panic("No input or output file name is provided")
	}

	pixels, w, h, err := cbf.ReadCBF(*fin)
	if err != nil {
		panic(err)
	}

	err = cbf.WritePNG(pixels, w, h, *fout)

	if err != nil {
		panic(err)
	}

	fmt.Println("created:", *fout)
}
