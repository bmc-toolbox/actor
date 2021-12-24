package routes

import (
	"net/http"
	"strconv"

	"github.com/bmc-toolbox/actor/internal/actions"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type (
	BladeByPosAPI struct {
		baseAPI
	}
)

func NewBladeByPosAPI(planMaker *actions.PlanMaker) *BladeByPosAPI {
	return &BladeByPosAPI{baseAPI{planMaker: planMaker}}
}

// ChassisBladePowerStatusByPosition checks the current power status of a blade in a given chassis
func (ba BladeByPosAPI) ChassisBladePowerStatusByPosition(ctx *gin.Context) {
	logger := log.WithField("method", "ChassisBladePowerStatusByPosition")

	host, bladePos, err := ba.getAndValidateParams(ctx)
	if err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	logger = log.WithField("ip", host).WithField("pos", bladePos)

	ba.powerStatus(ctx, map[string]interface{}{"host": host, "bladePos": bladePos}, logger)
}

// ChassisBladeExecuteActionsByPosition carries out the execution of the requested action-list for a blade in a given chassis
func (ba BladeByPosAPI) ChassisBladeExecuteActionsByPosition(ctx *gin.Context) {
	logger := log.WithField("method", "ChassisBladePowerStatusByPosition")

	host, bladePos, err := ba.getAndValidateParams(ctx)
	if err != nil {
		logger.Warn(err)
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	logger = log.WithField("ip", host).WithField("pos", bladePos)

	ba.executeActions(ctx, map[string]interface{}{"host": host, "bladePos": bladePos}, logger)
}

func (ba BladeByPosAPI) getAndValidateParams(ctx *gin.Context) (string, int, error) {
	host := ctx.Param("host")
	bladePosStr := ctx.Param("pos")

	for _, validateFn := range []func() error{
		func() error { return validateHost(host) },
		func() error { return validateBladePos(bladePosStr) },
	} {
		if err := validateFn(); err != nil {
			return "", 0, err
		}
	}

	bladePos, _ := strconv.Atoi(bladePosStr)

	return host, bladePos, nil
}
