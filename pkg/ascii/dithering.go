package ascii

import (
	"math"
)

type DitheringStrategy int

const (
	DitheringNone DitheringStrategy = iota
	DitheringFloydSteinberg
	DitheringAtkinson
	DitheringRiemersma
	DitheringOrdered2x2
	DitheringOrdered4x4
	DitheringThreshold
)

func (d DitheringStrategy) Apply(gray []float64, width, height, levels int) {
	switch d {
	case DitheringNone:
		return
	case DitheringFloydSteinberg:
		floydSteinberg(gray, width, height, levels)
	case DitheringAtkinson:
		atkinson(gray, width, height, levels)
	case DitheringRiemersma:
		riemersma(gray, width, height, levels)
	case DitheringOrdered2x2:
		ordered2x2(gray, width, height, levels)
	case DitheringOrdered4x4:
		ordered4x4(gray, width, height, levels)
	case DitheringThreshold:
		thresholdDither(gray, width, height, levels)
	}
}

func floydSteinberg(img []float64, width, height, levels int) {
	scale := 255.0 / float64(levels-1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*width + x

			oldPixel := img[i]
			newPixel := math.Round(oldPixel/scale) * scale
			err := oldPixel - newPixel

			img[i] = newPixel

			if x+1 < width {
				img[y*width+(x+1)] += err * 7.0 / 16.0
			}
			if y+1 < height {
				if x > 0 {
					img[(y+1)*width+(x-1)] += err * 3.0 / 16.0
				}
				img[(y+1)*width+x] += err * 5.0 / 16.0
				if x+1 < width {
					img[(y+1)*width+(x+1)] += err * 1.0 / 16.0
				}
			}
		}
	}

	for i := range img {
		if img[i] < 0 {
			img[i] = 0
		} else if img[i] > 255 {
			img[i] = 255
		}
	}
}

func atkinson(img []float64, width, height, levels int) {
	scale := 255.0 / float64(levels-1)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := y*width + x

			oldPixel := img[i]
			newPixel := math.Round(oldPixel/scale) * scale
			err := oldPixel - newPixel
			errFrac := err / 8.0

			img[i] = newPixel

			if x+1 < width {
				img[y*width+(x+1)] += errFrac
			}
			if x+2 < width {
				img[y*width+(x+2)] += errFrac
			}
			if y+1 < height {
				if x > 0 {
					img[(y+1)*width+(x-1)] += errFrac
				}
				img[(y+1)*width+x] += errFrac
				if x+1 < width {
					img[(y+1)*width+(x+1)] += errFrac
				}
			}
			if y+2 < height {
				img[(y+2)*width+x] += errFrac
			}
		}
	}

	for i := range img {
		if img[i] < 0 {
			img[i] = 0
		} else if img[i] > 255 {
			img[i] = 255
		}
	}
}

func riemersma(img []float64, width, height, levels int) {
	scale := 255.0 / float64(levels-1)
	total := width * height
	visited := make([]bool, total)
	var err float64

	directions := [][2]int{
		{0, 1},
		{1, 0},
		{0, -1},
		{-1, 0},
		{1, 1},
		{1, -1},
		{-1, -1},
		{-1, 1},
	}

	row, col := 0, 0
	visited[0] = true

	for range img {
		idx := row*width + col
		oldPixel := img[idx] + err
		newPixel := math.Round(oldPixel/scale) * scale
		err = oldPixel - newPixel
		img[idx] = newPixel

		for _, d := range directions {
			nr := row + d[0]
			nc := col + d[1]
			if nr >= 0 && nr < height && nc >= 0 && nc < width {
				nIdx := nr*width + nc
				if !visited[nIdx] {
					row, col = nr, nc
					visited[nIdx] = true
					break
				}
			}
		}
	}

	for i := range img {
		if img[i] < 0 {
			img[i] = 0
		} else if img[i] > 255 {
			img[i] = 255
		}
	}
}

func ordered2x2(img []float64, w, h, levels int) {
	matrix := [2][2]float64{
		{0.0 / 4.0, 2.0 / 4.0},
		{3.0 / 4.0, 1.0 / 4.0},
	}

	scale := 255.0 / float64(levels-1)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x

			old := img[i]
			threshold := (matrix[y%2][x%2] - 0.5) * scale
			val := old + threshold

			newVal := math.Round(val/scale) * scale
			if newVal < 0 {
				newVal = 0
			}
			if newVal > 255 {
				newVal = 255
			}
			img[i] = newVal
		}
	}
}

func ordered4x4(img []float64, w, h, levels int) {
	matrix := [4][4]float64{
		{0, 8, 2, 10},
		{12, 4, 14, 6},
		{3, 11, 1, 9},
		{15, 7, 13, 5},
	}
	for y := range matrix {
		for x := range matrix[y] {
			matrix[y][x] = (matrix[y][x] + 0.5) / 16.0
		}
	}

	scale := 255.0 / float64(levels-1)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			old := img[i]

			threshold := (matrix[y%4][x%4] - 0.5) * scale
			val := old + threshold

			newVal := math.Round(val/scale) * scale
			if newVal < 0 {
				newVal = 0
			}
			if newVal > 255 {
				newVal = 255
			}
			img[i] = newVal
		}
	}
}

func thresholdDither(img []float64, w, h, levels int) {
	scale := 255.0 / float64(levels-1)
	for i := range img {
		val := img[i]
		newVal := math.Round(val/scale) * scale
		if newVal < 0 {
			newVal = 0
		}
		if newVal > 255 {
			newVal = 255
		}
		img[i] = newVal
	}
}
