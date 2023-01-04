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
	"time"

	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
)

// InfraProviderApp provide functions to manage the infra providers
type InfraProviderApp struct {
	InfraProviderKubernetesRepo repo.InfraProviderRepository
	EndpointKubeFATERepo        repo.EndpointRepository
}

// InfraProviderEditableItem contains properties of a provider that should be provided by the user
type InfraProviderEditableItem struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Type        entity.InfraProviderType `json:"type"`
}

// InfraProviderItemBase is the base data structure item
type InfraProviderItemBase struct {
	InfraProviderEditableItem
	UUID      string    `json:"uuid"`
	CreatedAt time.Time `json:"created_at"`
}

// InfraProviderListItem is an item for listing the providers
type InfraProviderListItem struct {
	InfraProviderItemBase
	KubernetesProviderInfo InfraProviderListItemKubernetes `json:"kubernetes_provider_info"`
}

// InfraProviderListItemKubernetes contains list info related to a kubernetes infra provider
type InfraProviderListItemKubernetes struct {
	APIServer string `json:"api_server"`
}

// InfraProviderKubernetesConfig contains kubernetes provider details properties
type InfraProviderKubernetesConfig struct {
	KubeConfig         string                         `json:"kubeconfig_content"`
	Namespaces         []string                       `json:"namespaces_list"`
	IsInCluster        bool                           `json:"is_in_cluster"`
	RegistryConfigFATE valueobject.KubeRegistryConfig `json:"registry_config_fate"`
}

// InfraProviderCreationRequest represents a request to create an infra provider
type InfraProviderCreationRequest struct {
	InfraProviderEditableItem
	KubernetesProviderInfo InfraProviderKubernetesConfig `json:"kubernetes_provider_info"`
}

// InfraProviderUpdateRequest represents a request to update an infra provider
type InfraProviderUpdateRequest struct {
	InfraProviderEditableItem
	KubernetesProviderInfo InfraProviderKubernetesConfig `json:"kubernetes_provider_info"`
}

// InfraProviderInfoKubernetes contains information specific to kubernetes provider
type InfraProviderInfoKubernetes struct {
	InfraProviderListItemKubernetes
	InfraProviderKubernetesConfig
}

// InfraProviderDetail contains details info of a provider
type InfraProviderDetail struct {
	InfraProviderItemBase
	KubernetesProviderInfo InfraProviderInfoKubernetes `json:"kubernetes_provider_info"`
}

// GetProviderList returns provider list
func (app *InfraProviderApp) GetProviderList() ([]InfraProviderListItem, error) {

	var providerList []InfraProviderListItem

	domainProviderListInstance, err := app.InfraProviderKubernetesRepo.List()
	if err != nil {
		return nil, err
	}
	domainProviderList := domainProviderListInstance.([]entity.InfraProviderKubernetes)
	for _, p := range domainProviderList {
		providerList = append(providerList, InfraProviderListItem{
			InfraProviderItemBase: InfraProviderItemBase{
				InfraProviderEditableItem: InfraProviderEditableItem{
					Name:        p.Name,
					Description: p.Description,
					Type:        p.Type,
				},
				UUID:      p.UUID,
				CreatedAt: p.CreatedAt,
			},
			KubernetesProviderInfo: InfraProviderListItemKubernetes{
				APIServer: p.APIHost,
			},
		})
	}
	return providerList, nil
}

// TestKubernetesConnection validates the connection to the kubernetes cluster
func (app *InfraProviderApp) TestKubernetesConnection(kubeconfig *valueobject.KubeConfig) error {
	return kubeconfig.Validate()
}

