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

package kubefate

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/FederatedAI/FedLCM/pkg/kubernetes"
	"github.com/FederatedAI/FedLCM/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apiErr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util/podutils"
	"sigs.k8s.io/yaml"
)

// ClientManager provides methods to work with KubeFATE as a client
type ClientManager interface {
	// K8sClient returns the embedded K8s client
	K8sClient() kubernetes.Client
	// BuildClient retrieve the KubeFATE access info from the underlying K8s cluster and return a client instance
	BuildClient() (Client, error)
	// BuildPFClient build a KubeFATE client based on internal port-forwarding routine, caller must call the returned closerFn when no longer needs it
	BuildPFClient() (client Client, closerFn func(), err error)
}

// Manager extends ClientManager and provides methods to work with KubeFATE installation
type Manager interface {
	ClientManager
	// Install installs a KubeFATE deployment
	Install(bool) error
	// Uninstall uninstalls a KubeFATE installation
	Uninstall() error
	// InstallIngressNginxController installs a default ingress nginx controller
	InstallIngressNginxController() error
	// GetKubeFATEDeployment returns the Deployment object of KubeFATE
	GetKubeFATEDeployment() (*appv1.Deployment, error)
}

type manager struct {
	client kubernetes.Client
	meta   *InstallationMeta
}

// NewManager returns the KubeFATE manager
func NewManager(client kubernetes.Client, meta *InstallationMeta) Manager {
	return &manager{
		client: client,
		meta:   meta,
	}
}

// NewClientManager returns the KubeFATE client manager
func NewClientManager(client kubernetes.Client, meta *InstallationMeta) ClientManager {
	return &manager{
		client: client,
		meta:   meta,
	}
}

func (manager *manager) K8sClient() kubernetes.Client {
	return manager.client
}

func (manager *manager) Install(checkIngress bool) error {
	kubefateDeployment, err := manager.GetKubeFATEDeployment()
	if err == nil {
		return errors.New("kubefate is already installed")
	}
	if err := manager.client.ApplyOrDeleteYAML(manager.meta.yaml, false); err != nil {
		return errors.Wrapf(err, "failed to install kubefate, yaml: %s", manager.meta.yaml)
	}

	if err := utils.ExecuteWithTimeout(func() bool {
		log.Info().Msgf("checking kubefate deployment readiness...")
		kubefateDeployment, err = manager.GetKubeFATEDeployment()
		if err != nil {
			log.Err(err).Msgf("kubefate deployment installation error")
			return false
		}
		if kubefateDeployment.Status.ReadyReplicas > 0 {
			return true
		}
		log.Info().Msgf("kubefate installation not ready yet, status: %s", kubefateDeployment.Status.String())
		return false
	}, time.Minute*30, time.Second*10); err != nil {
		return errors.Wrapf(err, "error checking kubefate deployment")
	}
	if checkIngress {
		if err := utils.ExecuteWithTimeout(func() bool {
			log.Info().Msgf("checking kubefate ingress readiness...")
			ingress, err := manager.client.GetClientSet().NetworkingV1().Ingresses(manager.meta.namespace).Get(context.TODO(), manager.meta.kubefateIngressName, metav1.GetOptions{})
			if err != nil {
				log.Err(err).Msgf("kubefate ingress installation error")
				return false
			}
			if len(ingress.Status.LoadBalancer.Ingress) > 0 {
				return true
			}
			log.Info().Msgf("kubefate ingress installation not ready yet, status: %s", ingress.Status.String())
			return false
		}, time.Minute*30, time.Second*10); err != nil {
			return errors.Wrapf(err, "error checking kubefate ingress deployment")
		}
	}

	if err := utils.ExecuteWithTimeout(func() bool {
		log.Info().Msgf("verifying kubefate version...")
		kfc, err := manager.BuildClient()
		if err != nil {
			var closer func()
			kfc, closer, err = manager.BuildPFClient()
			if closer != nil {
				defer closer()
			}
			if err != nil {
				log.Err(err).Msg("error getting kubefate client instance")
				return false
			}
		}
		if _, err := kfc.CheckVersion(); err != nil {
			log.Err(err).Msg("error checking kubefate version")
			return false
		}
		return true
	}, time.Minute*20, time.Second*10); err != nil {
		return errors.Wrapf(err, "error checking kubefate service readiness")
	}
	log.Info().Msg("kubefate is installed and ready")
	return nil
}

