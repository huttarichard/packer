package packer

import (
	"bytes"
	"errors"
	"golang.org/x/image/bmp"
	"hash/crc64"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

// inputImage is the image wrapper that defines the position of the
type inputImage struct {
	Image draw.Image

	hash      uint64
	textureID int

	id                int
	duplicatedID      *int
	Name              string
	pos               image.Point
	size, sizeCurrent image.Rectangle
	crop              image.Rectangle

	cropped, rotated bool
}

type images struct {
	inputImages []*inputImage
	sortOrder   SortOrder
}

func (im *images) Less(i, j int) bool {
	switch im.sortOrder {
	case OrderByWidth:
		return compareImageByWidth(im.inputImages[i].Image.Bounds(), im.inputImages[j].Image.Bounds())
	case OrderByHeight:
		return compareImageByHeight(im.inputImages[i].Image.Bounds(), im.inputImages[j].Image.Bounds())
	case OrderByArea:
		return compareImageByArea(im.inputImages[i].Image.Bounds(), im.inputImages[j].Image.Bounds())
	case OrderByMax:
		return compareImageByMax(im.inputImages[i].Image.Bounds(), im.inputImages[j].Image.Bounds())
	}
	return false
}

func (im *images) Len() int {
	return len(im.inputImages)
}

func (im *images) Swap(i, j int) {
	im.inputImages[i], im.inputImages[j] = im.inputImages[j], im.inputImages[i]
}

// ImgEncoding is the image encoding enum
type ImgEncoding int

const (
	_ ImgEncoding = iota
	JPEG
	PNG
	BMP
	GIF
)

var (
	// ErrUnknownEncoding is an error that is thrown when the image is with unsupported encoding
	ErrUnknownEncoding = errors.New("Unknown image encoding provided")

	// ErrEmptyImage is an error thrown when the provided image is empty
	ErrEmptyImage = errors.New("Provided empty image")
)

// AddImageBytes add the image in the form of raw bytes
func (p *Packer) AddImageBytes(data []byte, enc ImgEncoding) error {
	return p.addImageBytes(data, enc)
}

// AddImage adds the image read from the reader
func (p *Packer) AddImage(r io.Reader, enc ImgEncoding) error {
	return p.addImage(r, enc)
}

// AddImage creates the new texture for the provided
func (p *Packer) addImageBytes(data []byte, enc ImgEncoding) error {
	buf := bytes.NewBuffer(data)
	return p.addImage(buf, enc)
}

func (p *Packer) addImage(r io.Reader, enc ImgEncoding) error {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(r, buf)
	t := &inputImage{}

	var (
		img image.Image
		err error
	)

	switch enc {
	case JPEG:
		img, err = jpeg.Decode(tee)
	case PNG:
		img, err = png.Decode(tee)
	case BMP:
		img, err = bmp.Decode(tee)
	case GIF:
		img, err = gif.Decode(tee)
	default:
		return ErrUnknownEncoding
	}
	if err != nil {
		return err
	}

	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		return ErrEmptyImage
	}

	dImg, ok := img.(draw.Image)
	if !ok {
		dImg = image.NewRGBA(img.Bounds())
		draw.Draw(dImg, dImg.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	t.Image = dImg
	t.hash = crc64.Checksum(buf.Bytes(), p.table)
	t.id = p.getID()
	t.size = dImg.Bounds()
	t.crop = p.crop(dImg)

	p.images.inputImages = append(p.images.inputImages, t)

	return nil
}
