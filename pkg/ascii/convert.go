package ascii

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/disintegration/imaging"
)

// Default ASCII ramps
const asciiChars = " .,:;i1tfLCG08@"
const asciiCharsInverted = "@80GCLft1i;:,. "

// ConvertConfig mirrors your Rust struct
type ConvertConfig struct {
	// Scale factor for output (0.01–1.0)
	Resolution float64
	// Contrast adjustment (0.1–3.0)
	Contrast float64
	// Brightness adjustment (0.1–3.0)
	Brightness float64
	// Invert the character mapping
	Inverted bool
	// Use colored output
	Colored bool
	// Dithering algorithm to use
	Dithering DitheringStrategy
	// Character set to use
	Charset CharSet
	// Custom charter ramp (if Charset is Custom)
	CustomRamp string
}

func DefaultConfig() ConvertConfig {
	return ConvertConfig{
		Resolution: 0.2,
		Contrast:   1.0,
		Brightness: 1.0,
		Inverted:   false,
		Colored:    true,
		Dithering:  DitheringNone,
		Charset:    CharSetPhoto,
	}
}

var (
	ErrInvalidResolution = errors.New("resolution must be in [0.01, 1.0]")
	ErrInvalidContrast   = errors.New("contrast must be in [0.1, 3.0]")
	ErrInvalidBrightness = errors.New("brightness must be in [0.1, 3.0]")
	ErrImageTooSmall     = errors.New("image too small after scaling")
)

func (c ConvertConfig) Validate() error {
	if c.Resolution < 0.01 || c.Resolution > 1.0 {
		return fmt.Errorf("%w: %f", ErrInvalidResolution, c.Resolution)
	}
	if c.Contrast < 0.1 || c.Contrast > 3.0 {
		return fmt.Errorf("%w: %f", ErrInvalidContrast, c.Contrast)
	}
	if c.Brightness < 0.1 || c.Brightness > 3.0 {
		return fmt.Errorf("%w: %f", ErrInvalidBrightness, c.Brightness)
	}
	return nil
}

func (c ConvertConfig) ramps() (normal, inverted string) {
	if strings.TrimSpace(c.CustomRamp) != "" {
		r := c.CustomRamp
		runes := []rune(r)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return r, string(runes)
	}

	switch c.Charset {
	case CharSetPhoto:
		return asciiPhoto, asciiPhotoInv
	case CharSetMinimal:
		return asciiMinimal, asciiMinimalInv
	case CharSetBlocks:
		return asciiBlocks, asciiBlocksInv
	case CharSetClassic:
		fallthrough
	default:
		return asciiClassic, asciiClassicInv
	}
}

type AsciiResult struct {
	Width, Height int
	Chars         []rune
	Colors        []color.NRGBA
	Colored       bool
}

func (r *AsciiResult) index(x, y int) int {
	return y*r.Width + x
}

