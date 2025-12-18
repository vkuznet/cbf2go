package httpapi

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"cbf2go/internal/cbf"
	"cbf2go/internal/embed"
	"cbf2go/internal/qdrant"
)

type Server struct {
	Qdrant   *qdrant.Client
	EmbedURL string
}

// default size of embedded vector
var defaultSize = 512

func (s *Server) Register(r *gin.Engine) {
	r.POST("/search_cbf", s.searchUpload)
	r.GET("/search_cbf_path", s.searchFile)
	r.POST("/hybdridsearch", s.hybridSearch)
}

func (s *Server) searchUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// size := defaultSize
	// if val, err := strconv.Atoi(c.Query("size")); err == nil {
	// 	size = val
	// }

	tmp := "/tmp/" + uuid.New().String() + ".cbf"
	c.SaveUploadedFile(file, tmp)

	//s.searchPath(c, tmp, size)
}

func (s *Server) searchFile(c *gin.Context) {
	path := c.Query("path")
	method := c.Query("method")
	size := defaultSize
	if val, err := strconv.Atoi(c.Query("size")); err == nil {
		size = val
	}
	s.searchPath(c, path, size, method)
}

func (s *Server) searchPath(c *gin.Context, path string, size int, method string) {
	// use verbose=0 for ReadCBF function call
	pixels, w, h, err := cbf.ReadCBF(path, 0)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	//log.Printf("qdrant search pixes=%v, w=%v h=%v", len(pixels), w, h)

	verbose := 0 // no verbose information
	var vec []float32
	if method == "resnet" {
		ec := embed.NewEmbedClient(s.EmbedURL)

		var floatPixels []float32
		for _, p := range pixels {
			floatPixels = append(floatPixels, float32(p))
		}

		vec, err = ec.EmbedPixels(floatPixels, h, w)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	} else {
		vec = embed.ImageToEmbedding(pixels, w, h, size, verbose)
	}
	hits, err := s.Qdrant.Search(vec, 10)
	// log.Printf("qdrant search %+v, error=%v", hits, err)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"hits": hits})
}

func (s *Server) hybridSearch(c *gin.Context) {
	text := c.PostForm("text")
	file, _ := c.FormFile("image")

	var imgBytes []byte
	if file != nil {
		f, _ := file.Open()
		imgBytes, _ = io.ReadAll(f)
	}

	var qdrant *qdrant.Client
	var clip *embed.CLIPClient

	hits, err := HybridSearch(qdrant, clip, imgBytes, text)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"hits": hits})
}
