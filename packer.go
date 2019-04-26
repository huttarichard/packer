package packer

import (
	"github.com/disintegration/imaging"
	"hash/crc64"
	"image"
	"image/color"
	"image/draw"
	"sort"
	"sync"
)

const (
	mininputImageSizeX = 32
	mininputImageSizeY = 32
)

// Packer is the image 2d bin packer
type Packer struct {
	images *images
	Root   *inputImage

	cfg                 *Config
	compare             int
	area, neededArea    int64
	missingImages       int
	mergedImages        int
	Ltr, merge, mergeBF bool
	MinFillRate         int
	cropThreshold       int
	Rotate              Rotation
	border              border

	bins []image.Rectangle

	nextID int

	table *crc64.Table

	lock sync.Mutex
}

// New creates the new Packer
func New(cfg *Config) *Packer {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Packer{
		cfg:    cfg,
		images: &images{sortOrder: cfg.SortOrder},
		table:  crc64.MakeTable(crc64.ECMA),
		border: border{l: cfg.Border, r: cfg.Border},
	}

	return p
}

// PackedImages gets the packed images
func (p *Packer) PackedImages() []draw.Image {
	p.pack(p.cfg.Heuristic, p.cfg.TextureWidth, p.cfg.TextureHeight)

	textures := []draw.Image{}

	for _, bin := range p.bins {
		// fmt.Printf("Creating bin: %s", bin)
		texture := image.NewRGBA(bin)
		for x := 0; x < bin.Dx(); x++ {
			for y := 0; y < bin.Dy(); y++ {
				texture.SetRGBA(x, y, color.RGBA{A: uint8(0)})
			}
		}
		textures = append(textures, texture)
	}

	// for j, texture := range textures {
	// 	for i, img := range p.images.inputImages {
	// 		if img.textureID != j {
	// 			continue
	// 		}
	// 		pos := image.Pt(img.pos.X+p.border.l+p.cfg.Extrude, img.pos.Y+p.border.t+p.cfg.Extrude)
	// 		var size, sizeOrig, crop image.Rectangle

	// 		sizeOrig = img.size
	// 		if p.cfg.CropThreshold == 0 {
	// 			size = img.size
	// 			crop = image.Rect(0, 0, size.Dx(), size.Dy())
	// 		} else {
	// 			size = img.crop
	// 			crop = img.crop
	// 		}

	// 		if img.rotated {
	// 			//transpose size
	// 			size.Max = image.Pt(size.Max.Y, size.Max.X)
	// 			crop = size
	// 		}

	// 	}
	// }

	for _, img := range p.images.inputImages {
		if img.duplicatedID != nil && p.cfg.Merge {
			continue
		}

		pos := image.Pt(img.pos.X+p.border.l, img.pos.Y+p.border.t)
		var size, crop image.Rectangle

		if p.cfg.CropThreshold == 0 {
			size = img.size
			crop = image.Rect(0, 0, size.Dx(), size.Dy())
		} else {
			size = img.crop
			crop = img.crop
		}

		if img.rotated {
			img.Image = imaging.Rotate90(img.Image)
			size.Max = image.Pt(size.Max.Y, size.Max.X)
			min := image.Pt(img.size.Dy()-crop.Min.Y-crop.Dy(), crop.Min.X)
			max := image.Pt(min.X+crop.Dy(), min.Y+crop.Dx())
			crop = image.Rectangle{min, max}
		}

		if img.textureID < len(p.bins) {
			// fmt.Printf("TextureID: %d\n", img.textureID)
			texture := textures[img.textureID]
			if p.cfg.Extrude != 0 {
				color1 := img.Image.At(crop.Min.X, crop.Min.Y)
				if p.cfg.Extrude == 1 {
					texture.Set(pos.X, pos.Y, color1)
				} else {
					m := image.NewRGBA(image.Rect(0, 0, p.cfg.Extrude-1-pos.X, p.cfg.Extrude-1-pos.Y))
					draw.Draw(m, m.Bounds(), &image.Uniform{color1}, image.ZP, draw.Src)
					draw.Draw(texture, image.Rect(pos.X, pos.Y, pos.X+p.cfg.Extrude-1, pos.Y+p.cfg.Extrude-1), m, pos, draw.Src)
				}

				color2 := img.Image.At(crop.Min.X, crop.Min.Y+crop.Max.Y-1)
				if p.cfg.Extrude == 1 {
					texture.Set(pos.X, pos.Y, color2)
				} else {
					m := image.NewRGBA(image.Rect(0, 0, p.cfg.Extrude-1-pos.X, p.cfg.Extrude-1-pos.Y))
					draw.Draw(m, m.Bounds(), &image.Uniform{color2}, image.ZP, draw.Src)
					draw.Draw(texture, image.Rect(pos.X, pos.Y, pos.X+p.cfg.Extrude-1, pos.Y+p.cfg.Extrude-1), m, pos, draw.Src)
				}
			} else {
				// fmt.Printf("Drawing at: %s, %s\n", pos, crop)

				draw.Draw(texture, image.Rectangle{image.Pt(pos.X, pos.Y), image.Pt(pos.X+img.crop.Dx(), pos.Y+img.crop.Dy())}, img.Image, image.ZP, draw.Src)
			}
		}

	}
	return textures
}

