// Copyright 2022 VMware, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { Component, OnInit, AfterViewChecked, ChangeDetectorRef, AfterContentChecked, OnDestroy } from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { SiteConfigureService } from '../../service/site-configure.service';
import { SiteService } from '../../service/site.service';
import '@cds/core/icon/register.js';
import { checkCircleIcon, checkIcon, ClarityIcons, errorStandardIcon, infoCircleIcon } from '@cds/core/icon';
import { Router, ActivatedRoute } from '@angular/router';
import { MessageService } from '../../components/message/message.service'
import { ValidatorGroup } from '../../../config/validators'

ClarityIcons.addIcons(checkIcon, errorStandardIcon, checkCircleIcon, infoCircleIcon);

@Component({
  selector: 'app-site-config',
  templateUrl: './site-config.component.html',
  styleUrls: ['./site-config.component.css']
})
export class SiteConfigComponent implements OnInit, AfterViewChecked, AfterContentChecked, OnDestroy {
  siteInfo: any = '';
  siteUpdatedInfo: any = {
    description: "",
    external_host: "",
    external_port: 0,
    https: false,
    fate_flow_grpc_port: 0,
    fate_flow_host: "",
    fate_flow_http_port: 0,
    fml_manager_endpoint: "",
    fml_manager_server_name: "",
    id: 0,
    kubeflow_config: {
      kubeconfig: "",
      minio_access_key: "",
      minio_endpoint: "",
      minio_region: "",
      minio_secret_key: "",
      minio_ssl_enabled: false,
    },
    name: "",
    party_id: 0,
    uuid: ""
  }

  // form for update site configuration
  form = this.fb.group(
    ValidatorGroup([
      {
        name: 'name',
        value: '',
        type: ['word'],
        max: 20,
        min: 2
      },
      {
        name: 'site_ip',
        type: ['notRequired', 'endpointWithoutPort'],
        value: ''
      },
      {
        name: 'site_https_string',
        type: [''],
        value: ''
      },
      {
        name: 'party_id',
        type: ['notRequired','number'],
        value: 0
      },
      {
        name: 'site_port',
        type: ['notRequired', 'number'],
        value: ''
      },
      {
        name: 'desc',
        type: ['']
      },
      {
        name: 'endpoint',
        type: ['notRequired', 'endpoint'],
        value: ''
      },
      {
        name: 'fml_https_string',
        type: [''],
        value: ''
      },
      {
        name: 'fml_manager_server_name',
        type: [''],
        value: ''
      },
      {
        name: 'fate_flow_host_ip',
        type: ['notRequired'],
        value: ''
      },
      {
        name: 'http_port',
        type: ['notRequired','number'],
        value: ''
      },
      {
        name: 'minio_endpoint',
        type: ['notRequired', 'endpoint'],
        value: ''
      },
      {
        name: 'access_key',
        type: [''],
        value: ''
      },
      {
        name: 'secret_key',
        type: [''],
        value: ''
      },
      {
        name: 'kubeconfig',
        type: [''],
        value: ''
      },
      {
        name: 'minio_ssl_enabled',
        type: ['notRequired', ''],
        value: false
      },
    ])
  );


  name: string = "";
  party_id = 0;
  site_ip: string = "";
  site_port = 0;
  site_https_string = "";
  desc: string = "";
  endpoint: string = "";
  fate_flow_host_ip: string = "";
  http_port = 0;
  minio_endpoint: string = "";
  access_key: string = "";
  secret_key: string = "";
  kubeconfig: string = "";
  minio_region: string = "";
  minio_ssl_enabled: boolean = false;
  // fml_manager_connected is the current status of fml manager connection for site
  fml_manager_connected: boolean = false;
  fml_https_string = "http"
  // fml_manager_endpoint_address is the full address (inclues http:// or https:// prefix)
  fml_manager_endpoint_address = "";
  fml_manager_server_name = ""
  fate_flow_connected: boolean = false;
  kubeflow_connected: boolean = false;
  isRegisterToFMLManagerSubmitted: boolean = false;
  isConnectFailed: boolean = false;
  isTestFATEFlowSubmit = false
  isTestFATEFlowFailed = false
  panelOpen1: boolean = true;
  panelOpen2: boolean = true;
  panelOpen3: boolean = true;
  unregisterModal = false
  errorMessage: string = "";
  isUpdateFailed: boolean = false;
  isUpdateSubmit: boolean = false;
  constructor(private fb: FormBuilder, private siteService: SiteService, private siteConfigService: SiteConfigureService, private route: ActivatedRoute,
    private router: Router, private msg: MessageService, private cdRef: ChangeDetectorRef) {
    this.getSiteConfiguration();
  }
  ngOnDestroy(): void {
    this.msg.close()
  }
  ngAfterContentChecked(): void {
    this.cdRef.detectChanges()
  }

