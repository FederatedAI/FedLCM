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

package kubernetes

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Client provides methods to interact with a Kubernetes API server
type Client interface {
	// GetClientSet returns the ClientSet for future use
	GetClientSet() kubernetes.Interface
	// GetConfig returns the *rest.Config for future use
	GetConfig() (*rest.Config, error)
	// ApplyOrDeleteYAML applies or delete the yaml content to/from the target cluster
	ApplyOrDeleteYAML(yamlStr string, delete bool) error
}

// client contains necessary info and object to work with a kubernetes cluster
type client struct {
	dynamicClient dynamic.Interface
	clientSet     kubernetes.Interface
	config        *rest.Config
}

// NewKubernetesClient returns a client struct based on the kubeconfig path
func NewKubernetesClient(kubeconfigPath string, kubeconfigContent string, inCluster bool) (Client, error) {
	var config *rest.Config
	var err error
	if kubeconfigPath != "" {
		log.Debug().Msg("build client with kubeconfigPath")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else if kubeconfigContent != "" {
		log.Debug().Msg("build client with kubeconfigContent")
		config, err = clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
	} else if inCluster {
		log.Debug().Msg("build client using inCluster config")
		config, err = rest.InClusterConfig()
	} else {
		err = errors.New("neither kubeconfigPath, kubeconfigContent nor inCluster specified")
	}
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &client{
		dynamicClient: dynamicClient,
		clientSet:     clientSet,
		config:        config,
	}, nil
}

func (c *client) GetClientSet() kubernetes.Interface {
	return c.clientSet
}

func (c *client) GetConfig() (*rest.Config, error) {
	return c.config, nil
}

func (c *client) ApplyOrDeleteYAML(yamlStr string, delete bool) error {
	groups, err := restmapper.GetAPIGroupResources(c.clientSet.Discovery())
	if err != nil {
		return err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groups)

	reader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader([]byte(yamlStr))))
	for {
		docBytes, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrapf(err, "failed to decode yaml content")
		}

		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{},
		}

		// Unmarshal the YAML document into the unstructured object.
		if err := yaml.Unmarshal(docBytes, &obj.Object); err != nil {
			return err
		}
		gvk := obj.GroupVersionKind()

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return err
		}
		resourceStr := fmt.Sprintf("%s(%s)", obj.GetName(), mapping.GroupVersionKind.String())
		if delete {
			log.Info().Msgf("Deleting %s with scope: %s", resourceStr, string(mapping.Scope.Name()))
		} else {
			log.Info().Msgf("Applying %s with scope: %s", resourceStr, string(mapping.Scope.Name()))
		}
		obj.SetManagedFields(nil)
		var dri dynamic.ResourceInterface
		if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
			if obj.GetNamespace() == "" {
				log.Debug().Msg("using default namespace")
				obj.SetNamespace("default")
			}
			dri = c.dynamicClient.Resource(mapping.Resource).Namespace(obj.GetNamespace())
		} else {
			dri = c.dynamicClient.Resource(mapping.Resource)
		}
		force := true
		if delete {
			err := dri.Delete(context.TODO(), obj.GetName(), v1.DeleteOptions{})
			if err != nil {
				if apierr.IsNotFound(err) {
					log.Info().Msgf("Resource %s not existing, continue", resourceStr)
				} else {
					return errors.Wrapf(err, "failed to delete resource %s", resourceStr)
				}
			} else {
				log.Info().Msgf("Done deleting %s with scope: %s", resourceStr, string(mapping.Scope.Name()))
			}
		} else {
			resp, err := dri.Patch(context.TODO(), obj.GetName(), types.ApplyPatchType, docBytes, v1.PatchOptions{
				Force:        &force,
				FieldManager: "fed-lifecycle-manager",
			})
			if err != nil {
				return err
			}
			log.Info().Msgf("Done applying %s with scope: %s, UID: %s", resourceStr, string(mapping.Scope.Name()), string(resp.GetUID()))
		}
	}
	return nil
}
