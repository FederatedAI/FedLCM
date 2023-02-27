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

package service

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"sync"

	site_portal_client "github.com/FederatedAI/FedLCM/pkg/site-portal-client"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/modules"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/yaml"
)

// ParticipantFATEExchangeUpgradeRequest is the exchange upgrade request
type ParticipantFATEExchangeUpgradeRequest struct {
	ParticipantFATEExchangeCreationRequest
	UpgradeVersion string `json:"upgrade_version"`
}

// ParticipantFATEClusterUpgradeRequest is the cluster upgrade request
type ParticipantFATEClusterUpgradeRequest struct {
	ParticipantFATEClusterCreationRequest
	UpgradeVersion string `json:"upgrade_version"`
}

// UpgradeExchange upgrade the FATE exchange, the returned *sync.WaitGroup can be used to wait for the completion of the async goroutine
func (s *ParticipantFATEService) UpgradeExchange(req *ParticipantFATEExchangeUpgradeRequest) (*entity.ParticipantFATE, *sync.WaitGroup, error) {
	// TODO
	return nil, nil, nil

	federationUUID := req.FederationUUID
	instance, err := s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationFATE)

	if exist, err := s.ParticipantFATERepo.IsExchangeCreatedByFederationUUID(federationUUID); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check exchange existence status")
	} else if exist {
		return nil, nil, errors.Errorf("an exchange is already deployed in federation %s", federationUUID)
	}

	if err := s.EndpointService.TestKubeFATE(req.EndpointUUID); err != nil {
		return nil, nil, err
	}

	instance, err = s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "faile to get chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeFATEExchange {
		return nil, nil, errors.Errorf("chart %s is not for FATE exchange deployment", chart.UUID)
	}

	var m map[string]interface{}
	err = yaml.Unmarshal([]byte(req.DeploymentYAML), &m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to unmarshal deployment yaml")
	}

	m["name"] = toDeploymentName(req.Name)
	m["namespace"] = req.Namespace

	finalYAMLBytes, err := yaml.Marshal(m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get final yaml content")
	}
	req.DeploymentYAML = string(finalYAMLBytes)

	if req.ProxyServerCertInfo.BindingMode == entity.CertBindingModeReuse {
		return nil, nil, errors.New("cannot re-use existing certificate")
	}

	var caCert *x509.Certificate
	if req.ProxyServerCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.FMLManagerServerCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.FMLManagerClientCertInfo.BindingMode == entity.CertBindingModeCreate {
		ca, err := s.CertificateService.DefaultCA()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get default CA")
		}
		caCert, err = ca.RootCert()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get CA cert")
		}
	}

	exchange := &entity.ParticipantFATE{
		Participant: entity.Participant{
			Name:           req.Name,
			Description:    req.Description,
			FederationUUID: req.FederationUUID,
			EndpointUUID:   req.EndpointUUID,
			ChartUUID:      req.ChartUUID,
			Namespace:      req.Namespace,
			DeploymentYAML: req.DeploymentYAML,
			IsManaged:      true,
			ExtraAttribute: entity.ParticipantExtraAttribute{
				IsNewNamespace:    false,
				UseRegistrySecret: req.RegistryConfig.UseRegistrySecret,
			},
		},
		Type:       entity.ParticipantFATETypeExchange,
		PartyID:    0,
		Status:     entity.ParticipantFATEStatusInstalling,
		AccessInfo: entity.ParticipantFATEModulesAccessMap{},
		CertConfig: entity.ParticipantFATECertConfig{
			ProxyServerCertInfo:      req.ProxyServerCertInfo,
			FMLManagerClientCertInfo: req.FMLManagerClientCertInfo,
			FMLManagerServerCertInfo: req.FMLManagerServerCertInfo,
		},
	}
	err = s.ParticipantFATERepo.Create(exchange)
	if err != nil {
		return nil, nil, err
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, "start creating exchange", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "installing fate exchange").Str("uuid", exchange.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("creating FATE exchange %s with UUID %s", exchange.Name, exchange.UUID)
		if err := func() error {
			endpointMgr, kfClient, kfClientCloser, err := s.buildKubeFATEMgrAndClient(req.EndpointUUID)
			if kfClientCloser != nil {
				defer kfClientCloser()
			}
			if err != nil {
				return err
			}

			if chart.Private {
				operationLog.Info().Msgf("making sure the chart is uploaded, name: %s, version: %s", chart.ChartName, chart.Version)
				if err := kfClient.EnsureChartExist(chart.ChartName, chart.Version, chart.ArchiveContent); err != nil {
					return errors.Wrapf(err, "error uploading FedLCM private chart")
				}
			}

			if exchange.ExtraAttribute.IsNewNamespace, err = ensureNSExisting(endpointMgr.K8sClient(), req.Namespace); err != nil {
				return err
			}
			if exchange.ExtraAttribute.IsNewNamespace {
				operationLog.Info().Msgf("created new namespace: %s", req.Namespace)
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
					return errors.Wrap(err, "failed to update exchange attribute")
				}
			} else {
				if clusters, err := kfClient.ListClusterByNamespace(req.Namespace); err != nil {
					return errors.Wrap(err, "failed to check existing fate or exchange installation")
				} else if len(clusters) > 0 {
					return errors.Errorf("cannot add new exchange as existing installations are running in the same namespace: %v", *clusters[0])
				}
			}

			if req.RegistryConfig.UseRegistrySecret {
				if err := createRegistrySecret(endpointMgr.K8sClient(), imagePullSecretsNameFATE, req.Namespace, req.RegistryConfig.RegistrySecretConfig); err != nil {
					return errors.Wrap(err, "failed to create registry secret")
				} else {
					operationLog.Info().Msgf("created registry secret %s with username %s for URL %s", imagePullSecretsNameFATE, req.RegistryConfig.RegistrySecretConfig.Username, req.RegistryConfig.RegistrySecretConfig.ServerURL)
				}
			}

			atsFQDN := req.ProxyServerCertInfo.CommonName
			if atsFQDN == "" {
				atsFQDN = fmt.Sprintf("proxy.%s", federation.Domain)
			}
			if req.ProxyServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.ProxyServerCertInfo.CommonName == "" {
					req.ProxyServerCertInfo.CommonName = atsFQDN
				}
				operationLog.Info().Msgf("creating certificate for the ATS service with CN: %s", req.ProxyServerCertInfo.CommonName)
				dnsNames := []string{req.ProxyServerCertInfo.CommonName}
				cert, pk, err := s.CertificateService.CreateCertificateSimple(req.ProxyServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create ATS certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", cert.SerialNumber, cert.Subject.CommonName)
				err = createATSSecret(endpointMgr.K8sClient(), req.Namespace, caCert, cert, pk)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(cert, entity.CertificateBindingServiceTypeATS, exchange.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				exchange.CertConfig.ProxyServerCertInfo.CommonName = req.ProxyServerCertInfo.CommonName
				exchange.CertConfig.ProxyServerCertInfo.UUID = cert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
					return errors.Wrap(err, "failed to update exchange cert info")
				}
			}

			fmlManagerFQDN := req.FMLManagerServerCertInfo.CommonName
			if fmlManagerFQDN == "" {
				fmlManagerFQDN = fmt.Sprintf("fmlmanager.server.%s", federation.Domain)
			}
			var serverCert, clientCert *entity.Certificate
			var serverPrivateKey, clientPrivateKey *rsa.PrivateKey
			if req.FMLManagerServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.FMLManagerServerCertInfo.CommonName == "" {
					req.FMLManagerServerCertInfo.CommonName = fmt.Sprintf("fmlmanager.server.%s", federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for FML Manager server with CN: %s", req.FMLManagerServerCertInfo.CommonName)
				dnsNames := []string{req.FMLManagerServerCertInfo.CommonName, "localhost"}
				serverCert, serverPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.FMLManagerServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create FML Manager server certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", serverCert.SerialNumber, serverCert.Subject.CommonName)
			}
			if req.FMLManagerClientCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.FMLManagerClientCertInfo.CommonName == "" {
					req.FMLManagerClientCertInfo.CommonName = fmt.Sprintf("fmlmanager.client.%s", federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for FML Manager client with CN: %s", req.FMLManagerClientCertInfo.CommonName)
				dnsNames := []string{req.FMLManagerClientCertInfo.CommonName}
				clientCert, clientPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.FMLManagerClientCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create FML Manager client certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", clientCert.SerialNumber, clientCert.Subject.CommonName)
			}
			if req.FMLManagerServerCertInfo.BindingMode != entity.CertBindingModeSkip && req.FMLManagerClientCertInfo.BindingMode != entity.CertBindingModeSkip {
				err = createTLSSecret(endpointMgr.K8sClient(), req.Namespace, serverCert, serverPrivateKey, clientCert, clientPrivateKey, caCert, entity.ParticipantFATESecretNameFMLMgr)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(serverCert, entity.CertificateBindingServiceFMLManagerServer, exchange.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				exchange.CertConfig.FMLManagerServerCertInfo.CommonName = req.FMLManagerServerCertInfo.CommonName
				exchange.CertConfig.FMLManagerServerCertInfo.UUID = serverCert.UUID
				if err := s.CertificateService.CreateBinding(clientCert, entity.CertificateBindingServiceFMLManagerClient, exchange.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				exchange.CertConfig.FMLManagerClientCertInfo.CommonName = req.FMLManagerClientCertInfo.CommonName
				exchange.CertConfig.FMLManagerClientCertInfo.UUID = clientCert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
					return errors.Wrap(err, "failed to update exchange cert info")
				}
			}

			jobUUID, err := kfClient.SubmitClusterInstallationJob(exchange.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "fail to submit cluster creation request")
			}
			exchange.JobUUID = jobUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
				return errors.Wrap(err, "failed to update exchange's job uuid")
			}
			operationLog.Info().Msgf("kubefate job created, uuid: %s", exchange.JobUUID)
			clusterUUID, err := kfClient.WaitClusterUUID(jobUUID)
			if err != nil {
				return errors.Wrapf(err, "fail to get cluster uuid")
			}
			exchange.ClusterUUID = clusterUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
				return errors.Wrap(err, "failed to update exchange cluster uuid")
			}
			operationLog.Info().Msgf("kubefate-managed cluster created, uuid: %s", exchange.ClusterUUID)

			job, err := kfClient.WaitJob(jobUUID)
			if err != nil {
				return err
			}
			if job.Status != modules.JobStatusSuccess {
				return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
			}
			operationLog.Info().Msgf("kubefate job succeeded")

			serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameNginx), "http")
			if err != nil {
				return errors.Wrapf(err, "fail to get nginx access info")
			}
			exchange.AccessInfo[entity.ParticipantFATEServiceNameNginx] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         false,
			}

			serviceType, host, port, err = getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameATS), "443")
			if err != nil {
				return errors.Wrapf(err, "fail to get ATS access info")
			}
			exchange.AccessInfo[entity.ParticipantFATEServiceNameATS] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         true,
				FQDN:        atsFQDN,
			}

			moduleList := m["modules"].([]interface{})
			for _, moduleName := range moduleList {
				if moduleName.(string) == "fmlManagerServer" {
					operationLog.Info().Msgf("fml-manager deployed, retrieving access info")
					serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameFMLMgr), "https-fml-manager-server")
					if err != nil {
						return errors.Wrapf(err, "fail to get fml manager access info")
					}
					exchange.AccessInfo[entity.ParticipantFATEServiceNameFMLMgr] = entity.ParticipantModulesAccess{
						ServiceType: serviceType,
						Host:        host,
						Port:        port,
						TLS:         true,
						FQDN:        fmlManagerFQDN,
					}
				}
			}

			exchange.Status = entity.ParticipantFATEStatusActive
			if err := s.BuildIngressInfoMap(exchange); err != nil {
				return errors.Wrapf(err, "failed to get ingress info")
			}
			return s.ParticipantFATERepo.UpdateInfoByUUID(exchange)
		}(); err != nil {
			operationLog.Error().Msgf(errors.Wrapf(err, "failed to install FATE exchange").Error())
			exchange.Status = entity.ParticipantFATEStatusFailed
			if updateErr := s.ParticipantFATERepo.UpdateStatusByUUID(exchange); updateErr != nil {
				operationLog.Error().Msgf(errors.Wrapf(updateErr, "failed to update FATE exchange status").Error())
			}
			return
		}
		operationLog.Info().Msgf("FATE exchange %s(%s) deployed", exchange.Name, exchange.UUID)
	}()

	return exchange, wg, nil
}

