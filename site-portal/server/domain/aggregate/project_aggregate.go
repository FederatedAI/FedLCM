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
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fmlmanager"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// ProjectAggregate is the aggregation of the concept of a "project"
// We tried to follow the practice suggested by many DDD articles, however we made
// some simplification - we typically only manipulate with only one "project data"
// or one "project participant" for a project. And if there are bulk changes, we
// simply work with the "repo" to persist the change as there is no special business
// logics to construct a complete "project".
type ProjectAggregate struct {
	Project         *entity.Project
	Participant     *entity.ProjectParticipant
	ProjectData     *entity.ProjectData
	ProjectRepo     repo.ProjectRepository
	ParticipantRepo repo.ProjectParticipantRepository
	InvitationRepo  repo.ProjectInvitationRepository
	DataRepo        repo.ProjectDataRepository
}

// ProjectInvitationContext is the context we use to issue project invitation
type ProjectInvitationContext struct {
	FMLManagerConnectionInfo *FMLManagerConnectionInfo
	SiteUUID                 string
	SitePartyID              uint
	SiteName                 string
	SiteDescription          string
}

// ProjectLocalDataAssociationContext is the context we use to create new local data association
type ProjectLocalDataAssociationContext struct {
	FMLManagerConnectionInfo *FMLManagerConnectionInfo
	LocalData                *entity.ProjectData
}

// ProjectLocalDataDismissalContext is the context we use to dismiss data association
type ProjectLocalDataDismissalContext struct {
	FMLManagerConnectionInfo *FMLManagerConnectionInfo
}

// ProjectRemoteDataAssociationContext is the context we use to create remote data association
type ProjectRemoteDataAssociationContext struct {
	LocalSiteUUID  string
	RemoteDataList []entity.ProjectData
}

// ProjectRemoteDataDismissalContext is the context we use to dismiss remote data association
type ProjectRemoteDataDismissalContext struct {
	LocalSiteUUID      string
	RemoteDataUUIDList []string
}

// ProjectSyncContext is the context we use to sync project info with fml-manager
type ProjectSyncContext struct {
	FMLManagerConnectionInfo *FMLManagerConnectionInfo
	LocalSiteUUID            string
}

// CountParticipant returns number of participants in the current project
func (aggregate *ProjectAggregate) CountParticipant() (int64, error) {
	if num, err := aggregate.ParticipantRepo.CountJoinedParticipantByProjectUUID(aggregate.Project.UUID); err != nil {
		return 0, err
	} else {
		// add the manager too
		return num + 1, nil
	}
}

// ListParticipant returns participants of the current project, or all participant in FML manager
func (aggregate *ProjectAggregate) ListParticipant(all bool, fmlManagerConnectionInfo *FMLManagerConnectionInfo) ([]entity.ProjectParticipant, error) {
	instanceList, err := aggregate.ParticipantRepo.GetByProjectUUID(aggregate.Project.UUID)
	if err != nil {
		return nil, err
	}
	allParticipants := instanceList.([]entity.ProjectParticipant)
	var participantList []entity.ProjectParticipant
	for _, participant := range allParticipants {
		if participant.Status == entity.ProjectParticipantStatusOwner ||
			participant.Status == entity.ProjectParticipantStatusJoined ||
			participant.Status == entity.ProjectParticipantStatusPending {
			participantList = append(participantList, participant)
		}
	}
	if !all {
		return participantList, nil
	}
	if !fmlManagerConnectionInfo.Connected || fmlManagerConnectionInfo.Endpoint == "" {
		return nil, errors.Errorf("not connected to FML manager")
	}
	uuidMap := make(map[string]interface{})
	for _, participant := range participantList {
		uuidMap[participant.SiteUUID] = nil
	}
	client := fmlmanager.NewFMLManagerClient(fmlManagerConnectionInfo.Endpoint, fmlManagerConnectionInfo.ServerName)
	siteList, err := client.GetAllSite()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to query site list from FML manager")
	}
	for _, site := range siteList {
		if _, ok := uuidMap[site.UUID]; !ok {
			participantList = append(participantList, entity.ProjectParticipant{
				Model:           gorm.Model{},
				UUID:            "",
				ProjectUUID:     aggregate.Project.UUID,
				SiteUUID:        site.UUID,
				SiteName:        site.Name,
				SitePartyID:     site.PartyID,
				SiteDescription: site.Description,
				Status:          entity.ProjectParticipantStatusUnknown,
			})
		}
	}
	return participantList, nil
}

