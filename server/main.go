// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/server/api"
	"github.com/FederatedAI/FedLCM/server/constants"
	"github.com/FederatedAI/FedLCM/server/infrastructure/gorm"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/utils/logging"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swag API
	_ "github.com/FederatedAI/FedLCM/server/docs"
)

// main starts the API server
//
// @title          lifecycle manager API service
// @version        v1
// @description    backend APIs of lifecycle manager service
// @termsOfService http://swagger.io/terms/
// @contact.name   FedLCM team
// @BasePath       /api/v1
// @in             header
func main() {
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		},
	).With().Caller().Stack().Logger().Level(zerolog.InfoLevel)
	debugLog, _ := strconv.ParseBool(viper.GetString("lifecyclemanager.debug"))
	if debugLog {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}

	if err := initDB(); err != nil {
		panic(err)
	}
	r := createGinEngine()
	initRouter(r)

	err := r.Run()
	if err != nil {
		log.Error().Err(err).Msg("gin run error, ")
		return
	}
}

func createGinEngine() *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(logger.SetLogger(
		logger.WithUTC(true),
		logger.WithLogger(logging.GetGinLogger)))

	return r
}

func initDB() error {
	var err error

	for i := 0; i < 3; i++ {
		err = gorm.InitDB()
		if err == nil {
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("initialization failed: %s", err)
}

func initRouter(r *gin.Engine) {

	v1 := r.Group("/api/" + constants.APIVersion)
	{
		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		v1.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"msg":                "The service is running",
				"api_version":        constants.APIVersion,
				"source_commit":      constants.Commit,
				"source_branch":      constants.Branch,
				"build_time":         constants.BuildTime,
				"experiment_enabled": viper.GetBool("lifecyclemanager.experiment.enabled"),
			})
		})

		// user management
		userRepo := &gorm.UserRepo{}
		userRepo.InitTable()
		userRepo.InitData()
		// create authMiddleware before any other controllers
		if err := api.CreateAuthMiddleware(userRepo); err != nil {
			panic(err)
		}
		api.NewUserController(userRepo).Route(v1)

		// infra provider management
		infraProviderKubernetesRepo := &gorm.InfraProviderKubernetesRepo{}
		infraProviderKubernetesRepo.InitTable()

		// endpoint management
		endpointKubeFATERepo := &gorm.EndpointKubeFATERepo{}
		endpointKubeFATERepo.InitTable()

		// chart management
		// TODO: replace this mock one with a complete implementation
		chartRepo := &gorm.ChartMockRepo{}
		api.NewChartController(chartRepo).Route(v1)

		// federation management
		federationFATERepo := &gorm.FederationFATERepo{}
		federationFATERepo.InitTable()
		federationOpenFLRepo := &gorm.FederationOpenFLRepo{}
		federationOpenFLRepo.InitTable()

		// participant management
		participantFATETRepo := &gorm.ParticipantFATERepo{}
		participantFATETRepo.InitTable()
		participantOpenFLRepo := &gorm.ParticipantOpenFLRepo{}
		participantOpenFLRepo.InitTable()

		// certificate management
		certificateAuthorityRepo := &gorm.CertificateAuthorityRepo{}
		certificateAuthorityRepo.InitTable()
		certificateRepo := &gorm.CertificateRepo{}
		certificateRepo.InitTable()
		certificateBindingRepo := &gorm.CertificateBindingRepo{}
		certificateBindingRepo.InitTable()

		// Event management
		eventRepo := &gorm.EventRepo{}
		eventRepo.InitTable()

		// Registration token management
		registrationTokenOpenFLRepo := &gorm.RegistrationTokenOpenFLRepo{}
		registrationTokenOpenFLRepo.InitTable()

		api.NewInfraProviderController(infraProviderKubernetesRepo, endpointKubeFATERepo).Route(v1)
		api.NewEndpointController(infraProviderKubernetesRepo, endpointKubeFATERepo, participantFATETRepo, participantOpenFLRepo, eventRepo).Route(v1)
		api.NewFederationController(infraProviderKubernetesRepo, endpointKubeFATERepo,
			federationFATERepo, federationOpenFLRepo, chartRepo, participantFATETRepo, participantOpenFLRepo, certificateAuthorityRepo,
			certificateRepo, certificateBindingRepo, registrationTokenOpenFLRepo, eventRepo).Route(v1)

		api.NewCertificateAuthorityController(certificateAuthorityRepo).Route(v1)
		api.NewCertificateController(certificateAuthorityRepo, certificateRepo, certificateBindingRepo, participantFATETRepo, participantOpenFLRepo, federationFATERepo, federationOpenFLRepo).Route(v1)
		api.NewEventController(eventRepo).Route(v1)
	}
}
