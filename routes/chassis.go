package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func connectToChassis(username string, password string, host string) (bmc devices.Cmc, err error) {
	conn, err := discover.ScanAndConnect(host, username, password)
	if err != nil {
		return bmc, err
	}

	if bmc, ok := conn.(devices.Cmc); ok {
		if bmc.IsActive() {
			return bmc, err
		}
		return bmc, fmt.Errorf("this is the passive device, so I won't trigger any action")
	}

	return bmc, fmt.Errorf("unknown device or vendor")
}

// ChassisPowerStatus checks the current power status of a given host
func ChassisPowerStatus(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	status, err := bmc.IsOn()
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": status, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"action": "ison", "status": status, "message": "ok"})
}

// ChassisExecuteActions carries out the execution of the requested action-list for a given chassis
func ChassisExecuteActions(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	json := &Request{}
	var response []gin.H
	if err := c.ShouldBindJSON(json); err == nil {
		for _, action := range json.ActionSequence {
			if strings.HasPrefix(action, "sleep") {
				err := actions.Sleep(action)
				if err != nil {
					response = append(response, gin.H{"action": action, "status": false, "error": err.Error()})
					c.JSON(http.StatusExpectationFailed, response)
					return
				}
				response = append(response, gin.H{"action": action, "status": true, "message": "ok"})
				continue
			}

			var status bool
			switch action {
			case actions.PowerCycle:
				status, err = bmc.PowerCycle()
			case actions.PowerOn:
				status, err = bmc.PowerOn()
			default:
				response = append(response, gin.H{"action": action, "error": "unknown action"})
				c.JSON(http.StatusBadRequest, response)
				return
			}

			if err != nil {
				response = append(response, gin.H{"action": action, "status": status, "error": err.Error()})
				c.JSON(http.StatusExpectationFailed, response)
				return
			}
			response = append(response, gin.H{"action": action, "status": status, "message": "ok"})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChassisBladePowerStatusByPosition checks the current power status of a blade in a given chassis
func ChassisBladePowerStatusByPosition(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	posString := c.Param("pos")
	pos, err := strconv.Atoi(posString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid position: %s", posString)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	status, err := bmc.IsOnBlade(pos)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": status, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"action": "ison", "status": status, "message": "ok"})
}

// ChassisBladeExecuteActionsByPosition carries out the execution of the requested action-list for a blade in a given chassis
func ChassisBladeExecuteActionsByPosition(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})

		return
	}

	posString := c.Param("pos")
	pos, err := strconv.Atoi(posString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid position: %s", posString)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	json := &Request{}
	var response []gin.H
	if err := c.ShouldBindJSON(json); err == nil {
		for _, action := range json.ActionSequence {
			if strings.HasPrefix(action, "sleep") {
				err := actions.Sleep(action)
				if err != nil {
					response = append(response, gin.H{"action": action, "status": false, "error": err.Error()})
					c.JSON(http.StatusExpectationFailed, response)
					return
				}
				response = append(response, gin.H{"action": action, "status": true, "message": "ok"})
				continue
			}

			var status bool
			switch action {
			case actions.PowerCycle:
				status, err = bmc.PowerCycleBlade(pos)
			case actions.IsOn:
				status, err = bmc.IsOnBlade(pos)
			case actions.PxeOnce:
				status, err = bmc.PxeOnceBlade(pos)
			case actions.PowerCycleBmc:
				status, err = bmc.PowerCycleBmcBlade(pos)
			case actions.PowerOn:
				status, err = bmc.PowerOnBlade(pos)
			case actions.PowerOff:
				status, err = bmc.PowerOffBlade(pos)
			case actions.Reseat:
				status, err = bmc.ReseatBlade(pos)
			default:
				response = append(response, gin.H{"action": action, "error": "unknown action"})
				c.JSON(http.StatusBadRequest, response)
				return
			}

			if err != nil {
				response = append(response, gin.H{"action": action, "status": status, "error": err.Error()})
				c.JSON(http.StatusExpectationFailed, response)
				return
			}
			response = append(response, gin.H{"action": action, "status": status, "message": "ok"})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChassisBladePowerStatusBySerial checks the current power status of a blade in a given chassis
func ChassisBladePowerStatusBySerial(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	serial := c.Param("serial")
	if serial == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid serial: %s", serial)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	pos, err := bmc.FindBladePosition(serial)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%s: %s", host, err)})
		return
	}

	status, err := bmc.IsOnBlade(pos)
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": status, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"action": "ison", "status": status, "message": "ok"})
}

// ChassisBladeExecuteActionsBySerial carries out the execution of the requested action-list for a blade in a given chassis
func ChassisBladeExecuteActionsBySerial(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	serial := c.Param("serial")
	if serial == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid serial: %s", serial)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	pos, err := bmc.FindBladePosition(serial)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%s: %s", host, err)})
		return
	}

	json := &Request{}
	var response []gin.H
	if err := c.ShouldBindJSON(json); err == nil {
		for _, action := range json.ActionSequence {
			if strings.HasPrefix(action, "sleep") {
				err := actions.Sleep(action)
				if err != nil {
					response = append(response, gin.H{"action": action, "status": false, "error": err.Error()})
					c.JSON(http.StatusExpectationFailed, response)
					return
				}
				response = append(response, gin.H{"action": action, "status": true, "message": "ok"})
				continue
			}

			var status bool
			switch action {
			case actions.PowerCycle:
				status, err = bmc.PowerCycleBlade(pos)
			case actions.IsOn:
				status, err = bmc.IsOnBlade(pos)
			case actions.PxeOnce:
				status, err = bmc.PxeOnceBlade(pos)
			case actions.PowerCycleBmc:
				status, err = bmc.PowerCycleBmcBlade(pos)
			case actions.PowerOn:
				status, err = bmc.PowerOnBlade(pos)
			case actions.PowerOff:
				status, err = bmc.PowerOffBlade(pos)
			case actions.Reseat:
				status, err = bmc.ReseatBlade(pos)
			default:
				response = append(response, gin.H{"action": action, "error": "unknown action"})
				c.JSON(http.StatusBadRequest, response)
				return
			}

			if err != nil {
				response = append(response, gin.H{"action": action, "status": status, "error": err.Error()})
				c.JSON(http.StatusExpectationFailed, response)
				return
			}
			response = append(response, gin.H{"action": action, "status": status, "message": "ok"})
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
