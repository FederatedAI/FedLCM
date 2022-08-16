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

package event

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Exchange is an interface to receive and post event
type Exchange interface {
	// PostEvent sends an event
	PostEvent(event Event) error
}

type selfHttpExchange struct {
	endpoint string
}

// NewSelfHttpExchange returns an Exchange for posting event to the site itself
func NewSelfHttpExchange() *selfHttpExchange {

	tlsEnabled := viper.GetBool("siteportal.tls.enabled")
	if tlsEnabled {
		tlsPort := viper.GetString("siteportal.tls.port")
		if tlsPort == "" {
			// this is the default listening port
			tlsPort = "8443"
		}
		return &selfHttpExchange{
			endpoint: fmt.Sprintf("https://localhost:%s", tlsPort),
		}
	}
	// gin's default port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return &selfHttpExchange{
		endpoint: fmt.Sprintf("http://localhost:%s", port),
	}
}

// PostEvent calls the http API to create the event
func (e *selfHttpExchange) PostEvent(event Event) error {
	urlStr := e.genURL(event.GetUrl())
	var payload []byte
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.Wrap(err, "Parse URL failed")
	}
	// if endpoint scheme is https, use tls
	if u.Scheme == "https" {
		caCertPath := viper.GetString("siteportal.tls.ca.cert")
		sitePortalClientCert := viper.GetString("siteportal.tls.client.cert")
		sitePortalClientKey := viper.GetString("siteportal.tls.client.key")
		pool := x509.NewCertPool()
		caCrt, err := ioutil.ReadFile(caCertPath)
		if err != nil {
			return errors.Wrapf(err, "read ca.crt file error")
		}
		pool.AppendCertsFromPEM(caCrt)
		clientCrt, err := tls.LoadX509KeyPair(sitePortalClientCert, sitePortalClientKey)
		if err != nil {
			return errors.Wrapf(err, "LoadX509KeyPair error")
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{clientCrt},
			},
		}
		client := &http.Client{Transport: tr}
		log.Info().Msg(fmt.Sprintf("Posting request to %s via HTTPs with body %s", urlStr, string(payload)))
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
	} else {
		log.Info().Msg(fmt.Sprintf("Posting request to %s via HTTP with body %s", urlStr, string(payload)))
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()
	_, err = e.parseResponse(resp)
	return err
}

func (e *selfHttpExchange) parseResponse(response *http.Response) ([]byte, error) {
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

func (e *selfHttpExchange) genURL(path string) string {
	return fmt.Sprintf("%s/api/v1/%s", e.endpoint, path)
}
