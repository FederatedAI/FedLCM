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
	"time"

	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/service"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/valueobject"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// ProjectApp provides functions to handle project related events
type ProjectApp struct {
	ProjectRepo     repo.ProjectRepository
	ParticipantRepo repo.ProjectParticipantRepository
	SiteRepo        repo.SiteRepository
	InvitationRepo  repo.ProjectInvitationRepository
	ProjectDataRepo repo.ProjectDataRepository
}

// ProjectInvitationRequest is an invitation for asking a site to join a project
type ProjectInvitationRequest struct {
	UUID                       string                   `json:"uuid"`
	SiteUUID                   string                   `json:"site_uuid"`
	SitePartyID                uint                     `json:"site_party_id"`
	ProjectUUID                string                   `json:"project_uuid"`
	ProjectName                string                   `json:"project_name"`
	ProjectDescription         string                   `json:"project_description"`
	ProjectAutoApprovalEnabled bool                     `json:"project_auto_approval_enabled"`
	ProjectManager             string                   `json:"project_manager"`
	ProjectManagingSiteName    string                   `json:"project_managing_site_name"`
	ProjectManagingSitePartyID uint                     `json:"project_managing_site_party_id"`
	ProjectManagingSiteUUID    string                   `json:"project_managing_site_uuid"`
	ProjectCreationTime        time.Time                `json:"project_creation_time"`
	AssociatedData             []ProjectDataAssociation `json:"associated_data"`
}

// ProjectDataAssociationBase contains the basic info of a project data association
type ProjectDataAssociationBase struct {
	DataUUID string `json:"data_uuid"`
}

// ProjectDataAssociation represents a data associated in a project
type ProjectDataAssociation struct {
	ProjectDataAssociationBase
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	SiteName       string    `json:"site_name"`
	SiteUUID       string    `json:"site_uuid"`
	SitePartyID    uint      `json:"site_party_id"`
	TableName      string    `json:"table_name"`
	TableNamespace string    `json:"table_namespace"`
	CreationTime   time.Time `json:"creation_time"`
	UpdateTime     time.Time `json:"update_time"`
}

// ProjectInfoWithStatus contains project basic information and the status inferred for certain participant
type ProjectInfoWithStatus struct {
	ProjectUUID                string               `json:"project_uuid"`
	ProjectName                string               `json:"project_name"`
	ProjectDescription         string               `json:"project_description"`
	ProjectAutoApprovalEnabled bool                 `json:"project_auto_approval_enabled"`
	ProjectManager             string               `json:"project_manager"`
	ProjectManagingSiteName    string               `json:"project_managing_site_name"`
	ProjectManagingSitePartyID uint                 `json:"project_managing_site_party_id"`
	ProjectManagingSiteUUID    string               `json:"project_managing_site_uuid"`
	ProjectCreationTime        time.Time            `json:"project_creation_time"`
	ProjectStatus              entity.ProjectStatus `json:"project_status"`
}

// ProcessInvitation handles new invitation request
func (app *ProjectApp) ProcessInvitation(req *ProjectInvitationRequest) error {
	project := &entity.Project{
		UUID:                req.ProjectUUID,
		Name:                req.ProjectName,
		Description:         req.ProjectDescription,
		AutoApprovalEnabled: req.ProjectAutoApprovalEnabled,
		ProjectCreatorInfo: &valueobject.ProjectCreatorInfo{
			Manager:             req.ProjectManager,
			ManagingSiteName:    req.ProjectManagingSiteName,
			ManagingSitePartyID: req.ProjectManagingSitePartyID,
			ManagingSiteUUID:    req.ProjectManagingSiteUUID,
		},
		Repo:  app.ProjectRepo,
		Model: gorm.Model{CreatedAt: req.ProjectCreationTime},
	}
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	invitationReq := &service.ProjectInvitationRequest{
		InvitationUUID: req.UUID,
		Project:        project,
		ManagingSite:   nil,
		TargetSite:     nil,
	}
	siteInstance, err := app.SiteRepo.GetByUUID(req.ProjectManagingSiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find managing site")
	}
	site := siteInstance.(*entity.Site)
	invitationReq.ManagingSite = &service.ProjectParticipantSiteInfo{
		Name:         site.Name,
		Description:  site.Description,
		UUID:         site.UUID,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   site.ServerName,
	}
	siteInstance, err = app.SiteRepo.GetByUUID(req.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find target site")
	}
	site = siteInstance.(*entity.Site)
	invitationReq.TargetSite = &service.ProjectParticipantSiteInfo{
		Name:         site.Name,
		Description:  site.Description,
		UUID:         site.UUID,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   site.ServerName,
	}
	projectDataList := make([]entity.ProjectData, len(req.AssociatedData))
	for index, data := range req.AssociatedData {
		projectDataList[index] = entity.ProjectData{
			Name:           data.Name,
			Description:    data.Description,
			ProjectUUID:    req.ProjectUUID,
			DataUUID:       data.DataUUID,
			SiteUUID:       data.SiteUUID,
			SiteName:       data.SiteName,
			SitePartyID:    data.SitePartyID,
			TableName:      data.TableName,
			TableNamespace: data.TableNamespace,
			CreationTime:   data.CreationTime,
			UpdateTime:     data.UpdateTime,
		}
	}
	invitationReq.AssociatedData = projectDataList
	return projectService.HandleInvitationRequest(invitationReq)
}

