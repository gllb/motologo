package main

import (
	"os"
	"image"
	"errors"
	"golang.org/x/image/bmp"
	"encoding/binary"
	"le-blc.com/motologo/motorun"
)

type MotologoHeader struct { // 11 bits
	Signature [9]byte
	ItemCount uint16
}

type Motologo struct { // unknown size (>1024 bits)
	Header MotologoHeader
	_ [2]byte
	MotorunMetaSet []MotorunMeta
	_ [115]byte // ??? 115 is arbitrary it only work with 28 MotorunMeta record
	MotorunSet []image.Image
}

type MotorunMeta struct { // 32 bits
	Name [24]byte
	Offset uint32
	Size   uint32
}

func check(e error) {
	if e != nil {
		//log.Fatal(e.Error())
		panic(e)
	}
}

func parseMotologoFile(f *os.File) (Motologo, error) {
	var motologo Motologo

	err := binary.Read(f, binary.LittleEndian, &motologo.Header)
	check(err)

	if string(motologo.Header.Signature[:]) != "MotoLogo" {
		return Motologo{}, errors.New("motologo: invalid format")
	}

	motorunMetaCount := (motologo.Header.ItemCount - 0xd)/0x20
	motologo.MotorunMetaSet = make([]MotorunMeta, motorunMetaCount)

	// skip 2 bytes garbage
	_, err = f.Seek(2, 1)
	check(err)

	err = binary.Read(f, binary.LittleEndian, &motologo.MotorunMetaSet)
	check(err)

	for _, motorunMeta := range motologo.MotorunMetaSet {
		_, err := f.Seek(int64(motorunMeta.Offset), 0)
		check(err)

		motorun, err := motorun.Decode(f)

		motologo.MotorunSet = append(motologo.MotorunSet, motorun)
	}

	return motologo, nil
}

func testDecode() {
	f, err := os.Open("./test/logo_battery.motorun")
	check(err)

	img, err := motorun.Decode(f)
	check(err)

	out, err := os.Create("./test/testDecode_logo_battery.bmp")
	check(err)

	err = bmp.Encode(out, img)
	check(err)

	err = f.Close()
	check(err)

	err = out.Close()
	check(err)
}

func testEncode() {
	f, err := os.Open("./test/logo_battery.bmp")
	check(err)

	img, err := bmp.Decode(f)
	check(err)

	out, err := os.Create("./test/testEncode_logo_battery.motorun")
	check(err)

	err = motorun.Encode(out, img)
	check(err)

	err = f.Close()
	check(err)

	err = out.Close()
	check(err)
}

func main() {
	f, err := os.Open("./test/logo_a.bin")
	check(err)
	parseMotologoFile(f)
	err = f.Close()
	check(err)
}