// UpgradeCluster upgrade a FATE cluster with exchange's access info, and will update exchange's route table
func (s *ParticipantFATEService) UpgradeCluster(req *ParticipantFATEClusterUpgradeRequest) (*entity.ParticipantFATE, *sync.WaitGroup, error) {
	// TODO
	return nil, nil, nil

	if err := s.CheckPartyIDConflict(req.FederationUUID, req.PartyID); err != nil {
		return nil, nil, err
	}
	federationUUID := req.FederationUUID
	instance, err := s.FederationRepo.GetByUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting federation info")
	}
	federation := instance.(*entity.FederationFATE)

	instance, err = s.ParticipantFATERepo.GetExchangeByFederationUUID(federationUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to check exchange existence status")
	}
	exchange := instance.(*entity.ParticipantFATE)

	if exchange.Status != entity.ParticipantFATEStatusActive {
		return nil, nil, errors.Errorf("exchange %v is not in active status", exchange.UUID)
	}
	if exchange.IsManaged {
		if err := s.EndpointService.TestKubeFATE(exchange.EndpointUUID); err != nil {
			return nil, nil, err
		}
	}

	if err := s.EndpointService.TestKubeFATE(req.EndpointUUID); err != nil {
		return nil, nil, err
	}

	instance, err = s.ChartRepo.GetByUUID(req.ChartUUID)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "faile to get chart")
	}
	chart := instance.(*entity.Chart)
	if chart.Type != entity.ChartTypeFATECluster {
		return nil, nil, errors.Errorf("chart %s is not for FATE cluster deployment", chart.UUID)
	}

	var m map[string]interface{}
	err = yaml.Unmarshal([]byte(req.DeploymentYAML), &m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to unmarshal deployment yaml")
	}

	m["name"] = toDeploymentName(req.Name)
	m["namespace"] = req.Namespace

	finalYAMLBytes, err := yaml.Marshal(m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get final yaml content")
	}
	req.DeploymentYAML = string(finalYAMLBytes)

	pulsarDomain, err := getPulsarDomainFromYAML(req.DeploymentYAML)
	if err != nil {
		return nil, nil, err
	}
	if pulsarDomain == "" {
		log.Warn().Msgf("using federation domain")
		pulsarDomain = federation.Domain
	}
	pulsarFQDN := fmt.Sprintf("%d.%s", req.PartyID, pulsarDomain)

	var caCert *x509.Certificate
	if req.PulsarServerCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.SitePortalClientCertInfo.BindingMode == entity.CertBindingModeCreate ||
		req.SitePortalServerCertInfo.BindingMode == entity.CertBindingModeCreate {
		ca, err := s.CertificateService.DefaultCA()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get default CA")
		}
		caCert, err = ca.RootCert()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed to get CA cert")
		}
	}

	cluster := &entity.ParticipantFATE{
		Participant: entity.Participant{
			Name:           req.Name,
			Description:    req.Description,
			FederationUUID: req.FederationUUID,
			EndpointUUID:   req.EndpointUUID,
			ChartUUID:      req.ChartUUID,
			Namespace:      req.Namespace,
			DeploymentYAML: req.DeploymentYAML,
			IsManaged:      true,
			ExtraAttribute: entity.ParticipantExtraAttribute{
				IsNewNamespace:    false,
				UseRegistrySecret: req.RegistryConfig.UseRegistrySecret,
			},
		},
		Type:       entity.ParticipantFATETypeCluster,
		PartyID:    req.PartyID,
		Status:     entity.ParticipantFATEStatusInstalling,
		AccessInfo: entity.ParticipantFATEModulesAccessMap{},
		CertConfig: entity.ParticipantFATECertConfig{
			PulsarServerCertInfo:     req.PulsarServerCertInfo,
			SitePortalServerCertInfo: req.SitePortalServerCertInfo,
			SitePortalClientCertInfo: req.SitePortalClientCertInfo,
		},
	}
	err = s.ParticipantFATERepo.Create(cluster)
	if err != nil {
		return nil, nil, err
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, "start creating cluster", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "installing fate cluster").Str("uuid", cluster.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("creating FATE cluster %s with UUID %s", cluster.Name, cluster.UUID)
		if err := func() error {
			endpointMgr, kfClient, closer, err := s.buildKubeFATEMgrAndClient(req.EndpointUUID)
			if closer != nil {
				defer closer()
			}
			if err != nil {
				return err
			}
			if chart.Private {
				operationLog.Info().Msgf("making sure the chart is uploaded, name: %s, version: %s", chart.ChartName, chart.Version)
				if err := kfClient.EnsureChartExist(chart.ChartName, chart.Version, chart.ArchiveContent); err != nil {
					return errors.Wrapf(err, "error uploading FedLCM private chart")
				}
			}

			if cluster.ExtraAttribute.IsNewNamespace, err = ensureNSExisting(endpointMgr.K8sClient(), req.Namespace); err != nil {
				return err
			}
			if cluster.ExtraAttribute.IsNewNamespace {
				operationLog.Info().Msgf("created new namespace: %s for this cluster", req.Namespace)
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
					return errors.Wrap(err, "failed to update cluster attribute")
				}
			} else {
				if clusters, err := kfClient.ListClusterByNamespace(req.Namespace); err != nil {
					return errors.Wrap(err, "failed to check existing fate or exchange installation")
				} else if len(clusters) > 0 {
					return errors.Errorf("cannot add new fate cluster as existing installations are running in the same namespace: %v", *clusters[0])
				}
			}

			if req.RegistryConfig.UseRegistrySecret {
				if err := createRegistrySecret(endpointMgr.K8sClient(), imagePullSecretsNameFATE, req.Namespace, req.RegistryConfig.RegistrySecretConfig); err != nil {
					return errors.Wrap(err, "failed to create registry secret")
				} else {
					operationLog.Info().Msgf("created registry secret %s with username %s for URL %s", imagePullSecretsNameFATE, req.RegistryConfig.RegistrySecretConfig.Username, req.RegistryConfig.RegistrySecretConfig.ServerURL)
				}
			}

			if req.PulsarServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				req.PulsarServerCertInfo.CommonName = fmt.Sprintf("%v.%s", req.PartyID, federation.Domain)
				operationLog.Info().Msgf("creating certificate for the pulsar service with CN: %s", req.PulsarServerCertInfo.CommonName)
				dnsNames := []string{req.PulsarServerCertInfo.CommonName}
				cert, pk, err := s.CertificateService.CreateCertificateSimple(req.PulsarServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create pulsar certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", cert.SerialNumber, cert.Subject.CommonName)
				err = createPulsarSecret(endpointMgr.K8sClient(), req.Namespace, caCert, cert, pk)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(cert, entity.CertificateBindingServiceTypePulsarServer, cluster.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				cluster.CertConfig.PulsarServerCertInfo.CommonName = req.PulsarServerCertInfo.CommonName
				cluster.CertConfig.PulsarServerCertInfo.UUID = cert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
					return errors.Wrap(err, "failed to update cluster cert info")
				}
			}

			sitePortalFQDN := fmt.Sprintf("site-%v.server.%s", req.PartyID, federation.Domain)
			var serverCert, clientCert *entity.Certificate
			var serverPrivateKey, clientPrivateKey *rsa.PrivateKey
			if req.SitePortalServerCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.SitePortalServerCertInfo.CommonName == "" {
					req.SitePortalServerCertInfo.CommonName = fmt.Sprintf("site-%v.server.%s", req.PartyID, federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for Site Portal server with CN: %s", req.SitePortalServerCertInfo.CommonName)
				// we need localhost in the dnsNames because the site portal will call itself during some workflows
				dnsNames := []string{req.SitePortalServerCertInfo.CommonName, "localhost"}
				serverCert, serverPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.SitePortalServerCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create Site Portal server certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", serverCert.SerialNumber, serverCert.Subject.CommonName)
			}
			if req.SitePortalClientCertInfo.BindingMode == entity.CertBindingModeCreate {
				if req.SitePortalClientCertInfo.CommonName == "" {
					req.SitePortalClientCertInfo.CommonName = fmt.Sprintf("site-%v.client.%s", req.PartyID, federation.Domain)
				}
				operationLog.Info().Msgf("creating certificate for Site Portal client with CN: %s", req.SitePortalClientCertInfo.CommonName)
				dnsNames := []string{req.SitePortalClientCertInfo.CommonName}
				clientCert, clientPrivateKey, err = s.CertificateService.CreateCertificateSimple(req.SitePortalClientCertInfo.CommonName, defaultCertLifetime, dnsNames)
				if err != nil {
					return errors.Wrapf(err, "failed to create Site Portal client certificate")
				}
				operationLog.Info().Msgf("got certificate with serial number: %v for CN: %s", clientCert.SerialNumber, clientCert.Subject.CommonName)
			}
			if req.SitePortalServerCertInfo.BindingMode != entity.CertBindingModeSkip && req.SitePortalClientCertInfo.BindingMode != entity.CertBindingModeSkip {
				err = createTLSSecret(endpointMgr.K8sClient(), req.Namespace, serverCert, serverPrivateKey, clientCert, clientPrivateKey, caCert, entity.ParticipantFATESecretNamePortal)
				if err != nil {
					return err
				}
				if err := s.CertificateService.CreateBinding(serverCert, entity.CertificateBindingServiceSitePortalServer, cluster.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				cluster.CertConfig.SitePortalServerCertInfo.CommonName = req.SitePortalServerCertInfo.CommonName
				cluster.CertConfig.SitePortalServerCertInfo.UUID = serverCert.UUID
				if err := s.CertificateService.CreateBinding(clientCert, entity.CertificateBindingServiceSitePortalClient, cluster.UUID, federationUUID, entity.FederationTypeFATE); err != nil {
					return err
				}
				cluster.CertConfig.SitePortalClientCertInfo.CommonName = req.SitePortalClientCertInfo.CommonName
				cluster.CertConfig.SitePortalClientCertInfo.UUID = clientCert.UUID
				if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
					return errors.Wrap(err, "failed to update cluster cert info")
				}
				operationLog.Info().Msgf("created cert secret and bindings for site-portal service")
			}

			// TODO: ensure there is no same-name cluster exists because we may create cluster-level resources using the cluster name
			jobUUID, err := kfClient.SubmitClusterInstallationJob(cluster.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "fail to submit cluster creation request")
			}
			cluster.JobUUID = jobUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
				return errors.Wrap(err, "failed to update cluster's job uuid")
			}
			operationLog.Info().Msgf("kubefate job created, uuid: %s", cluster.JobUUID)
			clusterUUID, err := kfClient.WaitClusterUUID(jobUUID)
			if err != nil {
				return errors.Wrapf(err, "fail to get cluster uuid")
			}
			cluster.ClusterUUID = clusterUUID
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
				return errors.Wrap(err, "failed to update cluster uuid")
			}
			operationLog.Info().Msgf("kubefate-managed cluster created, uuid: %s", cluster.ClusterUUID)
			job, err := kfClient.WaitJob(jobUUID)
			if err != nil {
				return err
			}
			if job.Status != modules.JobStatusSuccess {
				return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
			}

			serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNameNginx), "http")
			if err != nil {
				return errors.Wrapf(err, "fail to get nginx access info")
			}
			cluster.AccessInfo[entity.ParticipantFATEServiceNameNginx] = entity.ParticipantModulesAccess{
				ServiceType: serviceType,
				Host:        host,
				Port:        port,
				TLS:         false,
			}
			operationLog.Info().Msgf("kubefate job succeeded")

			if req.PulsarServerCertInfo.BindingMode != entity.CertBindingModeSkip {
				// the pulsar-public-tls service is always of type LoadBalancer, we try to use nodePort if no LoadBalancer IP is available
				serviceType, host, port, err = getServiceAccessWithFallback(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNamePulsar), "tls-port", true)
				if err != nil {
					return errors.Wrapf(err, "fail to get pulsar access info")
				}
				cluster.AccessInfo[entity.ParticipantFATEServiceNamePulsar] = entity.ParticipantModulesAccess{
					ServiceType: serviceType,
					Host:        host,
					Port:        port,
					TLS:         true,
					FQDN:        pulsarFQDN,
				}
			} else {
				pulsarHost, pulsarSSLPort, err := getPulsarInformationFromYAML(req.DeploymentYAML)
				if err != nil {
					return errors.Wrapf(err, "fail to get pulsar access info")
				}
				cluster.AccessInfo[entity.ParticipantFATEServiceNamePulsar] = entity.ParticipantModulesAccess{
					ServiceType: "External",
					Host:        pulsarHost,
					Port:        pulsarSSLPort,
					TLS:         true,
					FQDN:        pulsarFQDN,
				}
			}

			moduleList := m["modules"].([]interface{})
			for _, moduleName := range moduleList {
				if moduleName.(string) == "frontend" {
					operationLog.Info().Msgf("site-portal deployed, retrieving access info")
					serviceType, host, port, err := getServiceAccess(endpointMgr.K8sClient(), req.Namespace, string(entity.ParticipantFATEServiceNamePortal), "https-frontend")
					if err != nil {
						return errors.Wrapf(err, "fail to get site portal access info")
					}
					cluster.AccessInfo[entity.ParticipantFATEServiceNamePortal] = entity.ParticipantModulesAccess{
						ServiceType: serviceType,
						Host:        host,
						Port:        port,
						TLS:         true,
						FQDN:        sitePortalFQDN,
					}
					if fmlManagerInfo, ok := exchange.AccessInfo[entity.ParticipantFATEServiceNameFMLMgr]; ok {
						go func() {
							operationLog.Info().Msgf("automatically configure site portal and its connection with fml manager")
							password, err := cluster.GetSitePortalAdminPassword()
							if err != nil {
								operationLog.Err(err).Msg("failed to get site portal admin password")
								return
							}
							sitePortalClient, err := site_portal_client.NewClient(site_portal_client.Site{
								Username:             "Admin",
								Password:             password,
								Name:                 cluster.Name,
								Description:          cluster.Description,
								PartyID:              uint(cluster.PartyID),
								ExternalHost:         host,
								ExternalPort:         uint(port),
								HTTPS:                true,
								FMLManagerEndpoint:   fmt.Sprintf("https://%s:%d", fmlManagerInfo.Host, fmlManagerInfo.Port),
								FMLManagerServerName: fmlManagerInfo.FQDN,
								FATEFlowHost:         "fateflow",
								FATEFlowHTTPPort:     9380,
							})
							if err != nil {
								operationLog.Err(err).Msg("failed to get site portal client instance")
								return
							}
							if err := sitePortalClient.ConfigAndConnectSite(); err != nil {
								operationLog.Err(err).Msg("failed to configure site portal")
								return
							}
							operationLog.Info().Msg("configured site portal and it is connected to fml manager")
						}()
					} else {
						operationLog.Warn().Msgf("fml manager is not installed in the exchange, skipping configuration of site portal")
					}
				}
			}

			cluster.Status = entity.ParticipantFATEStatusActive
			if err := s.BuildIngressInfoMap(cluster); err != nil {
				return errors.Wrapf(err, "failed to get ingress info")
			}
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); err != nil {
				return errors.Wrap(err, "failed to save cluster info")
			}
			if exchange.IsManaged {
				operationLog.Info().Msg("rebuilding exchange route table")
				if err := s.rebuildRouteTable(exchange); err != nil {
					operationLog.Error().Msg(errors.Wrap(err, "error rebuilding route table while creating cluster").Error())
				}
			}
			return nil
		}(); err != nil {
			operationLog.Error().Msgf(errors.Wrap(err, "failed to install FATE cluster").Error())
			cluster.Status = entity.ParticipantFATEStatusFailed
			if updateErr := s.ParticipantFATERepo.UpdateStatusByUUID(cluster); updateErr != nil {
				operationLog.Error().Msgf(errors.Wrap(err, "failed to update FATE cluster status").Error())
			}
			return
		}
		operationLog.Info().Msgf("FATE cluster %s(%s) deployed", cluster.Name, cluster.UUID)
	}()
	return cluster, wg, nil
}
