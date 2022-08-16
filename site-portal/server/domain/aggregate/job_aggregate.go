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

package aggregate

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fateclient"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fateclient/template"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fmlmanager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// JobAggregate holds the job info and all the joined participants
type JobAggregate struct {
	Job                      *entity.Job
	Initiator                *entity.JobParticipant
	Participants             map[string]*entity.JobParticipant
	JobRepo                  repo.JobRepository
	ParticipantRepo          repo.JobParticipantRepository
	FMLManagerConnectionInfo FMLManagerConnectionInfo
	JobContext               JobContext
}

// JobContext contains necessary context for working with a job
type JobContext struct {
	AutoApprovalEnabled bool
	CurrentSiteUUID     string
}

// GenerateReaderConfigMaps returns maps whose key is index, and values are the reader's configurations, in
// specific, the table name and namespaces for each party.
func (aggregate *JobAggregate) GenerateReaderConfigMaps(hostUuidList []string) (hostMap,
	guestMap map[string]interface{}) {
	hostMap = map[string]interface{}{}
	guestMap = map[string]interface{}{}
	guestMap["0"] = map[string]interface{}{}
	guestMap["0"].(map[string]interface{})["reader_0"] = map[string]map[string]string{
		"table": {
			"name":      aggregate.Initiator.DataTableName,
			"namespace": aggregate.Initiator.DataTableNamespace,
		},
	}
	// The reason why we need the hostUuidList from the JobSubmissionRequest is that list can retain the sequence while
	// map cannot. Here if we only iterate the aggregate.Participants to generate the table name and namespace. The
	// sequence could be wrong and can mismatch the other configs we generate from the DAG. The good news is, the DAG's
	// hosts sequence matches the hostUuidList. So here we just need to iterate hostUuidList then things will be right.
	for i, hostUuid := range hostUuidList {
		hostMap[strconv.Itoa(i)] = map[string]interface{}{}
		hostMap[strconv.Itoa(i)].(map[string]interface{})["reader_0"] = map[string]map[string]string{
			"table": {
				"name":      aggregate.Participants[hostUuid].DataTableName,
				"namespace": aggregate.Participants[hostUuid].DataTableNamespace,
			},
		}
	}
	return hostMap, guestMap
}

// GenerateGeneralTrainingConf returns a string which contains the general conf information of a training job,
// including "dsl_version", "initiator", "role" and "job_parameters".
func (aggregate *JobAggregate) GenerateGeneralTrainingConf(hostUuidList []string) (string, error) {
	param := template.GeneralTrainingParam{
		Guest: template.PartyDataInfo{
			PartyID: strconv.Itoa(int(aggregate.Initiator.SitePartyID)),
		},
		Hosts: nil,
	}
	for _, hostUuid := range hostUuidList {
		param.Hosts = append(param.Hosts, template.PartyDataInfo{
			PartyID: strconv.Itoa(int(aggregate.Participants[hostUuid].SitePartyID)),
		})
	}
	generalConfJson, err := template.BuildTrainingConfGeneralStr(param)
	if err != nil {
		return "", err
	} else {
		return generalConfJson, nil
	}
}

// GenerateConfig returns the FATE job conf and dsl
func (aggregate *JobAggregate) GenerateConfig() (string, string, error) {
	switch aggregate.Job.Type {
	case entity.JobTypeTraining:
		return aggregate.generateTrainingConfig()
	case entity.JobTypePredict:
		return aggregate.generatePredictingConfig()
	case entity.JobTypePSI:
		return aggregate.generatePSIConfig()
	}
	return "", "", errors.New("unsupported job type")
}