// ProcessInvitationResponse handles invitation response
func (app *ProjectApp) ProcessInvitationResponse(invitationUUID string, accepted bool) error {
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	invitationInstance, err := app.InvitationRepo.GetByUUID(invitationUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find invitation")
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)

	projectInstance, err := app.ProjectRepo.GetByUUID(invitation.ProjectUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find project")
	}
	project := projectInstance.(*entity.Project)

	invitationReq := &service.ProjectInvitationRequest{
		InvitationUUID: invitation.UUID,
		Project:        project,
		ManagingSite:   nil,
		TargetSite:     nil,
	}

	siteInstance, err := app.SiteRepo.GetByUUID(project.ManagingSiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find managing site")
	}
	site := siteInstance.(*entity.Site)
	invitationReq.ManagingSite = &service.ProjectParticipantSiteInfo{
		Name:         site.Name,
		Description:  site.Description,
		UUID:         site.UUID,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   site.ServerName,
	}
	siteInstance, err = app.SiteRepo.GetByUUID(invitation.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find target site")
	}
	site = siteInstance.(*entity.Site)
	invitationReq.TargetSite = &service.ProjectParticipantSiteInfo{
		Name:         site.Name,
		Description:  site.Description,
		UUID:         site.UUID,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   site.ServerName,
	}

	if accepted {
		participantListInstance, err := app.ParticipantRepo.GetByProjectUUID(project.UUID)
		if err != nil {
			return err
		}
		participantList := participantListInstance.([]entity.ProjectParticipant)
		var otherSiteList []service.ProjectParticipantSiteInfo
		for _, participant := range participantList {
			if participant.Status == entity.ProjectParticipantStatusJoined && participant.SiteUUID != invitation.SiteUUID {
				siteInstance, err = app.SiteRepo.GetByUUID(participant.SiteUUID)
				if err != nil {
					return errors.Wrapf(err, "failed to find target site")
				}
				site = siteInstance.(*entity.Site)
				otherSiteList = append(otherSiteList, service.ProjectParticipantSiteInfo{
					Name:         site.Name,
					Description:  site.Description,
					UUID:         site.UUID,
					PartyID:      site.PartyID,
					ExternalHost: site.ExternalHost,
					ExternalPort: site.ExternalPort,
					HTTPS:        site.HTTPS,
					ServerName:   site.ServerName,
				})
			}
		}
		return projectService.HandleInvitationAcceptance(invitationReq, otherSiteList)
	} else {
		return projectService.HandleInvitationRejection(invitationReq)
	}
}

