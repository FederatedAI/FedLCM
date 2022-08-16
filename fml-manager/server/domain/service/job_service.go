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
	"fmt"
	"sync"

	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/FederatedAI/FedLCM/fml-manager/server/infrastructure/siteportal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// JobService processes job related requests
type JobService struct {
	JobRepo         repo.JobRepository
	ParticipantRepo repo.JobParticipantRepository
}

// JobCreationRequest holds the job info and all the joined participants
type JobCreationRequest struct {
	Job          *entity.Job
	Initiator    JobParticipantSiteInfo
	Participants map[string]JobParticipantSiteInfo
}

// JobApprovalResponse contains the info needed to process the job approval response
type JobApprovalResponse struct {
	Initiator     JobParticipantConnectionInfo
	Participants  map[string]JobParticipantConnectionInfo
	ApprovingSite *entity.JobParticipant
	Approved      bool
	JobUUID       string
}

// JobParticipantSiteInfo contains more detailed info of a participating site
type JobParticipantSiteInfo struct {
	*entity.JobParticipant
	JobParticipantConnectionInfo
}

// JobParticipantConnectionInfo contains connection info to the participating site
type JobParticipantConnectionInfo struct {
	ExternalHost string
	ExternalPort uint
	HTTPS        bool
	ServerName   string
}

// JobParticipantStatusInfo contains job participating status and the site connection info
type JobParticipantStatusInfo struct {
	entity.JobParticipantStatus
	JobParticipantConnectionInfo
}

// JobStatusUpdateContext contains info needed to update a job status
type JobStatusUpdateContext struct {
	JobUUID              string
	NewJobStatus         *entity.Job
	ParticipantStatusMap map[string]JobParticipantStatusInfo
	RequestJson          string
}

// HandleNewJobCreation process job creation request
func (s *JobService) HandleNewJobCreation(request *JobCreationRequest) error {
	if err := request.Job.Create(); err != nil {
		return err
	}
	if err := request.Initiator.Create(); err != nil {
		return err
	}
	for _, participant := range request.Participants {
		participant.JobUUID = request.Job.UUID
		if err := participant.Create(); err != nil {
			return errors.Wrapf(err, "failed to create participant: %s", participant.SiteUUID)
		}
	}

	wg := &sync.WaitGroup{}
	var failedSite []string
	for uuid := range request.Participants {
		wg.Add(1)
		go func(siteUUID string) {
			defer wg.Done()
			participant := request.Participants[siteUUID]
			sitePortalClient := siteportal.NewSitePortalClient(participant.ExternalHost, participant.ExternalPort, participant.HTTPS, participant.ServerName)
			if err := sitePortalClient.SendJobCreationRequest(request.Job.RequestJson); err != nil {
				log.Err(err).Str("job uuid", request.Job.UUID).Str("site uuid", siteUUID).Msg("fail to send job creation to site")
				failedSite = append(failedSite, fmt.Sprintf("%s(%s)", participant.SiteName, participant.SiteUUID))
			}
		}(uuid)
	}
	wg.Wait()
	if len(failedSite) != 0 {
		return errors.Errorf("failed to send job creation to some site(s): %v", failedSite)
	}
	return nil
}

// HandleJobApprovalResponse process job approval response
func (s *JobService) HandleJobApprovalResponse(response *JobApprovalResponse) error {
	var newStatus entity.JobParticipantStatus
	if response.Approved {
		newStatus = entity.JobParticipantStatusApproved
	} else {
		newStatus = entity.JobParticipantStatusRejected
	}
	statusInDB := response.ApprovingSite.GetStatus()
	log.Debug().Msgf("The status in DB is %d", statusInDB)
	if statusInDB == entity.JobParticipantStatusApproved {
		// This happens when the auto-approve thread comes, but there is a manual approve came earlier.
		// In such case, just ignore this approval response from the participant's site portal.
		return nil
	}
	if err := response.ApprovingSite.UpdateStatus(newStatus); err != nil {
		return err
	}
	sitePortalClient := siteportal.NewSitePortalClient(response.Initiator.ExternalHost, response.Initiator.ExternalPort, response.Initiator.HTTPS, response.Initiator.ServerName)
	if err := sitePortalClient.SendJobApprovalResponse(response.JobUUID, siteportal.JobApprovalContext{
		SiteUUID: response.ApprovingSite.SiteUUID,
		Approved: response.Approved,
	}); err != nil {
		return err
	}
	for uuid := range response.Participants {
		if uuid != response.ApprovingSite.SiteUUID {
			go func(siteUUID string) {
				participant := response.Participants[siteUUID]
				sitePortalClient := siteportal.NewSitePortalClient(participant.ExternalHost, participant.ExternalPort, participant.HTTPS, participant.ServerName)
				if err := sitePortalClient.SendJobApprovalResponse(response.JobUUID, siteportal.JobApprovalContext{
					SiteUUID: response.ApprovingSite.SiteUUID,
					Approved: response.Approved,
				}); err != nil {
					log.Err(err).Str("job uuid", response.JobUUID).Str("site uuid", siteUUID).Msg("fail to send job approval to site")
				}
			}(uuid)
		}
	}
	return nil
}

// HandleJobStatusUpdate process job status update
func (s *JobService) HandleJobStatusUpdate(context *JobStatusUpdateContext) error {
	for siteUUID, newInfo := range context.ParticipantStatusMap {
		jobParticipantInstance, err := s.ParticipantRepo.GetByJobAndSiteUUID(context.JobUUID, siteUUID)
		if err != nil {
			log.Err(err).Str("participant_uuid", siteUUID).Msg("failed to get participant info")
			continue
		}
		jobParticipant := jobParticipantInstance.(*entity.JobParticipant)
		jobParticipant.Repo = s.ParticipantRepo
		if jobParticipant.Status != newInfo.JobParticipantStatus {
			if err := jobParticipant.UpdateStatus(newInfo.JobParticipantStatus); err != nil {
				log.Err(err).Str("participant_uuid", siteUUID).Send()
			}
		}
		go func(uuid string) {
			client := siteportal.NewSitePortalClient(context.ParticipantStatusMap[uuid].ExternalHost, context.ParticipantStatusMap[uuid].ExternalPort, context.ParticipantStatusMap[uuid].HTTPS, context.ParticipantStatusMap[uuid].ServerName)
			if err := client.SendJobStatusUpdate(context.JobUUID, context.RequestJson); err != nil {
				log.Err(err).Str("job uuid", context.JobUUID).Str("site uuid", uuid).Msg("failed to send job status update")
			}
		}(siteUUID)
	}

	jobInstance, err := s.JobRepo.GetByUUID(context.JobUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query job")
	}
	job := jobInstance.(*entity.Job)
	job.Repo = s.JobRepo
	if err := job.Update(context.NewJobStatus); err != nil {
		return errors.Wrap(err, "failed to update job")
	}
	return nil
}
