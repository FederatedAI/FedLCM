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
	"strings"

	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/valueobject"
	"github.com/FederatedAI/FedLCM/fml-manager/server/infrastructure/siteportal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// ProjectService is the service to handle project related requests
type ProjectService struct {
	ProjectRepo     repo.ProjectRepository
	InvitationRepo  repo.ProjectInvitationRepository
	ParticipantRepo repo.ProjectParticipantRepository
	ProjectDataRepo repo.ProjectDataRepository
}

// ProjectInvitationRequest is an invitation for asking a site to join a project
type ProjectInvitationRequest struct {
	InvitationUUID string
	Project        *entity.Project
	ManagingSite   *ProjectParticipantSiteInfo
	TargetSite     *ProjectParticipantSiteInfo
	AssociatedData []entity.ProjectData
}

// ProjectParticipantSiteInfo contains more detailed info of a participating site
type ProjectParticipantSiteInfo struct {
	Name         string
	Description  string
	UUID         string
	PartyID      uint
	ExternalHost string
	ExternalPort uint
	HTTPS        bool
	ServerName   string
}

// HandleInvitationRequest creates/updates repo records and forward the invitation to the target site
func (s *ProjectService) HandleInvitationRequest(req *ProjectInvitationRequest) error {
	// create project if needed
	_, err := s.ProjectRepo.GetByUUID(req.Project.UUID)
	if err != nil {
		if errors.Is(err, repo.ErrProjectNotFound) {
			log.Info().Msgf("creating project %s(%s)", req.Project.Name, req.Project.UUID)
			req.Project.Model.ID = 0
			if err := s.ProjectRepo.Create(req.Project); err != nil {
				// race condition
				if !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
					return errors.Wrapf(err, "failed to create the project")
				} else {
					log.Info().Msgf("project %s(%s) already created, continue", req.Project.Name, req.Project.UUID)
				}
			}
			if err := s.ParticipantRepo.Create(&entity.ProjectParticipant{
				Model:           gorm.Model{},
				UUID:            uuid.NewV4().String(),
				ProjectUUID:     req.Project.UUID,
				SiteUUID:        req.ManagingSite.UUID,
				SiteName:        req.ManagingSite.Name,
				SitePartyID:     req.ManagingSite.PartyID,
				SiteDescription: req.ManagingSite.Description,
				Status:          entity.ProjectParticipantStatusOwner,
			}); err != nil {
				return errors.Wrapf(err, "failed to create the owner participant")
			}
			for _, data := range req.AssociatedData {
				if err := s.createOrUpdateData(&data); err != nil {
					return errors.Wrapf(err, "failed to create the owner data association for data(%s)", data.DataUUID)
				}
			}
		} else {
			return errors.Wrapf(err, "failed to query the project")
		}
	}
	// create invitation
	projectInvitation := &entity.ProjectInvitation{
		Model:       gorm.Model{},
		UUID:        req.InvitationUUID,
		ProjectUUID: req.Project.UUID,
		SiteUUID:    req.TargetSite.UUID,
		Status:      entity.ProjectInvitationStatusCreated,
	}
	if err := s.InvitationRepo.Create(projectInvitation); err != nil {
		return errors.Wrapf(err, "failed to create the invitation")
	}
	// create/update project participant info
	instance, err := s.ParticipantRepo.GetByProjectAndSiteUUID(req.Project.UUID, req.TargetSite.UUID)
	if err != nil {
		if errors.Is(err, repo.ErrProjectParticipantNotFound) {
			projectParticipant := &entity.ProjectParticipant{
				Model:           gorm.Model{},
				UUID:            uuid.NewV4().String(),
				ProjectUUID:     req.Project.UUID,
				SiteUUID:        req.TargetSite.UUID,
				SiteName:        req.TargetSite.Name,
				SitePartyID:     req.TargetSite.PartyID,
				SiteDescription: req.TargetSite.Description,
				Status:          entity.ProjectParticipantStatusPending,
			}
			if err := s.ParticipantRepo.Create(projectParticipant); err != nil {
				return errors.Wrapf(err, "failed to create the target participant")
			}
		} else {
			return errors.Wrapf(err, "failed to query the target participant")
		}
	} else {
		participant := instance.(*entity.ProjectParticipant)
		participant.Status = entity.ProjectParticipantStatusPending
		if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
			return errors.Wrapf(err, "failed to update the target participant status")
		}
	}
	// send invitation to site portal. this shouldn't be placed in a goroutine as we need to fail the original request if we hit error here
	log.Info().Msgf("sending invitation to target site: %s(%s)", req.TargetSite.Name, req.TargetSite.UUID)
	client := siteportal.NewSitePortalClient(req.TargetSite.ExternalHost, req.TargetSite.ExternalPort, req.TargetSite.HTTPS, req.TargetSite.ServerName)
	if err := client.SendInvitation(&siteportal.ProjectInvitationRequest{
		UUID:                       projectInvitation.UUID,
		SiteUUID:                   req.TargetSite.UUID,
		SitePartyID:                req.TargetSite.PartyID,
		ProjectUUID:                req.Project.UUID,
		ProjectName:                req.Project.Name,
		ProjectDescription:         req.Project.Description,
		ProjectAutoApprovalEnabled: req.Project.AutoApprovalEnabled,
		ProjectManager:             req.Project.Manager,
		ProjectManagingSiteName:    req.Project.ManagingSiteName,
		ProjectManagingSitePartyID: req.Project.ManagingSitePartyID,
		ProjectManagingSiteUUID:    req.Project.ManagingSiteUUID,
		ProjectCreationTime:        req.Project.CreatedAt,
	}); err != nil {
		return errors.Wrapf(err, "failed to forward the invitation")
	}
	// update invitation status
	projectInvitation.Status = entity.ProjectInvitationStatusSent
	if err := s.InvitationRepo.UpdateStatusByUUID(projectInvitation); err != nil {
		return errors.Wrapf(err, "failed to update the invitation status")
	}
	return nil
}

