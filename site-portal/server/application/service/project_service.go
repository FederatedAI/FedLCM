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

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/aggregate"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/service"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// ProjectApp provides interfaces for project management related APIs
type ProjectApp struct {
	ProjectRepo        repo.ProjectRepository
	ParticipantRepo    repo.ProjectParticipantRepository
	SiteRepo           repo.SiteRepository
	InvitationRepo     repo.ProjectInvitationRepository
	ProjectDataRepo    repo.ProjectDataRepository
	LocalDataRepo      repo.LocalDataRepository
	JobApp             *JobApp
	ProjectSyncService *service.ProjectSyncService
}

// ProjectCreationRequest is the request for creating a new local project
type ProjectCreationRequest struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	AutoApprovalEnabled bool   `json:"auto_approval_enabled"`
}

// ProjectInfo is the detailed project info
type ProjectInfo struct {
	ProjectListItemBase
	AutoApprovalEnabled bool `json:"auto_approval_enabled"`
}

// ProjectListItem contains basic info of a project plus data & job statistics
type ProjectListItem struct {
	ProjectListItemBase
	ParticipantsNum int64 `json:"participants_num"`
	LocalDataNum    int64 `json:"local_data_num"`
	RemoteDataNum   int64 `json:"remote_data_num"`
	RunningJobNum   int64 `json:"running_job_num"`
	SuccessJobNum   int64 `json:"success_job_num"`
	PendingJobExist bool  `json:"pending_job_exist"`
}

// ProjectListItemBase contains basic info of a project
type ProjectListItemBase struct {
	Name                string    `json:"name"`
	Description         string    `json:"description"`
	UUID                string    `json:"uuid"`
	CreationTime        time.Time `json:"creation_time"`
	Manager             string    `json:"manager"`
	ManagingSiteName    string    `json:"managing_site_name"`
	ManagingSitePartyID uint      `json:"managing_site_party_id"`
	ManagedByThisSite   bool      `json:"managed_by_this_site"`
}

// ProjectListItemClosed is a closed project
type ProjectListItemClosed struct {
	ProjectListItemBase
	ClosingStatus string `json:"closing_status"`
}

// ProjectList contains joined projects and pending projects
type ProjectList struct {
	JoinedProject  []ProjectListItem       `json:"joined_projects"`
	PendingProject []ProjectListItemBase   `json:"invited_projects"`
	ClosedProject  []ProjectListItemClosed `json:"closed_projects"`
}

// ProjectParticipant contains info of a project participant
type ProjectParticipant struct {
	ProjectParticipantBase
	CreationTime  time.Time                       `json:"creation_time"`
	Status        entity.ProjectParticipantStatus `json:"status"`
	IsCurrentSite bool                            `json:"is_current_site"`
}

