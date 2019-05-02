package packer

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPacker tests the image packer
func TestPacker(t *testing.T) {
	p := New(DefaultConfig())

	var files []string
	err := filepath.Walk("./tests", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "jpg") && !strings.HasPrefix(info.Name(), "joined") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)

	for _, file := range files {
		t.Logf("Reading: %s", file)
		f, err := os.Open(file)
		require.NoError(t, err)
		defer f.Close()
		_, err = p.AddImageReader(f)
		require.NoError(t, err)
	}

	require.NoError(t, p.Pack())

	for i, img := range p.OutputImages {
		t.Logf("Writing image: %d", i)
		f, err := os.Create(filepath.Join("tests", fmt.Sprintf("joined_%d.jpg", i)))
		require.NoError(t, err)
		defer f.Close()
		require.NoError(t, jpeg.Encode(f, img, &jpeg.Options{Quality: 100}))
	}
}

// BenchmarkPacker banches the packer
func BenchmarkPacker(b *testing.B) {

	var images [][]byte

	var files []string
	err := filepath.Walk("./tests", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "jpg") && !strings.HasPrefix(info.Name(), "joined") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(b, err)

	for _, file := range files {
		f, err := os.Open(file)
		require.NoError(b, err)
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		require.NoError(b, err)

		images = append(images, data)
	}

	getPacker := func(b *testing.B) *Packer {
		p := New(DefaultConfig())
		for _, image := range images {
			_, err := p.AddImageBytes(image)
			require.NoError(b, err)
		}
		return p
	}

	b.Run("Growing", func(b *testing.B) {
		b.Run("Square", func(b *testing.B) {
			b.Run("Default", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					p := getPacker(b)

					p.cfg.AutoGrow = true

					b.ResetTimer()
					b.StartTimer()
					require.NoError(b, p.Pack())
					b.StopTimer()
				}
			})

			b.Run("Double", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					p := getPacker(b)
					p.cfg.AutoGrow = true
					p.cfg.TextureHeight *= 2
					p.cfg.TextureWidth *= 2

					b.ResetTimer()
					b.StartTimer()
					require.NoError(b, p.Pack())
					b.StopTimer()
				}
			})

			b.Run("Quad", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					p := getPacker(b)
					p.cfg.AutoGrow = true
					p.cfg.TextureHeight *= 4
					p.cfg.TextureWidth *= 4

					b.ResetTimer()
					b.StartTimer()
					require.NoError(b, p.Pack())
					b.StopTimer()
				}
			})

		})

		b.Run("NonSquare", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p := getPacker(b)
				p.cfg.Square = false
				p.cfg.AutoGrow = true

				b.StartTimer()
				require.NoError(b, p.Pack())
				b.StopTimer()
			}
		})

	})

	b.Run("1024x1024", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := getPacker(b)
			p.cfg.TextureHeight = 1024
			p.cfg.TextureWidth = 1024
			p.cfg.AutoGrow = false

			b.StartTimer()
			require.NoError(b, p.Pack())
			b.StopTimer()
		}
	})

	b.Run("4096x4096", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := getPacker(b)
			p.cfg.TextureHeight = 4096
			p.cfg.TextureWidth = 4096
			p.cfg.AutoGrow = false

			b.StartTimer()
			require.NoError(b, p.Pack())
			b.StopTimer()
		}
	})

	// b.Run("Pack", func(b *testing.B) {
	// 	p := getPacker(b)
	// 	// p.
	// })

	b.Run("Pack", func(b *testing.B) {

		b.Run("pack", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				p := getPacker(b)
				b.StartTimer()

				require.NoError(b, p.pack(p.cfg.Heuristic, p.cfg.TextureWidth, p.cfg.TextureHeight))
			}
		})

		b.Run("createBinImages", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p := getPacker(b)

				p.pack(p.cfg.Heuristic, p.cfg.TextureWidth, p.cfg.TextureHeight)

				require.NoError(b, p.createBinImages())

			}
		})
	})
}