// HandleInvitationAcceptance updates the status in the DB and send the participants info to the joined sites
func (s *ProjectService) HandleInvitationAcceptance(req *ProjectInvitationRequest, otherSiteList []ProjectParticipantSiteInfo) error {
	// sanity check
	invitationInstance, err := s.InvitationRepo.GetByUUID(req.InvitationUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the invitation instance")
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}

	// update status
	invitation.Status = entity.ProjectInvitationStatusAccepted
	participantInstance, err := s.ParticipantRepo.GetByProjectAndSiteUUID(invitation.ProjectUUID, invitation.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the participant instance")
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusJoined
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return errors.Wrapf(err, "failed to update the participant status")
	}
	if err := s.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return errors.Wrapf(err, "failed to update the invitation status")
	}

	go func() {
		// send invitation acceptance to managing site
		log.Info().Msgf("forwarding response to owner site: %s(%s)", req.ManagingSite.Name, req.ManagingSite.UUID)
		client := siteportal.NewSitePortalClient(req.ManagingSite.ExternalHost, req.ManagingSite.ExternalPort, req.ManagingSite.HTTPS, req.ManagingSite.ServerName)
		if err := client.SendInvitationAcceptance(req.InvitationUUID); err != nil {
			log.Err(errors.Wrapf(err, "failed to redirect project invitation response")).Send()
		}

		// send new participant to all joined site
		for _, otherSite := range otherSiteList {
			go func(site ProjectParticipantSiteInfo) {
				log.Info().Msgf("sending new site info to site: %s(%s)", site.Name, site.UUID)
				if site.UUID != req.TargetSite.UUID {
					client = siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
					if err := client.SendProjectParticipants(invitation.ProjectUUID, []siteportal.ProjectParticipant{
						{
							UUID:            participant.UUID,
							ProjectUUID:     participant.ProjectUUID,
							SiteUUID:        participant.SiteUUID,
							SiteName:        participant.SiteName,
							SitePartyID:     participant.SitePartyID,
							SiteDescription: participant.SiteDescription,
							Status:          uint8(participant.Status),
						},
					}); err != nil {
						log.Err(err).Msgf("failed to send participants update to site: %s(%s), continue", site.Name, site.UUID)
					}
				}
			}(otherSite)
		}

		// send project participants to joining site
		instanceList, err := s.ParticipantRepo.GetByProjectUUID(req.Project.UUID)
		if err != nil {
			log.Err(errors.Wrapf(err, "failed to get participant list")).Send()
			return
		}
		participantList := instanceList.([]entity.ProjectParticipant)
		var joinedParticipantList []siteportal.ProjectParticipant
		for _, participant := range participantList {
			if participant.Status == entity.ProjectParticipantStatusJoined || participant.Status == entity.ProjectParticipantStatusOwner {
				joinedParticipantList = append(joinedParticipantList, siteportal.ProjectParticipant{
					UUID:            participant.UUID,
					ProjectUUID:     participant.ProjectUUID,
					SiteUUID:        participant.SiteUUID,
					SiteName:        participant.SiteName,
					SitePartyID:     participant.SitePartyID,
					SiteDescription: participant.SiteDescription,
					Status:          uint8(participant.Status),
				})
			}
		}
		log.Info().Msgf("sending participants sites info to new site: %s(%s)", req.TargetSite.Name, req.TargetSite.UUID)
		client = siteportal.NewSitePortalClient(req.TargetSite.ExternalHost, req.TargetSite.ExternalPort, req.TargetSite.HTTPS, req.TargetSite.ServerName)
		if err := client.SendProjectParticipants(invitation.ProjectUUID, joinedParticipantList); err != nil {
			log.Err(errors.Wrapf(err, "failed to send participant list to new site")).Send()
		}

		// send associated data to the newly joined site
		instanceList, err = s.ProjectDataRepo.GetListByProjectUUID(req.Project.UUID)
		if err != nil {
			log.Err(errors.Wrapf(err, "failed to get project data list")).Send()
			return
		}
		dataList := instanceList.([]entity.ProjectData)
		var associatedDataList []siteportal.ProjectData
		for _, data := range dataList {
			if data.Status == entity.ProjectDataStatusAssociated {
				associatedDataList = append(associatedDataList, siteportal.ProjectData{
					Name:           data.Name,
					Description:    data.Description,
					ProjectUUID:    data.ProjectUUID,
					DataUUID:       data.DataUUID,
					SiteUUID:       data.SiteUUID,
					SiteName:       data.SiteName,
					SitePartyID:    data.SitePartyID,
					TableName:      data.TableName,
					TableNamespace: data.TableNamespace,
					CreationTime:   data.CreationTime,
					UpdateTime:     data.UpdateTime,
				})
			}
		}
		log.Info().Msgf("sending project data info to new site: %s(%s)", req.TargetSite.Name, req.TargetSite.UUID)
		client = siteportal.NewSitePortalClient(req.TargetSite.ExternalHost, req.TargetSite.ExternalPort, req.TargetSite.HTTPS, req.TargetSite.ServerName)
		if err := client.SendProjectDataAssociation(invitation.ProjectUUID, associatedDataList); err != nil {
			log.Err(errors.Wrapf(err, "failed to send participant list to new site")).Send()
		}
	}()
	return nil
}

