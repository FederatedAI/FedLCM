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

package entity

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/event"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fateclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Job represents a FATE job
type Job struct {
	gorm.Model
	Name                   string    `json:"name" gorm:"type:varchar(255)"`
	Description            string    `json:"description" gorm:"type:text"`
	UUID                   string    `json:"uuid" gorm:"type:varchar(36)"`
	ProjectUUID            string    `json:"project_uuid" gorm:"type:varchar(36)"`
	Type                   JobType   `json:"type"`
	Status                 JobStatus `json:"status"`
	StatusMessage          string    `gorm:"type:text"`
	AlgorithmType          JobAlgorithmType
	AlgorithmComponentName string                      `json:"algorithm_component_name" gorm:"type:varchar(255)"`
	EvaluateComponentName  string                      `json:"evaluate_component_name" gorm:"type:varchar(255)"`
	AlgorithmConfig        valueobject.AlgorithmConfig `gorm:"type:text"`
	ModelName              string                      `json:"model_name" gorm:"type:varchar(255)"`
	PredictingModelUUID    string                      `gorm:"type:varchar(36)"`
	InitiatingSiteUUID     string                      `gorm:"type:varchar(36)"`
	InitiatingSiteName     string                      `gorm:"type:varchar(255)"`
	InitiatingSitePartyID  uint
	InitiatingUser         string `gorm:"type:varchar(255)"`
	IsInitiatingSite       bool
	FATEJobID              string `gorm:"type:varchar(255);column:fate_job_id"`
	FATEJobStatus          string `gorm:"type:varchar(36);column:fate_job_status"`
	FATEModelID            string `gorm:"type:varchar(255);column:fate_model_id"`
	FATEModelVersion       string `gorm:"type:varchar(255);column:fate_model_version"`
	FinishedAt             time.Time
	ResultJson             string             `gorm:"type:text"`
	Conf                   string             `gorm:"type:text"`
	DSL                    string             `gorm:"type:text"`
	RequestJson            string             `gorm:"type:text"`
	FATEFlowContext        FATEFlowContext    `gorm:"-"`
	Repo                   repo.JobRepository `gorm:"-"`
}

// FATEFlowContext currently contains FATE flow connection info
type FATEFlowContext struct {
	// FATEFlowHost is the host address of the service
	FATEFlowHost string
	// FATEFlowPort is the port of the service
	FATEFlowPort uint
	// FATEFlowIsHttps is whether the connection should be over TLS
	FATEFlowIsHttps bool
}

// JobStatus is the enum of job status
type JobStatus uint8

const (
	JobStatusUnknown JobStatus = iota
	JobStatusPending
	JobStatusRejected
	JobStatusRunning
	JobStatusFailed
	JobStatusSucceeded
	JobStatusDeploying
	JobStatusDeleted
)

func (s JobStatus) String() string {
	names := map[JobStatus]string{
		JobStatusUnknown:   "Unknown",
		JobStatusPending:   "Pending",
		JobStatusRejected:  "Rejected",
		JobStatusRunning:   "Running",
		JobStatusFailed:    "Failed",
		JobStatusSucceeded: "Succeeded",
	}
	return names[s]
}

// JobType is the enum of job type
type JobType uint8

const (
	JobTypeUnknown JobType = iota
	JobTypeTraining
	JobTypePredict
	JobTypePSI
)

func (t JobType) String() string {
	names := map[JobType]string{
		JobTypeUnknown:  "Unknown",
		JobTypeTraining: "Modeling",
		JobTypePredict:  "Predict",
		JobTypePSI:      "PSI",
	}
	return names[t]
}

// JobAlgorithmType is the enum of the job algorithm
type JobAlgorithmType uint8

const (
	JobAlgorithmTypeUnknown JobAlgorithmType = iota
	JobAlgorithmTypeHomoLR
	JobAlgorithmTypeHomoSBT
	JobAlgorithmTypeHeteroLR
	JobAlgorithmTypeHeteroSBT
)

const (
	jobResultModelEvaluation  = "model_evaluation"
	jobResultOutputData       = "output_data"
	jobResultOutputDataMeta   = "output_meta"
	jobResultComponentSummary = "component_summary"
)

