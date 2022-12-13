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
	"context"
	"crypto/rsa"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/pkg/kubefate"
	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/repo"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// This means the CA should support issuing certificate valid for at least 1 year
const defaultCertLifetime = time.Hour * 24 * 365

// ParticipantService provides functions to manage participants
type ParticipantService struct {
	FederationRepo     repo.FederationRepository
	ChartRepo          repo.ChartRepository
	EventService       EventServiceInt
	CertificateService ParticipantCertificateServiceInt
	EndpointService    ParticipantEndpointServiceInt
}

func (s *ParticipantService) buildKubeFATEMgrAndClient(endpointUUID string) (kubefate.ClientManager, kubefate.Client, func(), error) {
	var closer func()
	endpointMgr, err := s.EndpointService.buildKubeFATEClientManagerFromEndpointUUID(endpointUUID)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "failed to get endpoint manager")
	}
	kfClient, err := endpointMgr.BuildClient()
	if err != nil {
		kfClient, closer, err = endpointMgr.BuildPFClient()
		if err != nil {
			return nil, nil, closer, errors.Wrapf(err, "failed to get kubefate client")
		}
	}
	return endpointMgr, kfClient, closer, nil
}

// ParticipantDeploymentBaseInfo contains basic deployment information for a participant
type ParticipantDeploymentBaseInfo struct {
	Description    string `json:"description"`
	EndpointUUID   string `json:"endpoint_uuid"`
	DeploymentYAML string `json:"deployment_yaml"`
}

// For mocking purpose
var (
	getServiceAccessWithFallback = func(client kubernetes.Client, namespace, serviceName, portName string, lbFallbackToNodePort bool) (serviceType corev1.ServiceType, host string, port int, err error) {
		log.Info().Msgf("retrieving address for service: %s, port: %s in namespace: %s", serviceName, portName, namespace)
		service, err := client.GetClientSet().CoreV1().Services(namespace).Get(context.TODO(), serviceName, v1.GetOptions{})
		if err != nil {
			return
		}

		serviceYAML, _ := yaml.Marshal(service)
		log.Debug().Msgf("service yaml: %s", serviceYAML)

		serviceType = service.Spec.Type
		host = service.Spec.ClusterIP
		port = 0
		nodePort := 0
		for _, p := range service.Spec.Ports {
			if p.Name == portName {
				port = int(p.Port)
				nodePort = int(p.NodePort)
			}
		}

		getNodePortServiceAccess := func() {
			nl, err := client.GetClientSet().CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
			if err != nil {
				// TODO: use user specified host
				clientConfig, _ := client.GetConfig()
				u, _ := url.Parse(clientConfig.Host)
				host, _, _ = net.SplitHostPort(u.Host)
			} else {
				node := nl.Items[0]
				for _, addr := range node.Status.Addresses {
					if addr.Type == corev1.NodeExternalIP {
						host = addr.Address
						break
					} else if addr.Type == corev1.NodeInternalIP {
						host = addr.Address
					}
				}
			}

			port = nodePort
		}

		switch serviceType {
		case corev1.ServiceTypeNodePort:
			getNodePortServiceAccess()
		case corev1.ServiceTypeLoadBalancer:
			retry := 5
			for {
				service, err := client.GetClientSet().CoreV1().Services(namespace).Get(context.TODO(), serviceName, v1.GetOptions{})
				serviceYAML, _ := yaml.Marshal(service)
				log.Debug().Msgf("service yaml: %s", serviceYAML)
				if err != nil || len(service.Status.LoadBalancer.Ingress) == 0 {
					retry--
					if retry > 0 {
						log.Warn().Msgf("failed to get LoadBalancer address, retrying (%v remaining)", retry)
						time.Sleep(time.Second * 20)
						continue
					} else {
						break
					}
				}
				host = service.Status.LoadBalancer.Ingress[0].IP
				if host == "" {
					host = service.Status.LoadBalancer.Ingress[0].Hostname
				}
				break
			}
			if retry == 0 {
				if lbFallbackToNodePort {
					log.Info().Msg("fallback to acquiring the service address as type NodePort")
					serviceType = corev1.ServiceTypeNodePort
					getNodePortServiceAccess()
				} else {
					err = errors.New("failed to get load balancer address in time")
					return
				}
			}
		}
		log.Info().Msgf("%s(%s) type: %v, host: %s, port: %v", serviceName, portName, serviceType, host, port)
		if port == 0 || host == "" {
			err = errors.Wrapf(err, "failed to get port number or host address")
		}
		return
	}

	ensureNSExisting = func(client kubernetes.Client, namespace string) (bool, error) {
		created := false
		_, err := client.GetClientSet().CoreV1().Namespaces().Get(context.TODO(), namespace, v1.GetOptions{})
		if err != nil {
			if apierr.IsNotFound(err) {
				log.Info().Msgf("creating namespace %s", namespace)
				_, err := client.GetClientSet().CoreV1().Namespaces().Create(context.TODO(), &corev1.Namespace{
					ObjectMeta: v1.ObjectMeta{
						Name: namespace,
					},
				}, v1.CreateOptions{})
				if err != nil {
					return false, errors.Wrapf(err, "failed to create namespace %s", namespace)
				}
				created = true
			} else {
				return created, errors.Wrapf(err, "failed to query namespace %s", namespace)
			}
		}
		return created, nil
	}

	getIngressInfo = func(client kubernetes.Client, name, namespace string) (*entity.ParticipantFATEIngress, error) {
		ingress, err := client.GetClientSet().NetworkingV1().Ingresses(namespace).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "failed to get ingress info")
		}
		if len(ingress.Spec.Rules) == 0 {
			return nil, errors.New("ingress is not available")
		}
		ingressInfo := &entity.ParticipantFATEIngress{
			Hosts:     nil,
			Addresses: nil,
			TLS:       ingress.Spec.TLS != nil,
		}
		for _, rule := range ingress.Spec.Rules {
			ingressInfo.Hosts = append(ingressInfo.Hosts, rule.Host)
		}
		for _, lbIngress := range ingress.Status.LoadBalancer.Ingress {
			if lbIngress.Hostname != "" {
				ingressInfo.Addresses = append(ingressInfo.Addresses, lbIngress.Hostname)
			}
			if lbIngress.IP != "" {
				ingressInfo.Addresses = append(ingressInfo.Addresses, lbIngress.IP)
			}
		}
		return ingressInfo, nil
	}

	deletePodWithPrefix = func(client kubernetes.Client, namespace, podPrefix string) error {
		var err error
		pods, _ := client.GetClientSet().CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
		for _, pod := range pods.Items {
			if strings.HasPrefix(pod.Name, podPrefix) {
				if deleteErr := client.GetClientSet().CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, v1.DeleteOptions{}); deleteErr != nil {
					log.Err(deleteErr).Msgf("failed to delete pod")
					err = deleteErr
				}
			}
		}
		return err
	}

	createSecret = func(client kubernetes.Client, namespace string, secret *corev1.Secret) error {
		_, err := client.GetClientSet().CoreV1().Secrets(namespace).Get(context.TODO(), secret.Name, v1.GetOptions{})
		if err == nil {
			log.Warn().Msgf("deleting stale secret with name %s in namespace %s", secret.Name, namespace)
			err := client.GetClientSet().CoreV1().Secrets(namespace).Delete(context.TODO(), secret.Name, v1.DeleteOptions{})
			if err != nil {
				return errors.Wrapf(err, "failed to delete stale secret")
			}
		} else if !apierr.IsNotFound(err) {
			return errors.Wrapf(err, "failed to check secret existence")
		}
		_, err = client.GetClientSet().CoreV1().Secrets(namespace).Create(context.TODO(), secret, v1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to create secret")
		}
		return nil
	}

	createConfigMap = func(client kubernetes.Client, namespace string, cm *corev1.ConfigMap) error {
		_, err := client.GetClientSet().CoreV1().ConfigMaps(namespace).Get(context.TODO(), cm.Name, v1.GetOptions{})
		if err == nil {
			log.Warn().Msgf("deleting stale configmap with name %s in namespace %s", cm.Name, namespace)
			err := client.GetClientSet().CoreV1().ConfigMaps(namespace).Delete(context.TODO(), cm.Name, v1.DeleteOptions{})
			if err != nil {
				return errors.Wrapf(err, "failed to delete stale configmap")
			}
		} else if !apierr.IsNotFound(err) {
			return errors.Wrapf(err, "failed to check configmap existence")
		}
		_, err = client.GetClientSet().CoreV1().ConfigMaps(namespace).Create(context.TODO(), cm, v1.CreateOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to create configmap")
		}
		return nil
	}

	createRegistrySecret = func(client kubernetes.Client, secretName string, namespace string, registrySecretConfig valueobject.KubeRegistrySecretConfig) error {
		secret, err := registrySecretConfig.BuildKubeSecret(secretName, namespace)
		if err != nil {
			return err
		}
		return createSecret(client, namespace, secret)
	}
)