// ProcessInvitationRevocation handles invitation revocation request
func (app *ProjectApp) ProcessInvitationRevocation(invitationUUID string) error {
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	invitationInstance, err := app.InvitationRepo.GetByUUID(invitationUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find invitation")
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)

	projectInstance, err := app.ProjectRepo.GetByUUID(invitation.ProjectUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find project")
	}
	project := projectInstance.(*entity.Project)

	invitationReq := &service.ProjectInvitationRequest{
		InvitationUUID: invitation.UUID,
		Project:        project,
		ManagingSite:   nil,
		TargetSite:     nil,
	}

	siteInstance, err := app.SiteRepo.GetByUUID(project.ManagingSiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find managing site")
	}
	site := siteInstance.(*entity.Site)
	invitationReq.ManagingSite = &service.ProjectParticipantSiteInfo{
		Name:         site.Name,
		Description:  site.Description,
		UUID:         site.UUID,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   site.ServerName,
	}
	siteInstance, err = app.SiteRepo.GetByUUID(invitation.SiteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find target site")
	}
	site = siteInstance.(*entity.Site)
	invitationReq.TargetSite = &service.ProjectParticipantSiteInfo{
		Name:         site.Name,
		Description:  site.Description,
		UUID:         site.UUID,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   site.ServerName,
	}
	return projectService.HandleInvitationRevocation(invitationReq)
}

// ProcessParticipantInfoUpdate handles sites info update event
func (app *ProjectApp) ProcessParticipantInfoUpdate(siteUUID string) error {
	siteInstance, err := app.SiteRepo.GetByUUID(siteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find updated site")
	}
	updatedSite := siteInstance.(*entity.Site)

	list, err := app.SiteRepo.GetSiteList()
	if err != nil {
		return errors.Wrapf(err, "failed to list all sites")
	}
	allSites := list.([]entity.Site)

	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	allSitesInfo := make([]service.ProjectParticipantSiteInfo, len(allSites))
	for index, site := range allSites {
		allSitesInfo[index] = service.ProjectParticipantSiteInfo{
			Name:         site.Name,
			Description:  site.Description,
			UUID:         site.UUID,
			PartyID:      site.PartyID,
			ExternalHost: site.ExternalHost,
			ExternalPort: site.ExternalPort,
			HTTPS:        site.HTTPS,
			ServerName:   site.ServerName,
		}
	}

	return projectService.HandleParticipantInfoUpdate(service.ProjectParticipantSiteInfo{
		Name:         updatedSite.Name,
		Description:  updatedSite.Description,
		UUID:         updatedSite.UUID,
		PartyID:      updatedSite.PartyID,
		ExternalHost: updatedSite.ExternalHost,
		ExternalPort: updatedSite.ExternalPort,
		HTTPS:        updatedSite.HTTPS,
		ServerName:   updatedSite.ServerName,
	}, allSitesInfo)
}

// ProcessParticipantLeaving handles participate leaving
func (app *ProjectApp) ProcessParticipantLeaving(projectUUID, siteUUID string) error {
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	otherSiteList, err := app.getPeerParticipantList(projectUUID, siteUUID)
	if err != nil {
		return errors.New("failed to get peer participant list")
	}
	return projectService.HandleParticipantLeaving(projectUUID, siteUUID, otherSiteList)
}

// ProcessParticipantDismissal handles participate dismissal
func (app *ProjectApp) ProcessParticipantDismissal(projectUUID, siteUUID string) error {
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}

	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find project")
	}
	project := projectInstance.(*entity.Project)

	otherSiteList, err := app.getPeerParticipantList(projectUUID, project.ManagingSiteUUID)
	if err != nil {
		return errors.New("failed to get peer participant list")
	}

	siteInstance, err := app.SiteRepo.GetByUUID(siteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find target site")
	}
	site := siteInstance.(*entity.Site)
	targetSite := service.ProjectParticipantSiteInfo{
		Name:         site.Name,
		Description:  site.Description,
		UUID:         site.UUID,
		PartyID:      site.PartyID,
		ExternalHost: site.ExternalHost,
		ExternalPort: site.ExternalPort,
		HTTPS:        site.HTTPS,
		ServerName:   site.ServerName,
	}
	return projectService.HandleParticipantDismissal(projectUUID, targetSite, otherSiteList)
}

