package routes

import (
	"fmt"
	"net/http"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type (
	baseAPI struct {
		planMaker *actions.PlanMaker
	}
)

func (ba baseAPI) powerStatus(ctx *gin.Context, params map[string]interface{}, logger *logrus.Entry) {
	plan, err := ba.planMaker.MakePlan([]string{actions.IsOn}, params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	results, err := plan.Run()
	responses := actionResultsToResponses(results)

	if len(responses) == 0 {
		err = fmt.Errorf("actions have been executed but no response returned: %w", err)
		logger.WithError(err).Error("error carrying out action")
		ctx.JSON(http.StatusInternalServerError, newErrorResponse(err))
		return
	}
	response := responses[0]
	if err != nil {
		ctx.JSON(http.StatusExpectationFailed, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (ba baseAPI) executeActions(ctx *gin.Context, params map[string]interface{}, logger *logrus.Entry) {
	req, err := unmarshalRequest(ctx)
	if err != nil {
		logger.WithError(err).Error("failed to unmarshal request")
		ctx.JSON(http.StatusBadRequest, newErrorResponse(fmt.Errorf("failed to unmarshal request: %w", err)))
		return
	}

	plan, err := ba.planMaker.MakePlan(req.ActionSequence, params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newErrorResponse(err))
		return
	}

	results, err := plan.Run()
	responses := actionResultsToResponses(results)

	if err != nil {
		ctx.JSON(http.StatusExpectationFailed, responses)
		return
	}

	ctx.JSON(http.StatusOK, responses)
	return
}
