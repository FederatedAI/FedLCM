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
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Job represents a FATE job
type Job struct {
	gorm.Model
	Name                  string    `json:"name" gorm:"type:varchar(255)"`
	Description           string    `json:"description" gorm:"type:text"`
	UUID                  string    `json:"uuid" gorm:"type:varchar(36)"`
	ProjectUUID           string    `json:"project_uuid" gorm:"type:varchar(36)"`
	Type                  JobType   `json:"type"`
	Status                JobStatus `json:"status"`
	StatusMessage         string    `gorm:"type:text"`
	AlgorithmType         JobAlgorithmType
	AlgorithmConfig       AlgorithmConfig `gorm:"type:text"`
	ModelName             string          `json:"model_name" gorm:"type:varchar(255)"`
	PredictingModelUUID   string          `gorm:"type:varchar(36)"`
	InitiatingSiteUUID    string          `gorm:"type:varchar(36)"`
	InitiatingSiteName    string          `gorm:"type:varchar(255)"`
	InitiatingSitePartyID uint
	InitiatingUser        string `gorm:"type:varchar(255)"`
	FATEJobID             string `gorm:"type:varchar(255);column:fate_job_id"`
	FATEJobStatus         string `gorm:"type:varchar(36);column:fate_job_status"`
	FATEModelID           string `gorm:"type:varchar(255);column:fate_model_id"`
	FATEModelVersion      string `gorm:"type:varchar(255);column:fate_model_version"`
	Conf                  string `gorm:"type:text"`
	DSL                   string `gorm:"type:text"`
	RequestJson           string `gorm:"type:text"`
	FinishedAt            time.Time
	Repo                  repo.JobRepository `gorm:"-"`
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
)

// AlgorithmConfig contains algorithm configuration settings for the job
type AlgorithmConfig struct {
	TrainingValidationEnabled     bool     `json:"training_validation_enabled"`
	TrainingValidationSizePercent uint     `json:"training_validation_percent"`
	TrainingComponentsToDeploy    []string `json:"training_component_list_to_deploy"`
}

func (c AlgorithmConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *AlgorithmConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// Create initializes the job and save into the repo.
func (job *Job) Create() error {
	if job.UUID == "" {
		return errors.New("job must contains a valid uuid")
	}
	job.Status = JobStatusPending
	job.Model = gorm.Model{}
	job.FATEModelID = ""
	job.FATEJobID = ""
	job.FATEJobStatus = ""
	if err := job.Repo.Create(job); err != nil {
		return err
	}
	return nil
}

// Update updates the job info, including the fate job status.
func (job *Job) Update(newStatus *Job) error {
	if job.FATEJobID == "" || job.FATEJobStatus != newStatus.FATEJobStatus {
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
	job.Status = status
	if err := job.Repo.UpdateStatusByUUID(job); err != nil {
		return errors.Wrap(err, "failed to update job status")
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