func (r *AsciiResult) ToPlainText() string {
	var b strings.Builder
	b.Grow(r.Width*r.Height + r.Height)

	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			ch := r.Chars[r.index(x, y)]
			b.WriteRune(ch)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func (r *AsciiResult) ToMarkdown() string {
	var b strings.Builder

	b.WriteString("```text\n")
	b.WriteString(r.ToPlainText())
	if !strings.HasSuffix(b.String(), "\n") {
		b.WriteString("\n")
	}
	b.WriteString("```\n")

	return b.String()
}

func (r *AsciiResult) ToMarkdownColored() string {
	var b strings.Builder

	escape := func(s string) string {
		s = strings.ReplaceAll(s, "&", "&amp;")
		s = strings.ReplaceAll(s, "<", "&lt;")
		s = strings.ReplaceAll(s, ">", "&gt;")
		return s
	}

	b.WriteString(`<pre style="font-family: 'Courier New', monospace; font-size: 8px; line-height: 1; letter-spacing: 0.1em; background-color: #0f0f1a; color: #ffffff; padding: 20px; border-radius: 8px;">` + "\n")

	if r.Colored {
		for y := 0; y < r.Height; y++ {
			for x := 0; x < r.Width; x++ {
				i := y*r.Width + x
				ch := string(r.Chars[i])
				col := r.Colors[i]
				charEsc := escape(ch)
				fmt.Fprintf(&b,
					`<span style="color:rgb(%d,%d,%d)">%s</span>`,
					col.R, col.G, col.B, charEsc,
				)
			}
			b.WriteString("<br/>\n")
		}
	} else {
		for y := 0; y < r.Height; y++ {
			for x := 0; x < r.Width; x++ {
				ch := string(r.Chars[y*r.Width+x])
				b.WriteString(escape(ch))
			}
			b.WriteString("<br/>\n")
		}
	}

	b.WriteString("</pre>\n")
	return b.String()
}

func (r *AsciiResult) ToANSI() string {
	if !r.Colored {
		return r.ToPlainText()
	}

	var b strings.Builder

	for y := 0; y < r.Height; y++ {
		for x := 0; x < r.Width; x++ {
			i := r.index(x, y)
			ch := r.Chars[i]
			col := r.Colors[i]
			fmt.Fprintf(&b, "\x1b[38;2;%d;%d;%dm%s",
				col.R, col.G, col.B, string(ch))
		}
		b.WriteByte('\n')
	}
	b.WriteString("\x1b[0m")
	return b.String()
}

func (r *AsciiResult) ToHTML() string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>ASCII Art</title>
<style>
body {
  background-color: #1a1a2e;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  margin: 0;
  padding: 20px;
  box-sizing: border-box;
}
pre {
  font-family: 'Courier New', Courier, monospace;
  font-size: 9px;
  line-height: 1.45;
  letter-spacing: 0.05em;
  white-space: pre;
  background-color: #0f0f1a;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0,0,0,0.5);
}
</style>
</head>
<body>
<pre>`)

	escape := func(s string) string {
		s = strings.ReplaceAll(s, "&", "&amp;")
		s = strings.ReplaceAll(s, "<", "&lt;")
		s = strings.ReplaceAll(s, ">", "&gt;")
		return s
	}

	if r.Colored {
		for y := 0; y < r.Height; y++ {
			for x := 0; x < r.Width; x++ {
				i := r.index(x, y)
				ch := string(r.Chars[i])
				col := r.Colors[i]
				b.WriteString(fmt.Sprintf(
					`<span style="color:rgb(%d,%d,%d)">%s</span>`,
					col.R, col.G, col.B, escape(ch),
				))
			}
			b.WriteString("\n")
		}
	} else {
		for y := 0; y < r.Height; y++ {
			for x := 0; x < r.Width; x++ {
				ch := string(r.Chars[r.index(x, y)])
				b.WriteString(escape(ch))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString(`</pre>
</body>
</html>`)

	return b.String()
}

func getBrightness(r, g, b uint8) float64 {
	return 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
}

func adjustPixel(value, contrast, brightness float64) float64 {
	adjusted := (value-128.0)*contrast + 128.0
	adjusted = adjusted * brightness
	if adjusted < 0 {
		return 0
	}
	if adjusted > 255 {
		return 255
	}
	return adjusted
}

func ConvertImage(img image.Image, cfg ConvertConfig) (*AsciiResult, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	b := img.Bounds()
	origW, origH := b.Dx(), b.Dy()

	newW := int(float64(origW) * cfg.Resolution)
	newH := int(float64(origH) * cfg.Resolution * 0.5) // chars are ~2:1 height:width

	if newW < 1 || newH < 1 {
		return nil, ErrImageTooSmall
	}

	// Lanczos resize (like Rust)
	resized := imaging.Resize(img, newW, newH, imaging.Lanczos)
	rgbImg := imaging.Clone(resized) // ensure concrete type

	grayscale := make([]float64, 0, newW*newH)
	colors := make([]color.NRGBA, 0, newW*newH)

	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			c := color.NRGBAModel.Convert(rgbImg.At(x, y)).(color.NRGBA)

			r := adjustPixel(float64(c.R), cfg.Contrast, cfg.Brightness)
			g := adjustPixel(float64(c.G), cfg.Contrast, cfg.Brightness)
			b := adjustPixel(float64(c.B), cfg.Contrast, cfg.Brightness)

			gray := getBrightness(uint8(r), uint8(g), uint8(b))
			grayscale = append(grayscale, gray)
			colors = append(colors, color.NRGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: 255,
			})
		}
	}

	normalRamp, invertedRamp := cfg.ramps()

	var charsRamp string
	if cfg.Inverted {
		charsRamp = invertedRamp
	} else {
		charsRamp = normalRamp
	}
	ramp := []rune(charsRamp)
	levels := len(ramp)

	cfg.Dithering.Apply(grayscale, newW, newH, levels)

	asciiChars := make([]rune, len(grayscale))
	for i, v := range grayscale {
		idx := int(math.Round((v / 255.0) * float64(levels-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= levels {
			idx = levels - 1
		}
		asciiChars[i] = ramp[idx]
	}

	return &AsciiResult{
		Width:   newW,
		Height:  newH,
		Chars:   asciiChars,
		Colors:  colors,
		Colored: cfg.Colored,
	}, nil
}
