package motologo

import (
	"os"
	"errors"
	"bytes"
	"encoding/binary"
	"golang.org/x/image/bmp"
	"github.com/gllb/motologo/pkg/motorun"
)
func paddingBlock(size int, blockSize int) int {
	if r := size % blockSize; r != 0 {
		return blockSize - r
	}
	return 0
}

// EncodeMotologo write the motologo structure to f
func EncodeMotologo(f *os.File, m Motologo) error {
	if string(m.Header.Signature[:8]) != "MotoLogo" {
		return errors.New("motologo: invalid format")
	}
	if err := binary.Write(f, binary.LittleEndian, m.Header); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, m.MotorunMetaSet); err != nil {
		return err
	}

	headerPaddingSize := 1024 - (13 + len(m.MotorunMetaSet) * 32)

	if headerPaddingSize > 0 {
		headerPadding := bytes.Repeat([]byte{'\xff'}, headerPaddingSize)
		if err := binary.Write(f, binary.LittleEndian, headerPadding); err != nil {
			return err
		}
	} else {
		return errors.New("motologo: too many Motorun record")
	}

	for i, motorunBuf := range(m.MotorunSet) {

		motorunPaddingSize := paddingBlock(int(m.MotorunMetaSet[i].Size), 1024)

		if _, err := motorunBuf.WriteTo(f); err != nil {
			return err
		}

		if motorunPaddingSize > 0 {
			motorunPadding := bytes.Repeat([]byte{'\xff'}, motorunPaddingSize)
			if err := binary.Write(f, binary.LittleEndian, motorunPadding); err != nil {
				return err
			}
		}
	}

	return nil
}

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

		m.MotorunMetaSet[i].Size = uint32(m.MotorunSet[i].Len())

		padding := paddingBlock(int(m.MotorunMetaSet[i].Size), 1024)
		offset += m.MotorunMetaSet[i].Size + uint32(padding)
	}

	return m, nil
}