// InviteParticipant send project invitation to certain site
func (aggregate *ProjectAggregate) InviteParticipant(invitationContext *ProjectInvitationContext) error {
	if !invitationContext.FMLManagerConnectionInfo.Connected {
		return errors.New("FML manager not connected")
	}
	if aggregate.Project.Type == entity.ProjectTypeRemote {
		return errors.New("project not managed by current site")
	}
	if aggregate.Project.ManagingSiteUUID == invitationContext.SiteUUID {
		return errors.Errorf("invalid targeting site uuid %s", invitationContext.SiteUUID)
	}

	var dataList []entity.ProjectData
	if aggregate.Project.Type == entity.ProjectTypeLocal {
		log.Info().Msgf("building local data list for creating the project to FML Manager, project: %s(%s)", aggregate.Project.Name, aggregate.Project.UUID)
		aggregate.Project.Type = entity.ProjectTypeFederatedLocal
		dataListInstance, err := aggregate.DataRepo.GetListByProjectUUID(aggregate.Project.UUID)
		if err != nil {
			return errors.Wrap(err, "failed to query project data")
		}
		dataList = dataListInstance.([]entity.ProjectData)
	}
	// create invitation
	invitation := &entity.ProjectInvitation{
		UUID:        uuid.NewV4().String(),
		ProjectUUID: aggregate.Project.UUID,
		SiteUUID:    invitationContext.SiteUUID,
		Status:      entity.ProjectInvitationStatusCreated,
	}
	if err := aggregate.InvitationRepo.Create(invitation); err != nil {
		return err
	}
	// send invitation to fml manager
	client := fmlmanager.NewFMLManagerClient(invitationContext.FMLManagerConnectionInfo.Endpoint, invitationContext.FMLManagerConnectionInfo.ServerName)
	associatedDataList := make([]fmlmanager.ProjectDataAssociation, len(dataList))
	for index, localData := range dataList {
		associatedDataList[index] = fmlmanager.ProjectDataAssociation{
			ProjectDataAssociationBase: fmlmanager.ProjectDataAssociationBase{
				DataUUID: localData.DataUUID,
			},
			Name:           localData.Name,
			Description:    localData.Description,
			SiteName:       localData.SiteName,
			SiteUUID:       localData.SiteUUID,
			SitePartyID:    localData.SitePartyID,
			TableName:      localData.TableName,
			TableNamespace: localData.TableNamespace,
			CreationTime:   localData.CreationTime,
			UpdateTime:     localData.UpdateTime,
		}
	}
	if err := client.SendInvitation(fmlmanager.ProjectInvitation{
		UUID:                       invitation.UUID,
		SiteUUID:                   invitationContext.SiteUUID,
		SitePartyID:                invitationContext.SitePartyID,
		ProjectUUID:                aggregate.Project.UUID,
		ProjectName:                aggregate.Project.Name,
		ProjectDescription:         aggregate.Project.Description,
		ProjectAutoApprovalEnabled: aggregate.Project.AutoApprovalEnabled,
		ProjectManager:             aggregate.Project.Manager,
		ProjectManagingSiteName:    aggregate.Project.ManagingSiteName,
		ProjectManagingSitePartyID: aggregate.Project.ManagingSitePartyID,
		ProjectManagingSiteUUID:    aggregate.Project.ManagingSiteUUID,
		AssociatedData:             associatedDataList,
	}); err != nil {
		log.Err(err).Msg("failed to send invitation to FML manager")
		return err
	}
	// create/update participant item
	// XXX: site name and party ID info should be queried from FML manager
	participant := &entity.ProjectParticipant{
		UUID:            uuid.NewV4().String(),
		ProjectUUID:     aggregate.Project.UUID,
		SiteUUID:        invitationContext.SiteUUID,
		SiteName:        invitationContext.SiteName,
		SitePartyID:     invitationContext.SitePartyID,
		SiteDescription: invitationContext.SiteDescription,
		Status:          entity.ProjectParticipantStatusPending,
	}
	if err := aggregate.CreateOrUpdateParticipant(participant); err != nil {
		return err
	}
	// update status
	invitation.Status = entity.ProjectInvitationStatusSent
	if err := aggregate.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return err
	}
	// update project type
	if err := aggregate.ProjectRepo.UpdateTypeByUUID(aggregate.Project); err != nil {
		return errors.Wrapf(err, "failed to update project type")
	}
	return nil
}

