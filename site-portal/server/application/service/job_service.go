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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/constants"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/aggregate"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/entity"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// JobApp provides interface for job related API handling routines
type JobApp struct {
	SiteRepo        repo.SiteRepository
	JobRepo         repo.JobRepository
	ParticipantRepo repo.JobParticipantRepository
	ProjectRepo     repo.ProjectRepository
	ProjectDataRepo repo.ProjectDataRepository
	ModelRepo       repo.ModelRepository
}

// JobInfoBase contains the basic info of a job
type JobInfoBase struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        entity.JobType `json:"type"`
	ProjectUUID string         `json:"project_uuid"`
}

// JobListItemBase contains info of a job for displaying in a list view
type JobListItemBase struct {
	JobInfoBase
	UUID                  string           `json:"uuid"`
	Status                entity.JobStatus `json:"status"`
	StatusStr             string           `json:"status_str"`
	CreationTime          time.Time        `json:"creation_time"`
	FinishTime            time.Time        `json:"finish_time"`
	InitiatingSiteUUID    string           `json:"initiating_site_uuid"`
	InitiatingSiteName    string           `json:"initiating_site_name"`
	InitiatingSitePartyID uint             `json:"initiating_site_party_id"`
	PendingOnThisSite     bool             `json:"pending_on_this_site"`
	FATEJobID             string           `json:"fate_job_id"`
	FATEJobStatus         string           `json:"fate_job_status"`
	FATEModelName         string           `json:"fate_model_name"`
	IsInitiator           bool             `json:"is_initiator"`
	Username              string           `json:"username"`
}

// JobConf is the configuration content for a FATE job
type JobConf struct {
	ConfJson string `json:"conf_json"`
	DSLJson  string `json:"dsl_json"`
}

// JobDataBase is the basic info of a job data
type JobDataBase struct {
	DataUUID  string `json:"data_uuid"`
	LabelName string `json:"label_name"`
}

// JobData contains detailed info of a data used in a job
type JobData struct {
	JobDataBase
	Name                 string                      `json:"name"`
	Description          string                      `json:"description"`
	ProvidingSiteUUID    string                      `json:"providing_site_uuid"`
	ProvidingSiteName    string                      `json:"providing_site_name"`
	ProvidingSitePartyID uint                        `json:"providing_site_party_id"`
	IsLocal              bool                        `json:"is_local"`
	Status               entity.JobParticipantStatus `json:"site_status"`
	StatusStr            string                      `json:"site_status_str"`
}

// JobParticipantInfoBase contains basic information of a job participant
type JobParticipantInfoBase struct {
	SiteUUID    string `json:"site_uuid"`
	SiteName    string `json:"site_name"`
	SitePartyID uint   `json:"site_party_id"`
}

// TrainingJobResultInfo is a key-value map containing the trained model evaluation info
type TrainingJobResultInfo struct {
	EvaluationInfo map[string]string `json:"evaluation_info"`
}

// PredictingJobResultInfo contains predicting job result data
type PredictingJobResultInfo struct {
	Header []string        `json:"header"`
	Data   [][]interface{} `json:"data"`
	Count  int             `json:"count"`
}

// IntersectionJobResultInfo contains PSI job result data
type IntersectionJobResultInfo struct {
	Header          []string        `json:"header"`
	Data            [][]interface{} `json:"data"`
	IntersectNumber int             `json:"intersect_number"`
	IntersectRate   float64         `json:"intersect_rate"`
}

// JobAlgorithmConf is the algorithm configuration for a job
type JobAlgorithmConf struct {
	ValidationEnabled      bool                    `json:"training_validation_enabled"`
	ValidationSizePercent  uint                    `json:"training_validation_percent"`
	ModelName              string                  `json:"training_model_name"`
	AlgorithmType          entity.JobAlgorithmType `json:"training_algorithm_type"`
	AlgorithmComponentName string                  `json:"algorithm_component_name"`
	ComponentsToDeploy     []string                `json:"training_component_list_to_deploy"`
	ModelUUID              string                  `json:"predicting_model_uuid"`
	EvaluateComponentName  string                  `json:"evaluate_component_name"`
}

// JobSubmissionRequest is the request for creating a job
type JobSubmissionRequest struct {
	JobConf
	JobInfoBase
	InitiatorData JobDataBase   `json:"initiator_data"`
	OtherData     []JobDataBase `json:"other_site_data"`
	JobAlgorithmConf
}

// RemoteJobCreationRequest is a request for creating a record of a job that is initiated by other sites
type RemoteJobCreationRequest struct {
	JobSubmissionRequest
	Username string `json:"username"`
	UUID     string `json:"uuid"`
}

// JobDetail contains detailed info of a job, including the result and status message
type JobDetail struct {
	JobListItemBase
	InitiatorData JobData       `json:"initiator_data"`
	OtherData     []JobData     `json:"other_site_data"`
	StatusMsg     string        `json:"status_message"`
	ResultInfo    JobResultInfo `json:"result_info"`
	JobConf
	JobAlgorithmConf
}