// Pack packs the images with provided heuristic
func (p *Packer) pack(heur Heuristic, w, h int) {

	p.sortImages(w, h)

	p.missingImages = 1
	p.mergedImages = 0
	p.area = 0

	p.bins = []image.Rectangle{}

	areaBuf := p.addImagesToBins(heur, w, h)

	// fmt.Printf("Bins: %d\n", len(p.bins))
	if areaBuf != 0 && p.missingImages == 0 {
		p.cropLastImage(heur, w, h, false)
	}

	if p.cfg.Merge {
		for _, text := range p.images.inputImages {
			if text.duplicatedID != nil {
				dup := p.find(*text.duplicatedID)
				text.pos = dup.pos
				text.textureID = dup.textureID
				p.mergedImages++
			}
		}
	}

	return
}

// sortImages sorts the images
func (p *Packer) sortImages(w, h int) {
	p.recalculateDuplicates()

	p.neededArea = 0
	var size image.Rectangle

	for _, texture := range p.images.inputImages {

		texture.pos = image.Pt(999999, 999999)
		if p.cropThreshold != 0 {
			size = texture.crop
		} else {
			size = texture.size
		}

		if size.Dx() == w {
			size.Max = image.Pt(size.Max.X-p.border.l-p.border.r-2*p.cfg.Extrude, size.Max.Y)
		}
		if size.Dy() == h {
			size.Max = image.Pt(size.Max.X, size.Dy()-p.border.t-p.border.b-2*p.cfg.Extrude)
		}

		size.Max = image.Pt(size.Dx()+p.border.t+p.border.b+2*p.cfg.Extrude, size.Dy()+p.border.l+p.border.r+2*p.cfg.Extrude)

		if p.Rotate == RWidthGreaterHeight && size.Dx() > size.Dy() ||
			p.Rotate == RWidthGreater2Height && size.Dx() > 2*size.Dy() ||
			p.Rotate == RHeightGreaterWidth && size.Dy() > size.Dx() ||
			p.Rotate == RH2WidthH && size.Dy() > size.Dx() && size.Dx()*2 > size.Dy() ||
			p.Rotate == RW2HeightW && size.Dx() > size.Dy() && size.Dy()*2 > size.Dx() ||
			p.Rotate == RHeightGreater2Width && size.Dy() > 2*size.Dx() {
			size.Max = image.Pt(size.Max.Y, size.Max.X)
			texture.rotated = true
		}

		texture.sizeCurrent = size
		if texture.duplicatedID == nil || !p.cfg.Merge {
			p.neededArea += int64(size.Dx() * size.Dy())
		}
	}
	sort.Sort(p.images)
}

func (p *Packer) addImagesToBins(heur Heuristic, w, h int) int {
	binIndex := len(p.bins) - 1
	var areaBuf int
	var lastAreaBuf int

	for {
		p.missingImages = 0
		p.bins = append(p.bins, image.Rect(0, 0, w, h))
		binIndex++
		lastAreaBuf = p.fillBin(heur, w, h, binIndex)
		if lastAreaBuf == 0 {
			// fmt.Printf("LastAreaBuf == 0\n")
			p.bins = p.bins[:len(p.bins)-1]
		}
		areaBuf += lastAreaBuf

		if !(p.missingImages != 0 && lastAreaBuf != 0) {
			break
		}
	}

	return areaBuf
}

func (p *Packer) cropLastImage(heur Heuristic, w, h int, wh bool) {
	p.missingImages = 0
	lastImages := p.images.inputImages
	lastBins := p.bins
	lastArea := p.area

	p.bins = p.bins[:len(p.bins)-1]
	p.clearBin(len(p.bins))

	if p.cfg.Square {
		w /= 2
		h /= 2
	} else {
		if wh {
			w /= 2
		} else {
			h /= 2
		}
		wh = !wh
	}

	binIndex := len(p.bins)
	p.missingImages = 0
	p.bins = append(p.bins, image.Rect(0, 0, w, h))
	p.fillBin(heur, w, h, binIndex)
	if p.missingImages != 0 {
		p.images.inputImages = lastImages
		p.bins = lastBins
		p.area = lastArea
		p.missingImages = 0
		if p.cfg.Square {
			w *= 2
			h *= 2
		} else {
			if !wh {
				w *= 2
			} else {
				h *= 2
			}
			wh = !wh
		}

		if p.cfg.Autosize {
			rate := p.getFillRate()
			if rate < float64(p.MinFillRate) && (w > mininputImageSizeX && h > mininputImageSizeY) {
				p.divideLastImage(heur, w, h, wh)
				if p.getFillRate() <= rate {
					p.images.inputImages = lastImages
					p.bins = lastBins
					p.area = lastArea
				}
			}
		}
	} else {
		p.cropLastImage(heur, w, h, wh)
	}
}

