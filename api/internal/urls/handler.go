package urls

import (
	"fmt"
	"net/http"

	"github.com/DanielJohn17/url-shortner/api/internal/config"
	"github.com/DanielJohn17/url-shortner/api/internal/helpers"
	"github.com/gin-gonic/gin"
)

type URLHandlerInt interface {
	Create(c *gin.Context)
	RedirectUrl(c *gin.Context)
}

type URLHandler struct {
	service *URLService
}

var _ URLHandlerInt = (*URLHandler)(nil)

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
		"short_url": fmt.Sprintf("%s/%s", config.Env.DomainName, shortUrl),
	})
}

// RedirectUrl godoc
// @Summary Resolve a short code
// @Description Looks up the long URL for a short code.
// @Tags urls
// @Produce json
// @Param shortUrl path string true "Short code to resolve"
// @Success 302 {object} map[string]interface{} "Redirect found"
// @Failure 404 {object} map[string]interface{} "Short code not found"
// @Router /url_shorter/{shortUrl} [get]
func (h *URLHandler) RedirectUrl(c *gin.Context) {
	shortUrl := c.Param("shortUrl")

	longUrl, err := h.service.GetUrl(c, shortUrl)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.IndentedJSON(http.StatusFound, gin.H{
		"success":  true,
		"long_url": longUrl,
	})
}
