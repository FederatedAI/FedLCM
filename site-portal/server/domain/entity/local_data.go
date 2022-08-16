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
	"bytes"
	"database/sql/driver"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/FederatedAI/FedLCM/site-portal/server/domain/repo"
	"github.com/FederatedAI/FedLCM/site-portal/server/domain/valueobject"
	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/fateclient"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	once    sync.Once
	baseDir string
)

// LocalData represents an uploaded data file
type LocalData struct {
	gorm.Model
	UUID string `gorm:"type:varchar(36);index;unique"`
	// Name is the name to reference the data
	Name string `gorm:"type:varchar(255);not null"`
	// Description contains more text about the data
	Description string `json:"description" gorm:"type:text"`
	// Column is a list of headers in this data
	Column Headers `json:"column" gorm:"type:text"`
	// TableName is the name of the data in FATE system
	TableName string `json:"table_name" gorm:"type:varchar(255)"`
	// TableNamespace is the namespace of the data in FATE system
	TableNamespace string `json:"table_namespace" gorm:"type:varchar(255)"`
	// Count is the number of the samples in the data
	Count uint64 `json:"count"`
	// Features is feature name list
	Features Headers `json:"feature_size" gorm:"type:text"`
	// Preview is the first 10 lines of data in this data
	Preview string `json:"preview" gorm:"type:text"`
	// IDMetaInfo is the meta data describing the ID column
	IDMetaInfo *valueobject.IDMetaInfo `json:"id_meta_info" gorm:"type:text;column:id_meta_info"`
	// JobID is the related FATE upload job id
	JobID string `json:"-" gorm:"type:varchar(255);column:job_id"`
	// JobConf is the related FATE upload job conf
	JobConf string `json:"-" gorm:"type:text;column:job_conf"`
	// JobStatus is the current status of the data in FATE system
	JobStatus UploadJobStatus `json:"status"`
	// JobErrorString records the error message if the job failed
	JobErrorMsg string `json:"job_error_msg" gorm:"type:text"`
	// LocalFilePath is the file path relative to the baseDir
	LocalFilePath string `json:"-" gorm:"type:varchar(255)"`
	// UploadContext contains info needed to finish the upload job
	UploadContext UploadContext `json:"-" gorm:"-"`
	// Repo is used to store the necessary data into the storage
	Repo repo.LocalDataRepository `json:"-" gorm:"-"`
}

// UploadContext currently contains FATE flow connection info
type UploadContext struct {
	// FATEFlowHost is the host address of the service
	FATEFlowHost string
	// FATEFlowPort is the port of the service
	FATEFlowPort uint
	// FATEFlowIsHttps is whether the connection should be over TLS
	FATEFlowIsHttps bool
}

// UploadJobStatus is the status of the data in FATE system
type UploadJobStatus uint8

const (
	UploadJobStatusToBeCreated UploadJobStatus = iota
	UploadJobStatusCreating
	UploadJobStatusRunning
	UploadJobStatusFailed
	UploadJobStatusSucceeded
)

// MarshalJSON convert Cluster status to string
func (s *UploadJobStatus) MarshalJSON() ([]byte, error) {
	names := map[UploadJobStatus]string{
		UploadJobStatusToBeCreated: `"ToBeCreated"`,
		UploadJobStatusCreating:    `"Creating"`,
		UploadJobStatusRunning:     `"Running"`,
		UploadJobStatusFailed:      `"Failed"`,
		UploadJobStatusSucceeded:   `"Succeeded"`,
	}
	return bytes.NewBufferString(names[*s]).Bytes(), nil
}

// Headers are local data column names
type Headers []string

func (h Headers) Value() (driver.Value, error) {
	bJson, err := json.Marshal(h)
	return bJson, err
}

func (h *Headers) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), h)
}