// JobApprovalContext is the context used for a job approval response
type JobApprovalContext struct {
	SiteUUID string `json:"site_uuid"`
	Approved bool   `json:"approved"`
}

// JobStatusUpdateContext is the context used for updating a job status
type JobStatusUpdateContext struct {
	Status               entity.JobStatus                       `json:"status"`
	StatusMessage        string                                 `json:"status_message"`
	FATEJobID            string                                 `json:"fate_job_id"`
	FATEJobStatus        string                                 `json:"fate_job_status"`
	FATEModelID          string                                 `json:"fate_model_id"`
	FATEModelVersion     string                                 `json:"fate_model_version"`
	ParticipantStatusMap map[string]entity.JobParticipantStatus `json:"participant_status_map"`
}

// JobResultInfo contains result information for all types of jobs
type JobResultInfo struct {
	IntersectionResult IntersectionJobResultInfo `json:"intersection_result"`
	TrainingResult     map[string]string         `json:"training_result"`
	PredictingResult   PredictingJobResultInfo   `json:"predicting_result"`
}

// JobRawDagJson describes the DAG the user draw, also contains the configuration of each job component
type JobRawDagJson struct {
	RawJson string `json:"raw_json"`
}

// GenerateJobConfRequest contains 2 parts of the information: 1. the info of the job, like the participants. 2. the
// info of each job component's configuration, such as what penalty function an algorithm use.
type GenerateJobConfRequest struct {
	JobConf JobSubmissionRequest `json:"job_conf"`
	DagJson JobRawDagJson        `json:"dag_json"`
}

// List returns a list of jobs in the specified project
func (app *JobApp) List(projectUUID string) ([]JobListItemBase, error) {
	site, err := app.loadSite()
	if err != nil {
		return nil, err
	}

	jobListInstance, err := app.JobRepo.GetListByProjectUUID(projectUUID)
	if err != nil {
		return nil, err
	}
	domainJobList := jobListInstance.([]entity.Job)

	jobList := make([]JobListItemBase, 0)
	for _, job := range domainJobList {
		if job.Status != entity.JobStatusUnknown && job.Status != entity.JobStatusDeleted {
			jobParticipantInstance, err := app.ParticipantRepo.GetByJobAndSiteUUID(job.UUID, site.UUID)
			if err != nil {
				if err == repo.ErrJobParticipantNotFound {
					continue
				}
				return nil, errors.Wrap(err, "failed to get participant info")
			}
			jobParticipant := jobParticipantInstance.(*entity.JobParticipant)

			jobList = append(jobList, JobListItemBase{
				JobInfoBase: JobInfoBase{
					Name:        job.Name,
					Description: job.Description,
					Type:        job.Type,
					ProjectUUID: job.ProjectUUID,
				},
				UUID:                  job.UUID,
				Status:                job.Status,
				StatusStr:             job.Status.String(),
				CreationTime:          job.CreatedAt,
				FinishTime:            job.FinishedAt,
				InitiatingSiteUUID:    job.InitiatingSiteUUID,
				InitiatingSiteName:    job.InitiatingSiteName,
				InitiatingSitePartyID: job.InitiatingSitePartyID,
				PendingOnThisSite:     jobParticipant.Status == entity.JobParticipantStatusPending,
				FATEJobID:             job.FATEJobID,
				FATEJobStatus:         job.FATEJobStatus,
				FATEModelName:         job.FATEModelID + "#" + job.FATEModelVersion,
				IsInitiator:           job.InitiatingSiteUUID == site.UUID,
				Username:              job.InitiatingUser,
			})
		}
	}
	return jobList, nil
}

