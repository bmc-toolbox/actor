package server

import (
	"fmt"
	"html/template"

	rice "github.com/GeertJohan/go.rice"
	"github.com/bmc-toolbox/actor/routes"
	"github.com/gin-gonic/gin"
)

type (
	Server struct {
		config *Config
		router *gin.Engine
	}

	Config struct {
		IsDebug           bool
		Address           string
		ScreenshotStorage string
	}

	APIs struct {
		HostAPI          *routes.HostAPI
		ChassisAPI       *routes.ChassisAPI
		BladeByPosAPI    *routes.BladeByPosAPI
		BladeBySerialAPI *routes.BladeBySerialAPI
	}
)

// New creates a new Server
func New(config *Config, middlewares []gin.HandlerFunc, apis *APIs) (*Server, error) {
	if !config.IsDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	if err := setupDoc(router); err != nil {
		return nil, fmt.Errorf("failed to set up doc: %w", err)
	}

	if err := setupStatic(router, config.ScreenshotStorage); err != nil {
		return nil, fmt.Errorf("failed to set up static: %w", err)
	}

	setupRoutes(router, apis)

	router.Use(middlewares...)

	return &Server{config: config, router: router}, nil
}

// Serve start and build the webservice binding on unix socket
func (s *Server) Serve() error {
	return s.router.Run(s.config.Address)
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

func setupStatic(router *gin.Engine, screenshotStorage string) error {
	staticBox, err := rice.FindBox("static")
	if err != nil {
		return err
	}

	router.StaticFS("/static", staticBox.HTTPBox())
	router.Static("/screenshot", screenshotStorage)

	return nil
}

func setupRoutes(router *gin.Engine, apis *APIs) {
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "doc.tmpl", gin.H{})
	})

	// Host level actions
	router.GET("/host/:host", apis.HostAPI.HostPowerStatus)
	router.POST("/host/:host", apis.HostAPI.HostExecuteActions)

	// Chassis level actions
	router.GET("/chassis/:host", apis.ChassisAPI.ChassisPowerStatus)
	router.POST("/chassis/:host", apis.ChassisAPI.ChassisExecuteActions)

	// Blade action on chassis level by position
	router.GET("/chassis/:host/position/:pos", apis.BladeByPosAPI.ChassisBladePowerStatusByPosition)
	router.POST("/chassis/:host/position/:pos", apis.BladeByPosAPI.ChassisBladeExecuteActionsByPosition)

	// Blade action on chassis level by serial
	router.GET("/chassis/:host/serial/:serial", apis.BladeBySerialAPI.ChassisBladePowerStatusBySerial)
	router.POST("/chassis/:host/serial/:serial", apis.BladeBySerialAPI.ChassisBladeExecuteActionsBySerial)
}
