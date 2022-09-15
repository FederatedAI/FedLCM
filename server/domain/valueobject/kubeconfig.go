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
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/rs/zerolog/log"
)

// KubeConfig contains necessary information needed to work with a kubernetes cluster
type KubeConfig struct {
	// KubeConfigContent stores the kubeconfig file of a K8s cluster
	KubeConfigContent string `json:"kubeconfig_content"`
	// NamespacesList stores namespaces the user in KubeConfigContent can access
	NamespacesList []string `json:"namespaces_list"`
}

func (c KubeConfig) Value() (driver.Value, error) {
	bJson, err := json.Marshal(c)
	return bJson, err
}

func (c *KubeConfig) Scan(v interface{}) error {
	return json.Unmarshal([]byte(v.(string)), c)
}

// Validate checks if the config can be used to connect to a K8s cluster and if the user has enough privilege
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

	if len(c.NamespacesList) != 0 {
		log.Info().Msgf("check admin privilege by getting rolebindings in namespaces %v", c.NamespacesList)
		for _, namespace := range c.NamespacesList {
			roleBindingList, err := client.GetClientSet().RbacV1().RoleBindings(namespace).List(context.Background(), metav1.ListOptions{})
			if err != nil {
				log.Error().Msgf("cannot get role bindings in namespace: %s", namespace)
				return errors.Wrapf(err, "not enough privileges in namespace: %s", namespace)
			}
			log.Info().Msgf("got %d role bindings in namespace: %s", len(roleBindingList.Items), namespace)
		}
	} else if clusterRoleBindingList, err := client.GetClientSet().RbacV1().ClusterRoleBindings().List(context.Background(), metav1.ListOptions{}); err != nil {
		log.Error().Msgf("no namespace was provided")
		return errors.New("expect namespaces")
	} else {
		log.Info().Msgf("get %d cluster role bindings success", len(clusterRoleBindingList.Items))
	}
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