// GetJobDetail returns the detail info of a job
func (app *JobApp) GetJobDetail(uuid string) (*JobDetail, error) {
	site, err := app.loadSite()
	if err != nil {
		return nil, err
	}

	jobAggregate, err := app.loadJobAggregate(uuid)
	if err != nil {
		return nil, err
	}

	jobRequest := &JobSubmissionRequest{}
	if err := json.Unmarshal([]byte(jobAggregate.Job.RequestJson), jobRequest); err != nil {
		return nil, errors.Wrap(err, "failed to recover job submission request")
	}

	jobDetail := &JobDetail{
		JobListItemBase: JobListItemBase{
			JobInfoBase: JobInfoBase{
				Name:        jobAggregate.Job.Name,
				Description: jobAggregate.Job.Description,
				Type:        jobAggregate.Job.Type,
				ProjectUUID: jobAggregate.Job.ProjectUUID,
			},
			UUID:                  jobAggregate.Job.UUID,
			Status:                jobAggregate.Job.Status,
			StatusStr:             jobAggregate.Job.Status.String(),
			CreationTime:          jobAggregate.Job.CreatedAt,
			FinishTime:            jobAggregate.Job.UpdatedAt,
			InitiatingSiteUUID:    jobAggregate.Initiator.SiteUUID,
			InitiatingSiteName:    jobAggregate.Initiator.SiteName,
			InitiatingSitePartyID: jobAggregate.Initiator.SitePartyID,
			PendingOnThisSite:     false,
			FATEJobID:             jobAggregate.Job.FATEJobID,
			FATEJobStatus:         jobAggregate.Job.FATEJobStatus,
			IsInitiator:           jobAggregate.Initiator.SiteUUID == site.UUID,
			Username:              jobAggregate.Job.InitiatingUser,
			FATEModelName:         jobAggregate.Job.FATEModelID + "#" + jobAggregate.Job.FATEModelVersion,
		},
		InitiatorData: JobData{
			JobDataBase: JobDataBase{
				DataUUID:  jobAggregate.Initiator.DataUUID,
				LabelName: jobAggregate.Initiator.DataLabelName,
			},
			Name:                 jobAggregate.Initiator.DataName,
			Description:          jobAggregate.Initiator.DataDescription,
			ProvidingSiteUUID:    jobAggregate.Initiator.SiteUUID,
			ProvidingSiteName:    jobAggregate.Initiator.SiteName,
			ProvidingSitePartyID: jobAggregate.Initiator.SitePartyID,
			IsLocal:              jobAggregate.Initiator.SiteUUID == site.UUID,
			Status:               jobAggregate.Initiator.Status,
			StatusStr:            jobAggregate.Initiator.Status.String(),
		},
		OtherData: nil,
		StatusMsg: app.buildJobStatusMessage(jobAggregate),
		JobConf: JobConf{
			ConfJson: jobAggregate.Job.Conf,
			DSLJson:  jobAggregate.Job.DSL,
		},
		JobAlgorithmConf: JobAlgorithmConf{
			ValidationEnabled:      jobRequest.ValidationEnabled,
			ValidationSizePercent:  jobRequest.ValidationSizePercent,
			ModelName:              jobAggregate.Job.ModelName,
			AlgorithmType:          jobRequest.AlgorithmType,
			AlgorithmComponentName: jobRequest.AlgorithmComponentName,
			ComponentsToDeploy:     jobRequest.ComponentsToDeploy,
			ModelUUID:              jobRequest.ModelUUID,
		},
	}
	switch jobAggregate.Job.Type {
	case entity.JobTypeTraining:
		jobDetail.ResultInfo.TrainingResult = jobAggregate.Job.GetTrainingResultSummary()
	case entity.JobTypePredict:
		header, data, count := jobAggregate.Job.GetPredictingResultPreview()
		jobDetail.ResultInfo.PredictingResult = PredictingJobResultInfo{
			Header: header,
			Data:   data,
			Count:  count,
		}
	case entity.JobTypePSI:
		header, data, intersectNumber, intersectRate := jobAggregate.Job.GetIntersectionResult()
		jobDetail.ResultInfo.IntersectionResult = IntersectionJobResultInfo{
			Header:          header,
			Data:            data,
			IntersectNumber: intersectNumber,
			IntersectRate:   intersectRate,
		}
	}

	for _, participant := range jobAggregate.Participants {
		jobDetail.OtherData = append(jobDetail.OtherData, JobData{
			JobDataBase: JobDataBase{
				DataUUID:  participant.DataUUID,
				LabelName: "ignored",
			},
			Name:                 participant.DataName,
			Description:          participant.DataDescription,
			ProvidingSiteUUID:    participant.SiteUUID,
			ProvidingSiteName:    participant.SiteName,
			ProvidingSitePartyID: participant.SitePartyID,
			IsLocal:              participant.SiteUUID == site.UUID,
			Status:               participant.Status,
			StatusStr:            participant.Status.String(),
		})
		if participant.SiteUUID == site.UUID && participant.Status == entity.JobParticipantStatusPending {
			jobDetail.PendingOnThisSite = true
		}
	}

	return jobDetail, nil
}

// GetDataResultDownloadRequest returns a request object to be used to download the result data
func (app *JobApp) GetDataResultDownloadRequest(uuid string) (*http.Request, error) {
	jobAggregate, err := app.loadJobAggregate(uuid)
	if err != nil {
		return nil, err
	}
	return jobAggregate.GetDataResultDownloadRequest()
}

// Approve approves the job
func (app *JobApp) Approve(uuid string) error {
	jobAggregate, err := app.loadJobAggregate(uuid)
	if err != nil {
		return err
	}
	return jobAggregate.ApproveJob()
}

// Reject rejects the job
func (app *JobApp) Reject(uuid string) error {
	jobAggregate, err := app.loadJobAggregate(uuid)
	if err != nil {
		return err
	}
	return jobAggregate.RejectJob()
}

// Refresh checks the latest job status
func (app *JobApp) Refresh(uuid string) error {
	jobAggregate, err := app.loadJobAggregate(uuid)
	if err != nil {
		return err
	}
	return jobAggregate.RefreshJob()
}

// GenerateConfig returns the job configuration content based on the job info
func (app *JobApp) GenerateConfig(username string, request *JobSubmissionRequest) (*JobConf, error) {
	jobAggregate, err := app.buildJobAggregate(username, request)
	if err != nil {
		return nil, err
	}

	config, dsl, err := jobAggregate.GenerateConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get job config")
	}
	return &JobConf{
		ConfJson: config,
		DSLJson:  dsl,
	}, nil
}

