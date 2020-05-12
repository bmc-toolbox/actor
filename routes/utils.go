package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("invalid host: %s", host)
	}
	return nil
}

func unmarshalRequest(c *gin.Context) (*request, error) {
	req := &request{}
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	return req, nil
}
