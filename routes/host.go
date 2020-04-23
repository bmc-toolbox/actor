package routes

import (
	"fmt"
	"net/http"

	"github.com/bmc-toolbox/actor/internal/actions"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// HostPowerStatus checks the current power status of a given host
func HostPowerStatus(c *gin.Context) {
	logger := log.WithFields(log.Fields{"method": "HostPowerStatus"})

	host := c.Param("host")
	if err := validateHost(host); err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "bmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}
	logger = log.WithFields(log.Fields{"ip": host})

	executor := newHostActionExecutor(host, viper.GetString("bmc_user"), viper.GetString("bmc_pass"), false, logger)
	defer executor.cleanup()

	if err := executor.buildExecutionPlan([]string{actions.IsOn}); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	responses, err := executor.run()
	if len(responses) == 0 {
		err = fmt.Errorf("actions have been executed but no response returned: %w", err)
		logger.WithError(err).Error("error carrying out action")
		c.JSON(http.StatusInternalServerError, newErrorResponse(err))
		return
	}
	response := responses[0]
	if err != nil {
		c.JSON(http.StatusExpectationFailed, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// HostExecuteActions carries out the execution of the requested action-list for a given host
func HostExecuteActions(c *gin.Context) {
	logger := log.WithFields(log.Fields{"method": "HostExecuteActions"})

	host := c.Param("host")
	if err := validateHost(host); err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "bmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}
	logger = log.WithFields(log.Fields{"ip": host})

	req, err := unmarshalRequest(c)
	if err != nil {
		logger.WithError(err).Error("failed to unmarshal request")
		c.JSON(http.StatusBadRequest, newErrorResponse(fmt.Errorf("failed to unmarshal request: %w", err)))
		return
	}

	executor := newHostActionExecutor(
		host, viper.GetString("bmc_user"), viper.GetString("bmc_pass"), viper.GetBool("s3.enabled"), logger,
	)
	defer executor.cleanup()

	if err := executor.buildExecutionPlan(req.ActionSequence); err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	response, err := executor.run()
	if err != nil {
		c.JSON(http.StatusExpectationFailed, response)
		return
	}

	c.JSON(http.StatusOK, response)
	return
}

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
