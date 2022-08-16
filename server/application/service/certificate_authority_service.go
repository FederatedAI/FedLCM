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
	"encoding/json"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

// CertificateAuthorityApp provides functions to manage the Certificate Authority
type CertificateAuthorityApp struct {
	CertificateAuthorityRepo repo.CertificateAuthorityRepository
}

// CertificateAuthorityDetail contains the detail info of Certificate Authority
type CertificateAuthorityDetail struct {
	CertificateAuthorityEditableItem
	UUID          string                     `json:"uuid"`
	CreatedAt     time.Time                  `json:"created_at"`
	Status        CertificateAuthorityStatus `json:"status"`
	StatusMessage string                     `json:"status_message"`
}

// CertificateAuthorityEditableItem contains properties of a CA that should be provided by the user
type CertificateAuthorityEditableItem struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description"`
	Type        entity.CertificateAuthorityType `json:"type"`
	Config      map[string]interface{}          `json:"config"`
}

// CertificateAuthorityStatus is the certificate authority status
type CertificateAuthorityStatus uint8

const (
	CertificateAuthorityStatusUnknown CertificateAuthorityStatus = iota
	CertificateAuthorityStatusUnhealthy
	CertificateAuthorityStatusHealthy
)

// Get returns the current CA configuration
func (app *CertificateAuthorityApp) Get() (*CertificateAuthorityDetail, error) {
	instance, err := app.CertificateAuthorityRepo.GetFirst()
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, nil
		}
		return nil, err
	}
	ca := instance.(*entity.CertificateAuthority)
	// check the ca status
	caStatus := CertificateAuthorityStatusUnknown
	caStatusMessage := ""
	switch ca.Type {
	case entity.CertificateAuthorityTypeStepCA:
		_, err = ca.Client()
		if err != nil {
			caStatus = CertificateAuthorityStatusUnhealthy
			caStatusMessage = err.Error()
		} else {
			caStatus = CertificateAuthorityStatusHealthy
		}
	}

	var config map[string]interface{}
	err = json.Unmarshal([]byte(ca.ConfigurationJSON), &config)
	if err != nil {
		return nil, err
	}
	return &CertificateAuthorityDetail{
		CertificateAuthorityEditableItem: CertificateAuthorityEditableItem{
			Name:        ca.Name,
			Description: ca.Description,
			Type:        ca.Type,
			Config:      config,
		},
		UUID:          ca.UUID,
		CreatedAt:     ca.CreatedAt,
		Status:        caStatus,
		StatusMessage: caStatusMessage,
	}, nil

}

// CreateCA creates a certificate authority
func (app *CertificateAuthorityApp) CreateCA(caInfo *CertificateAuthorityEditableItem) error {
	//check CA exist
	instance, _ := app.CertificateAuthorityRepo.GetFirst()
	if instance != nil {
		return errors.Errorf("certificate authority already exists")
	}
	switch caInfo.Type {
	case entity.CertificateAuthorityTypeStepCA:
		//Decode map, remove unmapped keys
		var config entity.CertificateAuthorityConfigurationStepCA
		err := mapstructure.Decode(caInfo.Config, &config)
		if err != nil {
			return err
		}
		// validate URL schema
		u, err := url.ParseRequestURI(config.ServiceURL)
		if err != nil || u.Scheme == "" && u.Host == "" {
			return errors.Errorf("Service URL is invalid: http:// or https:// schema is required")
		}
		// marshal to json
		caConfig, err := json.Marshal(config)
		if err != nil {
			return err
		}
		ca := &entity.CertificateAuthority{
			UUID:              uuid.NewV4().String(),
			Name:              caInfo.Name,
			Description:       caInfo.Description,
			Type:              caInfo.Type,
			ConfigurationJSON: string(caConfig),
		}
		// validate CA config info
		_, err = ca.Client()
		if err != nil {
			return err
		}
		return app.CertificateAuthorityRepo.Create(ca)
	}
	return errors.Errorf("unknown certificate authority type: %v", caInfo.Type)
}

// Update changes CA settings
func (app *CertificateAuthorityApp) Update(uuid string, caInfo *CertificateAuthorityEditableItem) error {
	switch caInfo.Type {
	case entity.CertificateAuthorityTypeStepCA:
		var config entity.CertificateAuthorityConfigurationStepCA
		// decode map, remove unmapped keys
		err := mapstructure.Decode(caInfo.Config, &config)
		if err != nil {
			return err
		}
		// validate URL schema
		u, err := url.ParseRequestURI(config.ServiceURL)
		if err != nil || u.Scheme == "" && u.Host == "" {
			return errors.Errorf("Service URL is invalid: http:// or https:// schema is required")
		}
		// marshal to json
		caConfig, err := json.Marshal(config)
		if err != nil {
			return err
		}
		ca := &entity.CertificateAuthority{
			UUID:              uuid,
			Name:              caInfo.Name,
			Description:       caInfo.Description,
			Type:              caInfo.Type,
			ConfigurationJSON: string(caConfig),
		}
		// validate CA config info
		_, err = ca.Client()
		if err != nil {
			return err
		}
		return app.CertificateAuthorityRepo.UpdateByUUID(ca)
	}
	return errors.Errorf("unknown certificate authority type: %v", caInfo.Type)
}

// GetBuiltInCAConfig returns the config of built-in StepCA config
// Unnecessary to move to domain package due to simple logic
func (app *CertificateAuthorityApp) GetBuiltInCAConfig() (*entity.CertificateAuthorityConfigurationStepCA, error) {
	serverURL := viper.GetString("lifecyclemanager.builtinca.host")
	provisionerName := viper.GetString("lifecyclemanager.builtinca.provisioner.name")
	password := viper.GetString("lifecyclemanager.builtinca.provisioner.password")
	pemDir := viper.GetString("lifecyclemanager.builtinca.datadir")

	if serverURL == "" {
		return nil, errors.Errorf("fail to get the built-in CA config: missing server name")
	}
	if password == "" {
		return nil, errors.Errorf("fail to get the built-in CA config: missing provisioner password")
	}
	if provisionerName == "" {
		provisionerName = "admin"
	}
	pemPath := filepath.Join(pemDir, "certs/root_ca.crt")
	pem, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get the root certificate of built-in CA.")
	}
	config := &entity.CertificateAuthorityConfigurationStepCA{
		ServiceURL:            serverURL,
		ServiceCertificatePEM: string(pem),
		ProvisionerName:       provisionerName,
		ProvisionerPassword:   password,
	}
	return config, nil
}
