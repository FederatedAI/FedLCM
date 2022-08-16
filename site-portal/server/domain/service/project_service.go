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
	"strconv"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/aggregate"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/event"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fmlmanager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// ProjectService provides domain services for the project management context
// We created this because we are not sure if they can be put into the "aggregate"
// or the "entity" package - seems they are not owned by any entity.
// This can be revisited in the future.
type ProjectService struct {
	ProjectRepo     repo.ProjectRepository
	ParticipantRepo repo.ProjectParticipantRepository
	InvitationRepo  repo.ProjectInvitationRepository
	ProjectDataRepo repo.ProjectDataRepository
}

// ProcessProjectAcceptance process invitation acceptance response
func (s *ProjectService) ProcessProjectAcceptance(invitationUUID string) error {
	invitationInstance, err := s.InvitationRepo.GetByUUID(invitationUUID)
	if err != nil {
		return err
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}
	invitation.Status = entity.ProjectInvitationStatusAccepted
	participantInstance, err :=
		s.ParticipantRepo.GetByProjectAndSiteUUID(invitation.ProjectUUID, invitation.SiteUUID)
	if err != nil {
		return err
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusJoined
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return err
	}
	if err := s.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return err
	}
	return nil
}

// ProcessProjectRejection process invitation rejection response
func (s *ProjectService) ProcessProjectRejection(invitationUUID string) error {
	invitationInstance, err := s.InvitationRepo.GetByUUID(invitationUUID)
	if err != nil {
		return err
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}
	invitation.Status = entity.ProjectInvitationStatusRejected
	participantInstance, err :=
		s.ParticipantRepo.GetByProjectAndSiteUUID(invitation.ProjectUUID, invitation.SiteUUID)
	if err != nil {
		return err
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusRejected
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return err
	}
	if err := s.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return err
	}
	return nil
}

// ProcessInvitationRevocation removes the project and updates invitation status
func (s *ProjectService) ProcessInvitationRevocation(invitationUUID string) error {
	invitationInstance, err := s.InvitationRepo.GetByUUID(invitationUUID)
	if err != nil {
		return err
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}
	invitation.Status = entity.ProjectInvitationStatusRevoked

	// delete the project
	if err := s.ProjectRepo.DeleteByUUID(invitation.ProjectUUID); err != nil {
		return err
	}
	if err := s.ParticipantRepo.DeleteByProjectUUID(invitation.ProjectUUID); err != nil {
		return err
	}
	if err := s.ProjectDataRepo.DeleteByProjectUUID(invitation.ProjectUUID); err != nil {
		return err
	}
	if err := s.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return err
	}
	return nil
}

// ProcessParticipantDismissal processes participant dismissal event by updating the repo records
func (s *ProjectService) ProcessParticipantDismissal(projectUUID, siteUUID string, isCurrentSite bool) error {
	// dismiss data association
	dataListInstance, err := s.ProjectDataRepo.GetListByProjectAndSiteUUID(projectUUID, siteUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project data")
	}
	dataList := dataListInstance.([]entity.ProjectData)
	for _, data := range dataList {
		if data.Status == entity.ProjectDataStatusAssociated {
			data.Status = entity.ProjectDataStatusDismissed
			if err := s.ProjectDataRepo.UpdateStatusByUUID(&data); err != nil {
				return errors.Wrapf(err, "failed to dismiss data %s from site: %s", data.Name, data.SiteName)
			}
		}
	}
	// update participant status
	participantInstance, err := s.ParticipantRepo.GetByProjectAndSiteUUID(projectUUID, siteUUID)
	if err != nil {
		return err
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusDismissed
	if err := s.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return err
	}
	// mark project as dismissed for the dismissed site
	if isCurrentSite {
		projectInstance, err := s.ProjectRepo.GetByUUID(projectUUID)
		if err != nil {
			return err
		}
		project := projectInstance.(*entity.Project)
		project.Status = entity.ProjectStatusDismissed
		log.Info().Msgf("current site is dismissed from the project: %s(%s)", project.Name, project.UUID)
		return s.ProjectRepo.UpdateStatusByUUID(project)
	}
	return nil
}

// ProcessProjectClosing processes project closing event by update the repo records
func (s *ProjectService) ProcessProjectClosing(projectUUID string) error {
	projectInstance, err := s.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return err
	}
	project := projectInstance.(*entity.Project)
	project.Status = entity.ProjectStatusClosed

	if project.Type != entity.ProjectTypeRemote {
		return errors.New("project is not managed by other site")
	}

	if err := s.ProjectRepo.UpdateStatusByUUID(project); err != nil {
		return errors.Wrapf(err, "failed to update project status")
	}
	return nil
}

