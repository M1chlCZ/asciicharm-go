package tui

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/M1chlCZ/asciicharm-go/pkg/ascii"
	"github.com/disintegration/imaging"
)

func ditherName(d ascii.DitheringStrategy) string {
	switch d {
	case ascii.DitheringNone:
		return "None"
	case ascii.DitheringFloydSteinberg:
		return "FS"
	case ascii.DitheringAtkinson:
		return "Atkinson"
	case ascii.DitheringRiemersma:
		return "Riemersma"
	case ascii.DitheringOrdered2x2:
		return "Ord2x2"
	case ascii.DitheringOrdered4x4:
		return "Ord4x4"
	case ascii.DitheringThreshold:
		return "Thresh"
	default:
		return "?"
	}
}

func cycleDither(d ascii.DitheringStrategy) ascii.DitheringStrategy {
	switch d {
	case ascii.DitheringNone:
		return ascii.DitheringFloydSteinberg
	case ascii.DitheringFloydSteinberg:
		return ascii.DitheringAtkinson
	case ascii.DitheringAtkinson:
		return ascii.DitheringRiemersma
	case ascii.DitheringRiemersma:
		return ascii.DitheringOrdered2x2
	case ascii.DitheringOrdered2x2:
		return ascii.DitheringOrdered4x4
	case ascii.DitheringOrdered4x4:
		return ascii.DitheringThreshold
	case ascii.DitheringThreshold:
		fallthrough
	default:
		return ascii.DitheringNone
	}
}

func charsetName(cs ascii.CharSet) string {
	switch cs {
	case ascii.CharSetClassic:
		return "Classic"
	case ascii.CharSetPhoto:
		return "Photo"
	case ascii.CharSetMinimal:
		return "Minimal"
	case ascii.CharSetBlocks:
		return "Blocks"
	default:
		return "?"
	}
}

func LoadImage(path string) (image.Image, error) {
	img, err := imaging.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image: %w", err)
	}
	return img, nil
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func ListImageFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if isImageFile(e.Name()) {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}
