package main

import (
	"os"
	"fmt"
	"github.com/gllb/motologo/pkg/motologo"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	f, err := os.Open("./test/logo_a.bin")
	check(err)

	fmt.Println("--------- Decoding motologo file ./test/logo_a.bin")
	m, err := motologo.DecodeMotologoFile(f)
	check(err)

	err = f.Close()
	check(err)

	fmt.Println("--------- Extracting motologo in ./tmp/")
	err = motologo.Extract(m, "./tmp/")
	check(err)

	fmt.Println("--------- Creating motologo from ./tmp/")
	m, err = motologo.Create("./tmp/")
	check(err)

	f, err = os.Create("./test/testEncodeMotologo_logo_a.bin")
	check(err)

	fmt.Println("--------- Encoding motologo ./test/testEncodeMotologo_logo_a.bin")
	err = motologo.EncodeMotologo(f, m)
	check(err)
}