// Create initializes the job and save into the repo. The uuid will be automatically generated if not set
func (job *Job) Create() error {
	job.Status = JobStatusPending
	job.Model = gorm.Model{}
	if job.UUID == "" {
		job.UUID = uuid.NewV4().String()
	}
	if job.Type != JobTypePredict {
		job.FATEModelID = ""
		job.FATEModelVersion = ""
	}
	job.FATEJobID = ""
	job.FATEJobStatus = ""
	if err := job.Repo.Create(job); err != nil {
		return err
	}
	return nil
}

// Validate checks the job configuration
func (job *Job) Validate() error {
	if job.Type == JobTypeTraining {
		// XXX: this is a simple and strict check that training job must contain an evaluation component
		if !strings.Contains(job.DSL, "Evaluation") {
			return errors.New("training job must contain an Evaluation module")
		}
		if !(strings.Contains(job.DSL, "HomoLR") || strings.Contains(job.DSL, "HomoSecureBoost") || strings.Contains(job.DSL, "HeteroLR") || strings.Contains(job.DSL, "HeteroSecureBoost")) {
			return errors.Errorf("training job must contain at least one algorithm component")
		}
	} else if job.Type == JobTypePredict {
		if job.FATEModelID == "" || job.FATEModelVersion == "" {
			return errors.New("predicting job must contain model id and model version")
		}
	}
	return nil
}

// SubmitToFATE submits the job to FATE system and starts a monitoring routine
func (job *Job) SubmitToFATE(finishCB func()) error {
	if job.Conf == "" || job.DSL == "" {
		return errors.New("no conf or dsl")
	}
	fateClient := fateclient.NewFATEFlowClient(job.FATEFlowContext.FATEFlowHost, job.FATEFlowContext.FATEFlowPort, job.FATEFlowContext.FATEFlowIsHttps)
	fateJobID, modelInfo, submissionErr := fateClient.SubmitJob(job.Conf, job.DSL)
	if submissionErr != nil {
		if err := job.UpdateStatus(JobStatusFailed); err != nil {
			log.Err(err).Send()
		}
		if err := job.UpdateStatusMessage(fmt.Sprintf("failed to submit FATE job: %v", submissionErr)); err != nil {
			log.Err(err).Send()
		}
		return errors.Wrap(submissionErr, "failed to submit job to FATE")
	}
	job.FATEJobID = fateJobID
	job.FATEJobStatus = "started"
	if job.Type != JobTypePredict {
		job.FATEModelVersion = modelInfo.ModelVersion
		job.FATEModelID = modelInfo.ModelID
	}
	if err := job.Repo.UpdateFATEJobInfoByUUID(job); err != nil {
		return errors.Wrap(err, "failed to update FATE job info")
	}
	if err := job.UpdateStatus(JobStatusRunning); err != nil {
		return err
	}
	go job.waitForJobFinish(finishCB)
	return nil
}

// waitForJobFinish checks the job status until it is finished and calls the callback function
func (job *Job) waitForJobFinish(finishCB func()) {
	for job.Status == JobStatusRunning {
		if err := job.CheckFATEJobStatus(); err != nil {
			// TODO: exit after maximum number of retries
			log.Err(err).Str("job uuid", job.UUID).Str("fate job id", job.FATEJobID).Msg("failed to query job status")
		}
		log.Info().Str("job uuid", job.UUID).Str("fate job id", job.FATEJobID).Msg("job not finished, waiting")
		time.Sleep(20 * time.Second)
	}
	log.Info().Str("job uuid", job.UUID).Str("fate job id", job.FATEJobID).Msgf("job finished, call-back exists: %v", finishCB != nil)
	if finishCB != nil {
		finishCB()
	}
}

