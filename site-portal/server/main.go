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
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/api"
	"github.com/FederatedAI/FedLCM/site-portal/server/constants"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/gorm"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/utils/logging"
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// swag API
	_ "github.com/FederatedAI/FedLCM/site-portal/server/docs"
)

// FRONTEND_DIR is the folder where the frontend static file resides
var FRONTEND_DIR = getFrontendDir()

// main starts the API server
// @title site portal API service
// @version v1
// @description backend APIs of site portal service
// @termsOfService http://swagger.io/terms/
// @contact.name FedLCM team
// @BasePath /api/v1
// @in header
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
	debugLog, _ := strconv.ParseBool(viper.GetString("siteportal.debug"))
	if debugLog {
		log.Logger = log.Logger.Level(zerolog.DebugLevel)
	}

	if err := initDB(); err != nil {
		panic(err)
	}
	r := createGinEngine()
	initRouter(r)

	tlsEnabled := viper.GetBool("siteportal.tls.enabled")
	if tlsEnabled {
		sitePortalServerCert := viper.GetString("siteportal.tls.server.cert")
		sitePortalServerKey := viper.GetString("siteportal.tls.server.key")
		caCertPath := viper.GetString("siteportal.tls.ca.cert")
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
		tlsPort := viper.GetString("siteportal.tls.port")
		if tlsPort == "" {
			tlsPort = "8443"
		}
		server := http.Server{
			Addr:      ":" + tlsPort,
			Handler:   r,
			TLSConfig: tlsConfig,
		}
		log.Info().Msgf("Listening and serving HTTPs on %s", server.Addr)
		if err := server.ListenAndServeTLS(sitePortalServerCert, sitePortalServerKey); err != nil {
			log.Error().Err(err).Msg("server run error with TLS, ")
			return
		}
	} else {
		// Defining the listening port can facilitate development and debugging, and multiple services can be started in one place.
		port := viper.GetString("siteportal.port")
		if port == "" {
			port = "8080"
		}
		err := r.Run(fmt.Sprintf(":%s", port))
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

	r.NoRoute(func(c *gin.Context) {
		dir, file := path.Split(c.Request.RequestURI)
		ext := filepath.Ext(file)
		if file == "" || ext == "" {
			c.File(filepath.Join(FRONTEND_DIR, "index.html"))
		} else {
			c.File(filepath.Join(FRONTEND_DIR, dir, file))
		}
	})

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

		// user management
		userRepo := &gorm.UserRepo{}
		userRepo.InitTable()
		userRepo.InitData()
		// create authMiddleware before any other controllers
		if err := api.CreateAuthMiddleware(userRepo); err != nil {
			panic(err)
		}
		api.NewUserController(userRepo).Route(v1)

		// site management
		siteRepo := &gorm.SiteRepo{}
		siteRepo.InitTable()
		siteRepo.InitData()
		api.NewSiteController(siteRepo).Route(v1)

		// local data management repo
		localDataRepo := &gorm.LocalDataRepo{}
		localDataRepo.InitTable()

		// project management repo
		projectRepo := &gorm.ProjectRepo{}
		projectRepo.InitTable()
		projectParticipantRepo := &gorm.ProjectParticipantRepo{}
		projectParticipantRepo.InitTable()
		projectInvitationRepo := &gorm.ProjectInvitationRepo{}
		projectInvitationRepo.InitTable()
		projectDataRepo := &gorm.ProjectDataRepo{}
		projectDataRepo.InitTable()

		// job management repo
		jobRepo := &gorm.JobRepo{}
		jobRepo.InitTable()
		jobParticipantRepo := &gorm.JobParticipantRepo{}
		jobParticipantRepo.InitTable()

		// model management repo
		modelRepo := &gorm.ModelRepo{}
		modelRepo.InitTable()

		// model deployment repo
		modelDeploymentRepo := &gorm.ModelDeploymentRepo{}
		modelDeploymentRepo.InitTable()

		// local data management
		api.NewLocalDataController(localDataRepo, siteRepo, projectRepo, projectDataRepo).Route(v1)

		// project management
		api.NewProjectController(projectRepo, siteRepo, projectParticipantRepo,
			projectInvitationRepo, projectDataRepo, localDataRepo, jobRepo, jobParticipantRepo,
			modelRepo).Route(v1)

		// job management
		api.NewJobController(jobRepo, jobParticipantRepo, projectRepo, siteRepo, projectDataRepo, modelRepo).Route(v1)

		// model management
		api.NewModelController(modelRepo, modelDeploymentRepo, siteRepo, projectRepo).Route(v1)
	}
}

func getFrontendDir() (frontendDir string) {
	frontendDir = os.Getenv("FRONTEND_DIR")
	if frontendDir != "" {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(exe)
	frontendDir = filepath.Join(exePath, "frontend")
	return
}
