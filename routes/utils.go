package routes

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func validateHost(host string) error {
	if host == "" {
		return fmt.Errorf("invalid host: %q", host)
	}
	return nil
}

func validateBladePos(pos string) error {
	if _, err := strconv.Atoi(pos); err != nil {
		return fmt.Errorf("invalid pos: %q: %w", pos, err)
	}
	return nil
}

func validateBladeSerial(serial string) error {
	if serial == "" {
		return fmt.Errorf("invalid serial: %q", serial)
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