// CreateProvider creates a provider
func (app *InfraProviderApp) CreateProvider(providerInfo *InfraProviderCreationRequest) error {
	switch providerInfo.Type {
	case entity.InfraProviderTypeK8s:
		provider := &entity.InfraProviderKubernetes{
			InfraProviderBase: entity.InfraProviderBase{
				Name:        providerInfo.Name,
				Description: providerInfo.Description,
				Type:        providerInfo.Type,
			},
			Config: valueobject.KubeConfig{
				KubeConfigContent: providerInfo.KubernetesProviderInfo.KubeConfig,
				IsInCluster:       providerInfo.KubernetesProviderInfo.IsInCluster,
				NamespacesList:    providerInfo.KubernetesProviderInfo.Namespaces,
			},
			RegistryConfigFATE: providerInfo.KubernetesProviderInfo.RegistryConfigFATE,
			Repo:               app.InfraProviderKubernetesRepo,
		}
		return provider.Create()
	}
	return errors.Errorf("unknown provider type: %s", providerInfo.Type)
}

// DeleteProvider deletes a provider
func (app *InfraProviderApp) DeleteProvider(uuid string) error {
	// TODO：this should be in the domain layer
	endpointListInstance, err := app.EndpointKubeFATERepo.ListByInfraProviderUUID(uuid)
	if err != nil {
		return errors.Wrapf(err, "failed to query current infra's KubeFATE endpoint info")
	}
	domainEndpointList := endpointListInstance.([]entity.EndpointKubeFATE)
	if len(domainEndpointList) != 0 {
		return errors.Errorf("current infra provider %s still contains a KubeFATE endpoint", uuid)
	}
	return app.InfraProviderKubernetesRepo.DeleteByUUID(uuid)
}

// GetProviderDetail returns detailed info of a provider
func (app *InfraProviderApp) GetProviderDetail(uuid string) (*InfraProviderDetail, error) {
	domainProviderInstance, err := app.InfraProviderKubernetesRepo.GetByUUID(uuid)
	if err != nil {
		return nil, err
	}
	domainProvider := domainProviderInstance.(*entity.InfraProviderKubernetes)

	if err != nil {
		return nil, err
	}
	return &InfraProviderDetail{
		InfraProviderItemBase: InfraProviderItemBase{
			InfraProviderEditableItem: InfraProviderEditableItem{
				Name:        domainProvider.Name,
				Description: domainProvider.Description,
				Type:        domainProvider.Type,
			},
			UUID:      domainProvider.UUID,
			CreatedAt: domainProvider.CreatedAt,
		},
		KubernetesProviderInfo: InfraProviderInfoKubernetes{
			InfraProviderListItemKubernetes: InfraProviderListItemKubernetes{
				APIServer: domainProvider.APIHost,
			},
			InfraProviderKubernetesConfig: InfraProviderKubernetesConfig{
				KubeConfig:         domainProvider.Config.KubeConfigContent,
				Namespaces:         domainProvider.Config.NamespacesList,
				IsInCluster:        domainProvider.Config.IsInCluster,
				RegistryConfigFATE: domainProvider.RegistryConfigFATE,
			},
		},
	}, nil
}

// UpdateProvider changes provider settings
func (app *InfraProviderApp) UpdateProvider(uuid string, updateProviderInfo *InfraProviderUpdateRequest) error {
	switch updateProviderInfo.Type {
	case entity.InfraProviderTypeK8s:
		provider := &entity.InfraProviderKubernetes{
			InfraProviderBase: entity.InfraProviderBase{
				UUID:        uuid,
				Name:        updateProviderInfo.Name,
				Description: updateProviderInfo.Description,
				Type:        updateProviderInfo.Type,
			},
			Config: valueobject.KubeConfig{
				KubeConfigContent: updateProviderInfo.KubernetesProviderInfo.KubeConfig,
				NamespacesList:    updateProviderInfo.KubernetesProviderInfo.Namespaces,
				IsInCluster:       updateProviderInfo.KubernetesProviderInfo.IsInCluster,
			},
			RegistryConfigFATE: updateProviderInfo.KubernetesProviderInfo.RegistryConfigFATE,
			Repo:               app.InfraProviderKubernetesRepo,
		}
		return provider.Update()
	}
	return errors.Errorf("unknown provider type: %s", updateProviderInfo.Type)
}