func (p *Packer) updateCrop() {
	for _, t := range p.images.inputImages {
		t.crop = p.crop(t.Image)
	}
}

func (p *Packer) divideLastImage(heur Heuristic, w, h int, wh bool) {
	p.missingImages = 0
	lastImages := p.images.inputImages
	lastBins := p.bins
	lastArea := p.area

	p.bins = p.bins[:len(p.bins)]
	p.clearBin(len(p.bins))

	if p.cfg.Square {
		w /= 2
		h /= 2
	} else {
		if wh {
			w /= 2
		} else {
			h /= 2
		}
		wh = !wh
	}
	p.addImagesToBins(heur, w, h)
	if p.missingImages != 0 {
		p.images.inputImages = lastImages
		p.bins = lastBins
		p.area = lastArea
		p.missingImages = 0
	} else {
		p.cropLastImage(heur, w, h, wh)
	}
}

func (p *Packer) getFillRate() float64 {
	var binArea int64
	for _, bin := range p.bins {
		binArea += int64(bin.Dx() * bin.Dy())
	}
	return float64(p.area) / float64(binArea)
}

// fillBin fills the bin
func (p *Packer) fillBin(heur Heuristic, w, h, binIndex int) int {
	var (
		areaBuf int
		rects   = &maxRects{}
		mrn     = &maxRectsNode{}
	)
	mrn.r = image.Rect(0, 0, w, h)
	// fmt.Printf("Creating bin of size: %d, %d", w, h)
	rects.f = append(rects.f, mrn)
	rects.Heur = heur
	rects.leftToRight = p.Ltr
	rects.w = w
	rects.h = h
	rects.Rot = p.Rotate
	rects.border = &p.border

	for _, text := range p.images.inputImages {

		if !text.pos.Eq(image.Pt(999999, 999999)) {
			continue
		}
		// fmt.Printf("Adding image: %x to bin: %d\n", text.hash, binIndex)

		if text.duplicatedID == nil || !p.cfg.Merge {
			// fmt.Println("Inserting node")
			text.pos = rects.insertNode(text)
			text.textureID = binIndex
			// fmt.Printf("Pos: %v", text.pos)
			if text.pos.Eq(image.Pt(999999, 999999)) {
				p.missingImages++
			} else {

				areaBuf += text.sizeCurrent.Dx() * text.sizeCurrent.Dy()
				// fmt.Printf("Areabuf: %d\n", areaBuf)
				p.area += int64(text.sizeCurrent.Dx() * text.sizeCurrent.Dy())
			}
		}

	}

	return areaBuf
}

// clearBin clears the current image at index
func (p *Packer) clearBin(binIndex int) {
	for _, text := range p.images.inputImages {
		if text.textureID == binIndex {
			p.area -= int64(text.sizeCurrent.Dx() * text.sizeCurrent.Dy())
			text.pos = image.Pt(999999, 999999)
		}
	}
}

func (p *Packer) recalculateDuplicates() {
	for _, texture := range p.images.inputImages {
		texture.duplicatedID = nil
	}

	for i, texture := range p.images.inputImages {
		for k := i + 1; k < len(p.images.inputImages); k++ {
			textureK := p.images.inputImages[k]
			if textureK.duplicatedID == nil &&
				texture.hash == textureK.hash &&
				texture.size.Eq(textureK.size) &&
				texture.crop.Eq(textureK.size) {
				textureK.duplicatedID = &texture.id
			}
		}
	}
	return
}

func (p *Packer) removeID(id int) {
	var at int
	for i, texture := range p.images.inputImages {
		if texture.id == id {
			at = i
			break
		}
	}

	p.images.inputImages = append(p.images.inputImages[:at], p.images.inputImages[at+1:]...)
}

func (p *Packer) find(id int) *inputImage {
	for _, texture := range p.images.inputImages {
		if texture.id == id {
			return texture
		}
	}
	return nil
}

func compareImageByHeight(i, j image.Rectangle) bool {
	return (i.Dy()<<10)+i.Dx() > (j.Dy()<<10)+j.Dx()
}

func compareImageByWidth(i, j image.Rectangle) bool {
	return (i.Dx()<<10)+i.Dy() > (j.Dx()<<10)+j.Dy()
}

func compareImageByArea(i, j image.Rectangle) bool {
	return i.Dy()*i.Dx() > j.Dy()*j.Dx()
}

func compareImageByMax(i, j image.Rectangle) bool {
	var first, second int

	if i.Dy() > i.Dx() {
		first = i.Dy()
	} else {
		first = i.Dx()
	}

	if j.Dy() > j.Dx() {
		second = j.Dy()
	} else {
		second = j.Dx()
	}
	if first == second {
		return compareImageByArea(i, j)
	}
	return first > second
}

type border struct {
	t, b, l, r int
}

// getID gets the nextID
func (p *Packer) getID() int {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.nextID++
	return p.nextID
}