// ProcessInvitation handle's invitation request from FML manager and saves it into repo
func (aggregate *ProjectAggregate) ProcessInvitation(invitation *entity.ProjectInvitation, siteUUID string) error {
	if invitation.SiteUUID != siteUUID {
		return errors.Errorf("target site uuid: %s is not current site: %s", invitation.UUID, siteUUID)
	}
	aggregate.Project.Status = entity.ProjectStatusPending
	aggregate.Project.Type = entity.ProjectTypeRemote
	if err := aggregate.CreateOrUpdateProject(); err != nil {
		return err
	}
	if err := aggregate.ParticipantRepo.DeleteByProjectUUID(aggregate.Project.UUID); err != nil {
		return errors.Wrapf(err, "failed to clear participants")
	}
	if err := aggregate.DataRepo.DeleteByProjectUUID(aggregate.Project.UUID); err != nil {
		return errors.Wrap(err, "failed to clear project data")
	}
	// create invitation
	if invitation.UUID == "" {
		return errors.New("invalid invitation")
	}
	invitation.Status = entity.ProjectInvitationStatusSent
	if err := aggregate.InvitationRepo.Create(invitation); err != nil {
		return err
	}
	return nil
}

// CreateOrUpdateProject creates the project in the repo or update its status
func (aggregate *ProjectAggregate) CreateOrUpdateProject() error {
	projectInstance, err := aggregate.ProjectRepo.GetByUUID(aggregate.Project.UUID)
	if err != nil {
		if errors.Is(err, repo.ErrProjectNotFound) {
			if err := aggregate.Project.Create(); err != nil {
				return errors.Wrapf(err, "failed to create project")
			}
		} else {
			return err
		}
	} else {
		project := projectInstance.(*entity.Project)
		project.Status = aggregate.Project.Status
		aggregate.Project = project
		if err := aggregate.ProjectRepo.UpdateStatusByUUID(project); err != nil {
			return errors.Wrapf(err, "failed to update project info")
		}
	}
	return nil
}

// CreateOrUpdateParticipant creates a participant record in the repo or update its status
func (aggregate *ProjectAggregate) CreateOrUpdateParticipant(newParticipant *entity.ProjectParticipant) error {
	instance, err := aggregate.ParticipantRepo.GetByProjectAndSiteUUID(aggregate.Project.UUID, newParticipant.SiteUUID)
	if err != nil {
		if errors.Is(err, repo.ErrProjectParticipantNotFound) {
			newParticipant.Model = gorm.Model{}
			if newParticipant.UUID == "" {
				newParticipant.UUID = uuid.NewV4().String()
			}
			if err := aggregate.ParticipantRepo.Create(newParticipant); err != nil {
				return errors.Wrapf(err, "failed to create participant info")
			}
		} else {
			return errors.Wrapf(err, "failed to query participant info")
		}
	} else {
		participant := instance.(*entity.ProjectParticipant)
		participant.Status = newParticipant.Status
		if err := aggregate.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
			return errors.Wrapf(err, "failed to update participant info")
		}
	}
	return nil
}

// JoinProject joins the project by sending invitation response
func (aggregate *ProjectAggregate) JoinProject(fmlManagerConnectionInfo *FMLManagerConnectionInfo) error {
	if !fmlManagerConnectionInfo.Connected || fmlManagerConnectionInfo.Endpoint == "" {
		return errors.Errorf("not connected to FML manager")
	}
	if aggregate.Project.Status != entity.ProjectStatusPending {
		return errors.Errorf("invalide project status: %d", aggregate.Project.Status)
	}
	aggregate.Project.Status = entity.ProjectStatusJoined
	invitationInstance, err := aggregate.InvitationRepo.GetByProjectUUID(aggregate.Project.UUID)
	if err != nil {
		return err
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}
	invitation.Status = entity.ProjectInvitationStatusAccepted
	// send join request to fml manager
	client := fmlmanager.NewFMLManagerClient(fmlManagerConnectionInfo.Endpoint, fmlManagerConnectionInfo.ServerName)
	if err := client.SendInvitationAcceptance(invitation.UUID); err != nil {
		return errors.Wrapf(err, "unable to send invitation response to FML manager")
	}
	// update project and invitation status
	if err := aggregate.ProjectRepo.UpdateStatusByUUID(aggregate.Project); err != nil {
		return err
	}
	if err := aggregate.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return err
	}
	return nil
}

