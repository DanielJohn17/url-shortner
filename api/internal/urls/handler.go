package urls

import (
	"net/http"

	"github.com/DanielJohn17/url-shortner/api/internal/helpers"
	"github.com/gin-gonic/gin"
)

type URLHandlerInt interface {
	Create(c *gin.Context)
	GetUrl(c *gin.Context)
}

type URLHandler struct {
	service *URLService
}

func NewUrlHandler(s *URLService) *URLHandler {
	return &URLHandler{service: s}
}

// Create godoc
// @Summary Shorten a long URL
// @Description Accepts a long URL and returns a short code. Retries on collisions.
// @Tags urls
// @Accept json
// @Produce json
// @Param payload body UrlShort true "Long URL to shorten"
// @Success 201 {object} map[string]interface{} "Shortened URL created"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /url_shorter [post]
func (h *URLHandler) Create(c *gin.Context) {
	var urlShort UrlShort

	if err := helpers.ParseJSON(c, &urlShort); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	shortUrl, err := h.service.Create(c, urlShort.LongUrl)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": "false",
			"error":   err.Error(),
		})
		return
	}

	c.IndentedJSON(http.StatusCreated, gin.H{
		"success":   true,
		"short_url": shortUrl,
	})
}
