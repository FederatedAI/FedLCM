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
	"database/sql/driver"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubeRegistryConfig contains registry configurations that can be used in K8s clusters
type KubeRegistryConfig struct {
	UseRegistry          bool                     `json:"use_registry" form:"use_registry" yaml:"useRegistry"`
	Registry             string                   `json:"registry" form:"registry" yaml:"registry"`
	UseRegistrySecret    bool                     `json:"use_registry_secret" form:"use_registry_secret" yaml:"useRegistrySecret"`
	RegistrySecretConfig KubeRegistrySecretConfig `json:"registry_secret_config" yaml:"registrySecretConfig"`
}

func (c KubeRegistryConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *KubeRegistryConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// KubeRegistrySecretConfig is the secret configuration that can be used to generate K8s imagePullSecret
type KubeRegistrySecretConfig struct {
	ServerURL string `json:"server_url" yaml:"serverURL"`
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
}

// BuildKubeSecret create a K8s secret containing the `.dockerconfigjson` for authenticating with remote registry
func (c KubeRegistrySecretConfig) BuildKubeSecret(name, namespace string) (*corev1.Secret, error) {
	// .dockerconfigjson auth is base64 encoded with format 'username:password'
	auth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.Username, c.Password)))
	secretObj := &dockerConfigAuth{
		AuthMap: map[string]registryCredentials{
			c.ServerURL: {
				Username: c.Username,
				Password: c.Password,
				Auth:     auth,
			},
		},
	}
	secretData, err := json.Marshal(secretObj)
	if err != nil {
		return nil, err
	}
	return &corev1.Secret{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: "kubernetes.io/dockerconfigjson",
		Data: map[string][]byte{
			".dockerconfigjson": secretData,
		},
	}, nil
}

// BuildSecretB64String generates the authentication string
func (c KubeRegistrySecretConfig) BuildSecretB64String() (string, error) {
	// .dockerconfigjson auth is base64 encoded with format 'username:password'
	auth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.Username, c.Password)))
	secretObj := &dockerConfigAuth{
		AuthMap: map[string]registryCredentials{
			c.ServerURL: {
				Username: c.Username,
				Password: c.Password,
				Auth:     auth,
			},
		},
	}
	secretData, err := json.Marshal(secretObj)
	if err != nil {
		return "", err
	}
	secretStr := b64.StdEncoding.EncodeToString([]byte(secretData))
	return secretStr, nil
}

// dockerConfigAuth is the dockerconfig containing only the auths filed
type dockerConfigAuth struct {
	AuthMap registryAuthMap `json:"auths"`
}

// registryAuthMap is a map of registries to their credentials
type registryAuthMap map[string]registryCredentials

// registryCredentials defines the fields stored per registry in the dockerconfig auths map
type registryCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
}