// ProjectParticipantBase contains the basic info of a participant
type ProjectParticipantBase struct {
	UUID        string `json:"uuid"`
	PartyID     uint   `json:"party_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProjectAutoApprovalStatus is a container for holding the auto-approval status value
type ProjectAutoApprovalStatus struct {
	Enabled bool `json:"enabled"`
}

// ProjectInvitationRequest is the request a site received for joining a project
type ProjectInvitationRequest struct {
	UUID                       string    `json:"uuid"`
	SiteUUID                   string    `json:"site_uuid"`
	SitePartyID                uint      `json:"site_party_id"`
	ProjectUUID                string    `json:"project_uuid"`
	ProjectName                string    `json:"project_name"`
	ProjectDescription         string    `json:"project_description"`
	ProjectAutoApprovalEnabled bool      `json:"project_auto_approval_enabled"`
	ProjectManager             string    `json:"project_manager"`
	ProjectManagingSiteName    string    `json:"project_managing_site_name"`
	ProjectManagingSitePartyID uint      `json:"project_managing_site_party_id"`
	ProjectManagingSiteUUID    string    `json:"project_managing_site_uuid"`
	ProjectCreationTime        time.Time `json:"project_creation_time"`
}

// ProjectData contains information of an associated project data
type ProjectData struct {
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	DataID               string    `json:"data_id"`
	CreationTime         time.Time `json:"creation_time"`
	UpdatedTime          time.Time `json:"update_time"`
	ProvidingSiteUUID    string    `json:"providing_site_uuid"`
	ProvidingSiteName    string    `json:"providing_site_name"`
	ProvidingSitePartyID uint      `json:"providing_site_party_id"`
	IsLocal              bool      `json:"is_local"`
}

// ProjectDataAssociationRequest is the request to associate a local data to a project
type ProjectDataAssociationRequest struct {
	Name     string `json:"name"`
	DataUUID string `json:"data_id"`
}

// ProjectResourceSyncRequest is the request to sync certain project resource
type ProjectResourceSyncRequest struct {
	ProjectUUID string `json:"project_uuid"`
}

// CreateLocalProject creates a project locally
func (app *ProjectApp) CreateLocalProject(req *ProjectCreationRequest, username string) error {
	project := &entity.Project{
		Name:                req.Name,
		Description:         req.Description,
		AutoApprovalEnabled: req.AutoApprovalEnabled,
		Type:                entity.ProjectTypeLocal,
		ProjectCreatorInfo:  valueobject.ProjectCreatorInfo{},
		Repo:                app.ProjectRepo,
	}
	site := entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return errors.Wrapf(err, "failed to load site info")
	}
	creatorInfo := valueobject.ProjectCreatorInfo{
		Manager:             username,
		ManagingSiteName:    site.Name,
		ManagingSitePartyID: site.PartyID,
		ManagingSiteUUID:    site.UUID,
	}
	project.ProjectCreatorInfo = creatorInfo
	if err := project.Create(); err != nil {
		return err
	}
	participant := &entity.ProjectParticipant{
		UUID:            uuid.NewV4().String(),
		ProjectUUID:     project.UUID,
		SiteUUID:        site.UUID,
		SiteName:        site.Name,
		SitePartyID:     site.PartyID,
		SiteDescription: site.Description,
		Status:          entity.ProjectParticipantStatusOwner,
	}
	return app.ParticipantRepo.Create(participant)
}

// List returns all projects this site joined or pending on this site
func (app *ProjectApp) List() (*ProjectList, error) {
	if err := app.ProjectSyncService.EnsureProjectListSynced(); err != nil {
		log.Err(err).Msg("failed to sync project list")
	}
	currentSite, err := app.LoadSite()
	if err != nil {
		return nil, err
	}
	instanceList, err := app.ProjectRepo.GetAll()
	if err != nil {
		return nil, err
	}
	projectList := &ProjectList{
		JoinedProject:  make([]ProjectListItem, 0),
		PendingProject: make([]ProjectListItemBase, 0),
	}
	domainProjectList := instanceList.([]entity.Project)
	for _, project := range domainProjectList {
		projectItemBase := ProjectListItemBase{
			Name:                project.Name,
			Description:         project.Description,
			UUID:                project.UUID,
			CreationTime:        project.CreatedAt,
			Manager:             project.Manager,
			ManagingSiteName:    project.ManagingSiteName,
			ManagingSitePartyID: project.ManagingSitePartyID,
			ManagedByThisSite:   project.Status == entity.ProjectStatusManaged,
		}
		switch project.Status {
		case entity.ProjectStatusManaged, entity.ProjectStatusJoined:
			projectAggregate := aggregate.ProjectAggregate{
				Project:         &project,
				Participant:     nil,
				ProjectRepo:     app.ProjectRepo,
				ParticipantRepo: app.ParticipantRepo,
				InvitationRepo:  app.InvitationRepo,
				DataRepo:        app.ProjectDataRepo,
			}
			participantsNum, err := projectAggregate.CountParticipant()
			if err != nil {
				return nil, err
			}

			joinedProjectListItem := ProjectListItem{
				ProjectListItemBase: projectItemBase,
				ParticipantsNum:     participantsNum,
				LocalDataNum:        0,
				RemoteDataNum:       0,
				RunningJobNum:       0,
				SuccessJobNum:       0,
				PendingJobExist:     false,
			}

			associatedDataList, err := app.ListData(joinedProjectListItem.UUID, "")
			if err != nil {
				return nil, err
			}
			for _, data := range associatedDataList {
				if data.ProvidingSiteUUID == currentSite.UUID {
					joinedProjectListItem.LocalDataNum++
				} else {
					joinedProjectListItem.RemoteDataNum++
				}
			}

			// XXX: this should be placed in the job context?
			jobList, err := app.JobApp.List(joinedProjectListItem.UUID)
			if err != nil {
				return nil, err
			}
			for _, job := range jobList {
				if job.PendingOnThisSite {
					joinedProjectListItem.PendingJobExist = true
				}
				if job.Status == entity.JobStatusRunning {
					joinedProjectListItem.RunningJobNum++
				} else if job.Status == entity.JobStatusSucceeded {
					joinedProjectListItem.SuccessJobNum++
				}
			}

			projectList.JoinedProject = append(projectList.JoinedProject, joinedProjectListItem)
		case entity.ProjectStatusPending:
			projectList.PendingProject = append(projectList.PendingProject, projectItemBase)
		case entity.ProjectStatusClosed, entity.ProjectStatusDismissed, entity.ProjectStatusLeft:
			closingStatusStr := func(status entity.ProjectStatus) string {
				switch status {
				case entity.ProjectStatusClosed:
					return "closed by the managing site"
				case entity.ProjectStatusLeft:
					return "left"
				case entity.ProjectStatusDismissed:
					return "dismissed by the managing site"
				default:
					return "unknown"
				}
			}(project.Status)
			projectList.ClosedProject = append(projectList.ClosedProject, ProjectListItemClosed{
				ProjectListItemBase: projectItemBase,
				ClosingStatus:       closingStatusStr,
			})
		}
	}

	joinedProjects := map[string]interface{}{}
	for _, project := range projectList.JoinedProject {
		joinedProjects[project.UUID] = nil
	}
	_ = app.ProjectSyncService.CleanupProject(joinedProjects)

	return projectList, nil
}

// ListParticipant returns participants of a site or all participant registered in FML manager
func (app *ProjectApp) ListParticipant(uuid string, all bool) ([]ProjectParticipant, error) {
	if err := app.ProjectSyncService.EnsureProjectParticipantSynced(uuid); err != nil {
		log.Err(err).Msg("failed to sync project participant")
	}
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return nil, err
	}
	project := &entity.Project{
		UUID: uuid,
		Repo: app.ProjectRepo,
	}
	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	participants, err := projectAggregate.ListParticipant(all, &aggregate.FMLManagerConnectionInfo{
		Connected:  site.FMLManagerConnected,
		Endpoint:   site.FMLManagerEndpoint,
		ServerName: site.FMLManagerServerName,
	})
	if err != nil {
		return nil, err
	}
	list := make([]ProjectParticipant, len(participants))
	for i, participant := range participants {
		list[i] = ProjectParticipant{
			ProjectParticipantBase: ProjectParticipantBase{
				UUID:        participant.SiteUUID,
				PartyID:     participant.SitePartyID,
				Name:        participant.SiteName,
				Description: participant.SiteDescription,
			},
			CreationTime:  participant.CreatedAt,
			Status:        participant.Status,
			IsCurrentSite: participant.SiteUUID == site.UUID,
		}
	}
	return list, nil
}

// InviteParticipant invites certain participant to join current project
func (app *ProjectApp) InviteParticipant(uuid string, targetSite *ProjectParticipantBase) error {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return err
	}
	if err := app.EnsureProjectIsOpen(uuid); err != nil {
		return err
	}
	projectInstance, err := app.ProjectRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	project := projectInstance.(*entity.Project)
	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	return projectAggregate.InviteParticipant(&aggregate.ProjectInvitationContext{
		FMLManagerConnectionInfo: &aggregate.FMLManagerConnectionInfo{
			Connected:  site.FMLManagerConnected,
			Endpoint:   site.FMLManagerEndpoint,
			ServerName: site.FMLManagerServerName,
		},
		SitePartyID:     targetSite.PartyID,
		SiteUUID:        targetSite.UUID,
		SiteName:        targetSite.Name,
		SiteDescription: targetSite.Description,
	})
}

// ProcessInvitation processes the invitation from FML manager
func (app *ProjectApp) ProcessInvitation(req *ProjectInvitationRequest) error {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return err
	}
	project := &entity.Project{
		UUID:                req.ProjectUUID,
		Name:                req.ProjectName,
		Description:         req.ProjectDescription,
		AutoApprovalEnabled: req.ProjectAutoApprovalEnabled,
		ProjectCreatorInfo: valueobject.ProjectCreatorInfo{
			Manager:             req.ProjectManager,
			ManagingSiteName:    req.ProjectManagingSiteName,
			ManagingSitePartyID: req.ProjectManagingSitePartyID,
			ManagingSiteUUID:    req.ProjectManagingSiteUUID,
		},
		Repo:  app.ProjectRepo,
		Model: gorm.Model{CreatedAt: req.ProjectCreationTime},
	}
	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	invitation := &entity.ProjectInvitation{
		UUID:        req.UUID,
		ProjectUUID: req.ProjectUUID,
		SiteUUID:    req.SiteUUID,
	}
	return projectAggregate.ProcessInvitation(invitation, site.UUID)
}

// ToggleAutoApprovalStatus changes the project's auto-approval status
func (app *ProjectApp) ToggleAutoApprovalStatus(uuid string, status *ProjectAutoApprovalStatus) error {
	if err := app.EnsureProjectIsOpen(uuid); err != nil {
		return err
	}
	project := &entity.Project{
		UUID:                uuid,
		AutoApprovalEnabled: status.Enabled,
	}
	return app.ProjectRepo.UpdateAutoApprovalStatusByUUID(project)
}

// GetProject returns detailed info of a project
func (app *ProjectApp) GetProject(uuid string) (*ProjectInfo, error) {
	projectInstance, err := app.ProjectRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	project := projectInstance.(*entity.Project)
	return &ProjectInfo{
		ProjectListItemBase: ProjectListItemBase{
			Name:                project.Name,
			Description:         project.Description,
			UUID:                project.UUID,
			CreationTime:        project.CreatedAt,
			Manager:             project.Manager,
			ManagingSiteName:    project.ManagingSiteName,
			ManagingSitePartyID: project.ManagingSitePartyID,
			ManagedByThisSite:   project.Status == entity.ProjectStatusManaged,
		},
		AutoApprovalEnabled: project.AutoApprovalEnabled,
	}, nil
}

// JoinOrRejectProject joins or refuses to join a pending project
func (app *ProjectApp) JoinOrRejectProject(uuid string, join bool) error {
	site, err := app.LoadSite()
	if err != nil {
		return err
	}
	projectInstance, err := app.ProjectRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	project := projectInstance.(*entity.Project)
	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	fmlManagerConnectionInfo := &aggregate.FMLManagerConnectionInfo{
		Connected:  site.FMLManagerConnected,
		Endpoint:   site.FMLManagerEndpoint,
		ServerName: site.FMLManagerServerName,
	}
	if join {
		return projectAggregate.JoinProject(fmlManagerConnectionInfo)
	} else {
		return projectAggregate.RejectProject(fmlManagerConnectionInfo)
	}
}

// LeaveProject removes the current site from the specified project
func (app *ProjectApp) LeaveProject(projectUUID string) error {
	site, err := app.LoadSite()
	if err != nil {
		return err
	}

	if err := app.ensureNoRunningJobs(projectUUID, site.UUID); err != nil {
		return err
	}

	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project")
	}
	project := projectInstance.(*entity.Project)

	participantInstance, err := app.ParticipantRepo.GetByProjectAndSiteUUID(projectUUID, site.UUID)
	if err != nil {
		return errors.Wrap(err, "failed to query participant")
	}
	participant := participantInstance.(*entity.ProjectParticipant)

	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     participant,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	fmlManagerConnectionInfo := &aggregate.FMLManagerConnectionInfo{
		Connected:  site.FMLManagerConnected,
		Endpoint:   site.FMLManagerEndpoint,
		ServerName: site.FMLManagerServerName,
	}

	return projectAggregate.LeaveProject(fmlManagerConnectionInfo)
}

// CloseProject closes the managed project
func (app *ProjectApp) CloseProject(projectUUID string) error {
	if err := app.ensureNoRunningJobs(projectUUID, ""); err != nil {
		return err
	}

	site, err := app.LoadSite()
	if err != nil {
		return err
	}

	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project")
	}
	project := projectInstance.(*entity.Project)

	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	fmlManagerConnectionInfo := &aggregate.FMLManagerConnectionInfo{
		Connected:  site.FMLManagerConnected,
		Endpoint:   site.FMLManagerEndpoint,
		ServerName: site.FMLManagerServerName,
	}
	return projectAggregate.CloseProject(fmlManagerConnectionInfo)
}

// ProcessInvitationResponse handles the invitation response
func (app *ProjectApp) ProcessInvitationResponse(uuid string, accepted bool) error {
	domainService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	if accepted {
		return domainService.ProcessProjectAcceptance(uuid)
	} else {
		return domainService.ProcessProjectRejection(uuid)
	}
}

// ProcessInvitationRevocation handles the invitation revocation
func (app *ProjectApp) ProcessInvitationRevocation(uuid string) error {
	domainService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	return domainService.ProcessInvitationRevocation(uuid)
}

// CreateRemoteProjectParticipants processes a list of participants to create from FML manager for a remote project
func (app *ProjectApp) CreateRemoteProjectParticipants(projectUUID string, participants []entity.ProjectParticipant) error {
	projectAggregate := aggregate.ProjectAggregate{
		Project: &entity.Project{
			UUID: projectUUID,
			Repo: app.ProjectRepo,
		},
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	return projectAggregate.CreateRemoteProjectParticipants(participants)
}

// LoadSite is a helper function to return site entity object
func (app *ProjectApp) LoadSite() (*entity.Site, error) {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return nil, errors.Wrapf(err, "failed to load site info")
	}
	return site, nil
}

// RemoveProjectParticipants removes joined participant or revoke invitation
func (app *ProjectApp) RemoveProjectParticipants(projectUUID string, siteUUID string) error {
	site, err := app.LoadSite()
	if err != nil {
		return err
	}
	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to get project")
	}
	project := projectInstance.(*entity.Project)
	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	fmlManagerConnectionInfo := &aggregate.FMLManagerConnectionInfo{
		Connected:  site.FMLManagerConnected,
		Endpoint:   site.FMLManagerEndpoint,
		ServerName: site.FMLManagerServerName,
	}

	participantInstance, err := app.ParticipantRepo.GetByProjectAndSiteUUID(projectUUID, siteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to query participant info")
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	if participant.Status == entity.ProjectParticipantStatusJoined {
		if err := app.ensureNoRunningJobs(projectUUID, siteUUID); err != nil {
			return err
		}
	}
	return projectAggregate.RemoveParticipant(siteUUID, fmlManagerConnectionInfo)
}

// ProcessParticipantInfoUpdate processes participant info update event by updating impacted repo records
func (app *ProjectApp) ProcessParticipantInfoUpdate(participant *ProjectParticipantBase) error {
	toUpdateProjectTemplate := &entity.Project{
		ProjectCreatorInfo: valueobject.ProjectCreatorInfo{
			ManagingSiteName:    participant.Name,
			ManagingSitePartyID: participant.PartyID,
			ManagingSiteUUID:    participant.UUID,
		},
	}
	if err := app.ProjectRepo.UpdateManagingSiteInfoBySiteUUID(toUpdateProjectTemplate); err != nil {
		return errors.Wrapf(err, "failed to update projects creator info")
	}
	toUpdateParticipantTemplate := &entity.ProjectParticipant{
		SiteUUID:        participant.UUID,
		SiteName:        participant.Name,
		SitePartyID:     participant.PartyID,
		SiteDescription: participant.Description,
	}
	if err := app.ParticipantRepo.UpdateParticipantInfoBySiteUUID(toUpdateParticipantTemplate); err != nil {
		return errors.Wrapf(err, "failed to update projects participants info")
	}
	toUpdateDataTemplate := &entity.ProjectData{
		SiteUUID:    participant.UUID,
		SiteName:    participant.Name,
		SitePartyID: participant.PartyID,
	}
	if err := app.ProjectDataRepo.UpdateSiteInfoBySiteUUID(toUpdateDataTemplate); err != nil {
		return errors.Wrapf(err, "failed to update projects data info")
	}
	return nil
}

// ProcessParticipantLeaving processes participant leaving event by updating the repo record
func (app *ProjectApp) ProcessParticipantLeaving(projectUUID string, siteUUID string) error {
	participantInstance, err := app.ParticipantRepo.GetByProjectAndSiteUUID(projectUUID, siteUUID)
	if err != nil {
		return err
	}
	participant := participantInstance.(*entity.ProjectParticipant)
	participant.Status = entity.ProjectParticipantStatusLeft
	if err := app.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
		return err
	}
	return nil
}

// ProcessParticipantDismissal processes participant dismissal event
func (app *ProjectApp) ProcessParticipantDismissal(projectUUID string, siteUUID string) error {
	if err := app.ensureNoRunningJobs(projectUUID, siteUUID); err != nil {
		return err
	}
	site, err := app.LoadSite()
	if err != nil {
		return err
	}
	domainService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	return domainService.ProcessParticipantDismissal(projectUUID, siteUUID, site.UUID == siteUUID)
}

// ListLocalData returns a list of local data that haven't been associated into the specified project
func (app *ProjectApp) ListLocalData(projectUUID string) ([]ProjectData, error) {
	site, err := app.LoadSite()
	if err != nil {
		return nil, err
	}

	instanceList, err := app.LocalDataRepo.GetAll()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load local data list")
	}
	localDataList := instanceList.([]entity.LocalData)

	dataListInstance, err := app.ProjectDataRepo.GetListByProjectUUID(projectUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query project data")
	}
	associatedDataList := dataListInstance.([]entity.ProjectData)

	associatedDataUUIDMap := map[string]interface{}{}
	for _, associatedData := range associatedDataList {
		if associatedData.Status != entity.ProjectDataStatusAssociated {
			continue
		}
		associatedDataUUIDMap[associatedData.DataUUID] = nil
	}

	availableDataList := make([]ProjectData, 0)
	for _, localData := range localDataList {
		if _, ok := associatedDataUUIDMap[localData.UUID]; !ok && localData.JobStatus == entity.UploadJobStatusSucceeded {
			availableDataList = append(availableDataList, ProjectData{
				Name:                 localData.Name,
				Description:          localData.Description,
				DataID:               localData.UUID,
				CreationTime:         localData.CreatedAt,
				UpdatedTime:          localData.UpdatedAt,
				ProvidingSiteUUID:    site.UUID,
				ProvidingSiteName:    site.Name,
				ProvidingSitePartyID: site.PartyID,
				IsLocal:              true,
			})
		}
	}
	return availableDataList, nil
}

// ListData returns a list of data, local and remote, of the specified project
func (app *ProjectApp) ListData(projectUUID, participantUUID string) ([]ProjectData, error) {
	if err := app.ProjectSyncService.EnsureProjectDataSynced(projectUUID); err != nil {
		log.Err(err).Msg("failed to sync project data")
	}
	site, err := app.LoadSite()
	if err != nil {
		return nil, err
	}

	var dataListInstance interface{}
	if participantUUID == "" {
		dataListInstance, err = app.ProjectDataRepo.GetListByProjectUUID(projectUUID)
	} else {
		dataListInstance, err = app.ProjectDataRepo.GetListByProjectAndSiteUUID(projectUUID, participantUUID)
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to query project data")
	}
	dataList := dataListInstance.([]entity.ProjectData)

	projectDataList := make([]ProjectData, 0)
	for _, data := range dataList {
		if data.Status != entity.ProjectDataStatusAssociated {
			continue
		}
		projectDataList = append(projectDataList, ProjectData{
			Name:                 data.Name,
			Description:          data.Description,
			DataID:               data.DataUUID,
			CreationTime:         data.CreationTime,
			UpdatedTime:          data.UpdatedAt,
			ProvidingSiteUUID:    data.SiteUUID,
			ProvidingSiteName:    data.SiteName,
			ProvidingSitePartyID: data.SitePartyID,
			IsLocal:              data.SiteUUID == site.UUID,
		})
	}

	return projectDataList, nil
}

// CreateDataAssociation associates local data into the specified project
func (app *ProjectApp) CreateDataAssociation(projectUUID string, request *ProjectDataAssociationRequest) error {
	if err := app.EnsureProjectIsOpen(projectUUID); err != nil {
		return err
	}
	dataInstance, err := app.LocalDataRepo.GetByUUID(request.DataUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query local data")
	}
	localData := dataInstance.(*entity.LocalData)

	site, err := app.LoadSite()
	if err != nil {
		return err
	}

	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project")
	}
	project := projectInstance.(*entity.Project)

	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	fmlManagerConnectionInfo := &aggregate.FMLManagerConnectionInfo{
		Connected:  site.FMLManagerConnected,
		Endpoint:   site.FMLManagerEndpoint,
		ServerName: site.FMLManagerServerName,
	}
	return projectAggregate.AssociateLocalData(&aggregate.ProjectLocalDataAssociationContext{
		FMLManagerConnectionInfo: fmlManagerConnectionInfo,
		LocalData: &entity.ProjectData{
			Name:           localData.Name,
			Description:    localData.Description,
			ProjectUUID:    project.UUID,
			DataUUID:       localData.UUID,
			SiteUUID:       site.UUID,
			SiteName:       site.Name,
			SitePartyID:    site.PartyID,
			Type:           entity.ProjectDataTypeLocal,
			Status:         entity.ProjectDataStatusAssociated,
			TableName:      localData.TableName,
			TableNamespace: localData.TableNamespace,
			CreationTime:   localData.CreatedAt,
			UpdateTime:     localData.UpdatedAt,
			Repo:           app.ProjectDataRepo,
		},
	})
}

// RemoveDataAssociation dismisses the local data association
func (app *ProjectApp) RemoveDataAssociation(projectUUID, dataUUID string) error {
	if err := app.EnsureProjectIsOpen(projectUUID); err != nil {
		return err
	}
	site, err := app.LoadSite()
	if err != nil {
		return err
	}

	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project")
	}
	project := projectInstance.(*entity.Project)

	projectDataInstance, err := app.ProjectDataRepo.GetByProjectAndDataUUID(projectUUID, dataUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project data")
	}
	projectData := projectDataInstance.(*entity.ProjectData)

	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     projectData,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	fmlManagerConnectionInfo := &aggregate.FMLManagerConnectionInfo{
		Connected:  site.FMLManagerConnected,
		Endpoint:   site.FMLManagerEndpoint,
		ServerName: site.FMLManagerServerName,
	}
	return projectAggregate.DismissAssociatedLocalData(&aggregate.ProjectLocalDataDismissalContext{
		FMLManagerConnectionInfo: fmlManagerConnectionInfo,
	})
}

// CreateRemoteProjectDataAssociation adds the passed remote data association to the specified project
func (app *ProjectApp) CreateRemoteProjectDataAssociation(projectUUID string, dataList []entity.ProjectData) error {
	site, err := app.LoadSite()
	if err != nil {
		return err
	}
	projectAggregate := aggregate.ProjectAggregate{
		Project: &entity.Project{
			UUID: projectUUID,
			Repo: app.ProjectRepo,
		},
		Participant:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	return projectAggregate.CreateRemoteProjectData(&aggregate.ProjectRemoteDataAssociationContext{
		LocalSiteUUID:  site.UUID,
		RemoteDataList: dataList,
	})
}

// DismissRemoteProjectDataAssociation removes the data from the specified project
func (app *ProjectApp) DismissRemoteProjectDataAssociation(projectUUID string, dataUUIDList []string) error {
	site, err := app.LoadSite()
	if err != nil {
		return err
	}
	projectAggregate := aggregate.ProjectAggregate{
		Project: &entity.Project{
			UUID: projectUUID,
			Repo: app.ProjectRepo,
		},
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	return projectAggregate.DeleteRemoteProjectData(&aggregate.ProjectRemoteDataDismissalContext{
		LocalSiteUUID:      site.UUID,
		RemoteDataUUIDList: dataUUIDList,
	})
}

// ProcessProjectClosing processes project closing event
func (app *ProjectApp) ProcessProjectClosing(projectUUID string) error {
	if err := app.ensureNoRunningJobs(projectUUID, ""); err != nil {
		return err
	}

	domainService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	return domainService.ProcessProjectClosing(projectUUID)
}

// ProcessParticipantUnregistration processes participant unregistration event
// if siteUUID is empty the current site uuid will be used
func (app *ProjectApp) ProcessParticipantUnregistration(siteUUID string) error {
	site, err := app.LoadSite()
	if err != nil {
		return err
	}

	if siteUUID == "" {
		siteUUID = site.UUID
	}

	domainService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	return domainService.ProcessParticipantUnregistration(siteUUID, siteUUID == site.UUID)
}

// SyncProjectParticipant sync participant status of a project from the fml manager
func (app *ProjectApp) SyncProjectParticipant(projectUUID string) error {
	if err := app.EnsureProjectIsOpen(projectUUID); err != nil {
		return err
	}
	site, err := app.LoadSite()
	if err != nil {
		return err
	}

	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project")
	}
	project := projectInstance.(*entity.Project)

	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	return projectAggregate.SyncParticipant(&aggregate.ProjectSyncContext{
		FMLManagerConnectionInfo: &aggregate.FMLManagerConnectionInfo{
			Connected:  site.FMLManagerConnected,
			Endpoint:   site.FMLManagerEndpoint,
			ServerName: site.FMLManagerServerName,
		},
		LocalSiteUUID: site.UUID,
	})
}

// SyncProjectData sync data association status of a project from the fml manager
func (app *ProjectApp) SyncProjectData(projectUUID string) error {
	if err := app.EnsureProjectIsOpen(projectUUID); err != nil {
		return err
	}
	site, err := app.LoadSite()
	if err != nil {
		return err
	}

	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project")
	}
	project := projectInstance.(*entity.Project)

	projectAggregate := aggregate.ProjectAggregate{
		Project:         project,
		Participant:     nil,
		ProjectData:     nil,
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		DataRepo:        app.ProjectDataRepo,
	}
	return projectAggregate.SyncDataAssociation(&aggregate.ProjectSyncContext{
		FMLManagerConnectionInfo: &aggregate.FMLManagerConnectionInfo{
			Connected:  site.FMLManagerConnected,
			Endpoint:   site.FMLManagerEndpoint,
			ServerName: site.FMLManagerServerName,
		},
		LocalSiteUUID: site.UUID,
	})
}

// SyncProject sync remote projects related to current site
func (app *ProjectApp) SyncProject() error {
	site, err := app.LoadSite()
	if err != nil {
		return err
	}
	domainService := service.ProjectService{
		ProjectRepo:     app.ProjectRepo,
		ParticipantRepo: app.ParticipantRepo,
		InvitationRepo:  app.InvitationRepo,
		ProjectDataRepo: app.ProjectDataRepo,
	}
	return domainService.ProcessProjectSyncRequest(&aggregate.ProjectSyncContext{
		FMLManagerConnectionInfo: &aggregate.FMLManagerConnectionInfo{
			Connected:  site.FMLManagerConnected,
			Endpoint:   site.FMLManagerEndpoint,
			ServerName: site.FMLManagerServerName,
		},
		LocalSiteUUID: site.UUID,
	})
}

func (app *ProjectApp) ensureNoRunningJobs(projectUUID, siteUUID string) error {
	// XXX: job checking logic should be placed in the job context?
	jobList, err := app.JobApp.List(projectUUID)
	if err != nil {
		return err
	}
	for _, jobItem := range jobList {
		if jobItem.Status != entity.JobStatusRejected &&
			jobItem.Status != entity.JobStatusFailed &&
			jobItem.Status != entity.JobStatusSucceeded &&
			jobItem.Status != entity.JobStatusDeleted &&
			jobItem.Status != entity.JobStatusUnknown {
			// filter jobs by siteUUID
			if siteUUID != "" {
				_, err := app.JobApp.ParticipantRepo.GetByJobAndSiteUUID(jobItem.UUID, siteUUID)
				if errors.Is(err, repo.ErrJobParticipantNotFound) {
					continue
				}
			}
			return errors.Errorf("at least one job is not in finished status - job: %s(%s)", jobItem.Name, jobItem.StatusStr)
		}
	}
	return nil
}

// EnsureProjectIsOpen returns error if the project is not in a "opened" status.
// This function can be used in other call-sites before start changing project resources.
func (app *ProjectApp) EnsureProjectIsOpen(projectUUID string) error {
	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return err
	}
	project := projectInstance.(*entity.Project)
	if project.Status == entity.ProjectStatusClosed ||
		project.Status == entity.ProjectStatusLeft ||
		project.Status == entity.ProjectStatusDismissed ||
		project.Status == entity.ProjectStatusRejected {
		return errors.Errorf(`project can not be accessed in status: %v`, project.Status)
	}
	return nil
}
