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
	"sync"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/utils"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/modules"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sigs.k8s.io/yaml"
)

// ParticipantFATEExchangeUpgradeRequest is the exchange upgrade request
type ParticipantFATEExchangeUpgradeRequest struct {
	ExchangeUUID   string `json:"exchange_uuid"`
	FederationUUID string `json:"federation_uuid"`
	UpgradeVersion string `json:"upgrade_version"`
}

// ParticipantFATEClusterUpgradeRequest is the cluster upgrade request
type ParticipantFATEClusterUpgradeRequest struct {
	ClusterUUID    string `json:"cluster_uuid"`
	FederationUUID string `json:"federation_uuid"`
	UpgradeVersion string `json:"upgrade_version"`
}

// UpgradeExchange upgrade the FATE exchange, the returned *sync.WaitGroup can be used to wait for the completion of the async goroutine
func (s *ParticipantFATEService) UpgradeExchange(req *ParticipantFATEExchangeUpgradeRequest) (*entity.ParticipantFATE, *sync.WaitGroup, error) {

	participantInstance, err := s.ParticipantFATERepo.GetByUUID(req.ExchangeUUID)
	if err != nil {
		return nil, nil, err
	}
	exchange := participantInstance.(*entity.ParticipantFATE)

	//Check whether it is a cluster managed by fedlcm, a cluster not managed by fedlcm cannot be upgraded
	if !exchange.IsManaged {
		return nil, nil, errors.New("The cluster not managed by FedLCM cannot be upgraded.")
	}

	if err := s.EndpointService.TestKubeFATE(exchange.EndpointUUID); err != nil {
		return nil, nil, err
	}

	ClusterChartVersion := utils.GetChartVersionFromDeploymentYAML(exchange.DeploymentYAML)
	ClusterChartName := utils.GetChartNameFromDeploymentYAML(exchange.DeploymentYAML)

	// checkUpgradeable
	if utils.CompareVersion(ClusterChartVersion, req.UpgradeVersion) >= 0 {
		return nil, nil, errors.Errorf("the version passed in cannot be upgraded, currentVersion %s, upgradeVersion: %s", ClusterChartVersion, req.UpgradeVersion)
	}

	instance, err := s.ChartRepo.GetByNameAndVersion(ClusterChartName, req.UpgradeVersion)
	if err != nil {
		log.Err(err).Strs("chartName and version", []string{ClusterChartName, req.UpgradeVersion}).Msg("GetByNameAndVersion err")
		return nil, nil, errors.Wrapf(err, "faile to get chart")
	}
	upgradeChart := instance.(*entity.Chart)
	if upgradeChart.Type != entity.ChartTypeFATEExchange {
		return nil, nil, errors.Errorf("chart %s is not for FATE exchange deployment", upgradeChart.UUID)
	}

	var m map[string]interface{}
	err = yaml.Unmarshal([]byte(exchange.DeploymentYAML), &m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to unmarshal deployment yaml")
	}

	m["chartVersion"] = upgradeChart.Version

	finalYAMLBytes, err := yaml.Marshal(m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get final yaml content")
	}
	previousDeploymentYAML := exchange.DeploymentYAML
	exchange.DeploymentYAML = string(finalYAMLBytes)
	log.Debug().Str("exchange.DeploymentYAML", exchange.DeploymentYAML).Msg("show DeploymentYAML")

	exchange.Status = entity.ParticipantFATEStatusUpgrading

	err = s.ParticipantFATERepo.UpdateInfoByUUID(exchange)
	if err != nil {
		return nil, nil, err
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, "start upgrading exchange", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "upgrading fate exchange").Str("uuid", exchange.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeExchange, exchange.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("upgrading FATE exchange %s with UUID %s", exchange.Name, exchange.UUID)
		if err := func() error {
			_, kfClient, kfClientCloser, err := s.buildKubeFATEMgrAndClient(exchange.EndpointUUID)
			if kfClientCloser != nil {
				defer kfClientCloser()
			}
			if err != nil {
				return err
			}

			if upgradeChart.Private {
				operationLog.Info().Msgf("making sure the chart is uploaded, name: %s, version: %s", upgradeChart.ChartName, upgradeChart.Version)
				if err := kfClient.EnsureChartExist(upgradeChart.ChartName, upgradeChart.Version, upgradeChart.ArchiveContent); err != nil {
					return errors.Wrapf(err, "error uploading FedLCM private chart")
				}
			}

			jobUUID, err := kfClient.SubmitClusterUpdateJob(exchange.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "fail to submit cluster upgrade request")
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
			if err := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); err != nil {
				return errors.Wrap(err, "failed to update exchange cluster uuid")
			}
			operationLog.Info().Msgf("kubefate-managed cluster upgraded, uuid: %s", exchange.ClusterUUID)

			job, err := kfClient.WaitJob(jobUUID)
			if err != nil {
				return err
			}
			if job.Status != modules.JobStatusSuccess {
				return errors.Errorf("job is %s, job info: %v", job.Status.String(), job)
			}
			exchange.ClusterUUID = clusterUUID
			exchange.ChartUUID = upgradeChart.UUID
			operationLog.Info().Msgf("kubefate job succeeded")

			exchange.Status = entity.ParticipantFATEStatusActive
			if err := s.BuildIngressInfoMap(exchange); err != nil {
				return errors.Wrapf(err, "failed to get ingress info")
			}
			return s.ParticipantFATERepo.UpdateInfoByUUID(exchange)
		}(); err != nil {
			operationLog.Error().Msgf(errors.Wrapf(err, "failed to upgrade FATE exchange").Error())
			// we still mark the exchange to be active as kubefate can roll back the failed upgrade
			exchange.Status = entity.ParticipantFATEStatusActive
			exchange.DeploymentYAML = previousDeploymentYAML
			if updateErr := s.ParticipantFATERepo.UpdateInfoByUUID(exchange); updateErr != nil {
				operationLog.Error().Msgf(errors.Wrapf(updateErr, "failed to update FATE exchange info").Error())
			}
			return
		}
		operationLog.Info().Msgf("FATE exchange %s(%s) upgraded", exchange.Name, exchange.UUID)
	}()

	return exchange, wg, nil
}

