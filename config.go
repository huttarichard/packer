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
	AutoGrow          bool
	Autosize          bool
	CropThreshold     int
	AutoSizeThreshold int
	MinTextureSizeX   int
	MinTextureSizeY   int
	Heuristic         Heuristic
}

// DefaultConfig returns the default config for the packer
func DefaultConfig() *Config {
	return &Config{
		TextureHeight:     512,
		TextureWidth:      512,
		Merge:             true,
		Crop:              true,
		Border:            0,
		Extrude:           0,
		Rotate:            false,
		Square:            true,
		AutoGrow:          true,
		Autosize:          true,
		CropThreshold:     1,
		AutoSizeThreshold: 100,
		SortOrder:         OrderByMax,
		MinTextureSizeX:   32,
		MinTextureSizeY:   32,
		Heuristic:         HBaf,
	}
}
