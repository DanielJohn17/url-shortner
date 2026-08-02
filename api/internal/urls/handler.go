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

func (h *URLHandler) Create(c *gin.Context) {
	var urlShort UrlShort

	if err := helpers.ParseJSON(c, &urlShort); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
	}

	shortUrl, err := h.service.CreateWithDelay(c, urlShort.LongUrl)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"success": "false",
			"error":   err.Error(),
		})
	}

	c.IndentedJSON(http.StatusCreated, gin.H{
		"success":   true,
		"short_url": shortUrl,
	})
}
