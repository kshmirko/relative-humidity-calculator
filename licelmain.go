package main

import (
	"fmt"

	"github.com/kshmirko/licel-go/licel/licelformat"
)

func main() {

	//a := licelformat.NewLicelProfile("1 0 1 16380 1 0000 7.50 00353.o 0 0 00 000 12 002001 0.100 BT1")
	a := licelformat.NewLicelPack("b*.*")
	//a := licelformat.LoadLicelFile("b2021019.223500")
	v := licelformat.SelectCertainWavelength2(&a, true, 408)
	fmt.Println(v)
	for key := range a {
		fmt.Println(key)
	}
}
