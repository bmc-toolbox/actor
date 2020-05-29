package routes

import (
	"net/http"

	"github.com/bmc-toolbox/actor/internal/actions"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type (
	ChassisAPI struct {
		baseAPI
	}
)

func NewChassisAPI(planMaker *actions.PlanMaker) *ChassisAPI {
	return &ChassisAPI{baseAPI{planMaker: planMaker}}
}

// ChassisPowerStatus checks the current power status of a given host
func (ca ChassisAPI) ChassisPowerStatus(ctx *gin.Context) {
	logger := log.WithField("method", "ChassisPowerStatus")

	host := ctx.Param("host")
	if err := validateHost(host); err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}
	logger = log.WithField("ip", host)

	ca.powerStatus(ctx, map[string]interface{}{"host": host}, logger)
}

// ChassisExecuteActions carries out the execution of the requested action-list for a given chassis
func (ca ChassisAPI) ChassisExecuteActions(ctx *gin.Context) {
	logger := log.WithField("method", "ChassisExecuteAction")

	host := ctx.Param("host")
	if err := validateHost(host); err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}
	logger = log.WithField("ip", host)

	ca.executeActions(ctx, map[string]interface{}{"host": host}, logger)
}
