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

package fateclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	gourl "net/url"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type client struct {
	// host address
	host string
	// port number
	port uint
	// https sets if https is enabled
	https bool
	// apiVersion is the version string of the api
	apiVersion string
}

// DataUploadRequest is the request to upload data
type DataUploadRequest struct {
	File      string `json:"file"`
	Head      int    `json:"head"`
	Partition int    `json:"partition"`
	WorkMode  int    `json:"work_mode"`
	Backend   int    `json:"backend"`
	Namespace string `json:"namespace"`
	TableName string `json:"table_name"`
	Drop      int    `json:"drop"`
}

// ComponentTrackingCommonRequest is the request to query the metric of a component
type ComponentTrackingCommonRequest struct {
	JobID         string `json:"job_id"`
	PartyID       uint   `json:"party_id"`
	Role          string `json:"role"`
	ComponentName string `json:"component_name"`
}

// ModelDeployRequest is the request to deploy a trained model
type ModelDeployRequest struct {
	ModelInfo
	ComponentList []string `json:"cpn_list"`
}

// CommonResponse represents common response from FATE flow
type CommonResponse struct {
	RetCode int    `json:"retcode"`
	RetMsg  string `json:"retmsg"`
}

// JobSubmissionResponseData contains returned data of job submission response
type JobSubmissionResponseData struct {
	ModelInfo ModelInfo `json:"model_info"`
}

// ModelDeployResponseData contains returned data of model deploy response
type ModelDeployResponseData struct {
	ModelInfo
}

// ModelInfo contains the key infos of a model
type ModelInfo struct {
	ModelID      string `json:"model_id"`
	ModelVersion string `json:"model_version"`
}

// HomoModelConversionRequest is the request to convert a homo model
type HomoModelConversionRequest struct {
	ModelInfo
	PartyID uint   `json:"party_id"`
	Role    string `json:"role"`
}

// HomoModelDeploymentRequest is the request to deploy a model to KFServing
type HomoModelDeploymentRequest struct {
	HomoModelConversionRequest
	ServiceID            string      `json:"service_id"`
	ComponentName        string      `json:"component_name"`
	DeploymentType       string      `json:"deployment_type"`
	DeploymentParameters interface{} `json:"deployment_parameters"`
}

// NewFATEFlowClient returns a fate flow client
func NewFATEFlowClient(host string, port uint, https bool) *client {
	return &client{
		host:  host,
		port:  port,
		https: https,
	}
}

// UploadData calls the /data/upload API to upload local data to FATE flow
func (c *client) UploadData(request DataUploadRequest) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	if fw, err := w.CreateFormFile("file", request.File); err != nil {
		return "", err
	} else {
		src, err := os.Open(request.File)
		if err != nil {
			return "", err
		}
		defer src.Close()
		if _, err = io.Copy(fw, src); err != nil {
			return "", err
		}
	}
	_ = w.Close()

	args, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	url := c.genURL(fmt.Sprintf("data/upload?%s", gourl.QueryEscape(string(args))))
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	log.Info().Msg(fmt.Sprintf("Posting data upload request to %s", url))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return "", err
	}
	type UploadDataResponse struct {
		CommonResponse
		JobID string `json:"jobId"`
	}
	var uploadDataResponse UploadDataResponse
	if err := json.Unmarshal(body, &uploadDataResponse); err != nil {
		return "", err
	}
	if uploadDataResponse.RetCode != 0 || uploadDataResponse.JobID == "" {
		responseError := errors.Errorf("failed to get job id, retmsg: %s", uploadDataResponse.RetMsg)
		log.Err(responseError)
		return "", responseError
	}
	return uploadDataResponse.JobID, nil
}

// DeleteTable deletes a table in fate
func (c *client) DeleteTable(tableNamespace, tableName string) error {
	resp, err := c.postJSON("table/delete",
		fmt.Sprintf(`{"namespace": "%s", "table_name": "%s"}`, tableNamespace, tableName))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return err
	}
	var tableDeletionResp CommonResponse
	err = json.Unmarshal(body, &tableDeletionResp)
	if err != nil {
		return err
	}
	if tableDeletionResp.RetCode != 0 && !strings.Contains(tableDeletionResp.RetMsg, "no find table") {
		return errors.Errorf("error return code: %d, msg: %s", tableDeletionResp.RetCode, tableDeletionResp.RetMsg)
	}
	return nil
}