func (aggregate *JobAggregate) generatePredictingConfig() (string, string, error) {
	switch aggregate.Job.AlgorithmType {
	case entity.JobAlgorithmTypeHomoLR, entity.JobAlgorithmTypeHomoSBT:
		if len(aggregate.Participants) > 0 {
			return "", "", errors.New("horizontal predicting job cannot have other participants")
		}
		param := template.HomoPredictingParam{
			Role:         string(aggregate.Initiator.SiteRole),
			ModelID:      aggregate.Job.FATEModelID,
			ModelVersion: aggregate.Job.FATEModelVersion,
			PartyDataInfo: template.PartyDataInfo{
				PartyID:        strconv.Itoa(int(aggregate.Initiator.SitePartyID)),
				TableName:      aggregate.Initiator.DataTableName,
				TableNamespace: aggregate.Initiator.DataTableNamespace,
			},
		}
		return template.BuildHomoPredictingConf(param)
	}
	return "", "", errors.Errorf("invalied algorithm type: %d", aggregate.Job.AlgorithmType)
}

func (aggregate *JobAggregate) generatePSIConfig() (string, string, error) {
	param := template.PSIParam{
		Guest: template.PartyDataInfo{
			PartyID:        strconv.Itoa(int(aggregate.Initiator.SitePartyID)),
			TableName:      aggregate.Initiator.DataTableName,
			TableNamespace: aggregate.Initiator.DataTableNamespace,
		},
		Hosts: nil,
	}
	for _, host := range aggregate.Participants {
		param.Hosts = append(param.Hosts, template.PartyDataInfo{
			PartyID:        strconv.Itoa(int(host.SitePartyID)),
			TableName:      host.DataTableName,
			TableNamespace: host.DataTableNamespace,
		})
	}
	return template.BuildPsiConf(param)
}

func (aggregate *JobAggregate) generateTrainingConfig() (string, string, error) {
	switch aggregate.Job.AlgorithmType {
	case entity.JobAlgorithmTypeHomoLR, entity.JobAlgorithmTypeHomoSBT:
		homoAlgorithmType := template.HomoAlgorithmTypeLR
		if aggregate.Job.AlgorithmType == entity.JobAlgorithmTypeHomoSBT {
			homoAlgorithmType = template.HomoAlgorithmTypeSBT
		}
		info := template.HomoTrainingParam{
			Guest: template.PartyDataInfo{
				PartyID:        strconv.Itoa(int(aggregate.Initiator.SitePartyID)),
				TableName:      aggregate.Initiator.DataTableName,
				TableNamespace: aggregate.Initiator.DataTableNamespace,
			},
			Hosts:             nil,
			LabelName:         aggregate.Initiator.DataLabelName,
			ValidationEnabled: aggregate.Job.AlgorithmConfig.TrainingValidationEnabled,
			ValidationPercent: aggregate.Job.AlgorithmConfig.TrainingValidationSizePercent,
			Type:              homoAlgorithmType,
		}
		for _, host := range aggregate.Participants {
			info.Hosts = append(info.Hosts, template.PartyDataInfo{
				PartyID:        strconv.Itoa(int(host.SitePartyID)),
				TableName:      host.DataTableName,
				TableNamespace: host.DataTableNamespace,
			})
		}
		return template.BuildHomoTrainingConf(info)
	}
	return "", "", errors.Errorf("invalid algorithm type: %d", aggregate.Job.AlgorithmType)
}

// GeneratePredictingJobParticipants returns a list of participant that should join new predicting job based on the job
func (aggregate *JobAggregate) GeneratePredictingJobParticipants() ([]*entity.JobParticipant, error) {
	if aggregate.Job.Type != entity.JobTypeTraining {
		return nil, errors.New("invalid job type")
	}
	switch aggregate.Job.AlgorithmType {
	case entity.JobAlgorithmTypeHomoSBT, entity.JobAlgorithmTypeHomoLR:
		if participant, ok := aggregate.Participants[aggregate.JobContext.CurrentSiteUUID]; ok {
			return []*entity.JobParticipant{
				participant,
			}, nil
		} else if aggregate.Initiator.SiteUUID == aggregate.JobContext.CurrentSiteUUID {
			return []*entity.JobParticipant{
				aggregate.Initiator,
			}, nil
		} else {
			return nil, errors.New("current site cannot participate in the predicting job")
		}
	}
	return nil, errors.New("cannot get predicting job participant list")
}

