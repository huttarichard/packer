package packer

// SortOrder is the enum that defines sorting order for the packer
type SortOrder int

const (
	OrderNone SortOrder = iota
	OrderByWidth
	OrderByHeight
	OrderByArea
	OrderByMax
)

// Heuristic defines the enum for the heuristic
type Heuristic int

const (
	HNone Heuristic = iota
	HTl
	HBaf
	HBssf
	HBlsf
	HMinw
	HMinh
)

// Rotation defines the enums for the rotation
type Rotation int

const (
	RNever Rotation = iota
	ROnlyWhenNeeded
	RH2WidthH
	RWidthGreaterHeight
	RWidthGreater2Height
	RW2HeightW
	RHeightGreaterWidth
	RHeightGreater2Width
)
