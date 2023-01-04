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

package siteportal

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

// Client is an interface to work with certain site portal service
type Client interface {
	// SendInvitation sends a project invitation to a target site
	SendInvitation(request *ProjectInvitationRequest) error
	// SendInvitationAcceptance sends invitation acceptance response to the site
	SendInvitationAcceptance(invitationUUID string) error
	// SendInvitationRejection sends invitation rejection to the site
	SendInvitationRejection(invitationUUID string) error
	// SendProjectParticipants sends a list of ProjectParticipant to the site
	SendProjectParticipants(projectUUID string, participants []ProjectParticipant) error
	// SendInvitationRevocation sends invitation revocation response to the site
	SendInvitationRevocation(invitationUUID string) error
	// SendParticipantInfoUpdateEvent sends site info update event to the site
	SendParticipantInfoUpdateEvent(event ProjectParticipantUpdateEvent) error
	// SendProjectDataAssociation sends new data association to the site
	SendProjectDataAssociation(projectUUID string, data []ProjectData) error
	// SendProjectDataDismissal sends data association dismissal to the site
	SendProjectDataDismissal(projectUUID string, data []string) error
	// SendJobCreationRequest asks the site portal to create a new job
	SendJobCreationRequest(request string) error
	// SendJobApprovalResponse sends the approval result of a job to the initiating site
	SendJobApprovalResponse(jobUUID string, context JobApprovalContext) error
	// SendJobStatusUpdate sends the job status update
	SendJobStatusUpdate(jobUUID string, context string) error
	// SendProjectParticipantLeaving sends the participant leaving event
	SendProjectParticipantLeaving(projectUUID string, siteUUID string) error
	// SendProjectParticipantDismissal sends the participant dismissal event
	SendProjectParticipantDismissal(projectUUID string, siteUUID string) error
	// SendProjectClosing sends the project closing event
	SendProjectClosing(projectUUID string) error
	// SendProjectParticipantUnregistration sends the participant unregistration event
	SendProjectParticipantUnregistration(siteUUID string) error
	// CheckSiteStatus checks the status of the site
	CheckSiteStatus() error
}

// NewSitePortalClient returns a site port Client instance
func NewSitePortalClient(host string, port uint, https bool, serverName string) Client {
	scheme := "http"
	if https {
		scheme += "s"
	}
	return &client{
		endpoint:   fmt.Sprintf("%s://%s:%d", scheme, host, port),
		serverName: serverName,
	}
}

type client struct {
	endpoint   string
	serverName string
}

func (c *client) SendInvitation(request *ProjectInvitationRequest) error {
	resp, err := c.postJSON("project/internal/invitation", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendInvitationAcceptance(invitationUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/invitation/%s/accept", invitationUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendInvitationRejection(invitationUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/invitation/%s/reject", invitationUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendInvitationRevocation(invitationUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/invitation/%s/revoke", invitationUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendProjectParticipants(projectUUID string, participants []ProjectParticipant) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/%s/participants", projectUUID), participants)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendParticipantInfoUpdateEvent(event ProjectParticipantUpdateEvent) error {
	resp, err := c.postJSON("project/internal/event/participant/update", event)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendProjectParticipantLeaving(projectUUID, siteUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/%s/participant/%s/leave", projectUUID, siteUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendProjectParticipantDismissal(projectUUID, siteUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/%s/participant/%s/dismiss", projectUUID, siteUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendProjectDataAssociation(projectUUID string, data []ProjectData) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/%s/data/associate", projectUUID), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendProjectDataDismissal(projectUUID string, data []string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/%s/data/dismiss", projectUUID), data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendJobCreationRequest(request string) error {
	resp, err := c.postJSON("job/internal/create", request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendJobApprovalResponse(jobUUID string, context JobApprovalContext) error {
	resp, err := c.postJSON(fmt.Sprintf("job/internal/%s/response", jobUUID), context)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendJobStatusUpdate(jobUUID string, context string) error {
	resp, err := c.postJSON(fmt.Sprintf("job/internal/%s/status", jobUUID), context)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendProjectClosing(projectUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/%s/close", projectUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
}

func (c *client) SendProjectParticipantUnregistration(siteUUID string) error {
	resp, err := c.postJSON(fmt.Sprintf("project/internal/all/participant/%s/unregister", siteUUID), "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	return err
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
		return nil, errors.Wrap(err, "parse URL failed")
	}
	//if endpoint scheme is https, use tls
	if u.Scheme == "https" {
		client, err := c.getHttpsClientWithCert()
		if err != nil {
			return nil, err
		}
		log.Info().Msg(fmt.Sprintf("Posting request to %s via HTTPS with body %s", urlStr, string(payload)))
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

func (c *client) CheckSiteStatus() error {
	urlStr := c.genURL("status")
	var resp *http.Response
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.Wrap(err, "Parse URL failed")
	}
	if u.Scheme == "https" {
		client, err := c.getHttpsClientWithCert()
		if err != nil {
			return err
		}
		resp, err = client.Get(urlStr)
		if err != nil {
			return err
		}
		log.Info().Msg(fmt.Sprintf("Getting %s via HTTPs", urlStr))
	} else {
		resp, err = http.Get(urlStr)
		if err != nil {
			return err
		}
		log.Info().Msg(fmt.Sprintf("Getting %s via HTTP", urlStr))
	}
	defer resp.Body.Close()
	body, err := c.parseResponse(resp)
	if err != nil {
		return err
	}
	m := make(map[string]interface{})
	if err := json.Unmarshal(body, &m); err != nil {
		return err
	}
	if m["msg"] != "The service is running" {
		return errors.New("The service is not running")
	}
	return nil
}

// getHttpsClientWithCert returns a https client use fml manager client cert
func (c *client) getHttpsClientWithCert() (*http.Client, error) {
	caCertPath := viper.GetString("fmlmanager.tls.ca.cert")
	fmlManagerClientCert := viper.GetString("fmlmanager.tls.client.cert")
	fmlManagerClientKey := viper.GetString("fmlmanager.tls.client.key")
	pool := x509.NewCertPool()
	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, errors.Wrapf(err, "read ca.crt file error")
	}
	pool.AppendCertsFromPEM(caCrt)
	clientCrt, err := tls.LoadX509KeyPair(fmlManagerClientCert, fmlManagerClientKey)
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
