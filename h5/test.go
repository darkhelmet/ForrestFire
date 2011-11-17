package main

import (
	"h5"
	"fmt"
	"flag"
	"os"
)

func main() {
	flag.Parse()
	n := 0
	rdr, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Println("Error: ", err)
		n = 1
	}
	p := h5.NewParser(rdr)
	err = p.Parse()
	if err != nil {
		fmt.Println("Error: ", err)
		n = 1
	}
	fmt.Println(p.Top)
	os.Exit(n)
}
