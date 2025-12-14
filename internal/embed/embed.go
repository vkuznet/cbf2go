package embed

import (
	"math"
)

func ImageToEmbedding(pixels []int32, w, h int, size int) []float32 {
	// resize using nearest neighbor
	scaleX := float64(w) / float64(size)
	scaleY := float64(h) / float64(size)

	vec := make([]float32, size*size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			idx := srcY*w + srcX
			vec[y*size+x] = float32(pixels[idx])
		}
	}

	// normalize (L2)
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	norm = math.Sqrt(norm) + 1e-8

	for i := range vec {
		vec[i] /= float32(norm)
	}

	return vec
}
