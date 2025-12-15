package httpapi

import (
	"io"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"cbf2go/internal/cbf"
	"cbf2go/internal/embed"
	"cbf2go/internal/qdrant"
)

type Server struct {
	Qdrant *qdrant.Client
}

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

	tmp := "/tmp/" + uuid.New().String() + ".cbf"
	c.SaveUploadedFile(file, tmp)

	s.searchPath(c, tmp)
}

func (s *Server) searchFile(c *gin.Context) {
	path := c.Query("path")
	s.searchPath(c, path)
}

func (s *Server) searchPath(c *gin.Context, path string) {
	// use verbose=0 for ReadCBF function call
	pixels, w, h, err := cbf.ReadCBF(path, 0)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	log.Printf("qdrant search pixes=%v, w=%v h=%v", len(pixels), w, h)
	//w, h = cbf.ReconcileDimensions(pixels, w, h)
	//log.Printf("qdrant search pixes=%v, w=%v h=%v", len(pixels), w, h)

	vec := embed.ImageToEmbedding(pixels, w, h, 224)
	hits, err := s.Qdrant.Search(vec, 10)
	log.Printf("qdrant search %+v, error=%v", hits, err)
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
