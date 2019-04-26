package packer

import (
	"image"
)

func (p *Packer) crop(img image.Image) image.Rectangle {

	var j, w, h, x, y int
	// var pix color.RGBA
	var breakB bool

	var t = true
	cmpf1 := func(x, y, a int) bool {
		temp := img.At(x, y)
		_, _, _, tempA := temp.RGBA()
		return int(tempA) > a
	}
	cmp := func(x, y, a int) {
		if cmpf1(x, y, a) {
			t = false
			breakB = true
		} else {
			if !t {
				breakB = true
			}
		}
	}

	// cmpf2 := func(x, y int) bool {
	// 	temp := img.At(x, y)
	// 	r, g, b, a := temp.RGBA()
	// 	r1, g1, b1, a1 := pix.RGBA()
	// 	return r != r1 || g != g1 || b != b1 || a != a1
	// }

	// cmp2 := func(x, y int) {
	// 	if cmpf2(x, y) {
	// 		t = false
	// 		breakB = true
	// 	} else {
	// 		if !t {
	// 			breakB = true
	// 		}
	// 	}
	// }

yl:
	for y = 0; y < img.Bounds().Dy(); y++ {
		for j := 0; j < img.Bounds().Dx(); j++ {
			cmp(j, y, p.cropThreshold)
			if breakB {
				breakB = false
				break yl
			}
		}
	}

	t = true
xl:
	for x = 0; x < img.Bounds().Dx(); x++ {
		for j := y; j < img.Bounds().Dy(); j++ {
			cmp(x, j, p.cropThreshold)
			if breakB {
				breakB = false
				break xl
			}
		}
	}

	t = true
wl:
	for w = img.Bounds().Dx(); w > 0; w-- {
		for j = y; j < img.Bounds().Dy(); j++ {
			cmp(w-1, j, p.cropThreshold)
			if breakB {
				breakB = false
				break wl
			}
		}
	}

	t = true
hl:
	for h = img.Bounds().Dy(); h > 0; h-- {
		for j = x; j < w; j++ {
			cmp(j, h-1, p.cropThreshold)
			if breakB {
				breakB = false
				break hl
			}
		}
	}

	w = w - x
	h = h - y
	if w < 0 {
		w = 0
	}

	if h < 0 {
		h = 0
	}
	// b := img.Bounds()
	// fmt.Printf("Original: %d, %d, %d, %d\n", b.Min.X, b.Min.Y, b.Max.X, b.Max.Y)
	// fmt.Printf("Cropped: %d, %d, %d, %d\n", x, y, w, h)

	return image.Rectangle{image.Pt(x, y), image.Pt(x+w, y+h)}

}