// HandleInvitationRejection updates the DB status and redirect the response to the owner site
func (s *ProjectService) HandleInvitationRejection(req *ProjectInvitationRequest) error {
	invitationInstance, err := s.InvitationRepo.GetByUUID(req.InvitationUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the invitation instance")
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}
	invitation.Status = entity.ProjectInvitationStatusRejected
	participantInstance, err := s.ParticipantRepo.GetByProjectAndSiteUUID(invitation.ProjectUUID, invitation.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the participant instance")
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusRejected
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return errors.Wrapf(err, "failed to update the participant status")
	}
	if err := s.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return errors.Wrapf(err, "failed to update the invitation status")
	}
	// send rejection to owner site
	go func() {
		log.Info().Msgf("forwarding reject response to owner site: %s(%s)", req.ManagingSite.Name, req.ManagingSite.UUID)
		client := siteportal.NewSitePortalClient(req.ManagingSite.ExternalHost, req.ManagingSite.ExternalPort, req.ManagingSite.HTTPS, req.ManagingSite.ServerName)
		if err := client.SendInvitationRejection(req.InvitationUUID); err != nil {
			log.Err(errors.Wrapf(err, "failed to redirect project invitation response")).Send()
		}
	}()
	return nil
}

// HandleInvitationRevocation updates the DB status and redirect the response to the target site
func (s *ProjectService) HandleInvitationRevocation(req *ProjectInvitationRequest) error {
	invitationInstance, err := s.InvitationRepo.GetByUUID(req.InvitationUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the invitation instance")
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}
	invitation.Status = entity.ProjectInvitationStatusRevoked
	participantInstance, err := s.ParticipantRepo.GetByProjectAndSiteUUID(invitation.ProjectUUID, invitation.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the participant instance")
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusRevoked

	// send revocation to target site
	log.Info().Msgf("sending invitation revocation to site: %s(%s)", req.TargetSite.Name, req.TargetSite.UUID)
	client := siteportal.NewSitePortalClient(req.TargetSite.ExternalHost, req.TargetSite.ExternalPort, req.TargetSite.HTTPS, req.TargetSite.ServerName)
	if err := client.SendInvitationRevocation(req.InvitationUUID); err != nil {
		return errors.Wrapf(err, "failed to redirect project invitation response")
	}

	// update DB records
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return errors.Wrapf(err, "failed to update the participant status")
	}
	if err := s.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return errors.Wrapf(err, "failed to update the invitation status")
	}
	return nil
}

