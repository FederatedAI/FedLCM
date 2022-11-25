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
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
)

// LocalDataApp provides local data management services
type LocalDataApp struct {
	LocalDataRepo   repo.LocalDataRepository
	SiteRepo        repo.SiteRepository
	ProjectRepo     repo.ProjectRepository
	ProjectDataRepo repo.ProjectDataRepository
}

// LocalDataUploadRequest contains basic upload request information
type LocalDataUploadRequest struct {
	Name        string `form:"name"`
	Description string `form:"description"`
	FileHeader  *multipart.FileHeader
}

// LocalDataAssociateRequest contains basic local data association request information
type LocalDataAssociateRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	TableNamespace string `json:"table_namespace"`
	TableName      string `json:"table_name"`
}

// LocalDataListItem is an item describing a data record
type LocalDataListItem struct {
	Name            string                 `json:"name"`
	DataID          string                 `json:"data_id"`
	CreationTime    time.Time              `json:"creation_time"`
	SampleSize      uint64                 `json:"sample_size"`
	FeatureSize     int                    `json:"feature_size"`
	UploadJobStatus entity.UploadJobStatus `json:"upload_job_status"`
}

// LocalDataDetail contains local data details
type LocalDataDetail struct {
	LocalDataListItem
	Description string                  `json:"description"`
	TableName   string                  `json:"table_name"`
	Filename    string                  `json:"filename"`
	IDMetaInfo  *valueobject.IDMetaInfo `json:"id_meta_info"`
	Features    []string                `json:"features_array"`
	Preview     string                  `json:"preview_array"`
	NotUploaded bool                    `json:"not_uploaded_locally"`
}

// LocalDataIDMetaInfoUpdateRequest contains basic upload request information
type LocalDataIDMetaInfoUpdateRequest struct {
	*valueobject.IDMetaInfo
}

// Upload loads FATE flow connection info and calls into local data domain object to
// upload the data into the FATE system
func (s *LocalDataApp) Upload(request *LocalDataUploadRequest) (string, error) {
	site := entity.Site{
		Repo: s.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return "", errors.Wrapf(err, "failed to load connection info of FATE flow")
	}
	context := entity.UploadContext{
		FATEFlowHost:    site.FATEFlowHost,
		FATEFlowPort:    site.FATEFlowHTTPPort,
		FATEFlowIsHttps: false,
	}
	data := entity.LocalData{
		Name:          request.Name,
		Description:   request.Description,
		UploadContext: context,
		Repo:          s.LocalDataRepo,
	}
	if err := data.Upload(request.FileHeader); err != nil {
		return "", err
	}
	return data.UUID, nil
}

// AssociateFlowTable creates a local data record associated with existing flow table
func (s *LocalDataApp) AssociateFlowTable(request *LocalDataAssociateRequest) (string, error) {
	site := entity.Site{
		Repo: s.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return "", errors.Wrapf(err, "failed to load connection info of FATE flow")
	}
	context := entity.UploadContext{
		FATEFlowHost:    site.FATEFlowHost,
		FATEFlowPort:    site.FATEFlowHTTPPort,
		FATEFlowIsHttps: false,
	}
	data := entity.LocalData{
		Name:           request.Name,
		Description:    request.Description,
		TableName:      request.TableName,
		TableNamespace: request.TableNamespace,
		UploadContext:  context,
		Repo:           s.LocalDataRepo,
	}
	if err := data.CreateFromExistingTable(); err != nil {
		return "", err
	}
	return data.UUID, nil
}

// List return the list of data uploaded historically
func (s *LocalDataApp) List() ([]LocalDataListItem, error) {
	instanceList, err := s.LocalDataRepo.GetAll()
	if err != nil {
		return nil, err
	}
	dataList := instanceList.([]entity.LocalData)
	publicDataList := make([]LocalDataListItem, len(dataList))
	for index, item := range dataList {
		publicDataList[index] = LocalDataListItem{
			Name:            item.Name,
			DataID:          item.UUID,
			CreationTime:    item.CreatedAt,
			SampleSize:      item.Count,
			FeatureSize:     len(item.Features),
			UploadJobStatus: item.JobStatus,
		}
	}
	return publicDataList, nil
}