// SubmitJob submits the job to the FATE system
func (aggregate *JobAggregate) SubmitJob() error {
	if aggregate.Job == nil {
		return errors.New("job not created")
	}
	if len(aggregate.Participants) > 0 && !aggregate.FMLManagerConnectionInfo.Connected {
		return errors.New("fml manager not connected")
	}
	if len(aggregate.Participants) == 0 {
		if aggregate.Job.Type == entity.JobTypeTraining && aggregate.Job.AlgorithmType == entity.JobAlgorithmTypeHomoLR {
			return errors.New("homo LR job cannot be launched with only one party")
		}
	}
	if aggregate.JobContext.CurrentSiteUUID != aggregate.Job.InitiatingSiteUUID {
		return errors.Errorf("initiating site %s(%s) is not the current site", aggregate.Job.InitiatingSiteName, aggregate.Job.InitiatingSiteUUID)
	}

	if aggregate.Job.Conf == "" || aggregate.Job.DSL == "" {
		log.Warn().Msgf("job %s has no DSL or Conf content, generating now", aggregate.Job.Name)
		conf, dsl, err := aggregate.GenerateConfig()
		if err != nil {
			return err
		}
		aggregate.Job.Conf = conf
		aggregate.Job.DSL = dsl
	}

	if err := aggregate.Job.Validate(); err != nil {
		return err
	}
	if err := aggregate.Job.Create(); err != nil {
		return err
	}
	aggregate.Initiator.JobUUID = aggregate.Job.UUID
	if err := func() error {
		if err := aggregate.Initiator.Create(); err != nil {
			return err
		}
		if len(aggregate.Participants) == 0 {
			log.Info().Msgf("launching local job %s", aggregate.Job.UUID)
			if err := aggregate.Job.SubmitToFATE(aggregate.updateJobResultInfo); err != nil {
				return errors.Wrap(err, "failed to submit job to FATE")
			}
			if aggregate.Job.Type == entity.JobTypePSI {
				return errors.New("PSI job cannot be launched with only one party")
			}
		} else {
			log.Info().Msgf("sending job to FML Manager, job: %s, uuid: %s", aggregate.Job.Name, aggregate.Job.UUID)
			for _, participant := range aggregate.Participants {
				participant.JobUUID = aggregate.Job.UUID
				if err := participant.Create(); err != nil {
					return errors.Wrapf(err, "failed to create participant: %s", participant.SiteUUID)
				}
			}

			client := fmlmanager.NewFMLManagerClient(aggregate.FMLManagerConnectionInfo.Endpoint, aggregate.FMLManagerConnectionInfo.ServerName)
			if err := client.SendJobCreationRequest(aggregate.Job.UUID, aggregate.Job.InitiatingUser, aggregate.Job.RequestJson); err != nil {
				return errors.Wrap(err, "failed to create job to fml manager")
			}
		}
		return nil
	}(); err != nil {
		_ = aggregate.Job.UpdateStatus(entity.JobStatusDeleted)
		return err
	}
	return nil
}

// ApproveJob mark the job as approved and notify FML manager
func (aggregate *JobAggregate) ApproveJob() error {
	if !aggregate.FMLManagerConnectionInfo.Connected {
		return errors.New("fml manager not connected")
	}
	participant, ok := aggregate.Participants[aggregate.JobContext.CurrentSiteUUID]
	if !ok {
		return errors.Errorf("cannot find participant %s", aggregate.JobContext.CurrentSiteUUID)
	}

	client := fmlmanager.NewFMLManagerClient(aggregate.FMLManagerConnectionInfo.Endpoint, aggregate.FMLManagerConnectionInfo.ServerName)
	if err := client.SendJobApprovalResponse(aggregate.Job.UUID, fmlmanager.JobApprovalContext{
		SiteUUID: aggregate.JobContext.CurrentSiteUUID,
		Approved: true,
	}); err != nil {
		return errors.Wrap(err, "failed to send job approval to fml manager")
	}
	if err := participant.UpdateStatus(entity.JobParticipantStatusApproved); err != nil {
		return err
	}
	return nil
}

