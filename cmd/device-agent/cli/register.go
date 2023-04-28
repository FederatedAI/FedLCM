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
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/FederatedAI/FedLCM/pkg/utils"
	"github.com/FederatedAI/FedLCM/server/domain/entity"
	"github.com/FederatedAI/FedLCM/server/domain/valueobject"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/util/homedir"
)

type envoyRegistrationConfig struct {
	Name                  string                         `json:"name"`
	Description           string                         `json:"description"`
	Namespace             string                         `json:"namespace"`
	ChartUUID             string                         `json:"chart_uuid" yaml:"chartUUID"`
	Labels                valueobject.Labels             `json:"labels"`
	SkipCommonPythonFiles bool                           `json:"skip_common_python_files" yaml:"skipCommonPythonFiles"`
	EnablePSP             bool                           `json:"enable_psp" yaml:"enablePSP"`
	LessPrivileged        bool                           `json:"less_privileged" yaml:"lessPrivileged"`
	RegistryConfig        valueobject.KubeRegistryConfig `json:"registry_config" yaml:"registryConfig"`
}

type envoyRegistrationRequest struct {
	// required
	KubeConfig string `json:"kubeconfig"`
	TokenStr   string `json:"token"`

	// optional
	envoyRegistrationConfig
	ConfigYAML string `json:"config_yaml"`
}

func RegisterCommand() *cli.Command {
	return &cli.Command{
		Name: "register",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "extra-config",
				Usage:    "YAML file containing optional extra configuration for the register command",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "kube-config",
				Aliases:  []string{"k"},
				Value:    filepath.Join(homedir.HomeDir(), ".kube", "config"),
				Usage:    "Kubeconfig file",
				Required: false,
			},
			&cli.StringFlag{
				Name:     "envoy-config",
				Usage:    "Optional Envoy configuration file path",
				Required: false,
			},
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
			&cli.BoolFlag{
				Name:     "wait",
				Aliases:  []string{"w"},
				Value:    false,
				Usage:    "Wait for the registration to finished",
				Required: false,
			},
		},
		Usage: "Register to the lifecycle manager",
		Action: func(c *cli.Context) error {

			req := &envoyRegistrationRequest{
				TokenStr: c.String("token"),
			}

			kubeConfigPath := c.String("kube-config")
			kubeconfigBytes, err := ioutil.ReadFile(kubeConfigPath)
			if err != nil {
				return err
			}
			var m map[string]interface{}
			if err := yaml.Unmarshal(kubeconfigBytes, &m); err != nil {
				return errors.Wrap(err, "invalid kube config")
			}
			req.KubeConfig = string(kubeconfigBytes)

			envoyConfigPath := c.String("envoy-config")
			if envoyConfigPath != "" {
				envoyConfigBytes, err := ioutil.ReadFile(envoyConfigPath)
				if err != nil {
					return err
				}
				var m map[string]interface{}
				if err := yaml.Unmarshal(envoyConfigBytes, &m); err != nil {
					return errors.Wrap(err, "invalid envoy config")
				}
				req.ConfigYAML = string(envoyConfigBytes)
			}

			configPath := c.String("extra-config")
			if configPath != "" {
				configBytes, err := ioutil.ReadFile(configPath)
				if err != nil {
					return err
				}
				log.Debugf("ReadFile Success, yaml: %s", string(configBytes))
				config := envoyRegistrationConfig{}
				if err := yaml.Unmarshal(configBytes, &config); err != nil {
					return err
				}
				req.envoyRegistrationConfig = config
			}

			resp, err := sendJSON("POST",
				getUrl(c.String("server-host"), c.Int("server-port"), c.Bool("tls"), "federation/openfl/envoy/register"),
				req)
			defer resp.Body.Close()
			body, err := parseResponse(resp)
			if err != nil {
				return err
			}
			type RegisterResponse struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Data    string `json:"data"`
			}
			var registerResponse RegisterResponse
			if err := json.Unmarshal(body, &registerResponse); err != nil {
				return err
			}
			log.Infof("New Envoy is being prepared, UUID: %s", registerResponse.Data)
			if c.Bool("wait") {
				log.Infof("Waiting for the preparation to finish")
				if err := utils.ExecuteWithTimeout(func() bool {
					envoy, err := getEnvoyInfo(req.TokenStr, registerResponse.Data, c.String("server-host"), c.Int("server-port"), c.Bool("tls"))
					if err != nil {
						log.Errorf("error getting Envoy info: %v, retry", err)
						return false
					}
					log.Infof("Envoy %s(%s) status is: %v", envoy.Name, envoy.UUID, envoy.Status)
					if envoy.Status == entity.ParticipantOpenFLStatusFailed ||
						envoy.Status == entity.ParticipantOpenFLStatusActive ||
						envoy.Status == entity.ParticipantOpenFLStatusUnknown ||
						envoy.Status == entity.ParticipantOpenFLStatusRemoving {
						return true
					}
					return false
				}, time.Hour, time.Second*2); err != nil {
					return err
				}
				log.Infof("Preparation finished")
			}
			return nil
		},
	}
}
