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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/FederatedAI/FedLCM/pkg/utils"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/modules"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/portforward"
	"sigs.k8s.io/yaml"
)

// Client can be used to access KubeFATE service
type Client interface {
	// CheckVersion returns the version of the KubeFATE service
	CheckVersion() (string, error)
	// EnsureChartExist makes sure the specified chart is downloaded and managed by kubefate
	EnsureChartExist(name, version string, content []byte) error
	// ListClusterByNamespace returns clusters list in the specified namespace
	ListClusterByNamespace(namespace string) ([]*modules.Cluster, error)

	SubmitClusterInstallationJob(yamlStr string) (string, error)
	SubmitClusterUpdateJob(yamlStr string) (string, error)
	SubmitClusterDeletionJob(clusterUUID string) (string, error)
	GetJobInfo(jobUUID string) (*modules.Job, error)
	WaitJob(jobUUID string) (*modules.Job, error)
	WaitClusterUUID(jobUUID string) (string, error)
	StopJob(jobUUID string) error

	IngressAddress() string
	IngressRuleHost() string
}

type client struct {
	apiVersion      string
	ingressRuleHost string
	ingressAddress  string
	tls             bool
	username        string
	password        string
}

type pfClient struct {
	client
	fw       *portforward.PortForwarder
	stopChan chan struct{}
}

// InstallationMeta describes the basic info of a KubeFATE installation
type InstallationMeta struct {
	namespace                    string
	isClusterAdmin               bool
	kubefateDeployName           string
	kubefateIngressName          string
	yaml                         string
	ingressControllerNamespace   string
	ingressControllerDeployName  string
	ingressControllerServiceName string
	ingressControllerYAML        string
}

// ClusterArgs is the args to manage a cluster
type ClusterArgs struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	ChartName    string `json:"chart_name"`
	ChartVersion string `json:"chart_version"`
	Cover        bool   `json:"cover"`
	Data         []byte `json:"data"`
}

type ClusterJobResponse struct {
	Data *modules.Job
	Msg  string
}

type ClusterInfoResponse struct {
	Data *modules.Cluster
	Msg  string
}

type ChartListResponse struct {
	Data []*modules.HelmChart
	Msg  string
}

type ClusterListResponse struct {
	Data []*modules.Cluster
	Msg  string
}

// BuildInstallationMetaFromYAML builds a meta description of a KubeFATE installation
func BuildInstallationMetaFromYAML(namespace, yamlStr, ingressControllerYAMLStr string) (*InstallationMeta, error) {
	// empty ingressControllerYAMLStr means we don't need to install ingress controller
	if yamlStr == "" {
		return nil, errors.New("No yaml provided")
	}

	meta := &InstallationMeta{
		namespace:                    namespace,
		isClusterAdmin:               false,
		kubefateDeployName:           "kubefate",
		kubefateIngressName:          "kubefate",
		yaml:                         yamlStr,
		ingressControllerNamespace:   "ingress-nginx",
		ingressControllerDeployName:  "ingress-nginx-controller",
		ingressControllerServiceName: "ingress-nginx-controller",
		ingressControllerYAML:        ingressControllerYAMLStr,
	}

	if namespace == "" {
		meta.namespace = "kube-fate"
		meta.isClusterAdmin = true
	}

	return meta, nil
}

func (c *client) IngressAddress() string {
	return c.ingressAddress
}

func (c *client) IngressRuleHost() string {
	return c.ingressRuleHost
}

func (c *client) CheckVersion() (string, error) {
	url := c.getUrl("version")
	body := bytes.NewReader(nil)
	log.Info().Msgf("request info: %s", url)
	request, err := http.NewRequest("GET", url, body)
	if err != nil {
		log.Err(err).Msgf("new request error")
		return "", err
	}

	token, err := c.getToken()
	if err != nil {
		return "", errors.Wrapf(err, "get token error")
	}
	authorization := fmt.Sprintf("Bearer %s", token)

	request.Header.Add("Authorization", authorization)
	request.Host = c.ingressRuleHost

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", errors.Wrapf(err, "http request error")
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "read resp body error")
	}
	log.Info().Msgf("request success, body: %s", string(respBody))

	type VersionResultMsg struct {
		Msg     string
		Version string
	}

	versionResult := new(VersionResultMsg)

	err = json.Unmarshal(respBody, &versionResult)
	if err != nil {
		return "", errors.Wrapf(err, "Unmarshal resp body error")
	}

	return versionResult.Version, nil
}