// RejectJob mark the job as rejected and notify the FML manager
func (aggregate *JobAggregate) RejectJob() error {
	if !aggregate.FMLManagerConnectionInfo.Connected {
		return errors.New("fml manager not connected")
	}
	participant, ok := aggregate.Participants[aggregate.JobContext.CurrentSiteUUID]
	if !ok {
		return errors.Errorf("cannot find participant %s", aggregate.JobContext.CurrentSiteUUID)
	}

	client := fmlmanager.NewFMLManagerClient(aggregate.FMLManagerConnectionInfo.Endpoint, aggregate.FMLManagerConnectionInfo.ServerName)
	if err := client.SendJobApprovalResponse(aggregate.Job.UUID, fmlmanager.JobApprovalContext{
		SiteUUID: aggregate.JobContext.CurrentSiteUUID,
		Approved: false,
	}); err != nil {
		return errors.Wrap(err, "failed to send job approval to fml manager")
	}

	if err := participant.UpdateStatus(entity.JobParticipantStatusRejected); err != nil {
		return err
	}
	if err := aggregate.Job.UpdateStatus(entity.JobStatusRejected); err != nil {
		return err
	}
	return nil
}

// HandleRemoteJobCreation handles creation of job initiated by other sites
func (aggregate *JobAggregate) HandleRemoteJobCreation() error {
	if aggregate.Job.UUID == "" {
		return errors.New("missing job uuid")
	}
	_, err := aggregate.JobRepo.GetByUUID(aggregate.Job.UUID)
	if err != nil {
		if errors.Is(err, repo.ErrJobNotFound) {
			if err := aggregate.Job.Create(); err != nil {
				return errors.Wrapf(err, "failed to create job")
			}
		} else {
			return errors.Wrap(err, "failed to query job")
		}
	} else {
		log.Warn().Msgf("updating existing job to pending status, job: %s, uuid: %s", aggregate.Job.Name, aggregate.Job.UUID)
		if err := aggregate.Job.UpdateStatus(entity.JobStatusPending); err != nil {
			return errors.Wrapf(err, "failed to update job info")
		}
	}

	aggregate.Initiator.JobUUID = aggregate.Job.UUID
	if err := aggregate.createOrUpdateParticipant(aggregate.Initiator); err != nil {
		return err
	}

	for _, participant := range aggregate.Participants {
		participant.JobUUID = aggregate.Job.UUID
		if err := aggregate.createOrUpdateParticipant(participant); err != nil {
			return err
		}
	}

	if aggregate.JobContext.AutoApprovalEnabled {
		go func() {
			// add a delay to make sure other sites are in-sync
			time.Sleep(time.Second * 10)
			log.Info().Msgf("auto approving job %s(%s)", aggregate.Job.Name, aggregate.Job.UUID)
			if err := aggregate.ApproveJob(); err != nil {
				log.Err(err).Msgf("failed to approve job: %s", aggregate.Job.UUID)
			}
		}()
	}
	return nil
}

// handleJobApproval process job approval, and may start the job if all participants approved it
func (aggregate *JobAggregate) handleJobApproval(siteUUID string) error {
	participant, ok := aggregate.Participants[siteUUID]
	if !ok {
		return errors.Errorf("cannot find participant %s", siteUUID)
	}

	if err := participant.UpdateStatus(entity.JobParticipantStatusApproved); err != nil {
		return err
	}

	go func() {
		// reload participants as others may have updated them
		if err := aggregate.reloadParticipant(); err != nil {
			log.Err(err).Msg("failed to reload participant")
		}
		aggregate.checkPendingJobParticipantStatus()
	}()
	return nil
}

// handleJobRejection marks the job as rejected
func (aggregate *JobAggregate) handleJobRejection(siteUUID string) error {
	participantInstance, err := aggregate.ParticipantRepo.GetByJobAndSiteUUID(aggregate.Job.UUID, siteUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query participant")
	}
	participant := participantInstance.(*entity.JobParticipant)

	participant.Repo = aggregate.ParticipantRepo
	if err := participant.UpdateStatus(entity.JobParticipantStatusRejected); err != nil {
		return err
	}
	if err := aggregate.Job.UpdateStatus(entity.JobStatusRejected); err != nil {
		return err
	}
	return nil
}

