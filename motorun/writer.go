package motorun

import (
	"encoding/binary"
	"image"
	"io"
	"image/color"
	"bytes"
	"errors"
)

type header struct {
	sigMotorun [8]byte
	width uint16
	height uint16
}

func encodeRow(w io.Writer, row []color.Color) error {
	var stream bytes.Buffer
	j := 0

	for j < len(row) {
		k := j
		for k < len(row) && row[j] == row[k] {
			k++
		}
		if (k - j) > 1 {
			stream.WriteByte(byte((0x80 | (k - j) >> 8)))
			stream.WriteByte(byte((k - j) & 0xFF))
			r, g, b, _ := row[j].RGBA()
			stream.WriteByte(byte(b))
			stream.WriteByte(byte(g))
			stream.WriteByte(byte(r))
			j = k
		} else {
			l := k
			var m int
			for ((l - k) < 3) && ((m - l) < 2) {
				k = l - 1
				for (l < len(row)) && (row[k] != row[l]) {
					k++
					l++
				}
				for (l < len(row)) && (row[k] == row[l]) {
					l++
				}
				if l == len(row) {
					break
				}
				m = l
				for (m < len(row)) && (row[l] == row[m]) {
					m++
				}
			}
			if (k - j) == 0 {
				stream.WriteByte(0x0)
				stream.WriteByte(0x1)
				r, g, b, _ := row[len(row) - 1].RGBA()
				stream.WriteByte(byte(b))
				stream.WriteByte(byte(g))
				stream.WriteByte(byte(r))
				break
			}
			if (k == (len(row) - 1)) {
				k++
			}

			stream.WriteByte(byte((k - j) >> 8))
			stream.WriteByte(byte((k - j) & 0xFF))
			for l = 0; l < (k - j); l++ {
				r, g, b, _ := row[j + l].RGBA()
				stream.WriteByte(byte(b))
				stream.WriteByte(byte(g))
				stream.WriteByte(byte(r))
			}
			j = k
		}
	}
	buf := stream.Bytes()
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

// Encode writes the image m to w in MotoRun format.
func Encode(w io.Writer, m image.Image) error {
	d := m.Bounds().Size()
	if d.X < 0 || d.Y < 0 {
		return errors.New("motorun: negative bounds")
	}
	h := &header{
		sigMotorun: [8]byte{'M', 'o', 't', 'o', 'R', 'u', 'n', '\x00'},
		width: uint16(d.X),
		height: uint16(d.Y),
	}

	if err := binary.Write(w, binary.BigEndian, h); err != nil {
		return err
	}

	row := make([]color.Color, d.Y)
	for y := 0; y < d.Y; y++ {
		for x := 0; x < d.X; x++ {
			row[x] = m.At(x, y)
		}
		if err := encodeRow(w, row); err != nil {
			return err
		}
	}
	return nil
}
