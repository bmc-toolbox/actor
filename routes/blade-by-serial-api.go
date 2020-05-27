package routes

import (
	"net/http"

	"github.com/bmc-toolbox/actor/internal/actions"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type (
	BladeBySerialAPI struct {
		baseAPI
	}
)

func NewBladeBySerialAPI(planMaker *actions.PlanMaker) *BladeBySerialAPI {
	return &BladeBySerialAPI{baseAPI{planMaker: planMaker}}
}

// ChassisBladePowerStatusBySerial checks the current power status of a blade in a given chassis
func (ba BladeBySerialAPI) ChassisBladePowerStatusBySerial(ctx *gin.Context) {
	logger := log.WithField("method", "ChassisBladePowerStatusBySerial")

	host, bladeSerial, err := ba.getAndValidateParams(ctx)
	if err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	logger = log.WithField("ip", host).WithField("serial", bladeSerial)

	ba.powerStatus(ctx, map[string]interface{}{"host": host, "bladeSerial": bladeSerial}, logger)
}

// ChassisBladeExecuteActionsBySerial carries out the execution of the requested action-list for a blade in a given chassis
func (ba BladeBySerialAPI) ChassisBladeExecuteActionsBySerial(ctx *gin.Context) {
	logger := log.WithField("method", "ChassisBladePowerStatusBySerial")

	host, bladeSerial, err := ba.getAndValidateParams(ctx)
	if err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	logger = log.WithField("ip", host).WithField("serial", bladeSerial)

	ba.executeActions(ctx, map[string]interface{}{"host": host, "bladeSerial": bladeSerial}, logger)
}

func (ba BladeBySerialAPI) getAndValidateParams(ctx *gin.Context) (string, string, error) {
	host := ctx.Param("host")
	bladeSerial := ctx.Param("serial")

	for _, validateFn := range []func() error{
		func() error { return validateHost(host) },
		func() error { return validateBladeSerial(bladeSerial) }} {
		if err := validateFn(); err != nil {
			return "", "", err
		}
	}

	return host, bladeSerial, nil
}
