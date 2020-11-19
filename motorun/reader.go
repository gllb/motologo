package motorun

import (
	"io"
	"image"
	"io/ioutil"
	"image/color"
	"errors"
)

func readUint16(b []byte) uint16 {
	return uint16(b[1]) | uint16(b[0])<<8
}

// Decode reads a Motorun image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	c, err := decodeConfig(r)
	if err != nil {
		return nil, err
	}
	rgba := image.NewRGBA(image.Rect(0, 0, c.Width, c.Height))
	if c.Width == 0 || c.Height == 0 {
		return rgba, nil
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	y := 0
	x := 0
	for y < c.Height {
		pixelCount := int(readUint16(b))
		b = b[2:]
		repeatFlag := (pixelCount & 0x8000) == 0x8000
		pixelCount &= 0x7fff
		if repeatFlag {
			blue := b[0]
			green := b[1]
			red := b[2]
			b = b[3:]
			for i := 0; i < pixelCount; i++ {
				rgba.Set(x, y, color.RGBA{red, blue, green, 255})
				x += 1
				if x != c.Width {
					continue
				}
				x = 0
				y += 1
				if y == c.Height {
					break
				}
			}
		} else {
			for i := 0; i < pixelCount; i++ {
				blue := b[0]
				green := b[1]
				red := b[2]
				b = b[3:]
				rgba.Set(x, y, color.RGBA{red, blue, green, 255})
				x += 1
				if x != c.Width {
					continue
				}
				x = 0
				y += 1
				if y == c.Height {
					break
				}
			}
		}

	}
	return rgba, nil
}

// DecodeConfig returns the color model and dimensions of motorun image without
// decoding the entire image.
// Limitation: Color model can only be RGBA.
func decodeConfig(r io.Reader) (config image.Config, err error) {
	var b [12]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return image.Config{}, err
	}
	if string(b[:7]) != "MotoRun" {
		return image.Config{}, errors.New("motorun: invalid format")
	}
	width := readUint16(b[8:10])
	height := readUint16(b[10:12])

	return image.Config{ColorModel: color.RGBAModel, Width: int(width), Height: int(height)}, nil
}