// TestConnection sends a dummy version query request to the fate flow service
func (c *client) TestConnection() error {
	_, err := c.GetFATEVersion()
	return err
}

// GetFATEVersion returns the FATE-flow version
func (c *client) GetFATEVersion() (string, error) {
	resp, err := c.postJSON("version/get", `{"module": "FATE"}`)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return "", err
	}

	type versionGetResp struct {
		Data map[string]string `json:"data"`
		CommonResponse
	}
	versionInfo := new(versionGetResp)
	err = json.Unmarshal(body, &versionInfo)
	if err != nil {
		log.Error().Err(err).Msg("unmarshal resp body error")
		return "", err
	}
	if versionInfo.RetCode != 0 || versionInfo.Data["FATE"] == "" {
		log.Error().Msgf("unexpected version query response")
		return "", fmt.Errorf("unexpected version query response")
	}
	return versionInfo.Data["FATE"], nil
}

// QueryJobStatus returns the job status
func (c *client) QueryJobStatus(jobID string) (string, error) {
	resp, err := c.postJSON("job/query", fmt.Sprintf(`{"job_id":"%s"}`, jobID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return "", err
	}
	type JobStatusData struct {
		JobStatus string `json:"f_status"`
	}
	type JobStatusResponse struct {
		CommonResponse
		Data []JobStatusData `json:"data"`
	}
	jobQueryResponse := JobStatusResponse{}
	err = json.Unmarshal(body, &jobQueryResponse)
	if err != nil {
		return "", err
	}
	if jobQueryResponse.Data == nil || len(jobQueryResponse.Data) == 0 {
		return "unknown", errors.Errorf("Failed to query the job with id %s", jobID)
	}
	return jobQueryResponse.Data[0].JobStatus, nil
}

// SubmitJob submit a new Job
func (c *client) SubmitJob(conf, dsl string) (string, *ModelInfo, error) {
	var confObj map[string]interface{}
	if err := json.Unmarshal([]byte(conf), &confObj); err != nil {
		return "", nil, err
	}
	var dslObj map[string]interface{}
	if err := json.Unmarshal([]byte(dsl), &dslObj); err != nil {
		return "", nil, err
	}
	jobSubmissionBody := map[string]interface{}{
		"job_dsl":          dslObj,
		"job_runtime_conf": confObj,
	}
	resp, err := c.postJSON("job/submit", jobSubmissionBody)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return "", nil, err
	}
	type JobSubmissionResponse struct {
		CommonResponse
		JobID string                    `json:"jobId"`
		Data  JobSubmissionResponseData `json:"data"`
	}
	var response JobSubmissionResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", nil, err
	}
	if response.RetCode != 0 || response.JobID == "" {
		responseError := errors.Errorf("failed to get job id, retmsg: %s", response.RetMsg)
		log.Err(responseError)
		return "", nil, responseError
	}
	return response.JobID, &response.Data.ModelInfo, nil
}