// LoadJobComponents returns s json string, UI can use this json to populate the FML components, so that a user
// can drag the components to define a job. The json contains the default configs for each component.
func (app *JobApp) LoadJobComponents() string {
	return constants.JobComponents
}

// GenerateDslFromDag returns a json string, which is the DSL configuration which can be consumed by fateflow. The
// input is the raw json string which represent the DAG the user has draw on UI.
func (app *JobApp) GenerateDslFromDag(rawJson string) (string, error) {
	var dagStruct map[string]interface{}
	res := make(map[string]interface{})
	res["components"] = make(map[string]interface{})
	dslComponents := res["components"].(map[string]interface{})

	err := json.Unmarshal([]byte(rawJson), &dagStruct)
	if err != nil {
		return "", err
	}

	for key, value := range dagStruct {
		/*
			an example of "value":
			{
				"attributeType": "common",
				"commonAttributes": {},
				"diffAttributes": {},
				"conditions": {
					"output": {
						"data": ["data"]
					}
				},
				"module": "Reader"
			}
		*/
		dslComponents[key] = make(map[string]interface{})
		dslComponentContent := dslComponents[key].(map[string]interface{})
		dagValue := value.(map[string]interface{})
		dslComponentContent["module"] = dagValue["module"].(string)
		input := dagValue["conditions"].(map[string]interface{})["input"]
		if input != nil {
			dslComponentContent["input"] = input
		}
		output := dagValue["conditions"].(map[string]interface{})["output"]
		if output != nil {
			dslComponentContent["output"] = output
		}
		/*
			an example of "dslComponentContent":
			{
				"module": "Reader",
				"output": {
				"data": [
				  "data"
				]
			}
		*/
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	resStr, err := app.generateIndentedJsonStr(string(resJson))
	return resStr, nil
}

// GenerateConfFromDag returns a map, which represents the configuration of each fate job component.
func (app *JobApp) GenerateConfFromDag(
	username string, generateJobConfRequest *GenerateJobConfRequest) (string, error) {
	var res map[string]interface{}
	jobAggregate, err := app.buildJobAggregate(username, &generateJobConfRequest.JobConf)
	numberOfHosts := len(jobAggregate.Participants)
	componentParameters, err := app.buildComponentParameters(
		generateJobConfRequest.DagJson.RawJson, numberOfHosts)
	if err != nil {
		return "", errors.Wrap(err, "failed to build the component parameters")
	}

	hostUuidList := make([]string, 0)
	// The "OtherData" below is a list, the sequence matches the host configurations in above "componentParameters",
	// which is generated from the DAG user drawn. For example, if there are 3 hosts in the "otherData",
	// then the 1st host's config must match the "0" config in "componentParameters". The 2nd host's config must match
	// the "1" config in "componentParameters", etc. So now the problem is, how to make the later-added reader info keep
	// the same sequence? This cannot be done by "jobAggregate" because its "Participants" attr is a map. So here we
	// need the sequence information from "OtherData". "OtherData" contains the "dataUUID" which can help us find the
	// "siteUUID" of each party, in the same sequence with the DAG configurations. Then we can use the siteUUID as the
	// key to fetch the table name and namespace from the "Participants" map of "jobAggregate".
	for _, party := range generateJobConfRequest.JobConf.OtherData {
		projectDataInstance, err := app.ProjectDataRepo.GetByDataUUID(party.DataUUID)
		if err != nil {
			return "", errors.Wrap(err, "failed to query project data")
		}
		projectData := projectDataInstance.(*entity.ProjectData)
		hostUuidList = append(hostUuidList, projectData.SiteUUID)
	}

	generalJobConf, err := jobAggregate.GenerateGeneralTrainingConf(hostUuidList)
	if err != nil {
		errStr := "failed to generate the general training configurations"
		return errStr, errors.Wrap(err, errStr)
	}
	// Load the general configurations into the res
	err = json.Unmarshal([]byte(generalJobConf), &res)
	if err != nil {
		errStr := "failed to unmarshal the general job conf"
		return errStr, errors.Wrap(err, errStr)
	}

	hostReaderConfigMap, guestReaderConfigMap := jobAggregate.GenerateReaderConfigMaps(hostUuidList)
	// Replace the readers' config in "componentParameters" with the ones in defaultHostConfMap and defaultGuestConfMap
	hostConfigs := componentParameters["component_parameters"].(map[string]interface{})["role"].(map[string]interface{})["host"].(map[string]interface{})
	guestConfigs := componentParameters["component_parameters"].(map[string]interface{})["role"].(map[string]interface{})["guest"].(map[string]interface{})
	// for above 2 maps, the key should be an index like "0", "1", etc.
	for index, hostConfig := range hostConfigs {
		hostConfig.(map[string]interface{})["reader_0"] = hostReaderConfigMap[index].(map[string]interface{})["reader_0"]
	}
	for index, guestConfig := range guestConfigs {
		guestConfig.(map[string]interface{})["reader_0"] = guestReaderConfigMap[index].(map[string]interface{})["reader_0"]
	}
	for key, value := range componentParameters {
		res[key] = value
	}
	resJson, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	resStr, err := app.generateIndentedJsonStr(string(resJson))
	return resStr, nil
}

// buildComponentParameters is a helper function that helps to handle the common parameters and diff parameters the
// use has chosen on the DAG, and generate the format Fateflow can understand.
func (app *JobApp) buildComponentParameters(dagJsonStr string, numberOfHosts int) (map[string]interface{}, error) {
	res := make(map[string]interface{})
	res["component_parameters"] = make(map[string]interface{})
	componentParameters := res["component_parameters"].(map[string]interface{})
	componentParameters["common"] = make(map[string]interface{})
	componentParameters["role"] = make(map[string]interface{})
	common := componentParameters["common"].(map[string]interface{})
	role := componentParameters["role"].(map[string]interface{})
	role["host"] = make(map[string]interface{})
	role["guest"] = make(map[string]interface{})
	host := role["host"].(map[string]interface{})
	guest := role["guest"].(map[string]interface{})
	// There must be only one guest
	guest["0"] = make(map[string]interface{})
	for i := 0; i < numberOfHosts; i++ {
		// There can be more than one host
		host[strconv.Itoa(i)] = make(map[string]interface{})
	}
	var dagMap map[string]interface{}
	err := json.Unmarshal([]byte(dagJsonStr), &dagMap)
	if err != nil {
		return nil, err
	}
	for dagKey, dagValue := range dagMap {
		// dagKey is something like: "HomoLR_0"
		// dagValue is a map of 5 fixed keys: attributeType, commonAttributes, diffAttributes, conditions, module
		dagConfs := dagValue.(map[string]interface{})
		if dagConfs["attributeType"] == "common" {
			common[dagKey] = make(map[string]interface{})
			dagCommonAttrs := dagConfs["commonAttributes"].(map[string]interface{})
			commonAttrMap := app.filterEmptyAttributes(dagCommonAttrs)
			common[dagKey] = commonAttrMap
		} else if dagConfs["attributeType"] == "diff" {
			dagDiffAttrs := dagConfs["diffAttributes"].(map[string]interface{})
			// for dagDiffAttrs, the keys will be like: guest, host_0, host_1, etc.
			if dagConfs["module"].(string) == "Reader" {
				// The reason for this special case is: The reader's configuration will be copied from the default
				// configuration later. UI will not pass the table name/namespace here, so no need to handle here.
				// Also, for reader there is no attributes to choose from UI, so we can skip below code.
				continue
			}
			// guest is the simple one, handle it first
			dagGuestDiffAttrs := dagDiffAttrs["guest"].(map[string]interface{})
			guestConfMap := role["guest"].(map[string]interface{})["0"].(map[string]interface{})
			guestConfMap[dagKey] = app.filterEmptyAttributes(dagGuestDiffAttrs)
			// remove the guest item from the dagDiffAttrs, make it easier to handle the hosts
			delete(dagDiffAttrs, "guest")
			for dagHostName, dagHostDiffAttrMap := range dagDiffAttrs {
				// dagHostName must be like host_0, host_1, host_2, etc.
				hostIndexStr := strings.Split(dagHostName, "_")[1]
				roleHostMap := role["host"].(map[string]interface{})
				// roleHostMap's key is "0", "1", "2", etc.
				hostConfMap := roleHostMap[hostIndexStr].(map[string]interface{})
				hostConfMap[dagKey] = app.filterEmptyAttributes(dagHostDiffAttrMap.(map[string]interface{}))
			}
		} else {
			return nil, errors.New("Undefined attribute type in the dag")
		}
	}
	return res, nil
}

// filterEmptyAttributes is a helper function that helps to filter out the empty attributes of a Fateflow component
func (app *JobApp) filterEmptyAttributes(source map[string]interface{}) map[string]interface{} {
	target := make(map[string]interface{})
	for k, v := range source {
		if !app.isEmpty(v) {
			target[k] = v
		}
	}
	return target
}

// isEmpty is a helper function to help check if an attribute of a Fateflow component is {}, [] or ""
func (app *JobApp) isEmpty(object interface{}) bool {
	if object == nil {
		return true
	}
	objectType := reflect.TypeOf(object)
	kind := objectType.Kind()
	switch kind {
	case reflect.Slice:
		return len(object.([]interface{})) == 0
	case reflect.Map:
		return len(object.(map[string]interface{})) == 0
	case reflect.String:
		return object == ""
	}
	return false
}

func (app *JobApp) generateIndentedJsonStr(originalJsonStr string) (string, error) {
	var prettyJson bytes.Buffer
	if err := json.Indent(&prettyJson, []byte(originalJsonStr), "", "  "); err != nil {
		return "", err
	}
	return prettyJson.String(), nil
}

// GeneratePredictingJobParticipants returns a list of participants that should be used in a predicting job
func (app *JobApp) GeneratePredictingJobParticipants(modelUUID string) ([]JobParticipantInfoBase, error) {
	if modelUUID == "" {
		return nil, errors.New("invalid model uuid")
	}
	modelEntityInstance, err := app.ModelRepo.GetByUUID(modelUUID)
	if err != nil {
		return nil, err
	}
	modelEntity := modelEntityInstance.(*entity.Model)
	jobUUID := modelEntity.JobUUID
	jobAggregate, err := app.loadJobAggregate(jobUUID)
	if err != nil {
		return nil, err
	}
	participantList, err := jobAggregate.GeneratePredictingJobParticipants()
	if err != nil {
		return nil, err
	}
	var participantInfoList []JobParticipantInfoBase
	for _, participant := range participantList {
		// TODO: query from project participant repo to get the latest site info
		participantInfoList = append(participantInfoList, JobParticipantInfoBase{
			SiteUUID:    participant.SiteUUID,
			SiteName:    participant.SiteName,
			SitePartyID: participant.SitePartyID,
		})
	}
	return participantInfoList, nil
}

// SubmitJob creates the job
func (app *JobApp) SubmitJob(username string, request *JobSubmissionRequest) (*JobListItemBase, error) {
	jobAggregate, err := app.buildJobAggregate(username, request)
	if err != nil {
		return nil, err
	}
	if err := jobAggregate.SubmitJob(); err != nil {
		return nil, err
	}

	return &JobListItemBase{
		JobInfoBase:           JobInfoBase{},
		UUID:                  jobAggregate.Job.UUID,
		Status:                jobAggregate.Job.Status,
		CreationTime:          jobAggregate.Job.CreatedAt,
		FinishTime:            time.Time{},
		InitiatingSiteUUID:    jobAggregate.Job.InitiatingSiteUUID,
		InitiatingSiteName:    jobAggregate.Job.InitiatingSiteName,
		InitiatingSitePartyID: jobAggregate.Job.InitiatingSitePartyID,
		PendingOnThisSite:     false,
		FATEJobID:             jobAggregate.Job.FATEJobID,
		FATEJobStatus:         jobAggregate.Job.FATEJobStatus,
		IsInitiator:           true,
		Username:              username,
	}, nil
}

// ProcessNewRemoteJob processes the remote job creation request
func (app *JobApp) ProcessNewRemoteJob(request *RemoteJobCreationRequest) error {
	jobAggregate, err := app.buildJobAggregate(request.Username, &request.JobSubmissionRequest)
	if err != nil {
		return err
	}
	jobAggregate.Job.UUID = request.UUID
	return jobAggregate.HandleRemoteJobCreation()
}

// ProcessJobResponse handles job approval response
func (app *JobApp) ProcessJobResponse(uuid string, context *JobApprovalContext) error {
	jobAggregate, err := app.loadJobAggregate(uuid)
	if err != nil {
		return err
	}
	return jobAggregate.HandleJobApprovalResponse(context.SiteUUID, context.Approved)
}

// ProcessJobStatusUpdate handles job status update requests
func (app *JobApp) ProcessJobStatusUpdate(uuid string, context *JobStatusUpdateContext) error {
	jobAggregate, err := app.loadJobAggregate(uuid)
	if err != nil {
		return err
	}
	newJobStatusTemplate := &entity.Job{
		Status:           context.Status,
		StatusMessage:    context.StatusMessage,
		FATEJobID:        context.FATEJobID,
		FATEJobStatus:    context.FATEJobStatus,
		FATEModelID:      context.FATEModelID,
		FATEModelVersion: context.FATEModelVersion,
	}
	return jobAggregate.HandleJobStatusUpdate(newJobStatusTemplate, context.ParticipantStatusMap)
}

// loadSite is a helper function to return site entity object
func (app *JobApp) loadSite() (*entity.Site, error) {
	site := &entity.Site{
		Repo: app.SiteRepo,
	}
	if err := site.Load(); err != nil {
		return nil, errors.Wrapf(err, "failed to load site info")
	}
	return site, nil
}

// buildJobAggregate is a helper function to build the job aggregate from the job submission request
func (app *JobApp) buildJobAggregate(username string, request *JobSubmissionRequest) (*aggregate.JobAggregate, error) {
	site, err := app.loadSite()
	if err != nil {
		return nil, err
	}

	projectDataInstance, err := app.ProjectDataRepo.GetByDataUUID(request.InitiatorData.DataUUID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get initiator data")
	}
	projectData := projectDataInstance.(*entity.ProjectData)
	if request.ProjectUUID == "" {
		log.Warn().Msgf("missing project uuid, using the latest record")
		request.ProjectUUID = projectData.ProjectUUID
	}
	project, err := app.loadProject(request.ProjectUUID)
	if err != nil {
		return nil, err
	}

	requestJsonByte, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return nil, err
	}
	requestJsonStr := string(requestJsonByte)

	jobAggregate := &aggregate.JobAggregate{
		Job: &entity.Job{
			Name:                   request.Name,
			Description:            request.Description,
			ProjectUUID:            request.ProjectUUID,
			Type:                   request.Type,
			AlgorithmType:          request.AlgorithmType,
			AlgorithmComponentName: request.AlgorithmComponentName,
			EvaluateComponentName:  request.EvaluateComponentName,
			AlgorithmConfig: valueobject.AlgorithmConfig{
				TrainingValidationEnabled:     request.ValidationEnabled,
				TrainingValidationSizePercent: request.ValidationSizePercent,
				TrainingComponentsToDeploy:    request.ComponentsToDeploy,
			},
			ModelName:             request.ModelName,
			PredictingModelUUID:   request.ModelUUID,
			InitiatingSiteUUID:    projectData.SiteUUID,
			InitiatingSiteName:    projectData.SiteName,
			InitiatingSitePartyID: projectData.SitePartyID,
			InitiatingUser:        username,
			IsInitiatingSite:      site.UUID == projectData.SiteUUID,
			Conf:                  request.ConfJson,
			DSL:                   request.DSLJson,
			RequestJson:           requestJsonStr,
			FATEFlowContext: entity.FATEFlowContext{
				FATEFlowHost:    site.FATEFlowHost,
				FATEFlowPort:    site.FATEFlowHTTPPort,
				FATEFlowIsHttps: false,
			},
			Repo: app.JobRepo,
		},
		Initiator: &entity.JobParticipant{
			SiteUUID:           projectData.SiteUUID,
			SiteName:           projectData.SiteName,
			SitePartyID:        projectData.SitePartyID,
			SiteRole:           entity.JobParticipantRoleGuest,
			DataUUID:           projectData.DataUUID,
			DataName:           projectData.Name,
			DataDescription:    projectData.Description,
			DataTableName:      projectData.TableName,
			DataTableNamespace: projectData.TableNamespace,
			DataLabelName:      request.InitiatorData.LabelName,
			Status:             entity.JobParticipantStatusInitiator,
			Repo:               app.ParticipantRepo,
		},
		Participants:    map[string]*entity.JobParticipant{},
		JobRepo:         app.JobRepo,
		ParticipantRepo: app.ParticipantRepo,
		FMLManagerConnectionInfo: aggregate.FMLManagerConnectionInfo{
			Connected:  site.FMLManagerConnected,
			Endpoint:   site.FMLManagerEndpoint,
			ServerName: site.FMLManagerServerName,
		},
		JobContext: aggregate.JobContext{
			AutoApprovalEnabled: project.AutoApprovalEnabled,
			CurrentSiteUUID:     site.UUID,
		},
	}

	for _, otherData := range request.OtherData {
		projectDataInstance, err = app.ProjectDataRepo.GetByDataUUID(otherData.DataUUID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get other site data: %s", otherData.DataUUID)
		}
		projectData = projectDataInstance.(*entity.ProjectData)
		jobAggregate.Participants[projectData.SiteUUID] = &entity.JobParticipant{
			SiteUUID:           projectData.SiteUUID,
			SiteName:           projectData.SiteName,
			SitePartyID:        projectData.SitePartyID,
			SiteRole:           entity.JobParticipantRoleHost,
			DataUUID:           projectData.DataUUID,
			DataName:           projectData.Name,
			DataDescription:    projectData.Description,
			DataTableName:      projectData.TableName,
			DataTableNamespace: projectData.TableNamespace,
			DataLabelName:      request.InitiatorData.LabelName,
			Status:             entity.JobParticipantStatusPending,
			Repo:               app.ParticipantRepo,
		}
	}

	if request.Type == entity.JobTypePredict && jobAggregate.Job.IsInitiatingSite {
		log.Info().Msgf("changing participants role and job info according to original job info")
		if request.ModelUUID == "" {
			return nil, errors.New("invalid model uuid")
		}
		modelEntityInstance, err := app.ModelRepo.GetByUUID(request.ModelUUID)
		if err != nil {
			return nil, err
		}
		modelEntity := modelEntityInstance.(*entity.Model)
		jobUUID := modelEntity.JobUUID

		participant, err := app.loadJobParticipant(jobUUID, jobAggregate.Initiator.SiteUUID)
		if err != nil {
			return nil, err
		}
		jobAggregate.Initiator.SiteRole = participant.SiteRole

		for siteUUID := range jobAggregate.Participants {
			participant, err = app.loadJobParticipant(jobUUID, siteUUID)
			if err != nil {
				return nil, err
			}
			jobAggregate.Participants[siteUUID].SiteRole = participant.SiteRole
		}

		jobInstance, err := app.JobRepo.GetByUUID(jobUUID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to query job")
		}
		job := jobInstance.(*entity.Job)

		jobAggregate.Job.FATEModelID = modelEntity.FATEModelID
		jobAggregate.Job.FATEModelVersion = modelEntity.FATEModelVersion
		jobAggregate.Job.ModelName = job.ModelName
		jobAggregate.Job.AlgorithmType = job.AlgorithmType
		jobAggregate.Job.AlgorithmComponentName = job.AlgorithmComponentName
	}

	return jobAggregate, nil
}

