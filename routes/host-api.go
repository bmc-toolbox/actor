package routes

import (
	"fmt"
	"net/http"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/executor"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type (
	HostAPI struct {
		planMaker *executor.PlanMaker
	}
)

func NewHostAPI(planMaker *executor.PlanMaker) *HostAPI {
	return &HostAPI{planMaker: planMaker}
}

// HostPowerStatus checks the current power status of a given host
func (h HostAPI) HostPowerStatus(c *gin.Context) {
	logger := log.WithField("method", "HostPowerStatus")

	host := c.Param("host")
	if err := validateHost(host); err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "bmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}
	logger = log.WithField("ip", host)

	plan, err := h.planMaker.MakePlan([]string{actions.IsOn}, map[string]interface{}{"host": host})
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	results, err := plan.Run()
	responses := actionResultsToResponses(results)

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
func (h HostAPI) HostExecuteActions(c *gin.Context) {
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

	plan, err := h.planMaker.MakePlan(req.ActionSequence, map[string]interface{}{"host": host})
	if err != nil {
		c.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	results, err := plan.Run()
	responses := actionResultsToResponses(results)

	if err != nil {
		c.JSON(http.StatusExpectationFailed, responses)
		return
	}

	c.JSON(http.StatusOK, responses)
	return
}
