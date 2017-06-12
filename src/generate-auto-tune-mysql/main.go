package main

// pull in the package andrew mentioned that deals with memory
// read command line for destination cnf
// call object.generate with mem library and destination string

import (
	sigar "github.com/cloudfoundry/gosigar"
	"fmt"
	"flag"
	"os"
)

var (
	targetPercentage float64
	outputFile string
)

func main() {
	flag.Float64Var(&targetPercentage, "P", 50.0,
			"Set this to an integer which represents the percentage of system RAM to reserve for InnoDB's buffer pool")
	flag.StringVar(&outputFile, "f", "",
		       "Target file for rendering MySQL option file")
	flag.Parse()

	mem := sigar.Mem{}
	mem.Get()
	totalMem := mem.Total
	fmt.Printf("Total memory in bytes: %d", mem.Total)

	file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	Generate(totalMem, targetPercentage, file)
}