// HandleJobApprovalResponse process job approval response
func (aggregate *JobAggregate) HandleJobApprovalResponse(siteUUID string, approved bool) error {
	if approved {
		return aggregate.handleJobApproval(siteUUID)
	} else {
		return aggregate.handleJobRejection(siteUUID)
	}
}

// RefreshJob checks the FATE job status
func (aggregate *JobAggregate) RefreshJob() error {
	return aggregate.Job.CheckFATEJobStatus()
}

// HandleJobStatusUpdate process job status update. If the job becomes running, then a monitoring routine will be started
func (aggregate *JobAggregate) HandleJobStatusUpdate(newJobStatus *entity.Job, participantStatusMap map[string]entity.JobParticipantStatus) error {
	for siteUUID, newStatus := range participantStatusMap {
		participant, ok := aggregate.Participants[siteUUID]
		if ok && participant.Status != newStatus {
			log.Info().Str("participant name", participant.SiteName).
				Str("job name", aggregate.Job.Name).Str("job uuid", aggregate.Job.UUID).
				Msgf("HandleJobStatusUpdate: change participant status")
			if err := participant.UpdateStatus(newStatus); err != nil {
				log.Err(err).Str("participant_uuid", siteUUID).Send()
			}
		}
	}
	if err := aggregate.Job.Update(newJobStatus); err != nil {
		return err
	}
	if aggregate.Job.Status == entity.JobStatusSucceeded {
		log.Info().Msgf("try to update result info for job: %s(%s)", aggregate.Job.Name, aggregate.Job.UUID)
		go aggregate.updateJobResultInfo()
	}
	return nil
}

// createOrUpdateParticipant is a helper function to create a pending participant or update an existing one's status of pending
func (aggregate *JobAggregate) createOrUpdateParticipant(participant *entity.JobParticipant) error {
	participant.JobUUID = aggregate.Job.UUID
	participantInstance, err := aggregate.ParticipantRepo.GetByJobAndSiteUUID(aggregate.Job.UUID, participant.SiteUUID)
	if err != nil {
		if errors.Is(err, repo.ErrJobParticipantNotFound) {
			if err := participant.Create(); err != nil {
				return errors.Wrapf(err, "failed to create participant: %s", participant.SiteUUID)
			}
		} else {
			return errors.Wrap(err, "failed to query participant")
		}
	} else {
		participant = participantInstance.(*entity.JobParticipant)
		log.Info().Msgf("changing participant %s(%s) status to pending for job %s(%s)", participant.SiteName, participant.SiteUUID, aggregate.Job.Name, aggregate.Job.UUID)
		participant.Repo = aggregate.ParticipantRepo
		participant.Status = entity.JobParticipantStatusPending
		if err := participant.Repo.UpdateStatusByUUID(participant); err != nil {
			return errors.Wrap(err, "failed to update participant status")
		}
	}
	return nil
}

// sendApprovedJobStatusUpdate sends an approved job status(running, succeeded, failed, etc) to FML manager
func (aggregate *JobAggregate) sendApprovedJobStatusUpdate() {
	statusUpdateContext := fmlmanager.JobStatusUpdateContext{
		Status:               uint8(aggregate.Job.Status),
		StatusMessage:        aggregate.Job.StatusMessage,
		FATEJobID:            aggregate.Job.FATEJobID,
		FATEJobStatus:        aggregate.Job.FATEJobStatus,
		FATEModelID:          aggregate.Job.FATEModelID,
		FATEModelVersion:     aggregate.Job.FATEModelVersion,
		ParticipantStatusMap: map[string]uint8{},
	}

	for siteUUID := range aggregate.Participants {
		statusUpdateContext.ParticipantStatusMap[siteUUID] = uint8(entity.JobParticipantStatusApproved)
	}
	client := fmlmanager.NewFMLManagerClient(aggregate.FMLManagerConnectionInfo.Endpoint, aggregate.FMLManagerConnectionInfo.ServerName)
	if err := client.SendJobStatusUpdate(aggregate.Job.UUID, statusUpdateContext); err != nil {
		log.Err(err).Str("job uuid", aggregate.Job.UUID).Msgf("failed to send job status update to FATE")
	}
}

