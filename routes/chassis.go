package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func connectToChassis(username string, password string, host string) (bmc devices.Cmc, err error) {
	conn, err := discover.ScanAndConnect(host, username, password)
	if err != nil {
		return bmc, err
	}

	if bmc, ok := conn.(devices.Cmc); ok {
		if bmc.IsActive() {
			metrics.IncrCounter([]string{"action", "cmc", "success", "connect"}, 1)
			return bmc, err
		}

		metrics.IncrCounter([]string{"errors", "cmc", "connect_passive"}, 1)
		return bmc, fmt.Errorf("this is the passive device, so I won't trigger any action")
	}

	return bmc, fmt.Errorf("unknown device or vendor")
}

// ChassisPowerStatus checks the current power status of a given host
func ChassisPowerStatus(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		log.WithFields(
			log.Fields{"method": "ChassisPowerStatus", "ip": host},
		).Warn("Invalid host")

		metrics.IncrCounter([]string{"errors", "invalid_host"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisPowerStatus", "operation": "connectToChassis", "ip": host, "err": err},
		).Warn("Failed to setup bmc connection")

		metrics.IncrCounter([]string{"errors", "cmc", "connect_fail"}, 1)
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	status, err := bmc.IsOn()
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisPowerStatus", "operation": "bmc.IsOn", "ip": host, "err": err},
		).Warn("Error determining power status")

		metrics.IncrCounter([]string{"action", "cmc", "fail", "ison"}, 1)
		c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": status, "message": err.Error()})
		return
	}

	metrics.IncrCounter([]string{"action", "cmc", "success", "ison"}, 1)
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
		log.WithFields(
			log.Fields{"method": "ChassisExecuteAction", "operation": "connectToChassis", "ip": host, "err": err},
		).Warn("Failed to setup connection")

		metrics.IncrCounter([]string{"errors", "cmc", "connect_fail"}, 1)
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
					metrics.IncrCounter([]string{"action", "cmc", "fail", "sleep"}, 1)
					c.JSON(http.StatusExpectationFailed, response)
					return
				}

				metrics.IncrCounter([]string{"action", "cmc", "success", "sleep"}, 1)
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
				log.WithFields(
					log.Fields{"method": "ChassisExecuteActions", "ip": host, "action": action},
				).Warn("Unknown action")

				metrics.IncrCounter([]string{"errors", "cmc", "unknown_action"}, 1)
				response = append(response, gin.H{"action": action, "error": "unknown action"})
				c.JSON(http.StatusBadRequest, response)
				return
			}

			if err != nil {
				log.WithFields(
					log.Fields{"method": "ChassisExecuteActions", "ip": host, "action": action, "err": err.Error()},
				).Warn("Error carrying out action")

				metrics.IncrCounter([]string{"action", "cmc", "fail", action}, 1)

				response = append(response, gin.H{"action": action, "status": status, "error": err.Error()})
				c.JSON(http.StatusExpectationFailed, response)
				return
			}

			metrics.IncrCounter([]string{"action", "cmc", "success", action}, 1)
			response = append(response, gin.H{"action": action, "status": status, "message": "ok"})
		}
	} else {
		log.WithFields(
			log.Fields{"method": "HostExecuteActions", "ip": host, "err": err.Error()},
		).Warn("Bad request")

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChassisBladePowerStatusByPosition checks the current power status of a blade in a given chassis
func ChassisBladePowerStatusByPosition(c *gin.Context) {
	host := c.Param("host")
	if host == "" {

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	posString := c.Param("pos")
	pos, err := strconv.Atoi(posString)
	if err != nil {
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid position: %s", posString)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladePowerStatusByPosition", "operation": "connectToChassis", "ip": host, "err": err.Error()},
		).Warn("Error connecting to chassis")

		metrics.IncrCounter([]string{"errors", "cmc", "connect_fail"}, 1)
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	status, err := bmc.IsOnBlade(pos)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladePowerStatusByPosition", "operation": "bmc.IsOnBlade", "ip": host, "err": err.Error()},
		).Warn("Error determining blade power status")

		metrics.IncrCounter([]string{"action", "cmc", "fail", "ison"}, 1)
		c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": status, "message": err.Error()})
		return
	}

	metrics.IncrCounter([]string{"action", "cmc", "success", "ison"}, 1)
	c.JSON(http.StatusOK, gin.H{"action": "ison", "status": status, "message": "ok"})
}