// RejectProject reject to join the project by sending the invitation response
func (aggregate *ProjectAggregate) RejectProject(fmlManagerConnectionInfo *FMLManagerConnectionInfo) error {
	if !fmlManagerConnectionInfo.Connected || fmlManagerConnectionInfo.Endpoint == "" {
		return errors.Errorf("not connected to FML manager")
	}
	if aggregate.Project.Status != entity.ProjectStatusPending {
		return errors.Errorf("invalide project status: %d", aggregate.Project.Status)
	}
	aggregate.Project.Status = entity.ProjectStatusRejected
	invitationInstance, err := aggregate.InvitationRepo.GetByProjectUUID(aggregate.Project.UUID)
	if err != nil {
		return err
	}
	invitation := invitationInstance.(*entity.ProjectInvitation)
	if invitation.Status != entity.ProjectInvitationStatusSent {
		return errors.Errorf("invalide invitation status: %d", invitation.Status)
	}
	invitation.Status = entity.ProjectInvitationStatusRejected
	// send reject request to fml manager
	client := fmlmanager.NewFMLManagerClient(fmlManagerConnectionInfo.Endpoint, fmlManagerConnectionInfo.ServerName)
	if err := client.SendInvitationRejection(invitation.UUID); err != nil {
		return errors.Wrapf(err, "unable to send invitation response to FML manager")
	}
	// update project and invitation status
	if err := aggregate.ProjectRepo.UpdateStatusByUUID(aggregate.Project); err != nil {
		return err
	}
	if err := aggregate.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
		return err
	}
	return nil
}

// LeaveProject leave the current remote project
func (aggregate *ProjectAggregate) LeaveProject(fmlManagerConnectionInfo *FMLManagerConnectionInfo) error {
	// sanity checks
	if aggregate.Project.Type != entity.ProjectTypeRemote {
		return errors.New("project is not managed by other site")
	}
	if aggregate.Project.Status != entity.ProjectStatusJoined || aggregate.Participant.Status != entity.ProjectParticipantStatusJoined {
		return errors.New("current site is not in the project")
	}
	if !fmlManagerConnectionInfo.Connected || fmlManagerConnectionInfo.Endpoint == "" {
		return errors.Errorf("not connected to FML manager")
	}

	// no data association should exist
	dataListInstance, err := aggregate.DataRepo.GetListByProjectAndSiteUUID(aggregate.Project.UUID, aggregate.Participant.SiteUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project data")
	}
	dataList := dataListInstance.([]entity.ProjectData)

	for _, data := range dataList {
		if data.Status == entity.ProjectDataStatusAssociated {
			return errors.Errorf("at least one data association exists, data: %s", data.Name)
		}
	}

	// send leaving event to fml manager
	client := fmlmanager.NewFMLManagerClient(fmlManagerConnectionInfo.Endpoint, fmlManagerConnectionInfo.ServerName)
	if err := client.SendProjectParticipantLeaving(aggregate.Project.UUID, aggregate.Participant.SiteUUID); err != nil {
		return errors.Wrapf(err, "error sending participant leaving request to FML manager")
	}

	// update project and participant status
	aggregate.Project.Status = entity.ProjectStatusLeft
	if err := aggregate.ProjectRepo.UpdateStatusByUUID(aggregate.Project); err != nil {
		return err
	}
	aggregate.Participant.Status = entity.ProjectParticipantStatusLeft
	if err := aggregate.ParticipantRepo.UpdateStatusByUUID(aggregate.Participant); err != nil {
		return err
	}
	return nil
}

