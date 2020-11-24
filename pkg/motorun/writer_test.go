package motorun

import (
	"os"
	"golang.org/x/image/bmp"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func TestEncode(t *testing.T) {
	f, err := os.Open("./test/logo_battery.bmp")
	check(err)

	img, err := bmp.Decode(f)
	check(err)

	out, err := os.Create("./test/testEncode_logo_battery.motorun")
	check(err)

	err = Encode(out, img)
	check(err)

	err = f.Close()
	check(err)

	err = out.Close()
	check(err)
}
