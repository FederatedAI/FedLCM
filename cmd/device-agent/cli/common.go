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

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/FederatedAI/FedLCM/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func sendJSON(method, urlStr string, body interface{}) (*http.Response, error) {
	var resp *http.Response
	if err := utils.RetryWithMaxAttempts(func() error {
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
		log.Debugf("%s request to %s with body %s", method, urlStr, string(payload))
		resp, err = http.DefaultClient.Do(req)
		return err
	}, 3, 10*time.Second); err != nil {
		return nil, err
	}
	return resp, nil
}

func getUrl(host string, port int, tls bool, path string) string {
	schemaStr := "http"
	if tls {
		schemaStr = "https"
	}
	return fmt.Sprintf("%s://%s:%v/api/v1/%s", schemaStr, host, port, path)
}

func parseResponse(response *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body error")
	}
	log.Debugf("response body: %v", string(body))
	if response.StatusCode != http.StatusOK {
		return body, errors.Errorf("request error: %s, body: %s", response.Status, string(body))
	}
	return body, nil
}
