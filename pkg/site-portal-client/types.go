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

// Site contains essential info to configure its connection with fml manager
type Site struct {
	Username             string `json:"-"`
	Password             string `json:"-"`
	Name                 string `json:"name"`
	Description          string `json:"description"`
	PartyID              uint   `json:"party_id"`
	ExternalHost         string `json:"external_host"`
	ExternalPort         uint   `json:"external_port"`
	HTTPS                bool   `json:"https"`
	FMLManagerEndpoint   string `json:"fml_manager_endpoint"`
	FMLManagerServerName string `json:"fml_manager_server_name"`
	FATEFlowHost         string `json:"fate_flow_host"`
	FATEFlowHTTPPort     uint   `json:"fate_flow_http_port"`
}

// FMLManagerConnectionInfo contains connection settings for the fml manager
type FMLManagerConnectionInfo struct {
	// Endpoint address starting with "http" or "https"
	Endpoint string `json:"endpoint"`
	// ServerName is used by Site Portal to verify FML Manager's certificate
	ServerName string `json:"server_name"`
}
