package httpapi

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	"cbf2go/internal/cbf"
	"cbf2go/internal/embed"
	"cbf2go/internal/qdrant"
)

// TODO: I may introduce cache map for qdrant clients
//       if we plan to use multiple collections it would be useful

// Server represents server struct
type Server struct {
	Qdrant   *qdrant.Client
	EmbedURL string
}

// default size of embedded vector and number of matches to seek
var defaultSize = 512
var defaultLimit = 10

func (s *Server) Register(r *gin.Engine) {
	r.GET("/search_cbf_path", s.searchFile)
	r.POST("/hybdridsearch", s.hybridSearch)
}

func (s *Server) searchFile(c *gin.Context) {
	path := c.Query("path")
	method := c.Query("method")
	collection := c.Query("collection")
	size := defaultSize
	if val, err := strconv.Atoi(c.Query("size")); err == nil {
		size = val
	}
	limit := defaultLimit
	if val, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = val
	}
	s.searchPath(c, collection, path, method, size, limit)
}

func (s *Server) searchPath(c *gin.Context, collection, path, method string, size, limit int) {
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
	var hits []map[string]any
	if collection != "" {
		// we need to use new client with that collection
		client, e := qdrant.NewQdrantClient(
			s.Qdrant.URL,
			collection,
			s.Qdrant.FileExtension,
			s.Qdrant.Verbose,
		)
		if e != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		hits, err = client.Search(vec, limit)
	} else {
		hits, err = s.Qdrant.Search(vec, limit)
	}
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
