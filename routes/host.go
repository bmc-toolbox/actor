package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bmc-toolbox/actor/internal/actions"
	"github.com/bmc-toolbox/actor/internal/providers/ipmi"
	"github.com/bmc-toolbox/actor/internal/screenshot"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// HostPowerStatus checks the current power status of a given host
func HostPowerStatus(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	conn, err := discover.ScanAndConnect(host, viper.GetString("bmc_user"), viper.GetString("bmc_pass"))
	if err != nil {
		if err != errors.ErrVendorNotSupported {
			c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
			return
		}

	}
	if bmc, ok := conn.(devices.Bmc); ok {
		defer bmc.Close()
		status, err := bmc.PowerState()
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"action": "ison", "status": status, "message": "ok"})
		return
	}

	bmc, err := ipmi.New(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
	if err != nil {
		c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
		return
	}
	status, err := bmc.IsOn()
	if err != nil {
		c.JSON(http.StatusExpectationFailed, gin.H{"action": "ison", "status": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"action": "ison", "status": status, "message": "ok"})
}

// HostExecuteActions carries out the execution of the requested action-list for a given host
func HostExecuteActions(c *gin.Context) {
	host := c.Param("host")
	if host == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("invalid host: %s", host)})
		return
	}

	conn, err := discover.ScanAndConnect(host, viper.GetString("bmc_user"), viper.GetString("bmc_pass"))
	if err != nil {
		if err != errors.ErrVendorNotSupported {
			c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
			return
		}

		bmc, err := ipmi.New(viper.GetString("bmc_user"), viper.GetString("bmc_pass"), host)
		if err != nil {
			c.JSON(http.StatusPreconditionFailed, gin.H{"message": err.Error()})
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
					status, err = bmc.PowerCycle()
				case actions.IsOn:
					status, err = bmc.IsOn()
				case actions.PxeOnce:
					status, err = bmc.PxeOnce()
				case actions.PowerCycleBmc:
					status, err = bmc.PowerCycleBmc()
				case actions.PowerOn:
					status, err = bmc.PowerOn()
				case actions.PowerOff:
					status, err = bmc.PowerOff()
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
		return
	}

	if bmc, ok := conn.(devices.Bmc); ok {
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
				message := "ok"
				switch action {
				case actions.PowerCycle:
					status, err = bmc.PowerCycle()
				case actions.IsOn:
					var answer string
					answer, err = bmc.PowerState()
					if answer == "on" {
						status = true
					}
				case actions.PxeOnce:
					status, err = bmc.PxeOnce()
				case actions.PowerCycleBmc:
					status, err = bmc.PowerCycleBmc()
				case actions.PowerOn:
					status, err = bmc.PowerOn()
				case actions.PowerOff:
					status, err = bmc.PowerOff()
				case actions.Screenshot:
					if viper.GetBool("s3.enabled") {
						message, status, err = screenshot.S3(bmc, host)
					} else {
						message, status, err = screenshot.Local(bmc, host)
					}
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
				response = append(response, gin.H{"action": action, "status": status, "message": message})
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, response)
		return
	}

	c.JSON(http.StatusPreconditionFailed, gin.H{"message": "unknown device or vendor"})
}