// Get returns the detailed information of a data record
func (s *LocalDataApp) Get(uuid string) (*LocalDataDetail, error) {
	instance, err := s.LocalDataRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	localData := instance.(*entity.LocalData)
	localDataDetail := &LocalDataDetail{
		LocalDataListItem: LocalDataListItem{
			Name:            localData.Name,
			DataID:          localData.UUID,
			CreationTime:    localData.CreatedAt,
			SampleSize:      localData.Count,
			FeatureSize:     len(localData.Features),
			UploadJobStatus: localData.JobStatus,
		},
		Description: localData.Description,
		TableName:   fmt.Sprintf("%s#%s", localData.TableNamespace, localData.TableName),
		Filename:    filepath.Base(localData.LocalFilePath),
		IDMetaInfo:  localData.IDMetaInfo,
		Features:    localData.Features,
		Preview:     localData.Preview,
	}
	if localData.LocalFilePath == "" {
		localDataDetail.Filename = localDataDetail.Name
		localDataDetail.NotUploaded = true
	}

	return localDataDetail, nil
}

// GetFilePath returns absolute file path of the stored local data file
func (s *LocalDataApp) GetFilePath(uuid string) (string, error) {
	instance, err := s.LocalDataRepo.GetByUUID(uuid)
	if err != nil {
		return "", err
	}
	localData := instance.(*entity.LocalData)
	return localData.GetAbsFilePath()
}

// GetDataDownloadRequest returns a request object to be used to download the table data
func (s *LocalDataApp) GetDataDownloadRequest(uuid string) (*http.Request, error) {
	site := entity.Site{
		Repo: s.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return nil, errors.Wrapf(err, "failed to load connection info of FATE flow")
	}
	instance, err := s.LocalDataRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	localData := instance.(*entity.LocalData)
	localData.UploadContext = entity.UploadContext{
		FATEFlowHost:    site.FATEFlowHost,
		FATEFlowPort:    site.FATEFlowHTTPPort,
		FATEFlowIsHttps: false,
	}
	return localData.GetFlowDataDownloadRequest()
}

// GetColumns returns a list of headers of the current data
func (s *LocalDataApp) GetColumns(uuid string) ([]string, error) {
	instance, err := s.LocalDataRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	localData := instance.(*entity.LocalData)
	var columns []string
	for _, header := range localData.Column {
		if strings.ToLower(header) != "id" {
			columns = append(columns, header)
		}
	}
	return columns, nil
}

// DeleteData deletes the specified data
func (s *LocalDataApp) DeleteData(uuid string) error {
	instance, err := s.LocalDataRepo.GetByUUID(uuid)
	if err != nil {
		return err
	}
	localData := instance.(*entity.LocalData)
	localData.Repo = s.LocalDataRepo

	site := entity.Site{
		Repo: s.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return errors.Wrapf(err, "failed to load connection info of FATE flow")
	}

	dataListInstance, err := s.ProjectDataRepo.GetListByDataUUID(uuid)
	if err != nil {
		return errors.Wrap(err, "failed to query project data association")
	}
	dataList := dataListInstance.([]entity.ProjectData)

	var projectNameList []string
	for _, projectData := range dataList {
		if projectData.Status == entity.ProjectDataStatusAssociated {
			projectInstance, err := s.ProjectRepo.GetByUUID(projectData.ProjectUUID)
			if err != nil {
				if err == repo.ErrProjectNotFound {
					continue
				}
				return errors.Wrapf(err, "failed to query project %s", projectData.ProjectUUID)
			}
			project := projectInstance.(*entity.Project)
			if project.Status == entity.ProjectStatusManaged || project.Status == entity.ProjectStatusJoined {
				projectNameList = append(projectNameList, project.Name)
			}
		}
	}
	if len(projectNameList) > 0 {
		return errors.Errorf("data is used by project(s): %v; dimiss the association before deleting the data",
			projectNameList)
	}

	localData.UploadContext = entity.UploadContext{
		FATEFlowHost:    site.FATEFlowHost,
		FATEFlowPort:    site.FATEFlowHTTPPort,
		FATEFlowIsHttps: false,
	}
	return localData.Destroy()
}

func (s *LocalDataApp) UpdateIDMetaInfo(uuid string, req *LocalDataIDMetaInfoUpdateRequest) error {
	return s.LocalDataRepo.UpdateIDMetaInfoByUUID(uuid, req.IDMetaInfo)
}