func (manager *manager) Uninstall() error {
	clientSet := manager.client.GetClientSet()
	if err := clientSet.RbacV1().ClusterRoleBindings().Delete(context.TODO(), manager.meta.clusterRoleBindingName, metav1.DeleteOptions{}); err != nil && !apiErr.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete clusterrolebinding")
	}
	log.Info().Msgf("clusterrolebinding %s deleted", manager.meta.clusterRoleBindingName)

	if err := clientSet.RbacV1().ClusterRoles().Delete(context.TODO(), manager.meta.clusterRoleName, metav1.DeleteOptions{}); err != nil && !apiErr.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete clusterrole")
	}
	log.Info().Msgf("clusterrole %s deleted", manager.meta.clusterRoleName)

	if err := clientSet.PolicyV1beta1().PodSecurityPolicies().Delete(context.TODO(), manager.meta.pspName, metav1.DeleteOptions{}); err != nil && !apiErr.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete psp")
	}
	log.Info().Msgf("psp %s deleted", manager.meta.pspName)

	// TODO: explicitly delete other namespaced resources

	if err := clientSet.CoreV1().Namespaces().Delete(context.TODO(), manager.meta.namespace, metav1.DeleteOptions{}); err != nil && !apiErr.IsNotFound(err) {
		return errors.Wrapf(err, "failed to delete namespace")
	}
	log.Info().Msgf("namespace %s deleted", manager.meta.namespace)

	if err := utils.ExecuteWithTimeout(func() bool {
		log.Info().Msgf("checking namespace %s removing result...", manager.meta.namespace)
		ns, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), manager.meta.namespace, metav1.GetOptions{})
		if err != nil {
			if apiErr.IsNotFound(err) {
				return true
			}
			log.Warn().Err(err).Msgf("error getting namespace status")
			return false
		}
		if ns != nil {
			log.Info().Msgf("ns %s not deleted yet, status: %s", ns.Name, ns.Status.String())
		}
		return false
	}, time.Minute*30, time.Second*10); err != nil {
		return errors.Wrapf(err, "error deleting namespace %s", manager.meta.namespace)
	}
	log.Info().Msg("kubefate deleted")
	return nil
}

func (manager *manager) BuildClient() (Client, error) {
	kubefateDeployment, err := manager.GetKubeFATEDeployment()
	if err != nil {
		return nil, err
	}
	if kubefateDeployment.Status.ReadyReplicas == 0 {
		return nil, errors.New("kubefate deployment is not ready")
	}
	envList := kubefateDeployment.Spec.Template.Spec.Containers[0].Env
	client := &client{
		apiVersion: "v1",
	}
	if client.username = manager.getValueFromEnvs(envList, "FATECLOUD_USER_USERNAME"); client.username == "" {
		log.Warn().Msgf("cannot get kubefate username for installation %v, using default", manager.meta)
		client.username = "admin"
	}

	if client.password = manager.getValueFromEnvs(envList, "FATECLOUD_USER_PASSWORD"); client.password == "" {
		log.Warn().Msgf("cannot get kubefate password for installation %v, using default", manager.meta)
		client.password = "admin"
	}

	ingress, err := manager.client.GetClientSet().NetworkingV1().Ingresses(manager.meta.namespace).Get(context.TODO(), manager.meta.kubefateIngressName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ingress info")
	}
	log.Debug().Msgf("ingress spec: %v, status: %v", ingress.Spec, ingress.Status)
	if len(ingress.Spec.Rules) == 0 || len(ingress.Status.LoadBalancer.Ingress) == 0 {
		return nil, errors.New("kubefate ingress is not available")
	}

	client.ingressRuleHost = ingress.Spec.Rules[0].Host
	if client.ingressAddress = ingress.Status.LoadBalancer.Ingress[0].Hostname; client.ingressAddress == "" {
		client.ingressAddress = ingress.Status.LoadBalancer.Ingress[0].IP
	}
	if client.ingressAddress == "" {
		return nil, errors.New("cannot get ingress access info")
	}
	client.tls = ingress.Spec.TLS != nil

	// try to use the node port address instead
	if manager.meta.ingressControllerYAML != "" {
		service, err := manager.client.GetClientSet().CoreV1().Services(manager.meta.ingressControllerNamespace).Get(context.TODO(),
			manager.meta.ingressControllerServiceName, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get ingress controller service")
		}
		if service.Spec.Type == corev1.ServiceTypeNodePort {
			log.Info().Msg("using ingress service node port address")
			client.ingressAddress, err = manager.getIngressControllerServiceNodePortAddress("http")
		}
	}
	if _, err := client.CheckVersion(); err != nil {
		return nil, errors.Wrapf(err, "cannot verify the client connection")
	}
	return client, nil
}

