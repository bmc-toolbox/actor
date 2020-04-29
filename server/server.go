package server

import (
	"fmt"
	"time"

	"html/template"

	rice "github.com/GeertJohan/go.rice"
	"github.com/bmc-toolbox/actor/routes"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/bmc-toolbox/gin-go-metrics/middleware"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Serve start and build the webservice binding on unix socket
func Serve() error {
	if !viper.GetBool("debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	if err := setupMetrics(router); err != nil {
		return fmt.Errorf("failed to set up metrics: %w", err)
	}

	if err := setupDoc(router); err != nil {
		return fmt.Errorf("failed to set up doc: %w", err)
	}

	if err := setupStatic(router); err != nil {
		return fmt.Errorf("failed to set up static: %w", err)
	}

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

	// Blade action on chassis level by serial
	router.GET("/chassis/:host/serial/:serial", routes.ChassisBladePowerStatusBySerial)
	router.POST("/chassis/:host/serial/:serial", routes.ChassisBladeExecuteActionsBySerial)

	return router.Run(viper.GetString("bind_to"))
}

func setupMetrics(router *gin.Engine) error {
	if !viper.GetBool("metrics.enabled") {
		return nil
	}

	err := metrics.Setup(
		viper.GetString("metrics.type"),
		viper.GetString("metrics.host"),
		viper.GetInt("metrics.port"),
		viper.GetString("metrics.prefix.server"),
		time.Minute,
	)
	if err != nil {
		return err
	}

	go metrics.Scheduler(time.Minute, metrics.GoRuntimeStats, []string{})
	go metrics.Scheduler(time.Minute, metrics.MeasureRuntime, []string{"uptime"}, time.Now())

	p := middleware.NewMetrics([]string{})
	router.Use(p.HandlerFunc([]string{"http"}, []string{"/"}, true))

	return nil
}

func setupDoc(router *gin.Engine) error {
	templateBox, err := rice.FindBox("templates")
	if err != nil {
		return err
	}

	doc, err := template.New("doc.tmpl").Parse(templateBox.MustString("doc.tmpl"))
	if err != nil {
		return err
	}

	router.SetHTMLTemplate(doc)

	return nil
}

func setupStatic(router *gin.Engine) error {
	staticBox, err := rice.FindBox("static")
	if err != nil {
		return err
	}

	router.StaticFS("/static", staticBox.HTTPBox())
	router.Static("/screenshot", viper.GetString("screenshot_storage"))

	return nil
}
