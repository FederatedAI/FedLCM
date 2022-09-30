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

package fmlmanager

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type client struct {
	// endpoint address
	endpoint   string
	serverName string
}

// NewFMLManagerClient returns a client to FML manager service
func NewFMLManagerClient(endpoint string, serverName string) *client {
	return &client{
		endpoint:   endpoint,
		serverName: serverName,
	}
}

// CreateSite registered a site to FML manager
func (c *client) CreateSite(site *Site) error {
	resp, err := c.postJSON("site", site)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// UnregisterSite unregistered the site from FML manager
func (c *client) UnregisterSite(siteUUID string) error {
	resp, err := c.delete(fmt.Sprintf("site/%s", siteUUID))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendProjectClosing sends a project closing request
func (c *client) SendProjectClosing(projectUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/%s/close", projectUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendInvitation sends an invitation request
func (c *client) SendInvitation(invitation ProjectInvitation) error {
	resp, err := c.postJSON("project/invitation", invitation)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendInvitationAcceptance sends invitation acceptance response
func (c *client) SendInvitationAcceptance(invitationUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/invitation/%s/accept", invitationUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendInvitationRejection sends invitation reject response
func (c *client) SendInvitationRejection(invitationUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/invitation/%s/reject", invitationUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendInvitationRevocation send invitation revocation request
func (c *client) SendInvitationRevocation(invitationUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/invitation/%s/revoke", invitationUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendProjectDataAssociation sends new project data association to FML manager
func (c *client) SendProjectDataAssociation(projectUUID string, association ProjectDataAssociation) error {
	resp, err := c.postJSON(fmt.Sprintf("project/%s/data/associate", projectUUID), association)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendProjectDataDismissal sends project data dismissal to FML manager
func (c *client) SendProjectDataDismissal(projectUUID string, association ProjectDataAssociationBase) error {
	resp, err := c.postJSON(fmt.Sprintf("project/%s/data/dismiss", projectUUID), association)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendProjectParticipantLeaving sends project participant leaving to FML manager
func (c *client) SendProjectParticipantLeaving(projectUUID, siteUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/%s/participant/%s/leave", projectUUID, siteUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendProjectParticipantDismissal sends project participant dismissal to FML manager
func (c *client) SendProjectParticipantDismissal(projectUUID, siteUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/%s/participant/%s/dismiss", projectUUID, siteUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendJobCreationRequest sends job creation request
func (c *client) SendJobCreationRequest(uuid, username string, originalRequest string) error {
	creationRequest := &JobRemoteJobCreationRequest{}
	if err := json.Unmarshal([]byte(originalRequest), creationRequest); err != nil {
		return err
	}
	creationRequest.UUID = uuid
	creationRequest.Username = username
	resp, err := c.postJSON("job/create", creationRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendJobApprovalResponse sends job approval response
func (c *client) SendJobApprovalResponse(jobUUID string, approvalContext JobApprovalContext) error {
	resp, err := c.postJSON(fmt.Sprintf("job/%s/response", jobUUID), approvalContext)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// SendJobStatusUpdate sends job status update
func (c *client) SendJobStatusUpdate(jobUUID string, context JobStatusUpdateContext) error {
	resp, err := c.postJSON(fmt.Sprintf("job/%s/status", jobUUID), context)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

// GetAllSite gets all sites from the FML manager
func (c *client) GetAllSite() ([]Site, error) {
	urlStr := c.genURL("site")
	var resp *http.Response
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Parse URL failed")
	}
	if u.Scheme == "https" {
		client, err := c.getHttpsClientWithCert()
		if err != nil {
			return nil, err
		}
		resp, err = client.Get(urlStr)
		if err != nil {
			return nil, err
		}
		log.Info().Msg(fmt.Sprintf("Getting %s via HTTPs", urlStr))
	} else {
		resp, err = http.Get(urlStr)
		if err != nil {
			return nil, err
		}
		log.Info().Msg(fmt.Sprintf("Getting %s via HTTP", urlStr))
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	type GetSiteResponse struct {
		CommonResponse
		Sites []Site `json:"data"`
	}
	var getSiteResponse GetSiteResponse
	if err := json.Unmarshal(body, &getSiteResponse); err != nil {
		return nil, err
	}
	return getSiteResponse.Sites, nil
}

// GetProjectDataAssociation returns a map of associated data items of a project, indexed by data uuid
func (c *client) GetProjectDataAssociation(projectUUID string) (map[string]ProjectDataAssociation, error) {
	resp, err := c.getJSON(fmt.Sprintf("project/%s/data", projectUUID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseByte, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	type Response struct {
		CommonResponse
		Data map[string]ProjectDataAssociation `json:"data"`
	}
	var getAssociationResponse Response
	if err := json.Unmarshal(responseByte, &getAssociationResponse); err != nil {
		return nil, err
	}
	return getAssociationResponse.Data, nil
}

// GetProjectParticipant returns a map of participant items of a project, indexed by participant uuid
func (c *client) GetProjectParticipant(projectUUID string) (map[string]ProjectParticipant, error) {
	resp, err := c.getJSON(fmt.Sprintf("project/%s/participant", projectUUID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseByte, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	type Response struct {
		CommonResponse
		Data map[string]ProjectParticipant `json:"data"`
	}
	var getParticipantResponse Response
	if err := json.Unmarshal(responseByte, &getParticipantResponse); err != nil {
		return nil, err
	}
	return getParticipantResponse.Data, nil
}

// GetProject returns a list of projects the current site is related to
func (c *client) GetProject(siteUUID string) (map[string]ProjectInfoWithStatus, error) {
	resp, err := c.getJSON("project", map[string]string{"participant": siteUUID})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseByte, err := c.parseResponse(resp)
	if err != nil {
		return nil, err
	}
	type Response struct {
		CommonResponse
		Data map[string]ProjectInfoWithStatus `json:"data"`
	}
	var getProjectResponse Response
	if err := json.Unmarshal(responseByte, &getProjectResponse); err != nil {
		return nil, err
	}
	return getProjectResponse.Data, nil
}

func (c *client) delete(path string) (*http.Response, error) {
	urlStr := c.genURL(path)
	baseUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	urlStr = baseUrl.String()
	var resp *http.Response

	req, err := http.NewRequest(http.MethodDelete, urlStr, nil)
	if err != nil {
		return nil, err
	}
	if baseUrl.Scheme == "https" {
		client, err := c.getHttpsClientWithCert()
		if err != nil {
			return nil, err
		}
		log.Info().Msg(fmt.Sprintf("Deleting %s via HTTPs", urlStr))
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	} else {
		log.Info().Msg(fmt.Sprintf("Deleting %s via HTTP", urlStr))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func (c *client) getJSON(path string, query map[string]string) (*http.Response, error) {
	urlStr := c.genURL(path)
	baseUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	for k, v := range query {
		params.Add(k, v)
	}
	baseUrl.RawQuery = params.Encode()

	urlStr = baseUrl.String()
	var resp *http.Response
	if baseUrl.Scheme == "https" {
		client, err := c.getHttpsClientWithCert()
		if err != nil {
			return nil, err
		}
		resp, err = client.Get(urlStr)
		if err != nil {
			return nil, err
		}
		log.Info().Msg(fmt.Sprintf("Getting %s via HTTPs", urlStr))
	} else {
		resp, err = http.Get(urlStr)
		if err != nil {
			return nil, err
		}
		log.Info().Msg(fmt.Sprintf("Getting %s via HTTP", urlStr))
	}
	return resp, nil
}

func (c *client) postJSON(path string, body interface{}) (*http.Response, error) {
	urlStr := c.genURL(path)
	var payload []byte
	if stringBody, ok := body.(string); ok {
		payload = []byte(stringBody)
	} else {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "Parse URL failed")
	}
	//if endpoint scheme is https, use tls
	if u.Scheme == "https" {
		client, err := c.getHttpsClientWithCert()
		if err != nil {
			return nil, err
		}
		log.Info().Msg(fmt.Sprintf("Posting request to %s via HTTPs with body %s", urlStr, string(payload)))
		return client.Do(req)
	} else {
		log.Info().Msg(fmt.Sprintf("Posting request to %s via HTTP with body %s", urlStr, string(payload)))
		return http.DefaultClient.Do(req)
	}
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

func (c *client) genURL(path string) string {
	return fmt.Sprintf("%s/api/v1/%s", c.endpoint, path)
}

// getHttpsClientWithCert returns a https client use siteportal client cert
func (c *client) getHttpsClientWithCert() (*http.Client, error) {
	caCertPath := viper.GetString("siteportal.tls.ca.cert")
	sitePortalClientCert := viper.GetString("siteportal.tls.client.cert")
	sitePortalClientKey := viper.GetString("siteportal.tls.client.key")
	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, errors.Wrapf(err, "read ca.crt file error")
	}
	pool.AppendCertsFromPEM(caCrt)
	clientCrt, err := tls.LoadX509KeyPair(sitePortalClientCert, sitePortalClientKey)
	if err != nil {
		return nil, errors.Wrapf(err, "LoadX509KeyPair error:")
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{clientCrt},
			ServerName:   c.serverName,
		},
	}
	return &http.Client{Transport: tr}, nil
}