// CheckFATEJobStatus issues job status query
func (job *Job) CheckFATEJobStatus() error {
	if job.Status != JobStatusRunning {
		return nil
	}
	fateClient := fateclient.NewFATEFlowClient(job.FATEFlowContext.FATEFlowHost, job.FATEFlowContext.FATEFlowPort, job.FATEFlowContext.FATEFlowIsHttps)

	var err error
	status := job.FATEJobStatus
	if status, err = fateClient.QueryJobStatus(job.FATEJobID); err != nil {
		log.Err(err).Str("job uuid", job.UUID).Str("fate job id", job.FATEJobID).Msg("failed to query job status")
		return err
	} else {
		log.Info().Str("job uuid", job.UUID).Str("fate job id", job.FATEJobID).Str("status", status).Send()
		if job.FATEJobStatus != status {
			job.FATEJobStatus = status
			if err := job.Repo.UpdateFATEJobStatusByUUID(job); err != nil {
				return errors.Wrap(err, "failed to update FATE job status")
			}
		}
		switch status {
		case "success":
			// for training job, the initiating site needs to deploy the model
			// other sites can only update its model info via the updates from the initiating site
			if job.Type == JobTypeTraining {
				if err := job.UpdateStatus(JobStatusDeploying); err != nil {
					return err
				}
				if job.IsInitiatingSite {
					log.Info().Str("job uuid", job.UUID).Msgf("start deploying trained model")
					if err := job.deployTrainedModel(); err != nil {
						return errors.Wrap(err, "failed to deploy trained model")
					}
					if err := job.UpdateStatus(JobStatusSucceeded); err != nil {
						return err
					}
				} else {
					log.Info().Str("job uuid", job.UUID).Msgf("FATE job finished, waiting for deployed model info")
				}
			} else {
				if err := job.UpdateStatus(JobStatusSucceeded); err != nil {
					return err
				}
			}
		case "canceled", "timeout", "failed":
			if err := job.UpdateStatus(JobStatusFailed); err != nil {
				return err
			}
			if err := job.UpdateStatusMessage("FATE Job status is: " + status); err != nil {
				return err
			}
		case "running", "waiting":
			if err := job.UpdateStatus(JobStatusRunning); err != nil {
				return err
			}
		default:
			log.Error().Msgf("unknown job status: %s", status)
		}
		if job.Status != JobStatusRunning {
			job.FinishedAt = time.Now()
			if err := job.Repo.UpdateFinishTimeByUUID(job); err != nil {
				log.Err(err).Str("job uuid", job.UUID).Str("fate job id", job.FATEJobID).Msg("failed to update job finish time")
			}
		}
	}
	return nil
}

// Update updates the job info, including the fate job status. If the job starts running, a monitoring routine is started
func (job *Job) Update(newStatus *Job) error {
	if job.IsInitiatingSite {
		return errors.New("current site is job initiating site")
	}
	if job.FATEJobID == "" || job.FATEJobStatus != newStatus.FATEJobStatus ||
		job.FATEModelID != newStatus.FATEModelID || job.FATEModelVersion != newStatus.FATEModelVersion {
		job.FATEJobID = newStatus.FATEJobID
		job.FATEJobStatus = newStatus.FATEJobStatus
		job.FATEModelID = newStatus.FATEModelID
		job.FATEModelVersion = newStatus.FATEModelVersion
		if err := job.Repo.UpdateFATEJobInfoByUUID(job); err != nil {
			return errors.Wrap(err, "failed to update FATE job info")
		}
	}
	if job.Status != newStatus.Status {
		if err := job.UpdateStatus(newStatus.Status); err != nil {
			return err
		}
		if job.Status == JobStatusRunning {
			log.Info().Msgf("job is started by the initiating site, waiting for it to finish...")
			go job.waitForJobFinish(nil)
		}
	}
	if job.StatusMessage != newStatus.StatusMessage {
		if err := job.UpdateStatusMessage(newStatus.StatusMessage); err != nil {
			return err
		}
	}
	return nil
}

// UpdateStatus updates the job's status
func (job *Job) UpdateStatus(status JobStatus) error {
	// do not modify deleted jobs
	if job.Status != status && job.Status != JobStatusDeleted {
		job.Status = status
		if err := job.Repo.UpdateStatusByUUID(job); err != nil {
			return errors.Wrap(err, "failed to update job status")
		}
	}
	return nil
}

// UpdateStatusMessage updates the job's status message
func (job *Job) UpdateStatusMessage(message string) error {
	job.StatusMessage = message
	if err := job.Repo.UpdateStatusMessageByUUID(job); err != nil {
		return errors.Wrap(err, "failed to update job status message")
	}
	return nil
}