func (manager *manager) BuildPFClient() (Client, func(), error) {
	kubefateDeployment, err := manager.GetKubeFATEDeployment()
	if err != nil {
		return nil, nil, err
	}
	if kubefateDeployment.Status.ReadyReplicas == 0 {
		return nil, nil, errors.New("kubefate deployment is not ready")
	}
	envList := kubefateDeployment.Spec.Template.Spec.Containers[0].Env
	client := &pfClient{
		client: client{
			apiVersion: "v1",
		},
	}
	if client.username = manager.getValueFromEnvs(envList, "FATECLOUD_USER_USERNAME"); client.username == "" {
		log.Warn().Msgf("cannot get kubefate username for installation %v, using default", manager.meta)
		client.username = "admin"
	}

	if client.password = manager.getValueFromEnvs(envList, "FATECLOUD_USER_PASSWORD"); client.password == "" {
		log.Warn().Msgf("cannot get kubefate password for installation %v, using default", manager.meta)
		client.password = "admin"
	}

	fw, stopChan, err := manager.setUpPortForwarder()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "fail to setup portt forwarder")
	}
	if err := utils.RetryWithMaxAttempts(func() error {
		ports, err := fw.GetPorts()
		if err != nil {
			return err
		}
		client.ingressAddress = fmt.Sprintf("localhost:%d", ports[0].Local)
		client.fw = fw
		client.stopChan = stopChan
		// we don't really care about the Host field when using port forwarder, but lets try our best to get the real one configured
		client.ingressRuleHost = "kubefate.net"
		ingress, err := manager.client.GetClientSet().NetworkingV1().Ingresses(manager.meta.namespace).Get(context.TODO(), manager.meta.kubefateIngressName, metav1.GetOptions{})
		if err == nil {
			log.Debug().Msgf("ingress spec: %v, status: %v", ingress.Spec, ingress.Status)
			if len(ingress.Spec.Rules) != 0 {
				client.ingressRuleHost = ingress.Spec.Rules[0].Host
			}
			log.Warn().Msgf("kubefate ingress spec contain empty rules")
		} else {
			log.Err(err).Msgf("failed to query kubefate ingress info, using default one")
		}
		return nil
	}, 10, 1*time.Second); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to setup port-forwarding for kubefate pod")
	}
	return client, client.Close, nil
}

func (manager *manager) InstallIngressNginxController() error {
	if manager.meta.ingressControllerYAML == "" {
		return errors.Errorf("no ingress controller yaml provided")
	}
	if err := manager.client.ApplyOrDeleteYAML(manager.meta.ingressControllerYAML, true); err != nil {
		return errors.Wrapf(err, "failed to clean-up ingress controller, yaml: %s", manager.meta.ingressControllerYAML)
	}

	if err := utils.ExecuteWithTimeout(func() bool {
		if err := manager.client.ApplyOrDeleteYAML(manager.meta.ingressControllerYAML, false); err != nil {
			log.Err(err).Msgf("failed to install ingress controller, yaml: %s", manager.meta.ingressControllerYAML)
			return false
		}
		return true
	}, time.Minute*30, time.Second*20); err != nil {
		return errors.Wrapf(err, "error installing ingress controller")
	}

	if err := utils.ExecuteWithTimeout(func() bool {
		log.Info().Msgf("checking ingress controller deployment readiness...")
		controllerDeployment, err := manager.GetIngressNginxControllerDeployment()
		if err != nil {
			log.Err(err).Msgf("ingress controller deployment installation error")
			return false
		}
		if controllerDeployment.Status.ReadyReplicas > 0 {
			return true
		}
		log.Info().Msgf("ingress controller installation not ready yet, status: %s", controllerDeployment.Status.String())
		return false
	}, time.Minute*30, time.Second*10); err != nil {
		return errors.Wrapf(err, "error checking ingress controller deployment")
	}
	return nil
}

