package helpers

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ParseJSON[K comparable](c *gin.Context, payload *K) error {
	validate := validator.New()

	if c.Request.Body == nil {
		return fmt.Errorf("Missing request body")
	}

	if err := json.NewDecoder(c.Request.Body).Decode(payload); err != nil {
		return fmt.Errorf("error decoding payload")
	}

	if err := validate.Struct(payload); err != nil {
		return fmt.Errorf("Validation error: %w", err)
	}

	return nil
}