// UpdateResultInfo gets the job result from FATE and updates it into the repo
func (job *Job) UpdateResultInfo(partyID uint, role JobParticipantRole) error {
	if job.Type == JobTypeTraining {
		fateClient := fateclient.NewFATEFlowClient(job.FATEFlowContext.FATEFlowHost, job.FATEFlowContext.FATEFlowPort, job.FATEFlowContext.FATEFlowIsHttps)

		metric, err := fateClient.GetComponentMetric(fateclient.ComponentTrackingCommonRequest{
			JobID:         job.FATEJobID,
			PartyID:       partyID,
			Role:          string(role),
			ComponentName: job.EvaluateComponentName,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to query metric data")
		}
		resultByte, err := json.Marshal(map[string]interface{}{
			jobResultModelEvaluation: metric,
		})
		if err != nil {
			return err
		}
		job.ResultJson = string(resultByte)
	} else if job.Type == JobTypePredict {
		fateClient := fateclient.NewFATEFlowClient(job.FATEFlowContext.FATEFlowHost, job.FATEFlowContext.FATEFlowPort, job.FATEFlowContext.FATEFlowIsHttps)
		data, meta, err := fateClient.GetComponentOutputDataSummary(fateclient.ComponentTrackingCommonRequest{
			JobID:         job.FATEJobID,
			PartyID:       partyID,
			Role:          string(role),
			ComponentName: job.AlgorithmComponentName,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to query output data")
		}
		resultByte, err := json.Marshal(map[string]interface{}{
			jobResultOutputData:     data,
			jobResultOutputDataMeta: meta,
		})
		if err != nil {
			return err
		}
		job.ResultJson = string(resultByte)
	} else if job.Type == JobTypePSI {
		fateClient := fateclient.NewFATEFlowClient(job.FATEFlowContext.FATEFlowHost, job.FATEFlowContext.FATEFlowPort, job.FATEFlowContext.FATEFlowIsHttps)
		componentTrackingCommonRequest := fateclient.ComponentTrackingCommonRequest{
			JobID:         job.FATEJobID,
			PartyID:       partyID,
			Role:          string(role),
			ComponentName: "intersection_0",
		}
		intersectionSummary, err := fateClient.GetComponentSummary(componentTrackingCommonRequest)
		dataSummary, meta, err := fateClient.GetComponentOutputDataSummary(componentTrackingCommonRequest)
		if err != nil {
			return errors.Wrapf(err, "failed to query output data")
		}
		resultByte, err := json.Marshal(map[string]interface{}{
			jobResultOutputData:       dataSummary,
			jobResultOutputDataMeta:   meta,
			jobResultComponentSummary: intersectionSummary,
		})
		if err != nil {
			return err
		}
		job.ResultJson = string(resultByte)
	} else {
		return errors.Errorf("unknown job type: %v", job.Type)
	}
	if err := job.Repo.UpdateResultInfoByUUID(job); err != nil {
		return err
	}
	if job.Type == JobTypeTraining {
		// For hetero job, we don't need to create the model on host side because predict job cannot be launched from the host side.
		if (job.AlgorithmType == JobAlgorithmTypeHeteroLR || job.AlgorithmType == JobAlgorithmTypeHeteroSBT) && job.IsInitiatingSite == false {
			return nil
		}
		eventExchange := event.NewSelfHttpExchange()
		return eventExchange.PostEvent(event.ModelCreationEvent{
			Name:                   job.ModelName,
			ModelID:                job.FATEModelID,
			ModelVersion:           job.FATEModelVersion,
			ComponentName:          job.AlgorithmComponentName,
			ProjectUUID:            job.ProjectUUID,
			JobUUID:                job.UUID,
			JobName:                job.Name,
			Role:                   string(role),
			PartyID:                partyID,
			Evaluation:             job.GetTrainingResultSummary(),
			ComponentAlgorithmType: uint8(job.AlgorithmType),
		})
	}
	return nil
}

// GetTrainingResultSummary returns the summary mapping of the training result. It is the evaluation info of the model
func (job *Job) GetTrainingResultSummary() map[string]string {
	if job.Status != JobStatusSucceeded || job.Type != JobTypeTraining || job.ResultJson == "" {
		return nil
	}
	EvaluationSummary := map[string]string{}
	type EvaluationMeta struct {
		MetricType string `json:"metric_type"`
		Name       string `json:"name"`
	}
	type EvaluationItem struct {
		Data [][]interface{} `json:"data"`
		Meta EvaluationMeta  `json:"meta"`
	}
	type EvaluationInfo struct {
		Train      map[string]EvaluationItem `json:"train"`
		Validation map[string]EvaluationItem `json:"validate"`
	}
	type EvaluationData struct {
		ModelEvaluation EvaluationInfo `json:"model_evaluation"`
	}
	evaluationData := EvaluationData{}
	if err := json.Unmarshal([]byte(job.ResultJson), &evaluationData); err != nil {
		log.Err(err).Msg("failed to unmarshal job result")
		return nil
	}
	buildSummary := func(itemMap map[string]EvaluationItem) {
		for _, item := range itemMap {
			if item.Meta.MetricType == "EVALUATION_SUMMARY" {
				for _, record := range item.Data {
					if len(record) == 2 {
						EvaluationSummary[fmt.Sprintf("%v", record[0])] = fmt.Sprintf("%v", record[1])
					}
				}
			}
		}
	}
	if len(evaluationData.ModelEvaluation.Validation) > 0 {
		buildSummary(evaluationData.ModelEvaluation.Validation)
	} else {
		buildSummary(evaluationData.ModelEvaluation.Train)
	}
	return EvaluationSummary
}

// GetPredictingResultPreview returns a fraction of the predicting result data
func (job *Job) GetPredictingResultPreview() ([]string, [][]interface{}, int) {
	if job.Status != JobStatusSucceeded || job.Type != JobTypePredict || job.ResultJson == "" {
		return nil, nil, -1
	}
	type OutputMeta struct {
		Header [][]string `json:"header"`
		Total  []int      `json:"total"`
	}
	type PredictResultInfo struct {
		OutputData [][][]interface{} `json:"output_data"`
		OutputMeta OutputMeta        `json:"output_meta"`
	}
	predictResultInfo := PredictResultInfo{}
	if err := json.Unmarshal([]byte(job.ResultJson), &predictResultInfo); err != nil {
		log.Err(err).Msg("failed to unmarshal job result")
		return nil, nil, -1
	}
	return predictResultInfo.OutputMeta.Header[0], predictResultInfo.OutputData[0], predictResultInfo.OutputMeta.Total[0]
}

// GetIntersectionResult parses the PSI job result and returns the preview data
func (job *Job) GetIntersectionResult() ([]string, [][]interface{}, int, float64) {
	if job.Status != JobStatusSucceeded || job.Type != JobTypePSI || job.ResultJson == "" {
		return nil, nil, -1, 0
	}
	type OutputMeta struct {
		Header [][]string `json:"header"`
		Total  []int      `json:"total"`
	}
	type ComponentSummary struct {
		IntersectNum  int     `json:"intersect_num"`
		IntersectRate float64 `json:"intersect_rate"`
	}
	type IntersectionResultInfo struct {
		OutputData       [][][]interface{} `json:"output_data"`
		OutputMeta       OutputMeta        `json:"output_meta"`
		ComponentSummary ComponentSummary  `json:"component_summary"`
	}
	intersectionResultInfo := IntersectionResultInfo{}
	if err := json.Unmarshal([]byte(job.ResultJson), &intersectionResultInfo); err != nil {
		log.Err(err).Msg("failed to unmarshal intersection result")
		return nil, nil, -1, 0
	}
	res := intersectionResultInfo
	if res.ComponentSummary.IntersectNum == 0 || res.ComponentSummary.IntersectRate == 0 {
		return nil, nil, 0, 0
	}
	return res.OutputMeta.Header[0], res.OutputData[0], res.ComponentSummary.IntersectNum, res.ComponentSummary.IntersectRate
}

// deployTrainedModel deploys the trained model and update job with the newly deployed model's info
func (job *Job) deployTrainedModel() error {
	if job.Type != JobTypeTraining {
		return nil
	}
	if job.FATEJobStatus != "success" {
		log.Warn().Str("job uuid", job.UUID).Msgf("fate job status is %s, cannot deploy model", job.FATEJobStatus)
		return nil
	}
	fateClient := fateclient.NewFATEFlowClient(job.FATEFlowContext.FATEFlowHost, job.FATEFlowContext.FATEFlowPort, job.FATEFlowContext.FATEFlowIsHttps)
	componentList := job.AlgorithmConfig.TrainingComponentsToDeploy
	if len(componentList) == 0 {
		log.Error().Msg("there should have been at least one component in the to-deploy list")
	}
	modelInfo, err := fateClient.DeployModel(fateclient.ModelDeployRequest{
		ModelInfo: fateclient.ModelInfo{
			ModelID:      job.FATEModelID,
			ModelVersion: job.FATEModelVersion,
		},
		ComponentList: componentList,
	})
	if err != nil {
		return errors.Wrap(err, "failed to ask FATE to deploy model")
	}
	job.FATEModelID = modelInfo.ModelID
	job.FATEModelVersion = modelInfo.ModelVersion
	return job.Repo.UpdateFATEJobInfoByUUID(job)
}