func getServiceAccess(client kubernetes.Client, namespace, serviceName, portName string) (corev1.ServiceType, string, int, error) {
	return getServiceAccessWithFallback(client, namespace, serviceName, portName, false)
}

func toDeploymentName(name string) string {
	// replace non-alphanumeric character with space
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9]`)
	name = re.ReplaceAllString(name, " ")

	// trim spaces and replace with dash
	name = strings.TrimSpace(name)
	re = regexp.MustCompile(`\s+`)
	name = re.ReplaceAllString(name, " ")
	name = strings.ReplaceAll(name, " ", "-")

	return name
}

// ParticipantCertificateServiceInt declares the methods of a certificate service that this service needs.
// "Caller defines interfaces"
type ParticipantCertificateServiceInt interface {
	DefaultCA() (*entity.CertificateAuthority, error)
	CreateCertificateSimple(commonName string, lifetime time.Duration, dnsNames []string) (cert *entity.Certificate, pk *rsa.PrivateKey, err error)
	CreateBinding(cert *entity.Certificate, serviceType entity.CertificateBindingServiceType, participantUUID string, federationUUID string, federationType entity.FederationType) error
	RemoveBinding(participantUUID string) error
}

// ParticipantEndpointServiceInt declares the methods of an endpoint service that participant service needs
type ParticipantEndpointServiceInt interface {
	TestKubeFATE(uuid string) error

	buildKubeFATEClientManagerFromEndpointUUID(uuid string) (kubefate.ClientManager, error)
	ensureEndpointExist(infraUUID string, namespace string, registryConfig valueobject.KubeRegistryConfig) (string, error)
}