// ProcessParticipantUnregistration handles participant unregistration event
func (app *ProjectApp) ProcessParticipantUnregistration(siteUUID string) error {
	list, err := app.SiteRepo.GetSiteList()
	if err != nil {
		return errors.Wrapf(err, "failed to list all sites")
	}
	allSites := list.([]entity.Site)

	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	allSitesInfo := make([]service.ProjectParticipantSiteInfo, len(allSites))
	for index, site := range allSites {
		allSitesInfo[index] = service.ProjectParticipantSiteInfo{
			Name:         site.Name,
			Description:  site.Description,
			UUID:         site.UUID,
			PartyID:      site.PartyID,
			ExternalHost: site.ExternalHost,
			ExternalPort: site.ExternalPort,
			HTTPS:        site.HTTPS,
			ServerName:   site.ServerName,
		}
	}

	return projectService.HandleParticipantUnregistration(siteUUID, allSitesInfo)
}

// ProcessDataAssociation handles new data association
func (app *ProjectApp) ProcessDataAssociation(projectUUID string, data *ProjectDataAssociation) error {
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	otherSiteList, err := app.getPeerParticipantList(projectUUID, data.SiteUUID)
	if err != nil {
		return errors.New("failed to get peer participant list")
	}
	return projectService.HandleDataAssociation(&entity.ProjectData{
		Name:           data.Name,
		Description:    data.Description,
		ProjectUUID:    projectUUID,
		DataUUID:       data.DataUUID,
		SiteUUID:       data.SiteUUID,
		SiteName:       data.SiteName,
		SitePartyID:    data.SitePartyID,
		TableName:      data.TableName,
		TableNamespace: data.TableNamespace,
		CreationTime:   data.CreationTime,
		UpdateTime:     data.UpdateTime,
	}, otherSiteList)
}

// ProcessDataDismissal handles data association dismissal
func (app *ProjectApp) ProcessDataDismissal(projectUUID string, baseData *ProjectDataAssociationBase) error {
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	dataInstance, err := app.ProjectDataRepo.GetByProjectAndDataUUID(projectUUID, baseData.DataUUID)
	if err != nil {
		if err == repo.ErrProjectDataNotFound {
			return nil
		}
		return errors.Wrap(err, "failed to query association")
	}
	data := dataInstance.(*entity.ProjectData)
	otherSiteList, err := app.getPeerParticipantList(data.ProjectUUID, data.SiteUUID)
	if err != nil {
		return errors.New("failed to get peer participant list")
	}
	return projectService.HandleDataDismissal(data.ProjectUUID, data.DataUUID, otherSiteList)
}

func (app *ProjectApp) getPeerParticipantList(projectUUID, siteUUID string) ([]service.ProjectParticipantSiteInfo, error) {
	participantListInstance, err := app.ParticipantRepo.GetByProjectUUID(projectUUID)
	if err != nil {
		return nil, err
	}
	participantList := participantListInstance.([]entity.ProjectParticipant)
	var otherSiteList []service.ProjectParticipantSiteInfo
	for _, participant := range participantList {
		if participant.SiteUUID != siteUUID &&
			(participant.Status == entity.ProjectParticipantStatusJoined ||
				participant.Status == entity.ProjectParticipantStatusOwner) {
			siteInstance, err := app.SiteRepo.GetByUUID(participant.SiteUUID)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to find site")
			}
			site := siteInstance.(*entity.Site)
			otherSiteList = append(otherSiteList, service.ProjectParticipantSiteInfo{
				Name:         site.Name,
				Description:  site.Description,
				UUID:         site.UUID,
				PartyID:      site.PartyID,
				ExternalHost: site.ExternalHost,
				ExternalPort: site.ExternalPort,
				HTTPS:        site.HTTPS,
				ServerName:   site.ServerName,
			})
		}
	}
	return otherSiteList, nil
}

// ProcessProjectClosing handles project closing event
func (app *ProjectApp) ProcessProjectClosing(projectUUID string) error {
	projectService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		InvitationRepo:  app.InvitationRepo,
		ParticipantRepo: app.ParticipantRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to find project")
	}
	project := projectInstance.(*entity.Project)

	otherSiteList, err := app.getPeerParticipantList(projectUUID, project.ManagingSiteUUID)
	if err != nil {
		return errors.New("failed to get peer participant list")
	}
	return projectService.HandleProjectClosing(projectUUID, otherSiteList)
}

