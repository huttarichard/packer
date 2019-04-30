package packer

import (
	"bytes"
	"errors"
	"hash/crc64"
	"image"
	"image/draw"
	"image/jpeg"
	"io"
)

// InputImage is the image wrapper that defines the position of the
type InputImage struct {
	image draw.Image

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

// PackedPosition gets the position of the image within the packed image
func (i *InputImage) PackedPosition() image.Point {
	return i.pos
}

// Hash gets the image hash value
func (i *InputImage) Hash() uint64 {
	return i.hash
}

// TextureID gets the target image (output) texture id
func (i *InputImage) TextureID() int {
	return i.textureID
}

type images struct {
	inputImages []*InputImage
	sortOrder   SortOrder
}

func (im *images) Less(i, j int) bool {
	switch im.sortOrder {
	case OrderByWidth:
		return compareImageByWidth(im.inputImages[i].image.Bounds(), im.inputImages[j].image.Bounds())
	case OrderByHeight:
		return compareImageByHeight(im.inputImages[i].image.Bounds(), im.inputImages[j].image.Bounds())
	case OrderByArea:
		return compareImageByArea(im.inputImages[i].image.Bounds(), im.inputImages[j].image.Bounds())
	case OrderByMax:
		return compareImageByMax(im.inputImages[i].image.Bounds(), im.inputImages[j].image.Bounds())
	}
	return false
}

func (im *images) Len() int {
	return len(im.inputImages)
}

func (im *images) Swap(i, j int) {
	im.inputImages[i], im.inputImages[j] = im.inputImages[j], im.inputImages[i]
}

var (
	// ErrUnknownEncoding is an error that is thrown when the image is with unsupported encoding
	ErrUnknownEncoding = errors.New("Unknown image encoding provided")

	// ErrEmptyImage is an error thrown when the provided image is empty
	ErrEmptyImage = errors.New("Provided empty image")
)

// AddImageBytes add the image in the form of raw bytes
func (p *Packer) AddImageBytes(data []byte) error {
	return p.addImageBytes(data)
}

// AddImageReader adds the image from the reader
func (p *Packer) AddImageReader(r io.Reader) error {
	return p.addImage(r)
}

// AddImage adds the image with the hash provided
func (p *Packer) AddImage(img image.Image, hash ...uint64) error {

	var h uint64
	if len(hash) == 0 || (len(hash) > 0 && hash[0] == 0) {

		buf := &bytes.Buffer{}
		if err := jpeg.Encode(buf, img, nil); err != nil {
			return err
		}
		h = crc64.Checksum(buf.Bytes(), p.table)
		buf.Reset()
	} else {
		h = hash[0]
	}
	return p.getInputImageData(img, h)
}

// AddImage creates the new texture for the provided
func (p *Packer) addImageBytes(data []byte) error {
	buf := bytes.NewBuffer(data)
	return p.addImage(buf)
}

func (p *Packer) addImage(r io.Reader) error {
	buf := &bytes.Buffer{}
	tee := io.TeeReader(r, buf)

	img, _, err := image.Decode(tee)
	if err != nil {
		return err
	}

	hash := crc64.Checksum(buf.Bytes(), p.table)

	return p.getInputImageData(img, hash)
}

func (p *Packer) getInputImageData(img image.Image, hash uint64) error {
	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		return ErrEmptyImage
	}

	dImg, ok := img.(draw.Image)
	if !ok {
		dImg = image.NewRGBA(img.Bounds())
		draw.Draw(dImg, dImg.Bounds(), img, img.Bounds().Min, draw.Src)
	}
	t := &InputImage{}
	t.image = dImg
	t.hash = hash
	t.id = p.getID()
	t.size = dImg.Bounds()
	t.crop = p.crop(dImg)

	p.images.inputImages = append(p.images.inputImages, t)

	return nil
}