func (manager *manager) getValueFromEnvs(envList []v1.EnvVar, key string) string {
	for _, env := range envList {
		if env.Name != key {
			continue
		}
		if env.Value != "" {
			return env.Value
		}
		if env.ValueFrom != nil {
			value, err := manager.getSecretRefValue(env.ValueFrom)
			if err != nil {
				log.Err(err).Msgf("failed to gett secret value for ")
				return ""
			}
			return value
		}
		return ""
	}
	return ""
}

func (manager *manager) getSecretRefValue(source *v1.EnvVarSource) (string, error) {
	secret, err := manager.client.GetClientSet().CoreV1().Secrets(manager.meta.namespace).Get(context.TODO(), source.SecretKeyRef.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if data, ok := secret.Data[source.SecretKeyRef.Key]; ok {
		return string(data), nil
	}
	return "", fmt.Errorf("key %s not found in secret %s", source.SecretKeyRef.Key, source.SecretKeyRef.Name)
}

func (manager *manager) GetKubeFATEDeployment() (*appv1.Deployment, error) {
	return manager.client.GetClientSet().AppsV1().Deployments(manager.meta.namespace).Get(context.TODO(), manager.meta.kubefateDeployName, metav1.GetOptions{})
}

func (manager *manager) GetIngressNginxControllerDeployment() (*appv1.Deployment, error) {
	return manager.client.GetClientSet().AppsV1().Deployments(manager.meta.ingressControllerNamespace).Get(context.TODO(), manager.meta.ingressControllerDeployName, metav1.GetOptions{})
}

func (manager *manager) getIngressControllerServiceNodePortAddress(portName string) (address string, err error) {
	host := ""
	log.Info().Msgf("retrieving address for service: %s, port: %s in namespace: %s", manager.meta.ingressControllerServiceName, portName, manager.meta.ingressControllerNamespace)
	service, err := manager.K8sClient().GetClientSet().CoreV1().Services(manager.meta.ingressControllerNamespace).
		Get(context.TODO(), manager.meta.ingressControllerServiceName, metav1.GetOptions{})
	if err != nil {
		return
	}

	serviceYAML, _ := yaml.Marshal(service)
	log.Debug().Msgf("service yaml: %s", serviceYAML)

	nodePort := 0
	for _, p := range service.Spec.Ports {
		if p.Name == portName {
			nodePort = int(p.NodePort)
		}
	}

	nl, _ := manager.K8sClient().GetClientSet().CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	node := nl.Items[0]
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			host = addr.Address
			break
		} else if addr.Type == corev1.NodeInternalIP {
			host = addr.Address
		}
	}
	if host == "" {
		err = errors.New("cannot find node address")
	}
	address = fmt.Sprintf("%s:%d", host, nodePort)
	log.Info().Msgf("ingress address: %v", address)
	return
}

func (manager *manager) setUpPortForwarder() (fw *portforward.PortForwarder, stopChan chan struct{}, err error) {
	deploy, err := manager.GetKubeFATEDeployment()
	if err != nil {
		return
	}

	config, err := manager.client.GetConfig()
	if err != nil {
		return
	}

	ns, selector, err := polymorphichelpers.SelectorsForObject(deploy)
	if err != nil {
		return
	}
	clientset, err := corev1client.NewForConfig(config)
	if err != nil {
		return
	}
	pod, _, err := polymorphichelpers.GetFirstPod(clientset, ns, selector.String(), time.Second*20, func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) })
	if err != nil {
		return
	}

	req := manager.client.GetClientSet().CoreV1().RESTClient().Post().Namespace(pod.GetNamespace()).
		Resource("pods").Name(pod.GetName()).SubResource("portforward")

	roundTripper, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: roundTripper}, http.MethodPost, req.URL())

	// 8080 is kubefate's listening port
	portMapping := []string{":8080"}
	stopChan = make(chan struct{}, 1)
	readyChannel := make(chan struct{})
	fw, err = portforward.New(dialer, portMapping, stopChan, readyChannel, os.Stdout, os.Stderr)
	if err != nil {
		return
	}

	go func() {
		err = fw.ForwardPorts()
		if err != nil {
			log.Err(err).Msgf("error forwarding ports for %v", fw)
		}
		log.Info().Msgf("shutdown port forwarder %v", fw)
	}()
	return
}