// ListProjectByParticipant returns information of projects related to the specified site
func (app *ProjectApp) ListProjectByParticipant(participantUUID string) (map[string]ProjectInfoWithStatus, error) {
	if participantUUID == "" {
		return nil, errors.New("missing participant uuid")
	}
	participantListInstance, err := app.ParticipantRepo.GetBySiteUUID(participantUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list participant %s", participantUUID)
	}
	participantList := participantListInstance.([]entity.ProjectParticipant)

	projectMap := map[string]ProjectInfoWithStatus{}
	for _, participant := range participantList {
		projectInstance, err := app.ProjectRepo.GetByUUID(participant.ProjectUUID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get project info for project %s", participant.ProjectUUID)
		}
		project := projectInstance.(*entity.Project)

		projectInfo := ProjectInfoWithStatus{
			ProjectUUID:                project.UUID,
			ProjectName:                project.Name,
			ProjectDescription:         project.Description,
			ProjectAutoApprovalEnabled: project.AutoApprovalEnabled,
			ProjectManager:             project.Manager,
			ProjectManagingSiteName:    project.ManagingSiteName,
			ProjectManagingSitePartyID: project.ManagingSitePartyID,
			ProjectManagingSiteUUID:    project.ManagingSiteUUID,
			ProjectCreationTime:        project.CreatedAt,
			ProjectStatus:              0,
		}

		// ignore participant status for closed project
		if project.Status == entity.ProjectStatusClosed {
			projectInfo.ProjectStatus = entity.ProjectStatusClosed
		} else {
			switch participant.Status {
			case entity.ProjectParticipantStatusPending:
				projectInfo.ProjectStatus = entity.ProjectStatusPending
			case entity.ProjectParticipantStatusJoined:
				projectInfo.ProjectStatus = entity.ProjectStatusJoined
			case entity.ProjectParticipantStatusDismissed:
				projectInfo.ProjectStatus = entity.ProjectStatusDismissed
			case entity.ProjectParticipantStatusLeft:
				projectInfo.ProjectStatus = entity.ProjectStatusLeft
			case entity.ProjectParticipantStatusRevoked:
				projectInfo.ProjectStatus = entity.ProjectStatusDismissed
			case entity.ProjectParticipantStatusRejected:
				projectInfo.ProjectStatus = entity.ProjectStatusRejected
			case entity.ProjectParticipantStatusOwner:
				projectInfo.ProjectStatus = entity.ProjectStatusManaged
			case entity.ProjectParticipantStatusUnknown:
				projectInfo.ProjectStatus = entity.ProjectStatusClosed
			}
		}
		projectMap[projectInfo.ProjectUUID] = projectInfo
	}
	return projectMap, err
}

// ListDataAssociationByProject returns current associated data for certain project
func (app *ProjectApp) ListDataAssociationByProject(projectUUID string) (map[string]ProjectDataAssociation, error) {
	if projectUUID == "" {
		return nil, errors.New("missing project uuid")
	}
	dataListInstance, err := app.ProjectDataRepo.GetListByProjectUUID(projectUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load data assoication")
	}
	dataList := dataListInstance.([]entity.ProjectData)

	projectDataMap := map[string]ProjectDataAssociation{}
	for _, data := range dataList {
		if data.Status == entity.ProjectDataStatusAssociated {
			projectDataMap[data.DataUUID] = ProjectDataAssociation{
				ProjectDataAssociationBase: ProjectDataAssociationBase{
					DataUUID: data.DataUUID,
				},
				Name:           data.Name,
				Description:    data.Description,
				SiteName:       data.SiteName,
				SiteUUID:       data.SiteUUID,
				SitePartyID:    data.SitePartyID,
				TableName:      data.TableName,
				TableNamespace: data.TableNamespace,
				CreationTime:   data.CreationTime,
				UpdateTime:     data.UpdateTime,
			}
		}
	}
	return projectDataMap, err
}

// ListParticipantByProject returns participant list in a project
func (app *ProjectApp) ListParticipantByProject(projectUUID string) (map[string]entity.ProjectParticipant, error) {
	participantListInstance, err := app.ParticipantRepo.GetByProjectUUID(projectUUID)
	if err != nil {
		return nil, err
	}
	participantList := participantListInstance.([]entity.ProjectParticipant)
	participantMap := map[string]entity.ProjectParticipant{}
	for index, participant := range participantList {
		participantMap[participant.SiteUUID] = participantList[index]
	}
	return participantMap, nil
}