// ProcessProjectSyncRequest sync project info from fml manager and may updates data and participant info if needed
func (s *ProjectService) ProcessProjectSyncRequest(context *aggregate.ProjectSyncContext) error {
	if !context.FMLManagerConnectionInfo.Connected {
		log.Warn().Msg("ProcessProjectSyncRequest: FML manager is not connected")
		return nil
	}

	log.Info().Msg("start syncing project list...")
	client := fmlmanager.NewFMLManagerClient(context.FMLManagerConnectionInfo.Endpoint, context.FMLManagerConnectionInfo.ServerName)
	projectMap, err := client.GetProject(context.LocalSiteUUID)
	if err != nil {
		return err
	}

	projectListInstance, err := s.ProjectRepo.GetAll()
	if err != nil {
		return err
	}
	projectList := projectListInstance.([]entity.Project)

	for _, project := range projectList {
		// only sync projects managed by other sites
		if project.Type != entity.ProjectTypeRemote {
			delete(projectMap, project.UUID)
			continue
		}

		if remoteProject, ok := projectMap[project.UUID]; ok {
			if project.Status != entity.ProjectStatus(remoteProject.ProjectStatus) {
				project.Status = entity.ProjectStatus(remoteProject.ProjectStatus)
				log.Warn().Msgf("changing stale project status, project: %s, status: %v", project.UUID, project.Status)
				if err := s.ProjectRepo.UpdateStatusByUUID(&project); err != nil {
					return errors.Wrapf(err, "failed to update project (%s) status", project.UUID)
				}
				// TODO: figure out if we need to sync data and participant
			}
			delete(projectMap, project.UUID)
		} else {
			log.Warn().Msgf("removing stale project %s", project.UUID)
			if err := s.ProjectRepo.DeleteByUUID(project.UUID); err != nil {
				return errors.Wrapf(err, "failed to delete project (%s)", project.UUID)
			}
			if err := s.ParticipantRepo.DeleteByProjectUUID(project.UUID); err != nil {
				return errors.Wrapf(err, "failed to delete participants from project %s", project.UUID)
			}
			if err := s.ProjectDataRepo.DeleteByProjectUUID(project.UUID); err != nil {
				return errors.Wrapf(err, "failed to delete project data from project %s", project.UUID)
			}
		}
	}

	for _, remoteProject := range projectMap {
		// ignore projects managed by current site
		if remoteProject.ProjectManagingSiteUUID == context.LocalSiteUUID ||
			entity.ProjectStatus(remoteProject.ProjectStatus) == entity.ProjectStatusManaged {
			continue
		}
		// XXX: this project is newly added but we don't know, but this should never happen, at least if the project is in joined status
		log.Warn().Msgf("adding project %v", remoteProject)
	}
	return nil
}

// ProjectSyncService contains project sync status and provides methods for ensuring projects are synced
// TODO: we should use db data to record each project's last sync time
type ProjectSyncService struct {
	disabled                   bool
	listSyncedAt               time.Time
	projectDataSyncedAt        map[string]time.Time
	projectParticipantSyncedAt map[string]time.Time
	interval                   time.Duration
}

// NewProjectSyncService returns a syncing service instance
func NewProjectSyncService() *ProjectSyncService {
	disabled, _ := strconv.ParseBool(viper.GetString("siteportal.project.sync.disabled"))
	intervalInt, err := strconv.Atoi(viper.GetString("siteportal.project.sync.interval"))
	if err != nil || intervalInt == 0 {
		intervalInt = 20
	}
	return &ProjectSyncService{
		disabled:                   disabled,
		listSyncedAt:               time.Time{},
		projectDataSyncedAt:        map[string]time.Time{},
		projectParticipantSyncedAt: map[string]time.Time{},
		interval:                   time.Duration(intervalInt) * time.Minute,
	}
}

// EnsureProjectListSynced makes sure the project list is synced
func (s *ProjectSyncService) EnsureProjectListSynced() error {
	if s.disabled {
		return nil
	}
	if s.listSyncedAt.Add(s.interval).Before(time.Now()) {
		s.listSyncedAt = time.Now()
		exchange := event.NewSelfHttpExchange()
		if err := exchange.PostEvent(event.ProjectListSyncEvent{}); err != nil {
			return err
		}
	}
	return nil
}

// EnsureProjectDataSynced makes sure the data association info in a project is synced
func (s *ProjectSyncService) EnsureProjectDataSynced(projectUUID string) error {
	if s.disabled {
		return nil
	}
	if syncedAt, ok := s.projectDataSyncedAt[projectUUID]; !ok || syncedAt.Add(s.interval).Before(time.Now()) {
		s.projectDataSyncedAt[projectUUID] = time.Now()
		exchange := event.NewSelfHttpExchange()
		if err := exchange.PostEvent(event.ProjectDataSyncEvent{
			ProjectUUID: projectUUID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// EnsureProjectParticipantSynced makes sure the participants' info in a project is synced
func (s *ProjectSyncService) EnsureProjectParticipantSynced(projectUUID string) error {
	if s.disabled {
		return nil
	}
	if syncedAt, ok := s.projectParticipantSyncedAt[projectUUID]; !ok || syncedAt.Add(s.interval).Before(time.Now()) {
		s.projectParticipantSyncedAt[projectUUID] = time.Now()
		exchange := event.NewSelfHttpExchange()
		if err := exchange.PostEvent(event.ProjectParticipantSyncEvent{
			ProjectUUID: projectUUID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// CleanupProject remove stale project map item
func (s *ProjectSyncService) CleanupProject(joinedProjects map[string]interface{}) error {
	for k := range s.projectDataSyncedAt {
		if _, ok := joinedProjects[k]; !ok {
			delete(s.projectDataSyncedAt, k)
			delete(s.projectParticipantSyncedAt, k)
			log.Info().Str("entity", "ProjectSyncService").Msgf("removed stale project %s", k)
		}
	}
	for k := range s.projectParticipantSyncedAt {
		if _, ok := joinedProjects[k]; !ok {
			delete(s.projectParticipantSyncedAt, k)
			log.Info().Str("entity", "ProjectSyncService").Msgf("removed stale project %s", k)
		}
	}
	return nil
}