func (c *client) EnsureChartExist(name, version string, content []byte) error {
	resp, err := c.sendJSON("GET", "chart", nil)
	if err != nil {
		return err
	}
	body, err := c.parseResponse(resp)
	if err != nil {
		return err
	}
	var response ChartListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	for _, chart := range response.Data {
		if chart.Name == name && chart.Version == version {
			// just log but continue the uploading
			log.Info().Msgf("chart %s:%s already exists, override", name, version)
			break
		}
	}
	_ = resp.Body.Close()

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", fmt.Sprintf("%s-%s.tgz", name, version))
	if err != nil {
		return err
	}

	_, err = fileWriter.Write(content)
	if err != nil {
		return errors.Wrapf(err, "failed to write content")
	}

	contentType := bodyWriter.FormDataContentType()
	log.Debug().Str("contentType", contentType).Msg("contentType")
	_ = bodyWriter.Close()

	if err := utils.RetryWithMaxAttempts(func() error {
		token, err := c.getToken()
		if err != nil {
			return errors.Wrapf(err, "get token error")
		}
		authorization := fmt.Sprintf("Bearer %s", token)

		urlStr := c.getUrl("chart")
		payload := bodyBuf.Bytes()
		req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", contentType)
		req.Header.Add("Authorization", authorization)
		req.Host = c.ingressRuleHost
		log.Info().Msg(fmt.Sprintf("Uploading file to %s", urlStr))
		resp, err = http.DefaultClient.Do(req)
		return err
	}, 3, 10*time.Second); err != nil {
		return err
	}

	_, err = c.parseResponse(resp)
	return err
}

func (c *client) ListClusterByNamespace(namespace string) ([]*modules.Cluster, error) {
	resp, err := c.sendJSON("GET", "cluster", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	var response ClusterListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var res []*modules.Cluster
	for _, cluster := range response.Data {
		if cluster.NameSpace == namespace {
			res = append(res, cluster)
			break
		}
	}
	return res, err
}

func (c *client) getToken() (string, error) {

	login := map[string]string{
		"username": c.username,
		"password": c.password,
	}

	loginJsonB, err := json.Marshal(login)

	body := bytes.NewReader(loginJsonB)
	loginUrl := c.getUrl("user/login")

	request, err := http.NewRequest("POST", loginUrl, body)
	if err != nil {
		return "", err
	}
	request.Host = c.ingressRuleHost

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{}

	err = json.Unmarshal(rbody, &result)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(fmt.Sprint(result["message"]))
	}

	token := fmt.Sprint(result["token"])

	return token, nil
}

