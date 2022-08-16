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
	"context"
	"database/sql/driver"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/FederatedAI/FedLCM/site-portal/server/infrastructure/kubernetes"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// KubeflowConfig contains necessary information needed to deploy model to Kubeflow
type KubeflowConfig struct {
	// MinIOEndpoint is the address for the MinIO server
	MinIOEndpoint string `json:"minio_endpoint"`
	// MinIOAccessKey is the access-key for the MinIO server
	MinIOAccessKey string `json:"minio_access_key"`
	// MinIOSecretKey is the secret-key for the MinIO server
	MinIOSecretKey string `json:"minio_secret_key"`
	// MinIOSSLEnabled is whether this connection should be over ssl
	MinIOSSLEnabled bool `json:"minio_ssl_enabled"`
	// MinIORegion is the region of the MinIO service
	MinIORegion string `json:"minio_region"`
	// KubeConfig is the content of the kubeconfig file to connect to kubernetes
	KubeConfig string `json:"kubeconfig"`
}

func (c KubeflowConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *KubeflowConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

func (c *KubeflowConfig) getTempKubeconfigFile() (string, error) {
	if c.KubeConfig == "" {
		// return empty file path so the client will try the in-cluster config
		return "", nil
	}

	tmpFile, err := ioutil.TempFile(os.TempDir(), "kubeconfig-")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()
	text := []byte(c.KubeConfig)
	if _, err = tmpFile.Write(text); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

// Validate checks the connection to the kubernetes, the installation of KFServing and the connection to MinIO
func (c *KubeflowConfig) Validate() error {
	if err := func() error {
		kubeconfigFile, err := c.getTempKubeconfigFile()
		if err != nil {
			return err
		}
		client, err := kubernetes.NewKubernetesClient(kubeconfigFile)
		if err != nil {
			return err
		}
		list, err := client.GetInferenceServiceList()
		if err != nil {
			return err
		}
		log.Debug().Msgf("got isvc list %s", list)
		return nil
	}(); err != nil {
		return errors.Wrap(err, "failed to check KFServing installation")
	}
	if err := c.validateMinIO(); err != nil {
		return errors.Wrap(err, "failed to check MinIO connection")
	}
	return nil
}

func (c *KubeflowConfig) validateMinIO() error {
	client, err := minio.New(c.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.MinIOAccessKey, c.MinIOSecretKey, ""),
		Secure: c.MinIOSSLEnabled,
		Region: c.MinIORegion,
	})
	if err != nil {
		return err
	}
	// seems there is no "test connection" API, just test a dummy bucket and ignore the response
	_, err = client.BucketExists(context.TODO(), "check")
	if err != nil {
		return err
	}
	return nil
}
