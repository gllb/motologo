package main

import (
	"os"
	"fmt"
	"errors"
	"bytes"
	"golang.org/x/image/bmp"
	"encoding/binary"
	"github.com/gllb/motologo/pkg/motologo"
	"github.com/gllb/motologo/pkg/motorun"
)

var MOTOLOGO_FILE_LIST = [...]string {"logo_boot", "logo_battery", "logo_unlocked",
	"logo_lowpower", "logo_charge", "ic_unlocked_bootloader",
	"arrows", "yes", "wififlash", "switchconsole", "qcom", "factory",
	"droid_operation", "bptools", "barcodes", "start", "restartbootloader",
	"recoverymode", "poweroff", "bootloaderlogs", "bg", "red_fastboot",
	"orange_continue", "yellow_continue", "Switch_tools_mode", "yellow",
	"orange", "red"}

var MOTOLOGO_FILE_LIST_LEN = len(MOTOLOGO_FILE_LIST)

type MotologoHeader struct { // 11 bytes
	Signature [9]byte
	ItemCount uint32
}

type Motologo struct { // unknown size (>1024 bytes)
	Header MotologoHeader
	MotorunMetaSet []MotorunMeta
	_ [115]byte // ??? 115 is arbitrary it only work with 28 MotorunMeta record
	MotorunSet []bytes.Buffer
}

type MotorunMeta struct { // 32 bytes
	Name [24]byte
	Offset uint32
	Size   uint32
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// paddingBlock return the number of byte needed to fit size in blockSize blocks
func paddingBlock(size int, blockSize int) int {
	if r := size % blockSize; r != 0 {
		return blockSize - r
	}
	return 0
}

// EncodeMotologo write the motologo structure to w
func EncodeMotologo(w io.Writer, m Motologo) error {
	if string(m.Header.Signature[:8]) != "MotoLogo" {
		return errors.New("motologo: invalid format")
	}
	if err := binary.Write(w, binary.LittleEndian, m.Header); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, m.MotorunMetaSet); err != nil {
		return err
	}

	headerPaddingSize := 1024 - (13 + len(m.MotorunMetaSet) * 32)

	if headerPaddingSize > 0 {
		headerPadding := bytes.Repeat([]byte{'\xff'}, headerPaddingSize)
		if err := binary.Write(w, binary.LittleEndian, headerPadding); err != nil {
			return err
		}
	} else {
		return errors.New("motologo: too many Motorun record")
	}

	for i, motorunBuf := range(m.MotorunSet) {

		motorunPaddingSize := paddingBlock(int(m.MotorunMetaSet[i].Size), 1024)

		if _, err := motorunBuf.WriteTo(w); err != nil {
			return err
		}

		if motorunPaddingSize > 0 {
			motorunPadding := bytes.Repeat([]byte{'\xff'}, motorunPaddingSize)
			if err := binary.Write(w, binary.LittleEndian, motorunPadding); err != nil {
				return err
			}
		}
	}

	return nil
}

// DecodeMotologoFile return a motologo structure according to the content of f
// func DecodeMotologoFile(f *os.File) (Motologo, error) {
// 	var motologo Motologo

// 	err := binary.Read(f, binary.LittleEndian, &motologo.Header)
// 	check(err)

// 	if string(motologo.Header.Signature[:8]) != "MotoLogo" {

// 		return Motologo{}, errors.New("motologo: invalid format")
// 	}

// 	motorunMetaCount := (motologo.Header.ItemCount - 0xd)/0x20
// 	motologo.MotorunMetaSet = make([]MotorunMeta, motorunMetaCount)
// 	motologo.MotorunSet = make([]bytes.Buffer, motorunMetaCount)

// 	err = binary.Read(f, binary.LittleEndian, &motologo.MotorunMetaSet)
// 	check(err)

// 	for i, motorunMeta := range motologo.MotorunMetaSet {
// 		_, err := f.Seek(int64(motorunMeta.Offset), 0)
// 		check(err)

// 		img, err := motorun.Decode(f)
// 		motorun.Encode(&motologo.MotorunSet[i], img)
// 	}

// 	return motologo, nil
// }

// Extract create BMP file from Motologo file archive in destDir.
// func Extract(m Motologo, destDir string) error {
// 	if err := os.MkdirAll(destDir, 0777); err != nil {
// 		return err
// 	}
// 	for i, motorunBuffer := range m.MotorunSet {
// 		name := string(m.MotorunMetaSet[i].Name[:])
// 		fmt.Println(name, m.MotorunSet[i].Len())
// 		file, err := os.Create(destDir + strings.Trim(name, "\x00") + ".bmp")
// 		if err != nil {
// 			return err
// 		}

// 		img, err := motorun.Decode(&motorunBuffer)
// 		if err != nil {
// 			return err
// 		}
// 		if err := bmp.Encode(file, img); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// Create a motologo structure from the content of srcDir
func Create(srcDir string) (Motologo, error) {
	var m Motologo

	m.Header.Signature = [9]byte{'M', 'o', 't', 'o', 'L', 'o', 'g', 'o', '\x00'}
	m.Header.ItemCount = uint32(MOTOLOGO_FILE_LIST_LEN * 0x20 + 0xd)

	m.MotorunMetaSet = make([]MotorunMeta, MOTOLOGO_FILE_LIST_LEN)
	m.MotorunSet = make([]bytes.Buffer, MOTOLOGO_FILE_LIST_LEN)

	offset := uint32(1024)
	for i, filename := range(MOTOLOGO_FILE_LIST) {
		f, err := os.Open(srcDir + filename + ".bmp")
		if err != nil {
			return Motologo{}, err
		}
		img, err := bmp.Decode(f)
		if err != nil {
			return Motologo{}, err
		}

		if err := motorun.Encode(&m.MotorunSet[i], img); err != nil {
			return Motologo{}, err
		}
		copy(m.MotorunMetaSet[i].Name[:], filename)

		m.MotorunMetaSet[i].Offset = offset

		fmt.Println(filename, m.MotorunSet[i].Len())
		m.MotorunMetaSet[i].Size = uint32(m.MotorunSet[i].Len())

		padding := paddingBlock(int(m.MotorunMetaSet[i].Size), 1024)
		offset += m.MotorunMetaSet[i].Size + uint32(padding)
	}

	return m, nil
}

func main() {
	f, err := os.Open("./test/logo_a.bin")
	check(err)

	fmt.Println("--------- Decoding motologo file ./test/logo_a.bin")
	m, err := DecodeMotologoFile(f)
	check(err)

	err = f.Close()
	check(err)

	fmt.Println("--------- Extracting motologo in ./tmp/")
	err = Extract(m, "./tmp/")
	check(err)

	fmt.Println("--------- Creating motologo from ./tmp/")
	m, err = Create("./tmp/")
	check(err)

	f, err = os.Create("./test/testEncodeMotologo_logo_a.bin")
	check(err)

	fmt.Println("--------- Encoding motologo ./test/testEncodeMotologo_logo_a.bin")
	err = EncodeMotologo(f, m)
	check(err)
}
