package packer

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type file struct {
	path string
	enc  ImgEncoding
}

// TestPacker tests the image packer
func TestPacker(t *testing.T) {
	p := New(DefaultConfig())

	var files []file
	err := filepath.Walk("./tests", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "jpg") && !strings.HasPrefix(info.Name(), "joined") {
			files = append(files, file{path: path, enc: JPEG})
		}
		return nil
	})
	require.NoError(t, err)

	for _, file := range files {
		t.Logf("Reading: %s", file.path)
		f, err := os.Open(file.path)
		require.NoError(t, err)
		defer f.Close()
		require.NoError(t, p.AddImage(f, file.enc))
		// i--
		// if i == 0 {
		// break//
		// }
	}

	images := p.PackedImages()
	for i, img := range images {
		t.Logf("Writing image: %d", i)
		f, err := os.Create(filepath.Join("tests", fmt.Sprintf("joined_%d.jpg", i)))
		require.NoError(t, err)
		defer f.Close()
		require.NoError(t, jpeg.Encode(f, img, &jpeg.Options{Quality: 100}))
	}
}