// CreateRemoteProjectParticipants adds the passed participants into the repo
func (aggregate *ProjectAggregate) CreateRemoteProjectParticipants(participants []entity.ProjectParticipant) error {
	for _, participant := range participants {
		participant.Model.ID = 0
		if err := aggregate.CreateOrUpdateParticipant(&participant); err != nil {
			return err
		}
	}
	return nil
}

// RemoveParticipant removes a joined participant or revoke an invitation to a pending site
func (aggregate *ProjectAggregate) RemoveParticipant(siteUUID string, fmlManagerConnectionInfo *FMLManagerConnectionInfo) error {
	if !fmlManagerConnectionInfo.Connected {
		return errors.New("FML manager not connected")
	}
	if aggregate.Project.Type == entity.ProjectTypeRemote {
		return errors.New("project not managed by current site")
	}

	// find participant info
	instance, err := aggregate.ParticipantRepo.GetByProjectAndSiteUUID(aggregate.Project.UUID, siteUUID)
	if err != nil {
		return errors.Wrapf(err, "failed to query participant info")
	}
	participant := instance.(*entity.ProjectParticipant)

	if participant.Status == entity.ProjectParticipantStatusPending {
		log.Info().Msgf("revoke invitation for participant: %s", siteUUID)
		participant.Status = entity.ProjectParticipantStatusRevoked
		// find invitation
		invitationInstance, err := aggregate.InvitationRepo.GetByProjectAndSiteUUID(aggregate.Project.UUID, siteUUID)
		if err != nil {
			return errors.Wrapf(err, "failed to get invitation")
		}
		invitation := invitationInstance.(*entity.ProjectInvitation)
		if invitation.Status != entity.ProjectInvitationStatusSent {
			return errors.Errorf("invalide invitation status: %d", invitation.Status)
		}
		invitation.Status = entity.ProjectInvitationStatusRevoked

		// send revocation request to FML manager
		client := fmlmanager.NewFMLManagerClient(fmlManagerConnectionInfo.Endpoint, fmlManagerConnectionInfo.ServerName)
		if err := client.SendInvitationRevocation(invitation.UUID); err != nil {
			return errors.Wrapf(err, "unable to send invitation revocation to FML manager")
		}

		// update status in the repo
		if err := aggregate.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
			return errors.Wrapf(err, "failed to update participant status")
		}
		if err := aggregate.InvitationRepo.UpdateStatusByUUID(invitation); err != nil {
			return errors.Wrapf(err, "failed to update invitation status")
		}
	} else if participant.Status == entity.ProjectParticipantStatusJoined {
		log.Info().Msgf("dismiss participant: %s", siteUUID)
		participant.Status = entity.ProjectParticipantStatusDismissed
		// send dismissal request to FML manager
		client := fmlmanager.NewFMLManagerClient(fmlManagerConnectionInfo.Endpoint, fmlManagerConnectionInfo.ServerName)
		if err := client.SendProjectParticipantDismissal(aggregate.Project.UUID, participant.SiteUUID); err != nil {
			return errors.Wrapf(err, "unable to send participant dismissal to FML manager")
		}

		// dismiss data association
		dataListInstance, err := aggregate.DataRepo.GetListByProjectAndSiteUUID(aggregate.Project.UUID, siteUUID)
		if err != nil {
			return errors.Wrap(err, "failed to query project data")
		}
		dataList := dataListInstance.([]entity.ProjectData)
		for _, data := range dataList {
			if data.Status == entity.ProjectDataStatusAssociated {
				data.Status = entity.ProjectDataStatusDismissed
				if err := aggregate.DataRepo.UpdateStatusByUUID(&data); err != nil {
					return errors.Wrapf(err, "failed to dismiss data %s from site: %s", data.Name, participant.SiteName)
				}
			}
		}

		// update participant status in the repo
		if err := aggregate.ParticipantRepo.UpdateStatusByUUID(participant); err != nil {
			return errors.Wrapf(err, "failed to update participant status")
		}
	} else {
		return errors.Errorf("invalid participant status: %d", participant.Status)
	}

	return nil
}

