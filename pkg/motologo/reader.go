package motologo

import (
	"os"
	"errors"
	"strings"
	"bytes"
	"golang.org/x/image/bmp"
	"encoding/binary"
	"github.com/gllb/motologo/pkg/motorun"
)

// DecodeMotologoFile return a motologo structure according to the content of f
func DecodeMotologoFile(f *os.File) (Motologo, error) {
	var motologo Motologo

	if err := binary.Read(f, binary.LittleEndian, &motologo.Header); err != nil {
		return Motologo{}, err
	}

	if string(motologo.Header.Signature[:8]) != "MotoLogo" {

		return Motologo{}, errors.New("motologo: invalid format")
	}

	motorunMetaCount := (motologo.Header.ItemCount - 0xd)/0x20
	motologo.MotorunMetaSet = make([]MotorunMeta, motorunMetaCount)
	motologo.MotorunSet = make([]bytes.Buffer, motorunMetaCount)

	if err := binary.Read(f, binary.LittleEndian, &motologo.MotorunMetaSet); err != nil {
		return Motologo{}, err
	}

	for i, motorunMeta := range motologo.MotorunMetaSet {
		if _, err := f.Seek(int64(motorunMeta.Offset), 0); err != nil {
			return Motologo{}, err
		}

		img, err := motorun.Decode(f)
		if err != nil {
			return Motologo{}, err
		}
		motorun.Encode(&motologo.MotorunSet[i], img)
	}

	return motologo, nil
}

// Extract create BMP file from Motologo file archive in destDir.
func Extract(m Motologo, destDir string) error {
	if err := os.MkdirAll(destDir, 0777); err != nil {
		return err
	}
	for i, motorunBuffer := range m.MotorunSet {
		name := string(m.MotorunMetaSet[i].Name[:])

		file, err := os.Create(destDir + strings.Trim(name, "\x00") + ".bmp")
		if err != nil {
			return err
		}

		img, err := motorun.Decode(&motorunBuffer)
		if err != nil {
			return err
		}
		if err := bmp.Encode(file, img); err != nil {
			return err
		}
	}
	return nil
}