// HandleParticipantInfoUpdate updates the site info in the repo and send such update to impacted sites
func (s *ProjectService) HandleParticipantInfoUpdate(newSiteInfo ProjectParticipantSiteInfo, allSites []ProjectParticipantSiteInfo) error {
	toUpdateProjectTemplate := &entity.Project{
		ProjectCreatorInfo: &valueobject.ProjectCreatorInfo{
			ManagingSiteName:    newSiteInfo.Name,
			ManagingSitePartyID: newSiteInfo.PartyID,
			ManagingSiteUUID:    newSiteInfo.UUID,
		},
	}
	if err := s.ProjectRepo.UpdateManagingSiteInfoBySiteUUID(toUpdateProjectTemplate); err != nil {
		return errors.Wrapf(err, "failed to update projects creator info")
	}

	toUpdateParticipantTemplate := &entity.ProjectParticipant{
		SiteUUID:        newSiteInfo.UUID,
		SiteName:        newSiteInfo.Name,
		SitePartyID:     newSiteInfo.PartyID,
		SiteDescription: newSiteInfo.Description,
	}
	if err := s.ParticipantRepo.UpdateParticipantInfoBySiteUUID(toUpdateParticipantTemplate); err != nil {
		return errors.Wrapf(err, "failed to update projects participants info")
	}

	toUpdateDataTemplate := &entity.ProjectData{
		SiteUUID:    newSiteInfo.UUID,
		SiteName:    newSiteInfo.Name,
		SitePartyID: newSiteInfo.PartyID,
	}
	if err := s.ProjectDataRepo.UpdateSiteInfoBySiteUUID(toUpdateDataTemplate); err != nil {
		return errors.Wrap(err, "failed to update project data site info")
	}

	go func() {
		// XXX: we are issuing the event to all sites. Better to only issue the event to "impacted" sites
		for _, targetSite := range allSites {
			go func(site ProjectParticipantSiteInfo) {
				if site.UUID == newSiteInfo.UUID {
					return
				}
				log.Info().Msgf("sending participant info update event to site: %s(%s)", site.Name, site.UUID)
				client := siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
				if err := client.SendParticipantInfoUpdateEvent(siteportal.ProjectParticipantUpdateEvent{
					UUID:        newSiteInfo.UUID,
					PartyID:     newSiteInfo.PartyID,
					Name:        newSiteInfo.Name,
					Description: newSiteInfo.Description,
				}); err != nil {
					log.Err(err).Msgf("failed to send site info update event to site: %s(%s)", site.Name, site.UUID)
				}
			}(targetSite)
		}
	}()
	return nil
}