// UpgradeCluster upgrade a FATE cluster with exchange's access info, and will update exchange's route table
func (s *ParticipantFATEService) UpgradeCluster(req *ParticipantFATEClusterUpgradeRequest) (*entity.ParticipantFATE, *sync.WaitGroup, error) {

	instance, err := s.ParticipantFATERepo.GetExchangeByFederationUUID(req.FederationUUID)
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

	participantInstance, err := s.ParticipantFATERepo.GetByUUID(req.ClusterUUID)
	if err != nil {
		return nil, nil, err
	}
	cluster := participantInstance.(*entity.ParticipantFATE)

	//Check whether it is a cluster managed by fedlcm, a cluster not managed by fedlcm cannot be upgraded
	if !exchange.IsManaged {
		return nil, nil, errors.New("The cluster not managed by FedLCM cannot be upgraded.")
	}

	if err := s.EndpointService.TestKubeFATE(cluster.EndpointUUID); err != nil {
		return nil, nil, err
	}

	ClusterChartVersion := utils.GetChartVersionFromDeploymentYAML(cluster.DeploymentYAML)
	ClusterChartName := utils.GetChartNameFromDeploymentYAML(cluster.DeploymentYAML)

	// checkUpgradeable
	if utils.CompareVersion(ClusterChartVersion, req.UpgradeVersion) >= 0 {
		return nil, nil, errors.Errorf("the version passed in cannot be upgraded, currentVersion %s, upgradeVersion: %s", ClusterChartVersion, req.UpgradeVersion)
	}

	instance, err = s.ChartRepo.GetByNameAndVersion(ClusterChartName, req.UpgradeVersion)
	if err != nil {
		log.Err(err).Strs("chartName and version", []string{ClusterChartName, req.UpgradeVersion}).Msg("GetByNameAndVersion err")
		return nil, nil, errors.Wrapf(err, "faile to get chart")
	}
	upgradeChart := instance.(*entity.Chart)
	if upgradeChart.Type != entity.ChartTypeFATECluster {
		return nil, nil, errors.Errorf("chart %s is not for FATE deployment", upgradeChart.UUID)
	}

	var m map[string]interface{}
	err = yaml.Unmarshal([]byte(cluster.DeploymentYAML), &m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to unmarshal deployment yaml")
	}

	m["chartVersion"] = upgradeChart.Version

	finalYAMLBytes, err := yaml.Marshal(m)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to get final yaml content")
	}
	previousDeploymentYAML := cluster.DeploymentYAML
	cluster.DeploymentYAML = string(finalYAMLBytes)
	log.Debug().Str("cluster.DeploymentYAML", cluster.DeploymentYAML).Msg("show DeploymentYAML")

	cluster.Status = entity.ParticipantFATEStatusUpgrading

	err = s.ParticipantFATERepo.UpdateInfoByUUID(cluster)
	if err != nil {
		return nil, nil, err
	}

	_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, "start upgrading cluster", entity.EventLogLevelInfo)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		operationLog := log.Logger.With().Timestamp().Str("action", "upgrading fate cluster").Str("uuid", cluster.UUID).Logger().
			Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, message string) {
				eventLvl := entity.EventLogLevelInfo
				if level == zerolog.ErrorLevel {
					eventLvl = entity.EventLogLevelError
				}
				_ = s.EventService.CreateEvent(entity.EventTypeLogMessage, entity.EntityTypeCluster, cluster.UUID, message, eventLvl)
			}))
		operationLog.Info().Msgf("upgrading FATE cluster %s with UUID %s", cluster.Name, cluster.UUID)
		if err := func() error {
			_, kfClient, closer, err := s.buildKubeFATEMgrAndClient(cluster.EndpointUUID)
			if closer != nil {
				defer closer()
			}
			if err != nil {
				return err
			}
			if upgradeChart.Private {
				operationLog.Info().Msgf("making sure the chart is uploaded, name: %s, version: %s", upgradeChart.ChartName, upgradeChart.Version)
				if err := kfClient.EnsureChartExist(upgradeChart.ChartName, upgradeChart.Version, upgradeChart.ArchiveContent); err != nil {
					return errors.Wrapf(err, "error uploading FedLCM private chart")
				}
			}

			jobUUID, err := kfClient.SubmitClusterUpdateJob(cluster.DeploymentYAML)
			if err != nil {
				return errors.Wrapf(err, "fail to submit cluster upgrade request")
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

			cluster.ClusterUUID = clusterUUID
			cluster.ChartUUID = upgradeChart.UUID
			operationLog.Info().Msgf("kubefate job succeeded")

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
					operationLog.Error().Msg(errors.Wrap(err, "error rebuilding route table while upgrade cluster").Error())
				}
			}
			return nil
		}(); err != nil {
			operationLog.Error().Msgf(errors.Wrap(err, "failed to upgrade FATE cluster").Error())
			// we still mark the cluster to be active as kubefate can roll back the failed upgrade
			cluster.Status = entity.ParticipantFATEStatusActive
			cluster.DeploymentYAML = previousDeploymentYAML
			if updateErr := s.ParticipantFATERepo.UpdateInfoByUUID(cluster); updateErr != nil {
				operationLog.Error().Msgf(errors.Wrap(err, "failed to update FATE cluster info").Error())
			}
			return
		}
		operationLog.Info().Msgf("FATE cluster %s(%s) upgraded", cluster.Name, cluster.UUID)
	}()
	return cluster, wg, nil
}