// loadJobAggregate is a helper function to build the job aggregate from the info in the repo
func (app *JobApp) loadJobAggregate(jobUUID string) (*aggregate.JobAggregate, error) {
	site, err := app.loadSite()
	if err != nil {
		return nil, err
	}

	jobInstance, err := app.JobRepo.GetByUUID(jobUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query job")
	}
	job := jobInstance.(*entity.Job)
	job.Repo = app.JobRepo
	job.FATEFlowContext = entity.FATEFlowContext{
		FATEFlowHost:    site.FATEFlowHost,
		FATEFlowPort:    site.FATEFlowHTTPPort,
		FATEFlowIsHttps: false,
	}

	project, err := app.loadProject(job.ProjectUUID)
	if err != nil {
		return nil, err
	}

	jobAggregate := &aggregate.JobAggregate{
		Job:             job,
		Initiator:       nil,
		Participants:    map[string]*entity.JobParticipant{},
		JobRepo:         app.JobRepo,
		ParticipantRepo: app.ParticipantRepo,
		FMLManagerConnectionInfo: aggregate.FMLManagerConnectionInfo{
			Connected:  site.FMLManagerConnected,
			Endpoint:   site.FMLManagerEndpoint,
			ServerName: site.FMLManagerServerName,
		},
		JobContext: aggregate.JobContext{
			AutoApprovalEnabled: project.AutoApprovalEnabled,
			CurrentSiteUUID:     site.UUID,
		},
	}

	participantListInstance, err := app.ParticipantRepo.GetListByJobUUID(jobUUID)
	if err != nil {
		return nil, err
	}
	participantList := participantListInstance.([]entity.JobParticipant)
	for index, participant := range participantList {
		participantList[index].Repo = app.ParticipantRepo
		if participant.SiteUUID == jobAggregate.Job.InitiatingSiteUUID {
			jobAggregate.Initiator = &participantList[index]
		} else {
			jobAggregate.Participants[participant.SiteUUID] = &participantList[index]
		}
	}
	return jobAggregate, nil
}