// DeployModel deploy a trained model
func (c *client) DeployModel(request ModelDeployRequest) (*ModelInfo, error) {
	resp, err := c.postJSON("model/deploy", request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	type ModelDeployResponse struct {
		CommonResponse
		Data ModelDeployResponseData `json:"data"`
	}
	var response ModelDeployResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if response.RetCode != 0 {
		responseError := errors.Errorf("failed to get job id, retmsg: %s", response.RetMsg)
		log.Err(responseError)
		return nil, responseError
	}
	return &response.Data.ModelInfo, nil
}

// GetComponentMetric returns metric data for the specified component
func (c *client) GetComponentMetric(request ComponentTrackingCommonRequest) (interface{}, error) {
	resp, err := c.postJSON("tracking/component/metric/all", request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return "", err
	}
	type ComponentMetricResponse struct {
		CommonResponse
		Data interface{} `json:"data"`
	}
	var response ComponentMetricResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	if response.RetCode != 0 {
		responseError := errors.Errorf("failed to get component metric, retmsg: %s", response.RetMsg)
		log.Err(responseError)
		return "", responseError
	}
	return response.Data, nil
}

// GetComponentOutputDataSummary returns part of (maximum 100 lines of records) of the output data
func (c *client) GetComponentOutputDataSummary(request ComponentTrackingCommonRequest) (interface{}, interface{}, error) {
	resp, err := c.postJSON("tracking/component/output/data", request)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return nil, nil, err
	}
	type ComponentOutputDataResponse struct {
		CommonResponse
		Data interface{} `json:"data"`
		Meta interface{} `json:"meta"`
	}
	var response ComponentOutputDataResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, nil, err
	}
	if response.RetCode != 0 {
		responseError := errors.Errorf("failed to get component output data, retmsg: %s", response.RetMsg)
		log.Err(responseError)
		return nil, nil, responseError
	}
	//	response.Data should be like:
	//	[
	//		[
	//			["542", 1, 0.049868, 1.07685, 0.004134, -0.095249, -1.155891, -0.742153, -0.53295, -0.07775, -0.289188, -0.797202],
	//          ...
	//		]
	//	]
	//	response.Meta should be like:
	//	{
	//		"header": [
	//			["id", "y", "x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9"]
	//		],
	//		"names": ["data"],
	//		"total": [551]
	//	}
	return response.Data, response.Meta, nil
}

// GetComponentSummary is used for getting the summary for an intersection component, it can get the intersection rate.
func (c *client) GetComponentSummary(request ComponentTrackingCommonRequest) (interface{}, error) {
	resp, err := c.postJSON("tracking/component/summary/download", request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	type ComponentSummaryResponse struct {
		CommonResponse
		Data interface{} `json:"data"`
		Meta interface{} `json:"meta"`
	}
	var response ComponentSummaryResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if response.RetCode != 0 {
		responseError := errors.Errorf("failed to get component summary data, retmsg: %s", response.RetMsg)
		log.Err(responseError)
		return nil, responseError
	}
	// response.Data should be like:
	//{
	//	"intersect_num": 551,
	//	"intersect_rate": 1,
	//}
	return response.Data, nil
}

// GetComponentOutputDataDownloadRequest returns a *http.Request object to be used to download the component data
func (c *client) GetComponentOutputDataDownloadRequest(request ComponentTrackingCommonRequest) (*http.Request, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("GET", c.genURL("tracking/component/output/data/download"), bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// ConvertHomoModel calls the model/homo/convert API to convert the model
func (c *client) ConvertHomoModel(request HomoModelConversionRequest) error {
	resp, err := c.postJSON("model/homo/convert", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return err
	}
	var response CommonResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	if response.RetCode != 0 {
		responseError := errors.Errorf("failed to convert homo model, retmsg: %s", response.RetMsg)
		log.Err(responseError)
		return responseError
	}
	return nil
}

// DeployHomoModel calls the FATE-Flow API to deploy the model
func (c *client) DeployHomoModel(request HomoModelDeploymentRequest) (string, error) {
	resp, err := c.postJSON("model/homo/deploy", request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return "", err
	}
	var response CommonResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return string(body), err
	}
	if response.RetCode != 0 {
		responseError := errors.Errorf("failed to deploy homo model, retmsg: %s", response.RetMsg)
		log.Err(responseError)
		return string(body), responseError
	}
	return string(body), nil
}

func (c *client) parseResponse(response *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error().Err(err).Msg("read response body error")
		return nil, err
	}
	log.Info().Str("body", string(body)).Msg("response body")
	if response.StatusCode != http.StatusOK {
		log.Error().Msgf("request error: %s", response.Status)
		return nil, errors.Errorf("request error: %s, body: %s", response.Status, string(body))
	}
	return body, nil
}

func (c *client) postJSON(path string, body interface{}) (*http.Response, error) {
	url := c.genURL(path)
	var payload []byte
	if stringBody, ok := body.(string); ok {
		payload = []byte(stringBody)
	} else {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	log.Info().Msg(fmt.Sprintf("Posting request to %s with body %s", url, string(payload)))
	return http.DefaultClient.Do(req)
}

func (c *client) genURL(path string) string {
	schemaStr := "http"
	if c.https {
		schemaStr = "https"
	}
	return fmt.Sprintf("%s://%s:%d/v1/%s", schemaStr, c.host, c.port, path)
}