  ngAfterViewChecked() {
    this.cdRef.detectChanges()
  }
  ngOnInit(): void {
  }

  // getSiteConfiguration is to get the site configuration
  getSiteConfiguration() {
    this.siteService.getSiteInfo()
      .subscribe((data: any) => {
        this.siteInfo = data.data;
        this.name = this.siteInfo.name;
        this.party_id = this.siteInfo.party_id;
        this.desc = this.siteInfo.description;
        this.fml_manager_connected = this.siteInfo.fml_manager_connected;
        this.fate_flow_connected = this.siteInfo.fate_flow_connected;
        this.kubeflow_connected = this.siteInfo.kubeflow_connected;
        this.site_ip = this.siteInfo.external_host;
        this.site_port = this.siteInfo.external_port;
        this.site_https_string = this.siteInfo.https ? "https" : "http";
        this.fate_flow_host_ip = this.siteInfo.fate_flow_host;
        this.http_port = this.siteInfo.fate_flow_http_port;
        this.kubeconfig = this.siteInfo.kubeflow_config.kubeconfig;
        this.access_key = this.siteInfo.kubeflow_config.minio_access_key;
        this.minio_endpoint = this.siteInfo.kubeflow_config.minio_endpoint;
        this.minio_region = this.siteInfo.kubeflow_config.minio_region;
        this.secret_key = this.siteInfo.kubeflow_config.minio_secret_key;
        this.minio_ssl_enabled = this.siteInfo.kubeflow_config.minio_ssl_enabled;
        this.fml_manager_server_name = this.siteInfo.fml_manager_server_name;
        this.fml_manager_endpoint_address = this.siteInfo.fml_manager_endpoint;
        this.parseFMLManagerEndpoint(this.fml_manager_endpoint_address);
        if (this.fml_manager_connected) {
          this.setDisbled(true)
        } else {
          this.setDisbled(false)
        }
      });
  }

  // parseFMLManagerEndpoint is to parse the endpoint tp get the ip without http or https header
  parseFMLManagerEndpoint(fml_manager_endpoint_address: string) {
    if (fml_manager_endpoint_address != "") {
      var strList = fml_manager_endpoint_address.split("://");
      this.fml_https_string = strList[0];
      this.endpoint = strList[1];
    }
  }

  // enableFMLManagerServerName is the flag to alert user to input the 'fml manager server name' when user choose to connect fml manager via https
  get enableFMLManagerServerName() {
    return this.fml_https_string === "https"
  }

