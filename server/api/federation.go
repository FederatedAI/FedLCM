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

package api

import (
	"net/http"
	"strconv"

	"github.com/FederatedAI/FedLCM/server/application/service"
	"github.com/FederatedAI/FedLCM/server/constants"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	domainService "github.com/FederatedAI/FedLCM/server/domain/service"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// FederationController handles federation related APIs
type FederationController struct {
	federationApp         *service.FederationApp
	participantAppService *service.ParticipantApp
}

// NewFederationController returns a controller instance to handle federation API requests
func NewFederationController(infraProviderKubernetesRepo repo.InfraProviderRepository,
	endpointKubeFATERepo repo.EndpointRepository,
	federationFATERepo repo.FederationRepository,
	federationOpenFLRepo repo.FederationRepository,
	chartRepo repo.ChartRepository,
	participantFATERepo repo.ParticipantFATERepository,
	participantOpenflRepo repo.ParticipantOpenFLRepository,
	certificateAuthorityRepo repo.CertificateAuthorityRepository,
	certificateRepo repo.CertificateRepository,
	certificateBindingRepo repo.CertificateBindingRepository,
	registrationTokenOpenFLRepo repo.RegistrationTokenRepository,
	eventRepo repo.EventRepository) *FederationController {
	return &FederationController{
		federationApp: &service.FederationApp{
			FederationFATERepo:          federationFATERepo,
			ParticipantFATERepo:         participantFATERepo,
			FederationOpenFLRepo:        federationOpenFLRepo,
			ParticipantOpenFLRepo:       participantOpenflRepo,
			RegistrationTokenOpenFLRepo: registrationTokenOpenFLRepo,
		},
		participantAppService: &service.ParticipantApp{
			ParticipantFATERepo:         participantFATERepo,
			FederationFATERepo:          federationFATERepo,
			FederationOpenFLRepo:        federationOpenFLRepo,
			ParticipantOpenFLRepo:       participantOpenflRepo,
			RegistrationTokenOpenFLRepo: registrationTokenOpenFLRepo,
			EndpointKubeFATERepo:        endpointKubeFATERepo,
			InfraProviderKubernetesRepo: infraProviderKubernetesRepo,
			ChartRepo:                   chartRepo,
			CertificateAuthorityRepo:    certificateAuthorityRepo,
			CertificateRepo:             certificateRepo,
			CertificateBindingRepo:      certificateBindingRepo,
			EventRepo:                   eventRepo,
		},
	}
}

// Route sets up route mappings to federation related APIs
func (controller *FederationController) Route(r *gin.RouterGroup) {
	federation := r.Group("federation")

	// we use the token string in the request for authentication
	federation.POST("/openfl/envoy/register", controller.registerOpenFLEnvoy)
	federation.GET("/openfl/envoy/:uuid", controller.getOpenFLEnvoyWithToken)

	federation.Use(authMiddleware.MiddlewareFunc())
	{
		federation.GET("", controller.list)
	}
	fate := federation.Group("fate")
	{
		fate.POST("", controller.createFATE)
		fate.GET("/:uuid", controller.getFATE)
		fate.DELETE("/:uuid", controller.deleteFATE)

		fate.GET("/exchange/yaml", controller.getFATEExchangeDeploymentYAML)
		fate.GET("/cluster/yaml", controller.getFATEClusterDeploymentYAML)

		fate.POST("/:uuid/exchange", controller.createFATEExchange)
		fate.POST("/:uuid/exchange/external", controller.createExternalFATEExchange)
		fate.POST("/:uuid/cluster", controller.createFATECluster)
		fate.POST("/:uuid/partyID/check", controller.checkFATEPartyID)
		fate.POST("/:uuid/cluster/external", controller.createExternalFATECluster)

		fate.DELETE("/:uuid/exchange/:exchangeUUID", controller.deleteFATEExchange)
		fate.DELETE("/:uuid/cluster/:clusterUUID", controller.deleteFATECluster)

		fate.GET("/:uuid/participant", controller.getFATEParticipant)

		fate.GET("/:uuid/exchange/:exchangeUUID", controller.getFATEExchange)
		fate.GET("/:uuid/cluster/:clusterUUID", controller.getFATECluster)

		fate.GET("/:uuid/exchange/:exchangeUUID/upgrade", controller.getFATEExchangeUpgrade)
		fate.GET("/:uuid/cluster/:clusterUUID/upgrade", controller.getFATEClusterUpgrade)

		fate.POST("/:uuid/exchange/:exchangeUUID/upgrade", controller.upgradeFATEExchange)
		fate.POST("/:uuid/cluster/:clusterUUID/upgrade", controller.upgradeFATECluster)

	}

	openfl := federation.Group("openfl")
	{
		openfl.POST("", controller.createOpenFL)
		openfl.GET("/:uuid", controller.getOpenFL)
		openfl.DELETE("/:uuid", controller.deleteOpenFL)

		token := openfl.Group("/:uuid/token")
		token.POST("", controller.createOpenFLToken)
		token.GET("", controller.listOpenFLToken)
		token.DELETE("/:tokenUUID", controller.deleteOpenFLToken)

		openfl.GET("/:uuid/participant", controller.getOpenFLParticipant)

		openfl.GET("/director/yaml", controller.getOpenFLDirectorDeploymentYAML)

		openfl.POST("/:uuid/director", controller.createOpenFLDirector)
		openfl.DELETE("/:uuid/director/:directorUUID", controller.deleteOpenFLDirector)
		openfl.GET("/:uuid/director/:directorUUID", controller.getOpenFLDirector)

		openfl.GET("/:uuid/envoy/:envoyUUID", controller.getOpenFLEnvoy)
		openfl.DELETE("/:uuid/envoy/:envoyUUID", controller.deleteOpenFLEnvoy)
	}
}

// list returns the federation list
//
// @Summary Return federation list,
// @Tags    Federation
// @Produce json
// @Success 200 {object} GeneralResponse{data=[]service.FederationListItem} "Success"
// @Failure 401 {object} GeneralResponse                                    "Unauthorized operation"
// @Failure 500 {object} GeneralResponse{code=int}                          "Internal server error"
// @Router  /federation [get]
func (controller *FederationController) list(c *gin.Context) {
	if federationList, err := func() ([]service.FederationListItem, error) {
		return controller.federationApp.List()
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
			Data:    nil,
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code:    constants.RespNoErr,
			Message: "",
			Data:    federationList,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createFATE creates a new FATE federation
//
// @Summary Create a new FATE federation
// @Tags    Federation
// @Produce json
// @Param   federation body     service.FederationFATECreationRequest true "The federation info"
// @Success 200        {object} GeneralResponse                       "Success, the data field is the created federation's uuid"
// @Failure 401        {object} GeneralResponse                       "Unauthorized operation"
// @Failure 500        {object} GeneralResponse{code=int}             "Internal server error"
// @Router  /federation/fate [post]
func (controller *FederationController) createFATE(c *gin.Context) {
	if uuid, err := func() (string, error) {
		creationInfo := &service.FederationFATECreationRequest{}
		if err := c.ShouldBindJSON(creationInfo); err != nil {
			return "", err
		}
		return controller.federationApp.CreateFATEFederation(creationInfo)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: uuid,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATE returns detailed information of a FATE federation
//
// @Summary Get specific info of a FATE federation
// @Tags    Federation
// @Produce json
// @Param   uuid path     string                                             true "federation UUID"
// @Success 200  {object} GeneralResponse{data=service.FederationFATEDetail} "Success"
// @Failure 401  {object} GeneralResponse                                    "Unauthorized operation"
// @Failure 500  {object} GeneralResponse{code=int}                          "Internal server error"
// @Router  /federation/fate/{uuid} [get]
func (controller *FederationController) getFATE(c *gin.Context) {
	uuid := c.Param("uuid")
	if info, err := controller.federationApp.GetFATEFederation(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: info,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// deleteFATE deletes the specified federation
//
// @Summary Delete a FATE federation
// @Tags    Federation
// @Produce json
// @Param   uuid path     string                    true "federation UUID"
// @Success 200  {object} GeneralResponse           "Success"
// @Failure 401  {object} GeneralResponse           "Unauthorized operation"
// @Failure 500  {object} GeneralResponse{code=int} "Internal server error"
// @Router  /federation/fate/{uuid} [delete]
func (controller *FederationController) deleteFATE(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := controller.federationApp.DeleteFATEFederation(uuid); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATEExchangeDeploymentYAML returns deployment yaml content for deploying FATE exchange
//
// @Summary Get FATE exchange deployment yaml
// @Tags    Federation
// @Produce json
// @Param   chart_uuid          query    string                       true "the chart uuid"
// @Param   name                query    string                       true "name of the deployment"
// @Param   namespace           query    string                       true "namespace of the deployment"
// @Param   service_type        query    int                          true "type of the service to be exposed 1: LoadBalancer 2: NodePort"
// @Param   registry            query    string                       true "FATE registry config saved in the infra provider"
// @Param   use_registry        query    bool                         true "choose if use the customized registry config"
// @Param   use_registry_secret query    bool                         true "choose if use the customized registry secret"
// @Param   enable_psp          query    bool                         true "choose if enable the podSecurityPolicy"
// @Success 200                 {object} GeneralResponse{data=string} "Success, the data field is the yaml content"
// @Failure 401                 {object} GeneralResponse              "Unauthorized operation"
// @Failure 500                 {object} GeneralResponse{code=int}    "Internal server error"
// @Router  /federation/fate/exchange/yaml [get]
func (controller *FederationController) getFATEExchangeDeploymentYAML(c *gin.Context) {
	if yaml, err := func() (string, error) {
		chartUUID := c.DefaultQuery("chart_uuid", "")
		name := c.DefaultQuery("name", "")
		namespace := c.DefaultQuery("namespace", "")
		if chartUUID == "" || name == "" || namespace == "" {
			return "", errors.New("missing necessary parameters")
		}
		serviceType, err := strconv.Atoi(c.DefaultQuery("service_type", "1"))
		if err != nil {
			return "", errors.New("invalid service type parameter")
		}
		useRegistry, err := strconv.ParseBool(c.DefaultQuery("use_registry", "false"))
		if err != nil {
			return "", err
		}
		registry := c.DefaultQuery("registry", "")
		if useRegistry && registry == "" {
			return "", errors.New("missing registry")
		}
		useRegistrySecretFATE, err := strconv.ParseBool(c.DefaultQuery("use_registry_secret", "false"))
		if err != nil {
			return "", err
		}
		enablePSP, err := strconv.ParseBool(c.DefaultQuery("enable_psp", "true"))
		if err != nil {
			return "", err
		}
		return controller.participantAppService.GetFATEExchangeDeploymentYAML(&domainService.ParticipantFATEExchangeYAMLCreationRequest{
			ChartUUID:   chartUUID,
			Name:        name,
			Namespace:   namespace,
			ServiceType: entity.ParticipantDefaultServiceType(serviceType),
			RegistryConfig: valueobject.KubeRegistryConfig{
				UseRegistry:       useRegistry,
				Registry:          registry,
				UseRegistrySecret: useRegistrySecretFATE,
			},
			EnablePSP: enablePSP,
		})
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: yaml,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createFATEExchange creates a new FATE exchange
//
// @Summary Create a new FATE exchange
// @Tags    Federation
// @Produce json
// @Param   uuid            path     string                                         true "federation UUID"
// @Param   creationRequest body     service.ParticipantFATEExchangeCreationRequest true "The creation requests"
// @Success 200             {object} GeneralResponse                                "Success, the data field is the created exchange's uuid"
// @Failure 401             {object} GeneralResponse                                "Unauthorized operation"
// @Failure 500             {object} GeneralResponse{code=int}                      "Internal server error"
// @Router  /federation/fate/:uuid/exchange [post]
func (controller *FederationController) createFATEExchange(c *gin.Context) {
	if exchangeUUID, err := func() (string, error) {
		federationUUID := c.Param("uuid")
		req := &domainService.ParticipantFATEExchangeCreationRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		req.FederationUUID = federationUUID
		return controller.participantAppService.CreateFATEExchange(req)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: exchangeUUID,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createExternalFATEExchange creates an external FATE exchange
//
// @Summary Create an external FATE exchange
// @Tags    Federation
// @Produce json
// @Param   uuid            path     string                                                       true "federation UUID"
// @Param   creationRequest body     domainService.ParticipantFATEExternalExchangeCreationRequest true "The creation requests"
// @Success 200             {object} GeneralResponse                                              "Success, the data field is the created exchange's uuid"
// @Failure 401             {object} GeneralResponse                                              "Unauthorized operation"
// @Failure 500             {object} GeneralResponse{code=int}                                    "Internal server error"
// @Router  /federation/fate/:uuid/exchange/external [post]
func (controller *FederationController) createExternalFATEExchange(c *gin.Context) {
	if exchangeUUID, err := func() (string, error) {
		federationUUID := c.Param("uuid")
		req := &domainService.ParticipantFATEExternalExchangeCreationRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		req.FederationUUID = federationUUID
		return controller.participantAppService.CreateExternalFATEExchange(req)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: exchangeUUID,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATEParticipant returns participant list of the specified federation
//
// @Summary Get participant list of the specified federation
// @Tags    Federation
// @Produce json
// @Param   uuid path     string                                                        true "federation UUID"
// @Success 200  {object} GeneralResponse{data=service.ParticipantFATEListInFederation} "Success"
// @Failure 401  {object} GeneralResponse                                               "Unauthorized operation"
// @Failure 500  {object} GeneralResponse{code=int}                                     "Internal server error"
// @Router  /federation/fate/{uuid}/participant [get]
func (controller *FederationController) getFATEParticipant(c *gin.Context) {
	if participants, err := func() (*service.ParticipantFATEListInFederation, error) {
		federationUUID := c.Param("uuid")
		return controller.participantAppService.GetFATEParticipantList(federationUUID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: participants,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// deleteFATEExchange deletes the specified FATE exchange
//
// @Summary Delete a FATE exchange
// @Tags    Federation
// @Produce json
// @Param   uuid         path     string                    true  "federation UUID"
// @Param   exchangeUUID path     string                    true  "exchange UUID"
// @Param   force        query    bool                      false "if set to true, will try to remove the exchange forcefully"
// @Success 200          {object} GeneralResponse           "Success"
// @Failure 401          {object} GeneralResponse           "Unauthorized operation"
// @Failure 500          {object} GeneralResponse{code=int} "Internal server error"
// @Router  /federation/fate/{uuid}/exchange/{exchangeUUID} [delete]
func (controller *FederationController) deleteFATEExchange(c *gin.Context) {
	exchangeUUID := c.Param("exchangeUUID")
	if err := func() error {
		force, err := strconv.ParseBool(c.DefaultQuery("force", "false"))
		if err != nil {
			return err
		}
		return controller.participantAppService.RemoveFATEExchange(exchangeUUID, force)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATEClusterDeploymentYAML returns deployment yaml content for deploying FATE cluster
//
// @Summary Get FATE cluster deployment yaml
// @Tags    Federation
// @Produce json
// @Param   chart_uuid                           query    string                       true  "the chart uuid"
// @Param   federation_uuid                      query    string                       true  "the federation uuid"
// @Param   party_id                             query    int                          true  "party id"
// @Param   name                                 query    string                       true  "name of the deployment"
// @Param   namespace                            query    string                       true  "namespace of the deployment"
// @Param   service_type                         query    int                          true  "type of the service to be exposed"
// @Param   use_registry                         query    bool                         true  "choose if use the FATE registry config saved in the infra provider"
// @Param   enable_external_spark                query    bool                         true  "enable link an external Spark"
// @Param   external_spark_cores_per_node        query    int                          true  "external Spark info"
// @Param   external_spark_node                  query    int                          true  "external Spark info"
// @Param   external_spark_master                query    string                       true  "external Spark info"
// @Param   external_spark_driverHost            query    string                       true  "external Spark info"
// @Param   external_spark_driverHostType        query    string                       true  "external Spark info"
// @Param   external_spark_portMaxRetries        query    int                          true  "external Spark info"
// @Param   external_spark_driverStartPort       query    int                          true  "external Spark info"
// @Param   external_spark_blockManagerStartPort query    int                          true  "external spark info"
// @Param   external_spark_pysparkPython         query    string                       false "external spark info"
// @Param   enable_external_hdfs                 query    bool                         true  "enable link an external HDFS"
// @Param   external_hdfs_name_node              query    string                       true  "external HDFS info"
// @Param   external_hdfs_path_prefix            query    string                       false "external HDFS info"
// @Param   enable_external_pulsar               query    bool                         true  "enable link an external Pulsar"
// @Param   external_pulsar_host                 query    string                       true  "external Pulsar info"
// @Param   external_pulsar_mng_port             query    int                          true  "external Pulsar info"
// @Param   external_pulsar_port                 query    int                          true  "external Pulsar info"
// @Param   external_pulsar_ssl_port             query    int                          true  "external Pulsar info"
// @Param   use_registry_secret                  query    bool                         true  "choose if use the FATE registry secret saved in the infra provider"
// @Param   registry                             query    string                       true  "FATE registry config saved in the infra provider"
// @Param   enable_persistence                   query    bool                         true  "choose if use the persistent volume"
// @Param   storage_class                        query    string                       true  "provide the name of StorageClass"
// @Param   enable_psp                           query    bool                         true  "choose if enable the podSecurityPolicy"
// @Success 200                                  {object} GeneralResponse{data=string} "Success, the data field is the yaml content"
// @Failure 401                                  {object} GeneralResponse              "Unauthorized operation"
// @Failure 500                                  {object} GeneralResponse{code=int}    "Internal server error"
// @Router  /federation/fate/cluster/yaml [get]
func (controller *FederationController) getFATEClusterDeploymentYAML(c *gin.Context) {
	if yaml, err := func() (string, error) {
		chartUUID := c.DefaultQuery("chart_uuid", "")
		federationUUID := c.DefaultQuery("federation_uuid", "")
		name := c.DefaultQuery("name", "")
		namespace := c.DefaultQuery("namespace", "")
		if chartUUID == "" || name == "" || namespace == "" || federationUUID == "" {
			return "", errors.New("missing necessary parameters")
		}
		partyID, err := strconv.Atoi(c.Query("party_id"))
		if err != nil {
			return "", err
		}

		serviceType, err := strconv.Atoi(c.DefaultQuery("service_type", "1"))
		if err != nil {
			return "", errors.New("invalid service type parameter")
		}
		useRegistry, err := strconv.ParseBool(c.DefaultQuery("use_registry", "false"))
		if err != nil {
			return "", err
		}
		registry := c.DefaultQuery("registry", "")
		if useRegistry && registry == "" {
			return "", errors.New("missing registry")
		}
		useRegistrySecret, err := strconv.ParseBool(c.DefaultQuery("use_registry_secret", "false"))
		if err != nil {
			return "", err
		}
		enablePersistence, err := strconv.ParseBool(c.DefaultQuery("enable_persistence", "false"))
		if err != nil {
			return "", err
		}
		storageClass := c.DefaultQuery("storage_class", "")
		if enablePersistence && storageClass == "" {
			return "", errors.New("missing storage class name")
		}
		if err != nil {
			return "", err
		}

		enablePSP, err := strconv.ParseBool(c.DefaultQuery("enable_psp", "true"))
		if err != nil {
			return "", err
		}

		// Spark
		enableExternalSpark, err := strconv.ParseBool(c.DefaultQuery("enable_external_spark", "false"))
		if err != nil {
			return "", err
		}
		externalSparkCoresPerNode, err := strconv.Atoi(c.DefaultQuery("external_spark_cores_per_node", "8"))
		if err != nil {
			return "", errors.Errorf("invalid external_spark_cores_per_node parameter: %s", err)
		}
		externalSparkNode, err := strconv.Atoi(c.DefaultQuery("external_spark_node", "1"))
		if err != nil {
			return "", errors.Errorf("invalid external_spark_node parameter: %s", err)
		}
		externalSparkMaster := c.DefaultQuery("external_spark_master", "spark://spark-master:7077")
		externalSparkDriverHost := c.DefaultQuery("external_spark_driverHost", "fateflow")
		externalSparkDriverHostType := c.DefaultQuery("external_spark_driverHostType", "NodePort")
		externalSparkPortMaxRetries, err := strconv.Atoi(c.DefaultQuery("external_spark_portMaxRetries", "30"))
		if err != nil {
			return "", errors.Errorf("invalid external_spark_portMaxRetries parameter: %s", err)
		}
		externalSparkDriverStartPort, err := strconv.Atoi(c.DefaultQuery("external_spark_driverStartPort", "31000"))
		if err != nil {
			return "", errors.Errorf("invalid external_spark_driverStartPort parameter: %s", err)
		}
		externalSparkBlockManagerStartPort, err := strconv.Atoi(c.DefaultQuery("external_spark_blockManagerStartPort", "31100"))
		if err != nil {
			return "", errors.Errorf("invalid external_spark_blockManagerStartPort parameter: %s", err)
		}
		externalSparkPysparkPython := c.DefaultQuery("external_spark_pysparkPython", "")
		// HDFS
		enableExternalHDFS, err := strconv.ParseBool(c.DefaultQuery("enable_external_hdfs", "false"))
		if err != nil {
			return "", err
		}
		externalHDFSNameNode := c.DefaultQuery("external_hdfs_name_node", "hdfs://namenode:9000")
		externalHDFSPathPrefix := c.DefaultQuery("external_hdfs_path_prefix", "")
		// Pulsar
		enableExternalPulsar, err := strconv.ParseBool(c.DefaultQuery("enable_external_pulsar", "false"))
		if err != nil {
			return "", err
		}
		externalPulsarHost := c.DefaultQuery("external_pulsar_host", "pulsar")
		externalPulsarMngPort, err := strconv.Atoi(c.DefaultQuery("external_pulsar_mng_port", "8080"))
		if err != nil {
			return "", err
		}
		externalPulsarPort, err := strconv.Atoi(c.DefaultQuery("external_pulsar_port", "6650"))
		if err != nil {
			return "", err
		}
		externalPulsarSSLPort, err := strconv.Atoi(c.DefaultQuery("external_pulsar_ssl_port", "6651"))
		if err != nil {
			return "", err
		}

		return controller.participantAppService.GetFATEClusterDeploymentYAML(&domainService.ParticipantFATEClusterYAMLCreationRequest{
			ParticipantFATEExchangeYAMLCreationRequest: domainService.ParticipantFATEExchangeYAMLCreationRequest{
				ChartUUID:   chartUUID,
				Name:        name,
				Namespace:   namespace,
				ServiceType: entity.ParticipantDefaultServiceType(serviceType),
				RegistryConfig: valueobject.KubeRegistryConfig{
					UseRegistry:       useRegistry,
					Registry:          registry,
					UseRegistrySecret: useRegistrySecret,
				},
				EnablePSP: enablePSP,
			},
			FederationUUID:    federationUUID,
			PartyID:           partyID,
			EnablePersistence: enablePersistence,
			StorageClass:      storageClass,
			ExternalSpark: domainService.ExternalSpark{
				Enable:                enableExternalSpark,
				Cores_per_node:        externalSparkCoresPerNode,
				Nodes:                 externalSparkNode,
				Master:                externalSparkMaster,
				DriverHost:            externalSparkDriverHost,
				DriverHostType:        externalSparkDriverHostType,
				PortMaxRetries:        externalSparkPortMaxRetries,
				DriverStartPort:       externalSparkDriverStartPort,
				BlockManagerStartPort: externalSparkBlockManagerStartPort,
				PysparkPython:         externalSparkPysparkPython,
			},
			ExternalHDFS: domainService.ExternalHDFS{
				Enable:      enableExternalHDFS,
				Name_node:   externalHDFSNameNode,
				Path_prefix: externalHDFSPathPrefix,
			},
			ExternalPulsar: domainService.ExternalPulsar{
				Enable:   enableExternalPulsar,
				Host:     externalPulsarHost,
				Mng_port: externalPulsarMngPort,
				Port:     externalPulsarPort,
				SSLPort:  externalPulsarSSLPort,
			},
		})
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: yaml,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createFATECluster creates a new FATE cluster
//
// @Summary Create a new FATE cluster
// @Tags    Federation
// @Produce json
// @Param   uuid            path     string                                        true "federation UUID"
// @Param   creationRequest body     service.ParticipantFATEClusterCreationRequest true "The creation requests"
// @Success 200             {object} GeneralResponse                               "Success, the data field is the created cluster's uuid"
// @Failure 401             {object} GeneralResponse                               "Unauthorized operation"
// @Failure 500             {object} GeneralResponse{code=int}                     "Internal server error"
// @Router  /federation/fate/:uuid/cluster [post]
func (controller *FederationController) createFATECluster(c *gin.Context) {
	if exchangeUUID, err := func() (string, error) {
		federationUUID := c.Param("uuid")
		req := &domainService.ParticipantFATEClusterCreationRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		req.FederationUUID = federationUUID
		return controller.participantAppService.CreateFATECluster(req)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: exchangeUUID,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// deleteFATECluster deletes the specified FATE cluster
//
// @Summary Delete a FATE cluster
// @Tags    Federation
// @Produce json
// @Param   uuid        path     string                    true  "federation UUID"
// @Param   clusterUUID path     string                    true  "cluster UUID"
// @Param   force       query    bool                      false "if set to true, will try to remove the cluster forcefully"
// @Success 200         {object} GeneralResponse           "Success"
// @Failure 401         {object} GeneralResponse           "Unauthorized operation"
// @Failure 500         {object} GeneralResponse{code=int} "Internal server error"
// @Router  /federation/fate/{uuid}/cluster/{clusterUUID} [delete]
func (controller *FederationController) deleteFATECluster(c *gin.Context) {
	clusterUUID := c.Param("clusterUUID")
	if err := func() error {
		force, err := strconv.ParseBool(c.DefaultQuery("force", "false"))
		if err != nil {
			return err
		}
		return controller.participantAppService.RemoveFATECluster(clusterUUID, force)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createExternalFATECluster creates an external FATE cluster
//
// @Summary Create an external FATE cluster
// @Tags    Federation
// @Produce json
// @Param   uuid            path     string                                                true "federation UUID"
// @Param   creationRequest body     service.ParticipantFATEExternalClusterCreationRequest true "The creation requests"
// @Success 200             {object} GeneralResponse                                       "Success, the data field is the created cluster's uuid"
// @Failure 401             {object} GeneralResponse                                       "Unauthorized operation"
// @Failure 500             {object} GeneralResponse{code=int}                             "Internal server error"
// @Router  /federation/fate/:uuid/cluster/external [post]
func (controller *FederationController) createExternalFATECluster(c *gin.Context) {
	if clusterUUID, err := func() (string, error) {
		federationUUID := c.Param("uuid")
		req := &domainService.ParticipantFATEExternalClusterCreationRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		req.FederationUUID = federationUUID
		return controller.participantAppService.CreateExternalFATECluster(req)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: clusterUUID,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// checkFATEPartyID checks if the party ID is available
//
// @Summary Check if the party ID is available
// @Tags    Federation
// @Produce json
// @Param   uuid     path     string                    true "federation UUID"
// @Param   party_id query    int                       true "party ID"
// @Success 200      {object} GeneralResponse           "Success"
// @Failure 401      {object} GeneralResponse           "Unauthorized operation"
// @Failure 500      {object} GeneralResponse{code=int} "Internal server error"
// @Router  /federation/fate/:uuid/partyID/check [post]
func (controller *FederationController) checkFATEPartyID(c *gin.Context) {
	if err := func() error {
		federationUUID := c.Param("uuid")
		partyID, err := strconv.Atoi(c.Query("party_id"))
		if err != nil {
			return err
		}
		return controller.participantAppService.CheckFATPartyID(federationUUID, partyID)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATEExchange returns detailed information of a FATE exchange
//
// @Summary Get specific info of FATE Exchange
// @Tags    Federation
// @Produce json
// @Param   uuid path     string                                           true "federation UUID"
// @Success 200  {object} GeneralResponse{data=service.FATEExchangeDetail} "Success"
// @Failure 401  {object} GeneralResponse                                  "Unauthorized operation"
// @Failure 500  {object} GeneralResponse{code=int}                        "Internal server error"
// @Router  /federation/fate/{uuid}/exchange/{exchangeUUID} [get]
func (controller *FederationController) getFATEExchange(c *gin.Context) {
	exchangeUUID := c.Param("exchangeUUID")
	if exchangeDetail, err := controller.participantAppService.GetFATEExchangeDetail(exchangeUUID); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: exchangeDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATECluster returns detailed information of a FATE cluster
//
// @Summary Get specific info of FATE cluster
// @Tags    Federation
// @Produce json
// @Param   uuid path     string                                          true "federation UUID"
// @Success 200  {object} GeneralResponse{data=service.FATEClusterDetail} "Success"
// @Failure 401  {object} GeneralResponse                                 "Unauthorized operation"
// @Failure 500  {object} GeneralResponse{code=int}                       "Internal server error"
// @Router  /federation/fate/{uuid}/cluster/{clusterUUID} [get]
func (controller *FederationController) getFATECluster(c *gin.Context) {
	clusterUUID := c.Param("clusterUUID")
	if clusterDetail, err := controller.participantAppService.GetFATEClusterDetail(clusterUUID); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: clusterDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATEExchangeUpgrade returns detailed information of a FATE cluster
//
//	@Summary	Get specific info of FATE cluster
//	@Tags		Federation
//	@Produce	json
//	@Param		uuid	path		string															true	"federation UUID"
//	@Success	200		{object}	GeneralResponse{data=service.FATEClusterUpgradeableVersionList}	"Success"
//	@Failure	401		{object}	GeneralResponse													"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}										"Internal server error"
//	@Router		/federation/fate/{uuid}/exchange/:exchangeUUID/upgrade [get]
func (controller *FederationController) getFATEExchangeUpgrade(c *gin.Context) {
	clusterUUID := c.Param("clusterUUID")
	if clusterDetail, err := controller.participantAppService.GetFATEExchangeUpgrade(clusterUUID); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: clusterDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// getFATEExchangeUpgrade returns detailed information of a FATE cluster
//
//	@Summary	Get specific info of FATE cluster
//	@Tags		Federation
//	@Produce	json
//	@Param		uuid	path		string															true	"federation UUID"
//	@Success	200		{object}	GeneralResponse{data=service.FATEClusterUpgradeableVersionList}	"Success"
//	@Failure	401		{object}	GeneralResponse													"Unauthorized operation"
//	@Failure	500		{object}	GeneralResponse{code=int}										"Internal server error"
//	@Router		/federation/fate/{uuid}/cluster/:clusterUUID/upgrade [get]
func (controller *FederationController) getFATEClusterUpgrade(c *gin.Context) {
	clusterUUID := c.Param("clusterUUID")
	if clusterDetail, err := controller.participantAppService.GetFATEClusterUpgrade(clusterUUID); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: clusterDetail,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// upgradeFATEExchange upgrade the FATE exchange
//
//	@Summary	Create a new FATE exchange
//	@Tags		Federation
//	@Produce	json
//	@Param		uuid			path		string						true	"federation UUID"
//	@Param		exchangeUUID	path		string						true	"exchange UUID"
//	@Param		upgradeVersion	body		string						true	"upgrade version"
//	@Success	200				{object}	GeneralResponse				"Success, the data field is the created exchange's uuid"
//	@Failure	401				{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500				{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/federation/fate/:uuid/exchange/:exchangeUUID/upgrade [post]
func (controller *FederationController) upgradeFATEExchange(c *gin.Context) {
	if exchangeUUID, err := func() (string, error) {
		federationUUID := c.Param("uuid")
		req := &domainService.ParticipantFATEExchangeUpgradeRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		req.FederationUUID = federationUUID
		return controller.participantAppService.UpgradeFATEExchange(req)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: exchangeUUID,
		}
		c.JSON(http.StatusOK, resp)
	}
}

// createFATEExchange upgrade the FATE exchange
//
//	@Summary	Create a new FATE exchange
//	@Tags		Federation
//	@Produce	json
//	@Param		uuid			path		string						true	"federation UUID"
//	@Param		clusteruuid		path		string						true	"cluster UUID"
//	@Param		upgradeVersion	body		string						true	"upgrade version"
//	@Success	200				{object}	GeneralResponse				"Success, the data field is the upgrade cluster's uuid"
//	@Failure	401				{object}	GeneralResponse				"Unauthorized operation"
//	@Failure	500				{object}	GeneralResponse{code=int}	"Internal server error"
//	@Router		/federation/fate/:uuid/cluster/:clusterUUID/upgrade [post]
func (controller *FederationController) upgradeFATECluster(c *gin.Context) {
	if exchangeUUID, err := func() (string, error) {
		federationUUID := c.Param("uuid")
		req := &domainService.ParticipantFATEClusterUpgradeRequest{}
		if err := c.ShouldBindJSON(req); err != nil {
			return "", err
		}
		req.FederationUUID = federationUUID
		return controller.participantAppService.UpgradeFATECluster(req)
	}(); err != nil {
		resp := &GeneralResponse{
			Code:    constants.RespInternalErr,
			Message: err.Error(),
		}
		c.JSON(http.StatusInternalServerError, resp)
	} else {
		resp := &GeneralResponse{
			Code: constants.RespNoErr,
			Data: exchangeUUID,
		}
		c.JSON(http.StatusOK, resp)
	}
}
