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

package valueobject

import (
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/rs/zerolog/log"
)

// KubeConfig contains necessary information needed to work with a kubernetes cluster
// Currently only the kubeconfig content is included
type KubeConfig struct {
	KubeConfigContent string `json:"kubeconfig_content"`
}

func (c KubeConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *KubeConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// Validate checks if the config can be used to connect to a K8s cluster
func (c *KubeConfig) Validate() error {
	client, err := kubernetes.NewKubernetesClient("", c.KubeConfigContent)
	if err != nil {
		return err
	}
	versionInfo, err := client.GetClientSet().Discovery().ServerVersion()
	if err != nil {
		return err
	}
	log.Info().Msgf("got k8s server version: %s", versionInfo.String())
	// TODO: check other conditions like permissions, ingress installation status etc.
	return nil
}

// APIHost returns the address for the API server connection
func (c *KubeConfig) APIHost() (string, error) {
	client, err := kubernetes.NewKubernetesClient("", c.KubeConfigContent)
	if err != nil {
		return "", err
	}
	config, err := client.GetConfig()
	if err != nil {
		return "", err
	}
	return config.Host, nil
}

// SHA2565 hashes the kubeconfig content and returns the hash string
func (c *KubeConfig) SHA2565() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(c.KubeConfigContent)))
}
