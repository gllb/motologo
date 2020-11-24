package motorun

import (
	"os"
	"testing"
	"golang.org/x/image/bmp"
)

func TestDecodeMotorun(t *testing.T) {
	f, err := os.Open("./test/logo_battery.motorun")
	check(err)

	img, err := Decode(f)
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