  // fmlManagerEndpointIsIP is to validate is fml manager endpoint address is IP address or FQDN (if is IP, alert user to input the 'fml manager server name')
  get fmlManagerEndpointIsIP() {
    const reg = /((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}/g
    return this.fml_manager_server_name === "" && reg.test(this.endpoint?.trim())
  }

  // registerToFMLManager is to send request to register to fml manager
  registerToFMLManager() {
    this.isRegisterToFMLManagerSubmitted = true;
    // check if the form is change, if yes, alert user to save change first
    this.formOnChange();
    if (this.isFormChanged) {
      this.openModal = true;
      return;
    }
    this.fml_manager_endpoint_address = this.endpoint === "" ? "" : this.fml_https_string + "://" + this.endpoint
    var connectInfo = {
      endpoint: this.fml_manager_endpoint_address?.replace(/\s/g, ""),
      server_name: this.fml_manager_server_name?.trim()
    }
    this.siteConfigService.connectFML(connectInfo)
      .subscribe(
        data => {
          this.isConnectFailed = false;
          this.reloadCurrentRoute();
        },
        err => {
          this.errorMessage = err.error.message;
          this.isConnectFailed = true;
        });
  }

  // unregisterToFMLManager is to send a deregistration request to the fml manager
  unregisterToFMLManager() {
    this.siteConfigService.unregisterFML().subscribe(
      data => {
        this.reloadCurrentRoute();
      },
      err => {
        this.errorMessage = err.error.message;
        this.unregisterModal = false
        this.isConnectFailed = true;
        this.isRegisterToFMLManagerSubmitted = true;
      }
    )
  }

  testFATEFlowSuccess = false
  // testFATEFlow is to test the FATEFlow connection
  testFATEFlow() {
    this.isTestFATEFlowSubmit = true
    this.isTestFATEFlowFailed = false
    this.testFATEFlowSuccess = false
    var https = false;
    this.siteConfigService.testFATEFlow(this.fate_flow_host_ip.replace(/\s/g, ""), https, Number(this.http_port))
      .subscribe(
        data => {
          this.testFATEFlowSuccess = true
          this.isTestFATEFlowSubmit = false
          this.msg.success('serverMessage.default200', 1000)
        },
        err => {
          this.errorMessage = err.error.message;
          this.isTestFATEFlowFailed = true;
          this.isTestFATEFlowSubmit = false
        });
  }

  isTestKubeFlowSubmit = false
  isTestKubeFlowFailed = false
  isTestKubeFlowSuccess = false
  // testKubeFlow is to test the connection with KubeFlow
  testKubeFlow() {
    this.isTestKubeFlowSubmit = true
    this.isTestKubeFlowFailed = false
    this.isTestKubeFlowSuccess = false
    this.siteConfigService.testKubeFlow(this.kubeconfig, this.access_key.replace(/\s/g, ""), this.minio_endpoint.replace(/\s/g, ""), this.minio_region, this.secret_key.replace(/\s/g, ""), this.minio_ssl_enabled)
      .subscribe(
        data => {
          this.msg.success('serverMessage.default200', 1000)
          this.isTestKubeFlowSuccess = true
          this.isTestKubeFlowSubmit = false
        },
        err => {
          this.errorMessage = err.error.message
          this.isTestKubeFlowFailed = true
          this.isTestKubeFlowSubmit = false
        });
  }

  isFormChanged: boolean = false;
  openModal: boolean = false;
  // formOnChange is to detect if the form is changed
  formOnChange() {
    if (this.name != this.siteInfo.name || this.party_id != this.siteInfo.party_id || this.site_ip != this.siteInfo.external_host || this.site_port != this.siteInfo.external_port) {
      this.isFormChanged = true;
    }
    if ((this.siteInfo.https && this.site_https_string === "http") || (!this.siteInfo.https && this.site_https_string === "https")) {
      this.isFormChanged = true;
    }
    this.fml_manager_endpoint_address = this.endpoint === "" ? "http://" : this.fml_https_string + "://" + this.endpoint
  }

  // saveSiteConfigUpdate is to save the updated site config
  saveSiteConfigUpdate() {
    this.isUpdateFailed = false;
    this.isUpdateSubmit = true;
    if (!this.form.valid) {
      this.errorMessage = "Invalid input";
      this.isUpdateFailed = true;
      return;
    }
    this.siteUpdatedInfo.name = this.name;
    this.siteUpdatedInfo.party_id = Number(this.party_id);
    this.siteUpdatedInfo.external_host = this.site_ip.replace(/\s/g, "");
    this.siteUpdatedInfo.external_port = Number(this.site_port);
    this.siteUpdatedInfo.https = this.site_https_string === "https" ? true : false;
    this.siteUpdatedInfo.description = this.desc;
    this.fml_manager_endpoint_address = this.endpoint === "" ? "" : this.fml_https_string + "://" + this.endpoint
    this.siteUpdatedInfo.fml_manager_endpoint = this.fml_manager_endpoint_address.replace(/\s/g, "");
    this.siteUpdatedInfo.fml_manager_server_name = this.fml_manager_server_name.trim();
    this.siteUpdatedInfo.fate_flow_host = this.fate_flow_host_ip.replace(/\s/g, "");
    this.siteUpdatedInfo.fate_flow_http_port = Number(this.http_port);
    this.siteUpdatedInfo.kubeflow_config.minio_endpoint = this.minio_endpoint.replace(/\s/g, "");
    this.siteUpdatedInfo.kubeflow_config.minio_access_key = this.access_key.replace(/\s/g, "");
    this.siteUpdatedInfo.kubeflow_config.minio_secret_key = this.secret_key.replace(/\s/g, "");
    this.siteUpdatedInfo.kubeflow_config.kubeconfig = this.kubeconfig;
    this.siteUpdatedInfo.kubeflow_config.minio_ssl_enabled = this.minio_ssl_enabled;
    this.siteConfigService.putConfigUpdate(this.siteUpdatedInfo)
      .subscribe(data => {
        this.msg.success('serverMessage.modify200', 1000)
        this.reloadCurrentRoute();
      },
        err => {
          this.errorMessage = err.error.message;
          this.isUpdateFailed = true;
        });
  }

  // setDisbled is to disable the selected input container (when site is successfully connected to fml manager, disabled the field below)
  setDisbled(opt: boolean) {
    var keys = ["site_https_string", "site_ip", "site_port"]
    for (const key of keys) {
      const val = this.form.get(key)
      if (val) {
        if (opt) {
          val.disable()
        } else {
          val.enable()
        }
      }
    }
  }
  
  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

}