// Upload save the file into local storage and uploaded it to FATE system
func (d *LocalData) Upload(fileHeader *multipart.FileHeader) error {
	if d.UploadContext.FATEFlowHost == "" || d.UploadContext.FATEFlowPort == 0 {
		return errors.Errorf("cannot find valid FATE flow connection info")
	}
	if err := d.Repo.CheckNameConflict(d.Name); err != nil {
		return err
	}
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	d.UUID = uuid.NewV4().String()
	parentDir := filepath.Join(getBaseDir(), d.UUID)
	if err = os.MkdirAll(parentDir, 0700); err != nil {
		return err
	}
	// check if the file name is valid
	if strings.TrimSpace(fileHeader.Filename) == "" {
		return errors.Errorf("file name can not be empty")
	}
	filename := strings.Split(fileHeader.Filename, ".")[0]
	var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9\s\d\-_/]*$`).MatchString
	if !isStringAlphabetic(filename) {
		return errors.Errorf("file name can not contain special characters")
	}
	if len(filename) < 2 {
		return errors.Errorf("file name is too short")
	}
	if len(filename) > 255 {
		return errors.Errorf("file name is too long")
	}
	// parse the csv file and record some meta data and previews
	if csvErr := func() error {
		csvFile, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer csvFile.Close()
		csvReader := csv.NewReader(csvFile)

		// check headers
		headers, err := csvReader.Read()
		if err != nil {
			return nil
		}
		d.Column = headers
		if features, containsID := func() ([]string, bool) {
			containsID := false
			features := Headers{}
			for _, header := range headers {
				if strings.ToLower(header) == "id" {
					containsID = true
				} else if strings.ToLower(header) != "y" {
					features = append(features, header)
				}
			}
			return features, containsID
		}(); containsID == false {
			return errors.New("data must contain an id field")
		} else {
			d.Features = features
		}

		// count data lines and keep the first 10 lines as preview
		if count, previewJsonStr, err := func() (uint64, string, error) {
			var records []map[string]string
			var lines uint64
			for {
				recordLine, err := csvReader.Read()
				if err != nil {
					break
				}
				if lines < 10 {
					record := map[string]string{}
					for i, value := range recordLine {
						record[headers[i]] = value
					}
					records = append(records, record)
				}
				lines++
			}
			if err != nil && err != io.EOF {
				return 0, "", err
			}
			jsonBytes, err := json.Marshal(records)
			if err != nil {
				return 0, "", err
			}
			return lines, string(jsonBytes), nil
		}(); err != nil {
			return errors.Wrap(err, "error parsing data records")
		} else {
			d.Count = count
			d.Preview = previewJsonStr
		}
		return nil
	}(); csvErr != nil {
		return errors.Wrap(csvErr, "error parsing the uploaded csv file")
	}

	d.LocalFilePath = filepath.Join(d.UUID, filepath.Base(fileHeader.Filename))
	dst, err := os.Create(filepath.Join(getBaseDir(), d.LocalFilePath))
	if err != nil {
		return err
	}
	log.Info().Msgf("saving uploaded data file %s to %s", fileHeader.Filename, dst.Name())
	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	d.JobStatus = UploadJobStatusToBeCreated
	d.IDMetaInfo = nil
	d.TableName = fmt.Sprintf("table-%s", d.UUID)
	d.TableNamespace = fmt.Sprintf("ns-%s", d.UUID)
	if err := d.Repo.Create(d); err != nil {
		return err
	}

	go func() {
		log.Info().Msgf("uploading data file %s to FATE", dst.Name())
		fateClient := fateclient.NewFATEFlowClient(d.UploadContext.FATEFlowHost, d.UploadContext.FATEFlowPort, d.UploadContext.FATEFlowIsHttps)
		d.ChangeJobStatus(UploadJobStatusCreating)
		// 2 stands for SPARK_PULSAR
		backend := 2
		if viper.GetBool("siteportal.fate.eggroll.enabled") {
			backend = 0
		}
		uploadConf := fateclient.DataUploadRequest{
			File:      dst.Name(),
			Head:      1,
			Partition: 8, // XXX: use viper configuration instead of a hard-code one
			WorkMode:  1,
			Backend:   backend,
			Namespace: d.TableNamespace,
			TableName: d.TableName,
			Drop:      1,
		}
		confBytes, _ := json.Marshal(uploadConf)
		d.JobConf = string(confBytes)
		if jobID, err := fateClient.UploadData(uploadConf); err != nil {
			log.Err(err).Msgf("failed to upload data %s to FATE", d.UUID)
			d.ChangeJobStatus(UploadJobStatusFailed)
			d.JobErrorMsg = err.Error()
		} else {
			log.Info().Msgf("uploading job ID is %s", jobID)
			d.ChangeJobStatus(UploadJobStatusRunning)
			d.JobID = jobID
		}
		if err := d.Repo.UpdateJobInfoByUUID(d); err != nil {
			log.Err(err).Str("data uuid", d.UUID).Msg("failed to update data job info")
			return
		}
		// TODO: we should use a cron-like task to monitor the job status, preferably even be able to survive service restarts
		func() {
			for d.JobStatus == UploadJobStatusRunning {
				if status, err := fateClient.QueryJobStatus(d.JobID); err != nil {
					// TODO: exit after maximum number of retries and set job status to failed
					log.Err(err).Str("data uuid", d.UUID).Msg("failed to query job status")
				} else if status == "success" {
					d.ChangeJobStatus(UploadJobStatusSucceeded)
				} else if status == "canceled" || status == "timeout" || status == "failed" {
					log.Error().Str("data uuid", d.UUID).Str("status", status).Send()
					d.ChangeJobStatus(UploadJobStatusFailed)
					return
				} else {
					log.Info().Str("data uuid", d.UUID).Str("status", status).Send()
				}
				time.Sleep(5 * time.Second)
			}
		}()
		log.Info().Str("data uuid", d.UUID).Str("job id", d.JobID).
			Msgf("finished monitoring uploading job")
	}()
	return nil
}

// ChangeJobStatus upload the data's upload job status
func (d *LocalData) ChangeJobStatus(newStatus UploadJobStatus) {
	d.JobStatus = newStatus
	if err := d.Repo.UpdateJobInfoByUUID(d); err != nil {
		log.Err(err).Str("data uuid", d.UUID).Msgf("failed to update data status")
	}
}

// GetAbsFilePath returns the absolute path the local date file
func (d *LocalData) GetAbsFilePath() (string, error) {
	absPath := filepath.Join(getBaseDir(), d.LocalFilePath)
	if _, err := os.Stat(absPath); err == nil {
		return absPath, nil
	} else if os.IsNotExist(err) {
		return "", errors.Errorf("file no longer exists")
	} else {
		return "", errors.Wrap(err, "error checking file stat")
	}
}

func (d *LocalData) Destroy() error {
	log.Info().Str("data uuid", d.UUID).Msgf("removing data")
	// remove the data file
	absPath := filepath.Join(getBaseDir(), d.LocalFilePath)
	if err := func() error {
		if _, err := os.Stat(absPath); err != nil {
			if os.IsNotExist(err) {
				return nil
			} else {
				return errors.Wrap(err, "error checking file stat")
			}
		}
		if err := os.RemoveAll(filepath.Dir(absPath)); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		return errors.Wrapf(err, "error deleting file")
	}
	log.Info().Msgf("deleting table %s namespace: %s from FATE", d.TableName, d.TableNamespace)
	fateClient := fateclient.NewFATEFlowClient(d.UploadContext.FATEFlowHost, d.UploadContext.FATEFlowPort, d.UploadContext.FATEFlowIsHttps)
	if err := fateClient.DeleteTable(d.TableNamespace, d.TableName); err != nil {
		// TODO: retry
		log.Err(err).Msgf("error deleting table from FATE")
	}
	// delete the record from repo
	return d.Repo.DeleteByUUID(d.UUID)
}

// UpdateIDMetaInfo changes the meta info of the id column
func (d *LocalData) UpdateIDMetaInfo(info *valueobject.IDMetaInfo) error {
	d.IDMetaInfo = info
	return d.Repo.UpdateIDMetaInfoByUUID(d.UUID, d.IDMetaInfo)
}

func getBaseDir() string {
	once.Do(func() {
		baseDir = viper.GetString("siteportal.localdata.basedir")
		if baseDir == "" {
			panic("empty base folder")
		}
		if !filepath.IsAbs(baseDir) {
			panic("baseFolder should be absolute path")
		}
		if err := os.MkdirAll(baseDir, 0700); err != nil {
			panic(err)
		}
	})
	return baseDir
}
