// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"time"

	"github.com/bmc-toolbox/actor/internal"
	"github.com/bmc-toolbox/actor/internal/executor"
	"github.com/bmc-toolbox/actor/routes"
	"github.com/bmc-toolbox/actor/server"
	metrics "github.com/bmc-toolbox/gin-go-metrics"
	"github.com/bmc-toolbox/gin-go-metrics/middleware"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start actor web service",
	Run: func(cmd *cobra.Command, args []string) {
		if err := setupMetrics(); err != nil {
			log.Fatal(err)
		}

		config := &server.Config{
			Address:           viper.GetString("bind_to"),
			IsDebug:           viper.GetBool("debug"),
			ScreenshotStorage: viper.GetString("screenshot_storage"),
		}

		middlewares := []gin.HandlerFunc{
			middleware.NewMetrics([]string{}).HandlerFunc([]string{"http"}, []string{"/"}, true),
		}

		bmcUsername := viper.GetString("bmc_user")
		bmcPassword := viper.GetString("bmc_pass")

		hostExecutorFactory := internal.NewHostExecutorFactory(
			&internal.HostExecutorFactoryConfig{
				IsS3: viper.GetBool("s3.enabled"), Username: bmcUsername, Password: bmcPassword,
			},
		)
		hostAPI := routes.NewHostAPI(executor.NewPlanMaker(hostExecutorFactory))

		chassisExecutorFactory := internal.NewChassisExecutorFactory(
			&internal.ChassisExecutorFactoryConfig{Username: bmcUsername, Password: bmcPassword},
		)
		chassisAPI := routes.NewChassisAPI(executor.NewPlanMaker(chassisExecutorFactory))

		bladeExecutorFactory := internal.NewBladeExecutorFactory(
			&internal.BladeExecutorFactoryConfig{Username: bmcUsername, Password: bmcPassword},
		)
		bladeAPI := routes.NewBladeAPI(executor.NewPlanMaker(bladeExecutorFactory))

		server, err := server.New(config, middlewares, hostAPI, chassisAPI, bladeAPI)
		if err != nil {
			log.Fatal(err)
		}

		if err := server.Serve(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func setupMetrics() error {
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

	return nil
}
