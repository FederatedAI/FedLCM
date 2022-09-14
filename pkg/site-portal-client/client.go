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

package site_portal_client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/FederatedAI/FedLCM/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Client provides interface to work with a site portal service
type Client interface {
	// ConfigAndConnectSite configures the site info and connect it to the fml manager
	ConfigAndConnectSite() error
}

type client struct {
	site       Site
	httpClient *http.Client
	jwtToken   string
}

var _ Client = (*client)(nil)

func NewClient(site Site) (Client, error) {
	newClient := &client{
		site:       site,
		httpClient: http.DefaultClient,
	}
	if site.HTTPS {
		newClient.httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	}
	err := newClient.genToken()
	if err != nil {
		return nil, err
	}
	return newClient, nil
}

func (c *client) ConfigAndConnectSite() error {
	resp, err := c.sendJSON("PUT", "site", c.site)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	if err != nil {
		return errors.Wrap(err, "failed to configure site")
	}

	err = c.connectSite()
	if err != nil {
		return errors.Wrap(err, "failed to connect site to fml-manager")
	}
	return nil
}

func (c *client) connectSite() error {
	resp, err := c.sendJSON("POST", "site/fmlmanager/connect", FMLManagerConnectionInfo{
		Endpoint:   c.site.FMLManagerEndpoint,
		ServerName: c.site.FMLManagerServerName,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = c.parseResponse(resp)
	if err != nil {
		return errors.Wrapf(err, "failed to configure site")
	}
	return err
}

func (c *client) genToken() error {

	login := map[string]string{
		"username": c.site.Username,
		"password": c.site.Password,
	}

	loginJsonB, err := json.Marshal(login)

	body := bytes.NewReader(loginJsonB)
	loginUrl := c.genURL("user/login")

	request, err := http.NewRequest("POST", loginUrl, body)
	if err != nil {
		return err
	}

	var resp *http.Response
	resp, err = c.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result := map[string]interface{}{}

	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(string(respBody))
	}

	c.jwtToken = fmt.Sprint(result["data"])
	return nil
}

func (c *client) sendJSON(method, path string, body interface{}) (*http.Response, error) {
	var resp *http.Response
	if err := utils.RetryWithMaxAttempts(func() error {
		err := c.genToken()
		if err != nil {
			return errors.Wrapf(err, "get token error")
		}
		authorization := fmt.Sprintf("Bearer %s", c.jwtToken)

		urlStr := c.genURL(path)
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
		log.Info().Msg(fmt.Sprintf("%s request to %s with body %s", method, urlStr, string(payload)))
		resp, err = c.httpClient.Do(req)
		return err
	}, 3, 10*time.Second); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) genURL(path string) string {
	scheme := "http"
	if c.site.HTTPS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d/api/v1/%s", scheme, c.site.ExternalHost, c.site.ExternalPort, path)
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