// ChassisBladeExecuteActionsByPosition carries out the execution of the requested action-list for a blade in a given chassis
func ChassisBladeExecuteActionsByPosition(c *gin.Context) {
	host := c.Param("host")
	if host == "" {

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	posString := c.Param("pos")
	pos, err := strconv.Atoi(posString)
	if err != nil {

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid position: %s", posString)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladeExecuteActionsByPosition", "operation": "connectToChassis", "ip": host, "err": err.Error()},
		).Warn("Error connecting to chassis")

		metrics.IncrCounter([]string{"errors", "cmc", "connect_fail"}, 1)
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

					metrics.IncrCounter([]string{"action", "cmc", "fail", "sleep"}, 1)
					response = append(response, gin.H{"action": action, "status": false, "error": err.Error()})
					c.JSON(http.StatusExpectationFailed, response)
					return
				}

				metrics.IncrCounter([]string{"action", "cmc", "success", "sleep"}, 1)
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
				log.WithFields(
					log.Fields{"method": "ChassisBladeExecuteActionsByPosition", "ip": host, "action": action},
				).Warn("Unknown action")

				metrics.IncrCounter([]string{"errors", "cmc", "unknown_action"}, 1)
				response = append(response, gin.H{"action": action, "error": "unknown action"})
				c.JSON(http.StatusBadRequest, response)
				return
			}

			if err != nil {
				log.WithFields(
					log.Fields{"method": "ChassisBladeExecuteActionsByPosition", "ip": host, "action": action, "err": err.Error()},
				).Warn("Error carrying out action")

				metrics.IncrCounter([]string{"action", "cmc", "fail", action}, 1)
				response = append(response, gin.H{"action": action, "status": status, "error": err.Error()})
				c.JSON(http.StatusExpectationFailed, response)
				return
			}

			metrics.IncrCounter([]string{"action", "cmc", "success", action}, 1)
			response = append(response, gin.H{"action": action, "status": status, "message": "ok"})
		}
	} else {
		log.WithFields(
			log.Fields{"method": "ChassisBladeExecuteActionsByPosition", "ip": host, "err": err.Error()},
		).Warn("Bad request")

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ChassisBladePowerStatusBySerial checks the current power status of a blade in a given chassis
func ChassisBladePowerStatusBySerial(c *gin.Context) {
	host := c.Param("host")
	if host == "" {

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	serial := c.Param("serial")
	if serial == "" {
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid serial: %s", serial)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladePowerStatusBySerial", "operation": "connectToChassis", "ip": host, "err": err.Error()},
		).Warn("Error connecting to chassis")

		metrics.IncrCounter([]string{"errors", "cmc", "connect_fail"}, 1)
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	pos, err := bmc.FindBladePosition(serial)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladePowerStatusBySerial", "operation": "bmc.FindBladePosition", "ip": host, "err": err.Error()},
		).Warn("Unable to determin blade position")

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("%s: %s", host, err)})
		return
	}

	status, err := bmc.IsOnBlade(pos)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladePowerStatusBySerial", "operation": "bmc.IsOnBlade", "ip": host, "err": err.Error()},
		).Warn("Unable to determine blade power status")

		metrics.IncrCounter([]string{"action", "cmc", "fail", "blade_ison"}, 1)
		c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": status, "message": err.Error()})
		return
	}

	metrics.IncrCounter([]string{"action", "cmc", "success", "blade_ison"}, 1)
	c.JSON(http.StatusOK, gin.H{"action": "ison", "status": status, "message": "ok"})
}

// ChassisBladeExecuteActionsBySerial carries out the execution of the requested action-list for a blade in a given chassis
func ChassisBladeExecuteActionsBySerial(c *gin.Context) {
	host := c.Param("host")
	if host == "" {

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	serial := c.Param("serial")
	if serial == "" {

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid serial: %s", serial)})
		return
	}

	bmc, err := connectToChassis(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladeExecuteActionsBySerial", "operation": "connectToChassis", "ip": host, "err": err.Error()},
		).Warn("Error connecting to chassis")

		metrics.IncrCounter([]string{"errors", "cmc", "connect_fail"}, 1)
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	defer bmc.Close()

	pos, err := bmc.FindBladePosition(serial)
	if err != nil {
		log.WithFields(
			log.Fields{"method": "ChassisBladePowerStatusBySerial", "operation": "bmc.FindBladePosition", "ip": host, "err": err.Error()},
		).Warn("Unable to determin blade position")

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
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

					metrics.IncrCounter([]string{"action", "cmc", "fail", "sleep"}, 1)
					response = append(response, gin.H{"action": action, "status": false, "error": err.Error()})
					c.JSON(http.StatusExpectationFailed, response)
					return
				}

				metrics.IncrCounter([]string{"action", "cmc", "success", "sleep"}, 1)
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
				log.WithFields(
					log.Fields{"method": "ChassisBladePowerStatusBySerial", "ip": host, "action": action},
				).Warn("Unknown action")

				metrics.IncrCounter([]string{"errors", "cmc", "unknown_action"}, 1)
				response = append(response, gin.H{"action": action, "error": "unknown action"})
				c.JSON(http.StatusBadRequest, response)
				return
			}

			if err != nil {

				log.WithFields(
					log.Fields{"method": "ChassisBladePowerStatusBySerial", "ip": host, "action": action, "err": err.Error()},
				).Warn("Error carrying out action")

				metrics.IncrCounter([]string{"action", "cmc", "fail", action}, 1)
				response = append(response, gin.H{"action": action, "status": status, "error": err.Error()})
				c.JSON(http.StatusExpectationFailed, response)
				return
			}

			metrics.IncrCounter([]string{"action", "cmc", "success", action}, 1)
			response = append(response, gin.H{"action": action, "status": status, "message": "ok"})
		}
	} else {
		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)
		log.WithFields(
			log.Fields{"method": "ChassisBladePowerStatusBySerial", "ip": host, "err": err.Error()},
		).Warn("Bad request")

		metrics.IncrCounter([]string{"errors", "cmc", "user_request_invalid"}, 1)

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