// AssociateLocalData associates local data into the project
func (aggregate *ProjectAggregate) AssociateLocalData(localDataAssociationCtx *ProjectLocalDataAssociationContext) error {
	if aggregate.Project.Type == entity.ProjectTypeRemote || aggregate.Project.Type == entity.ProjectTypeFederatedLocal {
		if !localDataAssociationCtx.FMLManagerConnectionInfo.Connected {
			return errors.New("project contains other parties but FML manager is not connected")
		}
		client := fmlmanager.NewFMLManagerClient(localDataAssociationCtx.FMLManagerConnectionInfo.Endpoint, localDataAssociationCtx.FMLManagerConnectionInfo.ServerName)
		if err := client.SendProjectDataAssociation(aggregate.Project.UUID, fmlmanager.ProjectDataAssociation{
			ProjectDataAssociationBase: fmlmanager.ProjectDataAssociationBase{
				DataUUID: localDataAssociationCtx.LocalData.DataUUID,
			},
			Name:           localDataAssociationCtx.LocalData.Name,
			Description:    localDataAssociationCtx.LocalData.Description,
			SiteName:       localDataAssociationCtx.LocalData.SiteName,
			SiteUUID:       localDataAssociationCtx.LocalData.SiteUUID,
			SitePartyID:    localDataAssociationCtx.LocalData.SitePartyID,
			TableName:      localDataAssociationCtx.LocalData.TableName,
			TableNamespace: localDataAssociationCtx.LocalData.TableNamespace,
			CreationTime:   localDataAssociationCtx.LocalData.CreationTime,
			UpdateTime:     localDataAssociationCtx.LocalData.UpdateTime,
		}); err != nil {
			return errors.Wrap(err, "failed to send project data association to FML manager")
		}
	}
	localDataAssociationCtx.LocalData.Type = entity.ProjectDataTypeLocal
	localDataAssociationCtx.LocalData.Status = entity.ProjectDataStatusAssociated
	if err := aggregate.CreateOrUpdateData(localDataAssociationCtx.LocalData); err != nil {
		return errors.Wrapf(err, "failed to create project data")
	}
	return nil
}

// CreateOrUpdateData creates the associated data record or updates its status
func (aggregate *ProjectAggregate) CreateOrUpdateData(newData *entity.ProjectData) error {
	instance, err := aggregate.DataRepo.GetByProjectAndDataUUID(aggregate.Project.UUID, newData.DataUUID)
	if err != nil {
		if errors.Is(err, repo.ErrProjectDataNotFound) {
			newData.Model.ID = 0
			if newData.UUID == "" {
				newData.UUID = uuid.NewV4().String()
			}
			if err := aggregate.DataRepo.Create(newData); err != nil {
				return errors.Wrapf(err, "failed to create data association")
			}
		} else {
			return errors.Wrapf(err, "failed to query data association")
		}
	} else {
		data := instance.(*entity.ProjectData)
		data.Status = newData.Status
		if err := aggregate.DataRepo.UpdateStatusByUUID(data); err != nil {
			return errors.Wrapf(err, "failed to update data association")
		}
	}
	return nil
}

// DismissAssociatedLocalData dismisses local data association
func (aggregate *ProjectAggregate) DismissAssociatedLocalData(context *ProjectLocalDataDismissalContext) error {
	if aggregate.Project.Type == entity.ProjectTypeRemote || aggregate.Project.Type == entity.ProjectTypeFederatedLocal {
		if !context.FMLManagerConnectionInfo.Connected {
			return errors.New("project contains other parties but FML manager is not connected")
		}
		client := fmlmanager.NewFMLManagerClient(context.FMLManagerConnectionInfo.Endpoint, context.FMLManagerConnectionInfo.ServerName)
		if err := client.SendProjectDataDismissal(aggregate.Project.UUID, fmlmanager.ProjectDataAssociationBase{
			DataUUID: aggregate.ProjectData.DataUUID,
		}); err != nil {
			return errors.Wrap(err, "failed to send project data dismissal to FML manager")
		}
	}
	if aggregate.ProjectData.Type != entity.ProjectDataTypeLocal {
		return errors.New("cannot dismiss data from other sites")
	}
	aggregate.ProjectData.Status = entity.ProjectDataStatusDismissed
	if err := aggregate.DataRepo.UpdateStatusByUUID(aggregate.ProjectData); err != nil {
		return errors.Wrapf(err, "failed to dismiss project data")
	}
	return nil
}

