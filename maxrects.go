package packer

import (
	"image"
	"math"
)

type maxRectsNode struct {
	r image.Rectangle
	b trb1 //border
}

func (m *maxRectsNode) String() string {
	return m.r.String()
}

type trb1 struct {
	t, r, b, l image.Point
}

type maxRects struct {
	f  []*maxRectsNode
	r  []image.Rectangle
	fr []*maxRectsNode

	Heur        Heuristic
	w, h        int
	Rot         Rotation
	leftToRight bool
	border      *border
}

func (mr *maxRects) insertNode(input *InputImage) image.Point {

	var i, m int
	minV := math.MaxInt32
	mini := -1

	img := input.sizeCurrent

	// fmt.Printf("Image: %s\n", img)
	if img.Dx() == 0 || img.Dy() == 0 {
		return image.Pt(0, 0)
	}

	var leftNeighbor, rightNeighbor, _leftNeighbor, _rightNeighbor, rotated, bestIsRotated bool

	var f *maxRectsNode
	// fmt.Printf("Mr.F: %v\n", mr.f)
	for i = 0; i < len(mr.f); i++ {

		f = mr.f[i]
		if (f.r.Dx() >= img.Dx() && f.r.Dy() >= img.Dy()) ||
			(f.r.Dx() >= img.Dy() && f.r.Dy() >= img.Dx()) {

			rotated = false
			m = 0

			if (f.r.Dx() >= img.Dy() && f.r.Dy() >= img.Dx()) &&
				!(f.r.Dx() >= img.Dx() && f.r.Dy() >= img.Dy()) {

				if mr.Rot == 0 {
					continue
				}

				img.Max = image.Pt(img.Max.Y, img.Max.X)
				rotated = true
				m += img.Dy()
			}

			// fmt.Println("SwitchHeu")
			switch mr.Heur {
			case HNone:
				mini = i
				i = len(mr.f)
				continue
			case HTl:

				m += f.r.Min.Y

				_leftNeighbor, _rightNeighbor = false, false

				for _, r := range mr.r {

					if math.Abs(float64(r.Min.Y)+float64(r.Dy())/2-float64(f.r.Min.Y)-float64(f.r.Dy())/2) <
						math.Max(float64(r.Dy()), float64(f.r.Dy()/2)) {

						if r.Min.X+r.Dx() == f.r.Min.X {
							m -= 5
							_leftNeighbor = true
						}

						if r.Min.X == f.r.Min.X+f.r.Dx() {
							m -= 5
							_rightNeighbor = true
						}
					}

				}

				if _leftNeighbor || !_rightNeighbor {
					if f.r.Min.X+f.r.Dx() == mr.w {
						// fmt.Println("First")
						m--
						_rightNeighbor = true
					}
					if f.r.Min.X == 0 {
						// fmt.Println("Second")
						m--
						_leftNeighbor = true
					}
				}
			case HBaf:
				m += f.r.Dx() * f.r.Dy()
			case HBssf:
				m += min(f.r.Dx()-img.Dx(), f.r.Dy()-img.Dy())
			case HBlsf:
				m += max(f.r.Dx()-img.Dx(), f.r.Dy()-img.Dy())
			case HMinw:
				m += f.r.Dx()
			case HMinh:
				m += f.r.Dy()
			}

			// fmt.Printf("M: %d\n", m)
			if m < minV {
				minV = m
				mini = i
				leftNeighbor = _leftNeighbor
				rightNeighbor = _rightNeighbor
				bestIsRotated = rotated
			}
			if rotated {
				img.Max = image.Pt(img.Max.Y, img.Max.X)
			}
		}
	}
	if bestIsRotated {
		img.Max = image.Pt(img.Max.Y, img.Max.X)
		input.rotated = !input.rotated
		input.sizeCurrent.Max = image.Pt(input.sizeCurrent.Max.Y, input.sizeCurrent.Max.X)
	}

	// fmt.Printf("Mini: %d\n", mini)
	if mini >= 0 {
		i = mini
		var n0 maxRectsNode
		min := image.Pt(mr.f[i].r.Min.X, mr.f[i].r.Min.Y)
		max := image.Pt(min.X+img.Dx(), min.Y+img.Dy())
		buf := image.Rectangle{min, max}

		if mr.Heur == HTl {

			if !leftNeighbor && mr.f[i].r.Min.X != 0 &&
				mr.f[i].r.Dx()+mr.f[i].r.Min.X == mr.w {

				min := image.Pt(mr.w-img.Dx(), mr.f[i].r.Min.Y)
				max := image.Pt(min.X+img.Dx(), min.Y+img.Dy())
				buf = image.Rectangle{min, max}
			}

			if !leftNeighbor && rightNeighbor {
				min := image.Pt(mr.f[i].r.Min.X+mr.f[i].r.Dx()-img.Dx(), mr.f[i].r.Min.Y)
				max := image.Pt(min.X+img.Dx(), min.Y+img.Dy())
				buf = image.Rectangle{min, max}
			}
		}

		// fmt.Printf("Buf: %s\n\n", buf)
		n0.r = buf
		mr.r = append(mr.r, buf)

		if mr.f[i].r.Dx() > img.Dx() {

			n := &maxRectsNode{}

			var x0 int
			if buf.Min.X == mr.f[i].r.Min.X {
				x0 = img.Dx()
			}

			min := image.Pt(mr.f[i].r.Min.X+x0, mr.f[i].r.Min.Y)
			max := image.Pt(min.X+mr.f[i].r.Dx()-img.Dx(), min.Y+mr.f[i].r.Dy())

			n.r = image.Rectangle{min, max}

			mr.f = append(mr.f, n)
		}

		if mr.f[i].r.Dy() > img.Dy() {

			n := &maxRectsNode{}

			min := image.Pt(mr.f[i].r.Min.X, mr.f[i].r.Min.Y+img.Dy())
			max := image.Pt(min.X+mr.f[i].r.Dx(), min.Y+mr.f[i].r.Dy()-img.Dy())
			n.r = image.Rectangle{min, max}

			mr.f = append(mr.f, n)
		}

		mr.f = append(mr.f[:i], mr.f[i+1:]...)

		for i = 0; i < len(mr.f); i++ {

			f := mr.f[i]

			if f.r.Overlaps(n0.r) {

				if n0.r.Min.X+n0.r.Dx() < f.r.Min.X+f.r.Dx() {
					// fmt.Println("1")
					n := &maxRectsNode{}

					min := image.Pt(n0.r.Dx()+n0.r.Min.X, f.r.Dy())
					max := image.Pt(min.X+f.r.Dx()+f.r.Min.X-n0.r.Dx()-n0.r.Min.X, min.Y+f.r.Dy())
					n.r = image.Rectangle{min, max}

					mr.f = append(mr.f, n)
				}

				if n0.r.Min.Y+n0.r.Dy() < f.r.Min.Y+f.r.Dy() {
					// fmt.Println("2")
					n := &maxRectsNode{}

					min := image.Pt(n0.r.Min.X, n0.r.Min.Y+n0.r.Dy())
					max := image.Pt(min.X+f.r.Dx(), min.Y+f.r.Dy()+f.r.Min.Y-n0.r.Dy()-n0.r.Min.Y)
					n.r = image.Rectangle{min, max}
					mr.f = append(mr.f, n)
				}

				if n0.r.Min.X > f.r.Min.X {
					// fmt.Println("3")
					n := &maxRectsNode{}
					min := image.Pt(f.r.Min.X, f.r.Min.Y)
					max := image.Pt(min.X+n0.r.Min.X-f.r.Min.X, min.Y+f.r.Dy())
					n.r = image.Rectangle{min, max}
					mr.f = append(mr.f, n)
				}

				if n0.r.Min.Y > f.r.Min.Y {
					// fmt.Println("4")
					n := &maxRectsNode{}
					min := image.Pt(f.r.Min.X, f.r.Min.Y)
					max := image.Pt(min.X+f.r.Dx(), min.Y+n0.r.Min.Y-f.r.Min.Y)
					n.r = image.Rectangle{min, max}
					mr.f = append(mr.f, n)
				}

				mr.f = append(mr.f[:i], mr.f[i+1:]...)
				i--
			}

		}

		for i = 0; i < len(mr.f); i++ {
			for j := i + 1; j < len(mr.f); j++ {
				if i != j && mr.f[j].r.In(mr.f[i].r) {
					// fmt.Printf("Removing last: %d\n", i)
					mr.f = append(mr.f[:j], mr.f[j+1:]...)
					j--
				}
			}
		}

		return n0.r.Min
	}

	return image.Pt(999999, 999999)
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}

func min(i, j int) int {
	if i < j {
		return i
	}
	return j
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}