// updateJobResultInfo updates job result
func (aggregate *JobAggregate) updateJobResultInfo() {
	if aggregate.Job.Status != entity.JobStatusSucceeded {
		return
	}
	role := aggregate.Initiator.SiteRole
	partyID := aggregate.Initiator.SitePartyID
	if participant, ok := aggregate.Participants[aggregate.JobContext.CurrentSiteUUID]; ok {
		role = participant.SiteRole
		partyID = participant.SitePartyID
	}

	if err := aggregate.Job.UpdateResultInfo(partyID, role); err != nil {
		log.Err(err).Str("job uuid", aggregate.Job.UUID).Msg("failed to update job result")
	}
}

// checkPendingJobParticipantStatus checks the participant status and may start the job
func (aggregate *JobAggregate) checkPendingJobParticipantStatus() {
	// only the initiator can start the job
	if aggregate.Job.Status != entity.JobStatusPending || aggregate.JobContext.CurrentSiteUUID != aggregate.Job.InitiatingSiteUUID {
		return
	}

	for _, participant := range aggregate.Participants {
		if participant.Status != entity.JobParticipantStatusApproved {
			log.Info().Str("job uuid", aggregate.Job.UUID).Msgf("job still pending on site %s(%s)", participant.SiteName, participant.SiteUUID)
			return
		}
	}
	log.Info().Str("job uuid", aggregate.Job.UUID).Msgf("job approved by all site, starting")

	// we firstly send the job status update to let other sites know the job has been approved and starts running
	// then after the job is finished, we, in the callback, send the update again to other sites to update the job status and model info
	// the initiating site will get the result in the callback
	// and the other sites will start to get the result when handling the status update
	if err := aggregate.Job.SubmitToFATE(func() {
		aggregate.sendApprovedJobStatusUpdate()
		aggregate.updateJobResultInfo()
	}); err != nil {
		log.Err(err).Str("job uuid", aggregate.Job.UUID).Msgf("failed to submit job to FATE")
	}
	aggregate.sendApprovedJobStatusUpdate()
}

// reloadParticipant is a helper function to retrieve the participant info from the repo
func (aggregate *JobAggregate) reloadParticipant() error {
	participantListInstance, err := aggregate.ParticipantRepo.GetListByJobUUID(aggregate.Job.UUID)
	if err != nil {
		return err
	}
	participantList := participantListInstance.([]entity.JobParticipant)
	for index, participant := range participantList {
		participantList[index].Repo = aggregate.ParticipantRepo
		if participant.SiteUUID != aggregate.Job.InitiatingSiteUUID {
			aggregate.Participants[participant.SiteUUID] = &participantList[index]
		}
	}
	return nil
}

// GetDataResultDownloadRequest returns a request object to be used to download the result data
func (aggregate *JobAggregate) GetDataResultDownloadRequest() (*http.Request, error) {
	var componentName string
	if aggregate.Job.Type == entity.JobTypePredict {
		componentName = aggregate.Job.AlgorithmComponentName
	} else if aggregate.Job.Type == entity.JobTypePSI {
		componentName = "intersection_0"
	} else {
		return nil, errors.New("job is not a predicting job nor a PSI job")
	}
	if aggregate.Job.Status != entity.JobStatusSucceeded {
		return nil, errors.New("job did not finished successfully")
	}
	role := aggregate.Initiator.SiteRole
	partyID := aggregate.Initiator.SitePartyID
	if participant, ok := aggregate.Participants[aggregate.JobContext.CurrentSiteUUID]; ok {
		role = participant.SiteRole
		partyID = participant.SitePartyID
	}

	fateClient := fateclient.NewFATEFlowClient(aggregate.Job.FATEFlowContext.FATEFlowHost,
		aggregate.Job.FATEFlowContext.FATEFlowPort, aggregate.Job.FATEFlowContext.FATEFlowIsHttps)
	return fateClient.GetComponentOutputDataDownloadRequest(fateclient.ComponentTrackingCommonRequest{
		JobID:         aggregate.Job.FATEJobID,
		PartyID:       partyID,
		Role:          string(role),
		ComponentName: componentName,
	})
}
