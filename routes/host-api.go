package routes

import (
	"net/http"

	"github.com/bmc-toolbox/actor/internal/actions"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type (
	HostAPI struct {
		baseAPI
	}
)

func NewHostAPI(planMaker *actions.PlanMaker) *HostAPI {
	return &HostAPI{baseAPI{planMaker: planMaker}}
}

// HostPowerStatus checks the current power status of a given host
func (ha HostAPI) HostPowerStatus(ctx *gin.Context) {
	logger := log.WithField("method", "HostPowerStatus")

	host := ctx.Param("host")
	if err := validateHost(host); err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "bmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}
	logger = log.WithField("ip", host)

	ha.powerStatus(ctx, map[string]interface{}{"host": host}, logger)
}

// HostExecuteActions carries out the execution of the requested action-list for a given host
func (ha HostAPI) HostExecuteActions(ctx *gin.Context) {
	logger := log.WithFields(log.Fields{"method": "HostExecuteActions"})

	host := ctx.Param("host")
	if err := validateHost(host); err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "bmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}
	logger = log.WithFields(log.Fields{"ip": host})

	ha.executeActions(ctx, map[string]interface{}{"host": host}, logger)
}
