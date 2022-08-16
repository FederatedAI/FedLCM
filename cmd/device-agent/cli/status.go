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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"

	"github.com/FederatedAI/FedLCM/server/application/service"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func StatusCommand() *cli.Command {
	return &cli.Command{
		Name: "status",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "server-host",
				Aliases:  []string{"s"},
				Value:    "",
				Usage:    "Host address of the lifecycle manager",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "server-port",
				Aliases:  []string{"p"},
				Usage:    "Port number of the lifecycle manager",
				Required: true,
			},
			&cli.BoolFlag{
				Name:     "tls",
				Value:    false,
				Usage:    "Enable TLS",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "token",
				Aliases:  []string{"t"},
				Value:    "",
				Usage:    "The registration token string",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "uuid",
				Aliases:  []string{"u"},
				Value:    "",
				Usage:    "The envoy UUID",
				Required: true,
			},
		},
		Usage: "Query the status of the Envoy",
		Action: func(c *cli.Context) error {
			token := c.String("token")
			uuid := c.String("uuid")
			envoy, err := getEnvoyInfo(token, uuid, c.String("server-host"), c.Int("server-port"), c.Bool("tls"))
			if err != nil {
				return err
			}
			log.Infof("Envoy status is: %v", envoy.Status)
			return nil
		},
	}
}

func getEnvoyInfo(token, uuid, host string, port int, tls bool) (*service.OpenFLEnvoyDetail, error) {
	resp, err := sendJSON("GET", getUrl(host, port, tls, fmt.Sprintf("federation/openfl/envoy/%s?token=%s", uuid, token)), nil)
	defer resp.Body.Close()
	body, err := parseResponse(resp)
	if err != nil {
		return nil, err
	}

	type Response struct {
		Code    int                        `json:"code"`
		Message string                     `json:"message"`
		Data    *service.OpenFLEnvoyDetail `json:"data"`
	}
	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	if response.Data == nil || response.Data.UUID != uuid {
		return nil, errors.Errorf("invalid response: %v", response)
	}
	return response.Data, nil
}