// CreateRemoteProjectData creates remote data association
func (aggregate *ProjectAggregate) CreateRemoteProjectData(context *ProjectRemoteDataAssociationContext) error {
	for _, data := range context.RemoteDataList {
		if data.SiteUUID == context.LocalSiteUUID {
			log.Warn().Str("data uuid", data.UUID).Msgf("data from this local site")
			continue
		}
		data.Type = entity.ProjectDataTypeRemote
		data.Status = entity.ProjectDataStatusAssociated
		err := aggregate.CreateOrUpdateData(&data)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteRemoteProjectData deletes remote data association
func (aggregate *ProjectAggregate) DeleteRemoteProjectData(context *ProjectRemoteDataDismissalContext) error {
	for _, dataUUID := range context.RemoteDataUUIDList {
		if dataUUID == context.LocalSiteUUID {
			log.Warn().Str("data uuid", dataUUID).Msgf("data from this local site")
			continue
		}
		instance, err := aggregate.DataRepo.GetByProjectAndDataUUID(aggregate.Project.UUID, dataUUID)
		if err != nil {
			if errors.Is(err, repo.ErrProjectDataNotFound) {
				log.Warn().Str("data uuid", dataUUID).Str("project uuid", aggregate.Project.UUID).Msg("data not associated in this project")
				continue
			} else {
				return errors.Wrapf(err, "failed to query data association")
			}
		} else {
			data := instance.(*entity.ProjectData)
			data.Status = entity.ProjectDataStatusDismissed
			if err := aggregate.DataRepo.UpdateStatusByUUID(data); err != nil {
				return errors.Wrapf(err, "failed to update data association")
			}
		}
	}
	return nil
}

// SyncDataAssociation sync the data association info with the fml manager
func (aggregate *ProjectAggregate) SyncDataAssociation(context *ProjectSyncContext) error {
	if aggregate.Project.Type == entity.ProjectTypeLocal {
		return nil
	}
	if !context.FMLManagerConnectionInfo.Connected {
		log.Warn().Msg("SyncDataAssociation: FML manager is not connected")
		return nil
	}
	log.Info().Msgf("start syncing project data, project: %s(%s)", aggregate.Project.Name, aggregate.Project.UUID)
	client := fmlmanager.NewFMLManagerClient(context.FMLManagerConnectionInfo.Endpoint, context.FMLManagerConnectionInfo.ServerName)
	associatedDataMap, err := client.GetProjectDataAssociation(aggregate.Project.UUID)
	if err != nil {
		return err
	}

	dataListInstance, err := aggregate.DataRepo.GetListByProjectUUID(aggregate.Project.UUID)
	if err != nil {
		return errors.Wrap(err, "failed to query project data")
	}
	dataList := dataListInstance.([]entity.ProjectData)
	for _, data := range dataList {
		oldStatus := data.Status
		if _, ok := associatedDataMap[data.DataUUID]; ok {
			data.Status = entity.ProjectDataStatusAssociated
			delete(associatedDataMap, data.DataUUID)
		} else {
			data.Status = entity.ProjectDataStatusDismissed
		}
		if oldStatus != data.Status {
			log.Warn().Msgf("changing stale data association status, data: %v", data)
			if err := aggregate.DataRepo.UpdateStatusByUUID(&data); err != nil {
				return errors.Wrapf(err, "failed to update data association")
			}
		}
	}
	for _, associatedData := range associatedDataMap {
		data := &entity.ProjectData{
			Name:           associatedData.Name,
			Description:    associatedData.Description,
			ProjectUUID:    aggregate.Project.UUID,
			DataUUID:       associatedData.DataUUID,
			SiteUUID:       associatedData.SiteUUID,
			SiteName:       associatedData.SiteName,
			SitePartyID:    associatedData.SitePartyID,
			Type:           entity.ProjectDataTypeRemote,
			Status:         entity.ProjectDataStatusAssociated,
			TableName:      associatedData.TableName,
			TableNamespace: associatedData.TableNamespace,
			CreationTime:   associatedData.CreationTime,
			UpdateTime:     associatedData.UpdateTime,
			Repo:           aggregate.DataRepo,
		}
		if associatedData.SiteUUID == context.LocalSiteUUID {
			data.Type = entity.ProjectDataTypeLocal
		}
		log.Warn().Msgf("adding or updating missing data association: %v", data)
		if err := aggregate.CreateOrUpdateData(data); err != nil {
			return err
		}
	}
	return nil
}

// SyncParticipant sync the participant status from fml manager
func (aggregate *ProjectAggregate) SyncParticipant(context *ProjectSyncContext) error {
	if aggregate.Project.Type == entity.ProjectTypeLocal {
		return nil
	}
	if !context.FMLManagerConnectionInfo.Connected {
		log.Warn().Msg("SyncParticipant: FML manager is not connected")
		return nil
	}
	log.Info().Msgf("start syncing project participants, project: %s(%s)", aggregate.Project.Name, aggregate.Project.UUID)
	client := fmlmanager.NewFMLManagerClient(context.FMLManagerConnectionInfo.Endpoint, context.FMLManagerConnectionInfo.ServerName)
	participantMap, err := client.GetProjectParticipant(aggregate.Project.UUID)
	if err != nil {
		return err
	}

	participantListInstance, err := aggregate.ParticipantRepo.GetByProjectUUID(aggregate.Project.UUID)
	if err != nil {
		return err
	}
	participantList := participantListInstance.([]entity.ProjectParticipant)

	for _, participant := range participantList {
		oldStatus := participant.Status
		if _, ok := participantMap[participant.SiteUUID]; ok {
			participant.Status = entity.ProjectParticipantStatus(participantMap[participant.SiteUUID].Status)
			// change pending participant status to dismissed for non-managing site
			if aggregate.Project.ManagingSiteUUID != context.LocalSiteUUID && participant.Status == entity.ProjectParticipantStatusPending {
				participant.Status = entity.ProjectParticipantStatusDismissed
			}
			delete(participantMap, participant.SiteUUID)
		} else {
			participant.Status = entity.ProjectParticipantStatusDismissed
		}
		if participant.Status != oldStatus {
			log.Warn().Msgf("changing stale participant status, participant: %v", participant)
			if err := aggregate.ParticipantRepo.UpdateStatusByUUID(&participant); err != nil {
				return errors.Wrapf(err, "failed to update participant info")
			}
		}
	}

	for _, participant := range participantMap {
		newParticipant := &entity.ProjectParticipant{
			ProjectUUID:     participant.ProjectUUID,
			SiteUUID:        participant.SiteUUID,
			SiteName:        participant.SiteName,
			SitePartyID:     participant.SitePartyID,
			SiteDescription: participant.SiteDescription,
			Status:          entity.ProjectParticipantStatus(participant.Status),
		}
		// change pending participant status to dismissed for non-managing site
		if aggregate.Project.ManagingSiteUUID != context.LocalSiteUUID && newParticipant.Status == entity.ProjectParticipantStatusPending {
			newParticipant.Status = entity.ProjectParticipantStatusDismissed
		}
		log.Warn().Msgf("adding or updating missing participant: %v", participant)
		if err := aggregate.CreateOrUpdateParticipant(newParticipant); err != nil {
			return err
		}
	}
	return nil
}

// CloseProject closes the current project managed by current site
func (aggregate *ProjectAggregate) CloseProject(fmlManagerConnectionInfo *FMLManagerConnectionInfo) error {
	if aggregate.Project.Type == entity.ProjectTypeRemote {
		return errors.New("project is managed by other site")
	}

	// work with fml manager to close the project
	if aggregate.Project.Type == entity.ProjectTypeFederatedLocal {
		if !fmlManagerConnectionInfo.Connected || fmlManagerConnectionInfo.Endpoint == "" {
			return errors.Errorf("not connected to FML manager")
		}
		// send closing event to fml manager
		client := fmlmanager.NewFMLManagerClient(fmlManagerConnectionInfo.Endpoint, fmlManagerConnectionInfo.ServerName)
		if err := client.SendProjectClosing(aggregate.Project.UUID); err != nil {
			return errors.Wrapf(err, "error sending project closing request to FML manager")
		}
	}
	aggregate.Project.Status = entity.ProjectStatusClosed
	if err := aggregate.ProjectRepo.UpdateStatusByUUID(aggregate.Project); err != nil {
		return errors.Wrapf(err, "failed to update project status")
	}
	return nil
}
