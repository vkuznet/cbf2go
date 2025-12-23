package cbf

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"
)

func viridis(t float64) color.RGBA {
	// clamp
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}

	// polynomial approximation of matplotlib viridis
	r := 0.2803 + t*(0.2331+t*(0.0916+t*(-0.1766)))
	g := 0.0739 + t*(1.1436+t*(-0.5182+t*(-0.0416)))
	b := 0.2923 + t*(0.5600+t*(0.1200+t*(-0.1766)))

	return color.RGBA{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
		A: 255,
	}
}

func WritePNGColor(pixels []int32, w, h int, outPath string) error {
	if len(pixels) != w*h {
		return fmt.Errorf("pixel count mismatch: %d vs %d", len(pixels), w*h)
	}

	// ------------------------------------------------------------
	// Convert to float64
	// ------------------------------------------------------------
	n := len(pixels)
	vals := make([]float64, n)
	for i, v := range pixels {
		if v < 0 {
			v = 0 // defensive (CBF can have masked negatives)
		}
		vals[i] = float64(v)
	}

	// ------------------------------------------------------------
	// Percentile clipping
	// ------------------------------------------------------------
	lo := percentile(vals, 0.5)
	hi := percentile(vals, 99.5)

	if hi <= lo {
		return fmt.Errorf("invalid clip range: lo=%f hi=%f", lo, hi)
	}

	scale := 1.0 / (hi - lo)

	// ------------------------------------------------------------
	// Create RGBA image
	// ------------------------------------------------------------
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			v := vals[i]

			if v < lo {
				v = lo
			}
			if v > hi {
				v = hi
			}

			t := (v - lo) * scale // 0..1
			img.SetRGBA(x, y, viridis(t))
		}
	}

	// ------------------------------------------------------------
	// Write PNG
	// ------------------------------------------------------------
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func WritePNG(pixels []int32, w, h int, outPath string) error {
	if len(pixels) != w*h {
		return fmt.Errorf("pixel count mismatch: %d vs %d", len(pixels), w*h)
	}

	// ------------------------------------------------------------
	// Convert to float64 for statistics
	// ------------------------------------------------------------
	n := len(pixels)
	vals := make([]float64, n)
	for i, v := range pixels {
		vals[i] = float64(v)
	}

	// ------------------------------------------------------------
	// Percentile clipping (match Python: 99.5)
	// ------------------------------------------------------------
	clip := 99.5
	lo := percentile(vals, 100.0-clip)
	hi := percentile(vals, clip)

	if hi <= lo {
		return fmt.Errorf("invalid clip range: lo=%f hi=%f", lo, hi)
	}

	// ------------------------------------------------------------
	// Create grayscale image
	// ------------------------------------------------------------
	img := image.NewGray(image.Rect(0, 0, w, h))
	scale := 255.0 / (hi - lo)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			v := vals[i]

			if v < lo {
				v = lo
			}
			if v > hi {
				v = hi
			}

			u := uint8((v - lo) * scale)
			img.SetGray(x, y, color.Gray{Y: u})
		}
	}

	// ------------------------------------------------------------
	// Write PNG
	// ------------------------------------------------------------
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func percentile(data []float64, p float64) float64 {
	if len(data) == 0 {
		return 0
	}

	cp := make([]float64, len(data))
	copy(cp, data)
	sort.Float64s(cp)

	if p <= 0 {
		return cp[0]
	}
	if p >= 100 {
		return cp[len(cp)-1]
	}

	pos := (p / 100.0) * float64(len(cp)-1)
	i := int(pos)
	f := pos - float64(i)

	if i+1 < len(cp) {
		return cp[i]*(1-f) + cp[i+1]*f
	}
	return cp[i]
}