// HandleParticipantLeaving updates the participant status in the repo and send such update to impacted sites
func (s *ProjectService) HandleParticipantLeaving(projectUUID, siteUUID string, otherSiteList []ProjectParticipantSiteInfo) error {
	participantInstance, err := s.ParticipantRepo.GetByProjectAndSiteUUID(projectUUID, siteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the participant instance")
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusLeft
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return errors.Wrapf(err, "failed to update the participant status")
	}
	// send such update to other sites
	go func() {
		for _, otherSite := range otherSiteList {
			if otherSite.UUID != siteUUID {
				go func(site ProjectParticipantSiteInfo) {
					log.Info().Msgf("sending project participant leaving to site: %s(%s)", site.Name, site.UUID)
					client := siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
					if err := client.SendProjectParticipantLeaving(projectUUID, siteUUID); err != nil {
						log.Err(err).Msgf("failed to send project participant leaving to site: %s(%s), continue", site.Name, site.UUID)
					}
				}(otherSite)
			}
		}
	}()
	return nil
}

// HandleParticipantDismissal sends the dismissal to the target site and update the participant status in the repo and send such update to other sites
func (s *ProjectService) HandleParticipantDismissal(projectUUID string, targetSite ProjectParticipantSiteInfo, otherSiteList []ProjectParticipantSiteInfo) error {
	// synchronously notify the target site, because maybe only the target site is running a job that no one is aware of
	// TODO: provide a "force" option so that we can ignore error if the cause is that the target site is no longer exists
	client := siteportal.NewSitePortalClient(targetSite.ExternalHost, targetSite.ExternalPort, targetSite.HTTPS, targetSite.ServerName)
	if err := client.SendProjectParticipantDismissal(projectUUID, targetSite.UUID); err != nil {
		return errors.Wrapf(err, "failed to send project participant dismissal to site: %s(%s)", targetSite.Name, targetSite.UUID)
	}
	// dismiss data association
	dataListInstance, err := s.ProjectDataRepo.GetListByProjectAndSiteUUID(projectUUID, targetSite.UUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project data")
	}
	dataList := dataListInstance.([]entity.ProjectData)
	for _, data := range dataList {
		if data.Status == entity.ProjectDataStatusAssociated {
			data.Status = entity.ProjectDataStatusDismissed
			if err := s.ProjectDataRepo.UpdateStatusByUUID(&data); err != nil {
				return errors.Wrapf(err, "failed to dismiss data %s from site: %s", data.Name, targetSite.Name)
			}
		}
	}
	// update participant status
	participantInstance, err := s.ParticipantRepo.GetByProjectAndSiteUUID(projectUUID, targetSite.UUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get the participant instance")
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusDismissed
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return errors.Wrapf(err, "failed to update the participant status")
	}
	// send such event to other sites
	go func() {
		for _, otherSite := range otherSiteList {
			if otherSite.UUID != targetSite.UUID {
				go func(site ProjectParticipantSiteInfo) {
					log.Info().Msgf("sending project participant dismissal to site: %s(%s)", site.Name, site.UUID)
					client := siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
					if err := client.SendProjectParticipantDismissal(projectUUID, targetSite.UUID); err != nil {
						log.Err(err).Msgf("failed to send project participant leaving to site: %s(%s), continue", site.Name, site.UUID)
					}
				}(otherSite)
			}
		}
	}()
	return nil
}

// HandleDataAssociation sends the new data association to other participating sites
func (s *ProjectService) HandleDataAssociation(newData *entity.ProjectData, otherSiteList []ProjectParticipantSiteInfo) error {
	if err := s.createOrUpdateData(newData); err != nil {
		return err
	}
	// inform other joined site of this newly associated data
	go func() {
		for _, otherSite := range otherSiteList {
			if otherSite.UUID != newData.SiteUUID {
				go func(site ProjectParticipantSiteInfo) {
					log.Info().Msgf("sending new project data info to site: %s(%s)", site.Name, site.UUID)
					client := siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
					if err := client.SendProjectDataAssociation(newData.ProjectUUID, []siteportal.ProjectData{
						{
							Name:           newData.Name,
							Description:    newData.Description,
							ProjectUUID:    newData.ProjectUUID,
							DataUUID:       newData.DataUUID,
							SiteUUID:       newData.SiteUUID,
							SiteName:       newData.SiteName,
							SitePartyID:    newData.SitePartyID,
							TableName:      newData.TableName,
							TableNamespace: newData.TableNamespace,
							CreationTime:   newData.CreationTime,
							UpdateTime:     newData.UpdateTime,
						},
					}); err != nil {
						log.Err(err).Msgf("failed to send new project data info to site: %s(%s), continue", site.Name, site.UUID)
					}
				}(otherSite)
			}
		}
	}()
	return nil
}