// buildJobStatusMessage builds the status message based on the job and participant status
func (app *JobApp) buildJobStatusMessage(jobAggregate *aggregate.JobAggregate) string {
	switch jobAggregate.Job.Status {
	case entity.JobStatusUnknown:
		return "Job status unknown"
	case entity.JobStatusSucceeded:
		return "Job succeeded"
	case entity.JobStatusFailed:
		return "Job failed: " + jobAggregate.Job.StatusMessage
	case entity.JobStatusRunning:
		return "Job is running"
	}
	if jobAggregate.Job.Status == entity.JobStatusRejected {
		rejectedParticipantStr := ""
		for _, participant := range jobAggregate.Participants {
			if participant.Status == entity.JobParticipantStatusRejected {
				participantStr := fmt.Sprintf("%s(%d)", participant.SiteName, participant.SitePartyID)
				if rejectedParticipantStr == "" {
					rejectedParticipantStr += participantStr
				} else {
					rejectedParticipantStr += ", " + participantStr
				}
			}
		}
		return "Job is rejected by " + rejectedParticipantStr
	}
	if jobAggregate.Job.Status == entity.JobStatusPending {
		pendingParticipantStr := ""
		for _, participant := range jobAggregate.Participants {
			if participant.Status == entity.JobParticipantStatusPending {
				participantStr := fmt.Sprintf("%s(%d)", participant.SiteName, participant.SitePartyID)
				if pendingParticipantStr == "" {
					pendingParticipantStr += participantStr
				} else {
					pendingParticipantStr += ", " + participantStr
				}
			}
		}
		if pendingParticipantStr == "" {
			return "Job is waiting to start"
		} else {
			return "Job is pending on " + pendingParticipantStr
		}
	}
	return fmt.Sprintf(`Job status "%s" unknown`, jobAggregate.Job.Status.String())
}

// loadProject loads and returns the project entity from the repo
func (app *JobApp) loadProject(projectUUID string) (*entity.Project, error) {
	projectInstance, err := app.ProjectRepo.GetByUUID(projectUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query project")
	}
	return projectInstance.(*entity.Project), nil
}

// loadJobParticipant loads and return the job participant entity
func (app *JobApp) loadJobParticipant(jobUUID, siteUUID string) (*entity.JobParticipant, error) {
	participantInstance, err := app.ParticipantRepo.GetByJobAndSiteUUID(jobUUID, siteUUID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query participant")
	}
	participant := participantInstance.(*entity.JobParticipant)
	return participant, nil
}

// DeleteJob mark the job status as deleted
func (app *JobApp) DeleteJob(jobUUID string) error {
	jobInstance, err := app.JobRepo.GetByUUID(jobUUID)
	if err != nil {
		return errors.Wrap(err, "failed to query job")
	}
	job := jobInstance.(*entity.Job)
	job.Repo = app.JobRepo
	return job.UpdateStatus(entity.JobStatusDeleted)
}
