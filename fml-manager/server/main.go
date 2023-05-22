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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/fml-manager/server/api"
	"github.com/FederatedAI/FedLCM/fml-manager/server/constants"
	"github.com/FederatedAI/FedLCM/fml-manager/server/infrastructure/gorm"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/utils/logging"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swag API
	_ "github.com/FederatedAI/FedLCM/fml-manager/server/docs"
)

// main starts the API server
//	@title			fml manager API service
//	@version		v1
//	@description	backend APIs of fml manager service
//	@termsOfService	http://swagger.io/terms/
//	@contact.name	FedLCM team
//	@BasePath		/api/v1
//	@in				header
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
	debugLog, _ := strconv.ParseBool(viper.GetString("fmlmanager.debug"))
	if debugLog {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}

	if err := initDB(); err != nil {
		panic(err)
	}
	r := createGinEngine()
	initRouter(r)

	tlsEnabled := viper.GetBool("fmlmanager.tls.enabled")
	if tlsEnabled {
		fmlManagerServerCert := viper.GetString("fmlmanager.tls.server.cert")
		fmlManagerServerKey := viper.GetString("fmlmanager.tls.server.key")
		caCertPath := viper.GetString("fmlmanager.tls.ca.cert")
		pool := x509.NewCertPool()
		caCrt, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			log.Error().Err(err).Msg("read ca.crt file error")
		}
		pool.AppendCertsFromPEM(caCrt)
		tlsConfig := &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  pool,
		}
		tlsPort := viper.GetString("fmlmanager.tls.port")
		if tlsPort == "" {
			tlsPort = "8443"
		}
		server := http.Server{
			Addr:      ":" + tlsPort,
			Handler:   r,
			TLSConfig: tlsConfig,
		}
		log.Info().Msgf("Listening and serving HTTPs on %s", server.Addr)
		if err := server.ListenAndServeTLS(fmlManagerServerCert, fmlManagerServerKey); err != nil {
			log.Error().Err(err).Msg("gin run error with TLS, ")
			return
		}
	} else {
		err := r.Run()
		if err != nil {
			log.Error().Err(err).Msg("gin run error, ")
			return
		}
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
				"msg":           "The service is running",
				"api_version":   constants.APIVersion,
				"source_commit": constants.Commit,
				"source_branch": constants.Branch,
				"build_time":    constants.BuildTime,
			})
		})

		// site management
		siteRepo := &gorm.SiteRepo{}
		siteRepo.InitTable()
		api.NewSiteController(siteRepo).Route(v1)

		// project management
		projectRepo := &gorm.ProjectRepo{}
		projectRepo.InitTable()
		projectParticipantRepo := &gorm.ProjectParticipantRepo{}
		projectParticipantRepo.InitTable()
		projectInvitationRepo := &gorm.ProjectInvitationRepo{}
		projectInvitationRepo.InitTable()
		projectDataRepo := &gorm.ProjectDataRepo{}
		projectDataRepo.InitTable()
		api.NewProjectController(projectRepo, siteRepo, projectParticipantRepo, projectInvitationRepo, projectDataRepo).Route(v1)

		// job management repo
		jobRepo := &gorm.JobRepo{}
		jobRepo.InitTable()
		jobParticipantRepo := &gorm.JobParticipantRepo{}
		jobParticipantRepo.InitTable()
		api.NewJobController(jobRepo, jobParticipantRepo, projectRepo, siteRepo, projectDataRepo).Route(v1)
	}
}
