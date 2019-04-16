package server

import (
	"fmt"
	"log"
	"os"
	"time"

	"html/template"

	"github.com/GeertJohan/go.rice"
	"github.com/bmc-toolbox/actor/routes"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/bmc-toolbox/gin-go-metrics/middleware"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const (
	poweroff   = "poweroff"
	poweron    = "poweron"
	powercycle = "powercycle"
	hardreset  = "hardreset"
	reseat     = "reseat"
	ison       = "ison"
)

// Serve start and build the webservice binding on unix socket
func Serve() {
	debug := viper.GetBool("debug")

	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	templateBox, err := rice.FindBox("templates")
	if err != nil {
		log.Fatal(err)
	}

	staticBox, err := rice.FindBox("static")
	if err != nil {
		log.Fatal(err)
	}

	doc, err := template.New("doc.tmpl").Parse(templateBox.MustString("doc.tmpl"))
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	if viper.GetBool("metrics.enabled") {
		err := metrics.Setup(
			viper.GetString("metrics.type"),
			viper.GetString("metrics.host"),
			viper.GetInt("metrics.port"),
			viper.GetString("metrics.prefix.server"),
			time.Minute,
		)
		if err != nil {
			fmt.Printf("Failed to set up monitoring: %s", err)
			os.Exit(1)
		}
		go metrics.Scheduler(time.Minute, metrics.GoRuntimeStats, []string{""})
		go metrics.Scheduler(time.Minute, metrics.MeasureRuntime, []string{"uptime"}, time.Now())
		p := middleware.NewMetrics([]string{})
		router.Use(p.HandlerFunc())
	}

	router.SetHTMLTemplate(doc)
	router.Static("/screenshot", viper.GetString("screenshot_storage"))
	router.StaticFS("/static", staticBox.HTTPBox())

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "doc.tmpl", gin.H{})
	})

	// Host level actions
	router.GET("/host/:host", routes.HostPowerStatus)
	router.POST("/host/:host", routes.HostExecuteActions)

	// Chassis level actions
	router.GET("/chassis/:host", routes.ChassisPowerStatus)
	router.POST("/chassis/:host", routes.ChassisExecuteActions)

	// Blade action on chassis level by position
	router.GET("/chassis/:host/position/:pos", routes.ChassisBladePowerStatusByPosition)
	router.POST("/chassis/:host/position/:pos", routes.ChassisBladeExecuteActionsByPosition)

	//  Blade action on chassis level by serial
	router.GET("/chassis/:host/serial/:serial", routes.ChassisBladePowerStatusBySerial)

	router.POST("/chassis/:host/serial/:serial", routes.ChassisBladeExecuteActionsBySerial)

	err = router.Run(viper.GetString("bind_to"))
	if err != nil {
		log.Print(err)
	}
}
