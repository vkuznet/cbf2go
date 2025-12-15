package embed

import (
	"fmt"
	"math"
)

func pixelsToUint8(pixels []int32) []uint8 {
	out := make([]uint8, len(pixels))
	for i, v := range pixels {
		out[i] = uint8(v) // identical to numpy astype(uint8)
	}
	return out
}

func resizeBilinear(src []uint8, w, h, size int) []uint8 {
	dst := make([]uint8, size*size)

	sx := float64(w) / float64(size)
	sy := float64(h) / float64(size)

	for y := 0; y < size; y++ {
		fy := float64(y) * sy
		y0 := int(fy)
		y1 := min(y0+1, h-1)
		wy := fy - float64(y0)

		for x := 0; x < size; x++ {
			fx := float64(x) * sx
			x0 := int(fx)
			x1 := min(x0+1, w-1)
			wx := fx - float64(x0)

			v00 := float64(src[y0*w+x0])
			v01 := float64(src[y0*w+x1])
			v10 := float64(src[y1*w+x0])
			v11 := float64(src[y1*w+x1])

			v0 := v00*(1-wx) + v01*wx
			v1 := v10*(1-wx) + v11*wx
			v := v0*(1-wy) + v1*wy

			dst[y*size+x] = uint8(v + 0.5)
		}
	}
	return dst
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ImageToEmbedding(pixels []int32, w, h, size int) []float32 {
	// 1) uint8 cast (matches numpy astype)
	u8 := pixelsToUint8(pixels)

	// 2) resize (bilinear)
	resized := resizeBilinear(u8, w, h, size)

	// 3) normalize to [0,1] and flatten
	vec := make([]float32, size*size)
	for i, v := range resized {
		vec[i] = float32(v) / 255.0
	}

	// 4) L2 normalize
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	norm = math.Sqrt(norm) + 1e-8

	for i := range vec {
		vec[i] /= float32(norm)
	}

	fmt.Println("embedded vector", vec[:10], "dim", len(vec))

	return vec
}
