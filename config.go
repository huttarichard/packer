package packer

// Config is the packer configuration
type Config struct {
	SortOrder         SortOrder
	TextureHeight     int
	TextureWidth      int
	Merge             bool
	Crop              bool
	Square            bool
	Rotate            bool
	Border            int
	Extrude           int
	Recursion         bool
	Autosize          bool
	CropThreshold     int
	AutoSizeThreshold int
	MinTextureSizeX   int
	MinTextureSizeY   int
	Heuristic         Heuristic
	OutputFormat      ImgEncoding
}

// DefaultConfig returns the default config for the packer
func DefaultConfig() *Config {
	return &Config{
		TextureHeight:     16384,
		TextureWidth:      16384,
		Merge:             true,
		Crop:              true,
		Border:            0,
		Extrude:           0,
		Rotate:            false,
		Recursion:         true,
		Square:            true,
		Autosize:          true,
		CropThreshold:     1,
		AutoSizeThreshold: 100,
		SortOrder:         OrderByMax,
		MinTextureSizeX:   32,
		MinTextureSizeY:   32,
		Heuristic:         HBaf,
		OutputFormat:      PNG,
	}
}