// HandleDataDismissal sends data dismissal event to other participating site
func (s *ProjectService) HandleDataDismissal(projectUUID, dataUUID string, otherSiteList []ProjectParticipantSiteInfo) error {
	providingSiteUUID := ""
	instance, err := s.ProjectDataRepo.GetByProjectAndDataUUID(projectUUID, dataUUID)
	if err != nil {
		if errors.Is(err, repo.ErrProjectDataNotFound) {
			log.Warn().Str("data uuid", dataUUID).Str("project uuid", projectUUID).Msg("data not associated in this project")
		} else {
			return errors.Wrapf(err, "failed to query data association")
		}
	} else {
		data := instance.(*entity.ProjectData)
		data.Status = entity.ProjectDataStatusDismissed
		providingSiteUUID = data.SiteUUID
		if err := s.ProjectDataRepo.UpdateStatusByUUID(data); err != nil {
			return errors.Wrapf(err, "failed to update data association")
		}
	}
	// inform other joined site of this dismissed associated data
	go func() {
		for _, otherSite := range otherSiteList {
			if otherSite.UUID != providingSiteUUID {
				go func(site ProjectParticipantSiteInfo) {
					log.Info().Msgf("sending project data dismissal to site: %s(%s)", site.Name, site.UUID)
					client := siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
					if err := client.SendProjectDataDismissal(projectUUID, []string{dataUUID}); err != nil {
						log.Err(err).Msgf("failed to send project data dismissal to site: %s(%s), continue", site.Name, site.UUID)
					}
				}(otherSite)
			}
		}
	}()

	return nil
}

// createOrUpdateData creates or updates the project data records
func (s *ProjectService) createOrUpdateData(newData *entity.ProjectData) error {
	instance, err := s.ProjectDataRepo.GetByProjectAndDataUUID(newData.ProjectUUID, newData.DataUUID)
	if err != nil {
		if errors.Is(err, repo.ErrProjectDataNotFound) {
			newData.Model.ID = 0
			newData.UUID = uuid.NewV4().String()
			newData.Status = entity.ProjectDataStatusAssociated
			if err := s.ProjectDataRepo.Create(newData); err != nil {
				return errors.Wrapf(err, "failed to create data association")
			}
		} else {
			return errors.Wrapf(err, "failed to query data association")
		}
	} else {
		data := instance.(*entity.ProjectData)
		data.Status = entity.ProjectDataStatusAssociated
		if err := s.ProjectDataRepo.UpdateStatusByUUID(data); err != nil {
			return errors.Wrapf(err, "failed to update data association")
		}
	}
	return nil
}

// HandleProjectClosing updates project status and sends the event to other site
func (s *ProjectService) HandleProjectClosing(projectUUID string, otherSiteList []ProjectParticipantSiteInfo) error {

	projectInstance, err := s.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find project")
	}
	project := projectInstance.(*entity.Project)

	project.Status = entity.ProjectStatusClosed
	if err := s.ProjectRepo.UpdateStatusByUUID(project); err != nil {
		return errors.Wrapf(err, "failed to update project status")
	}

	go func() {
		for _, otherSite := range otherSiteList {
			go func(site ProjectParticipantSiteInfo) {
				log.Info().Msgf("sending project closing to site: %s(%s)", site.Name, site.UUID)
				client := siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
				if err := client.SendProjectClosing(projectUUID); err != nil {
					log.Err(err).Msgf("failed to send project closing to site: %s(%s), continue", site.Name, site.UUID)
				}
			}(otherSite)
		}
	}()
	return nil
}