func (c *client) SubmitClusterInstallationJob(yamlStr string) (string, error) {
	body, err := c.buildClusterRequestBody(yamlStr)
	if err != nil {
		return "", err
	}
	resp, err := c.postJSON("cluster", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return c.getClusterJobUUID(resp)
}

func (c *client) SubmitClusterUpdateJob(yamlStr string) (string, error) {
	body, err := c.buildClusterRequestBody(yamlStr)
	if err != nil {
		return "", err
	}
	resp, err := c.putJSON("cluster", body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return c.getClusterJobUUID(resp)
}

func (c *client) SubmitClusterDeletionJob(clusterUUID string) (string, error) {
	_, err := c.getClusterInfo(clusterUUID)
	if err != nil {
		if strings.ContainsAny(err.Error(), "record not found") {
			log.Info().Msg("cluster record not found, skip deletion")
			return "", nil
		}
		log.Error().Err(err).Msg("get cluster info error")
		return "", err
	}
	resp, err := c.sendJSON("DELETE", "cluster/"+clusterUUID, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	jobUUID, err := c.getClusterJobUUID(resp)
	if err != nil {
		if strings.ContainsAny(err.Error(), "record not found") {
			log.Info().Msg("cluster record not found error, ignore")
			return "", nil
		}
	}
	return jobUUID, err
}

func (c *client) getClusterInfo(clusterUUID string) (*modules.Cluster, error) {
	resp, err := c.sendJSON("GET", "cluster/"+clusterUUID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	var response ClusterInfoResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response.Data, err
}

func (c *client) GetJobInfo(jobUUID string) (*modules.Job, error) {
	resp, err := c.sendJSON("GET", "job/"+jobUUID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	var response ClusterJobResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response.Data, err
}

func (c *client) WaitJob(jobUUID string) (*modules.Job, error) {
	for {
		job, err := c.GetJobInfo(jobUUID)
		if err != nil {
			return nil, err
		}
		switch job.Status {
		case modules.JobStatusRunning, modules.JobStatusPending:
			time.Sleep(time.Second * 5)
			continue
		default:
			log.Info().Msgf("job (%s) status: %v", job.Uuid, job.Status)
			return job, nil
		}
	}
}

func (c *client) StopJob(jobUUID string) error {
	job, err := c.GetJobInfo(jobUUID)
	if err != nil {
		return err
	}
	if job.Status == modules.JobStatusRunning {
		resp, err := c.sendJSON("PUT", "job/"+jobUUID+"?jobStatus=stop", nil)
		if err != nil {
			return errors.Wrapf(err, "failed to stop the running job: %s", jobUUID)
		}
		defer resp.Body.Close()
		log.Info().Msg("Stop Job Success")
	}
	return nil
}

func (c *client) WaitClusterUUID(jobUUID string) (string, error) {
	for {
		job, err := c.GetJobInfo(jobUUID)
		if err != nil {
			return "", err
		}
		if job.ClusterId != "" {
			return job.ClusterId, nil
		}
		switch job.Status {
		case modules.JobStatusFailed, modules.JobStatusTimeout, modules.JobStatusCanceled, modules.JobStatusStopping:
			return "", errors.Errorf("failed to get ClusterUUID: job status: %s, job info: %v", job.Status.String(), job)
		default:
			log.Info().Msgf("waiting for ClusterUUID of job (%s) status: %v", job.Uuid, job.Status)
			time.Sleep(time.Second * 5)
			continue
		}
	}
}

func (c *client) getClusterJobUUID(resp *http.Response) (string, error) {
	bodyBytes, err := c.parseResponse(resp)
	if err != nil {
		return "", err
	}
	var response ClusterJobResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", err
	}
	return response.Data.Uuid, err
}

func (c *client) buildClusterRequestBody(yamlStr string) (string, error) {
	var m map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &m)
	if err != nil {
		return "", err
	}

	name, ok := m["name"]
	if !ok {
		return "", errors.New("name not found")
	}

	namespace, ok := m["namespace"]
	if !ok {
		return "", errors.New("namespace not found")
	}

	chartVersion, ok := m["chartVersion"]
	if !ok {
		return "", errors.New("chartVersion not found")
	}

	chartName, ok := m["chartName"]
	if !ok {
		chartName = ""
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	valBJ, err := json.Marshal(m)
	log.Info().Msg(fmt.Sprintf("Cluster operation request valBJ: %s", string(valBJ)))
	if err != nil {
		return "", err
	}

	args := ClusterArgs{
		Name:         name.(string),
		Namespace:    namespace.(string),
		ChartName:    chartName.(string),
		ChartVersion: chartVersion.(string),
		Cover:        true,
		Data:         valBJ,
	}

	body, err := json.Marshal(args)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *client) getUrl(path string) string {
	address := c.ingressAddress
	apiVersion := c.apiVersion
	schemaStr := "http"
	if c.tls {
		schemaStr = "https"
	}
	return schemaStr + "://" + address + "/" + apiVersion + "/" + path
}

func (c *client) sendJSON(method, path string, body interface{}) (*http.Response, error) {
	var resp *http.Response
	if err := utils.RetryWithMaxAttempts(func() error {
		token, err := c.getToken()
		if err != nil {
			return errors.Wrapf(err, "get token error")
		}
		authorization := fmt.Sprintf("Bearer %s", token)

		urlStr := c.getUrl(path)
		var payload []byte
		if stringBody, ok := body.(string); ok {
			payload = []byte(stringBody)
		} else {
			var err error
			if payload, err = json.Marshal(body); err != nil {
				return err
			}
		}
		req, err := http.NewRequest(method, urlStr, bytes.NewBuffer(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", authorization)
		req.Host = c.ingressRuleHost
		log.Info().Msg(fmt.Sprintf("%s request to %s with body %s", method, urlStr, string(payload)))
		resp, err = http.DefaultClient.Do(req)
		return err
	}, 3, 10*time.Second); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) putJSON(path string, body interface{}) (*http.Response, error) {
	return c.sendJSON("PUT", path, body)
}

func (c *client) postJSON(path string, body interface{}) (*http.Response, error) {
	return c.sendJSON("POST", path, body)
}

func (c *client) parseResponse(response *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error().Err(err).Msg("read response body error")
		return body, err
	}
	log.Debug().Str("body", string(body)).Msg("response body")
	if response.StatusCode != http.StatusOK {
		log.Error().Msgf("request error: %s", response.Status)
		m := make(map[string]string)
		if err := json.Unmarshal(body, &m); err != nil {
			log.Warn().Err(err).Msg("unable to unmarshal error body")
		} else if errorMessage, ok := m["error"]; ok {
			return body, errors.Errorf("request error: %s with error message: %s", response.Status, errorMessage)
		}
		return body, errors.Errorf("request error: %s with unspecified error message", response.Status)
	}
	return body, nil
}

func (c *pfClient) Close() {
	if c.stopChan != nil {
		log.Info().Msgf("closing port forwarder %v", c.fw)
		close(c.stopChan)
		c.stopChan = nil
	}
}
