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

import { Component, OnInit } from '@angular/core';
import { FormGroup, FormBuilder } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { EndpointService } from 'src/app/services/common/endpoint.service';
import { FedService } from 'src/app/services/federation-fate/fed.service';
import { ENDPOINTSTATUS, CHARTTYPE, constantGather } from 'src/utils/constant';
import { ChartService } from 'src/app/services/common/chart.service';
import { ValidatorGroup } from 'src/utils/validators'
import { InfraService } from 'src/app/services/common/infra.service';
import { EndpointType } from 'src/app/view/endpoint/endpoint-model'

@Component({
  selector: 'app-exchange-new',
  templateUrl: './exchange-new.component.html',
  styleUrls: ['./exchange-new.component.scss']
})
export class ExchangeNewComponent implements OnInit {
  form: FormGroup;

  constructor(private formBuilder: FormBuilder, private fedservice: FedService, private router: Router, private route: ActivatedRoute, private endpointService: EndpointService, private chartservice: ChartService, private infraservice: InfraService) {
    //form contains configuration to create a new exchange
    this.form = this.formBuilder.group({
      info: this.formBuilder.group({
        name: [''],
        description: [''],
        exchangeType: [''],
      }),
      external: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'trafficServerHost',
            type: ['ip'],
            value: ""
          },
          {
            name: 'trafficServerPort',
            type: ['number'],
            value: 0
          },
          {
            name: 'nginxHost',
            type: ['ip'],
            value: ""
          },
          {
            name: 'nginxPort',
            type: ['number'],
            value: 0
          }
        ])
      ),
      endpoint: this.formBuilder.group({
        endpoint_uuid: ['']
      }),
      chart: this.formBuilder.group({
        chart_uuid: [''],
      }),
      namespace: this.formBuilder.group({
        namespace: ['fate-exchange'],
      }),
      certificate: this.formBuilder.group({
        cert: ['skip'],
        fml_manager_client_cert_mode: [1],
        fml_manager_client_cert_uuid: [''],
        fml_manager_server_cert_mode: [1],
        fml_manager_server_cert_uuid: [''],
        proxy_server_cert_mode: [1],
        proxy_server_cert_uuid: ['']
      }),
      serviceType: this.formBuilder.group({
        serviceType: [null],
      }),
      registry: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'useRegistry',
            type: [''],
            value: false
          },
          {
            name: 'useRegistrySecret',
            type: [''],
            value: false
          },
          {
            name: 'registryConfig',
            type: [''],
            value: ''
          },
          {
            name: 'server_url',
            type: ['internet'],
            value: ''
          },
          {
            name: 'username',
            type: [''],
            value: ''
          },
          {
            name: 'password',
            type: [''],
            value: ''
          },
        ])
      ),
      psp: this.formBuilder.group({
        enablePSP: [false],
      }),
      yaml: this.formBuilder.group({
        yaml: [''],
      })
    });
    this.showEndpointList();
    this.showChartList();
  }
  ngOnInit(): void {
  }

  fed_uuid = String(this.route.snapshot.paramMap.get('id'));

  //support to create two type of exchange: create a new one or add an external
  get exchangeType() {
    return this.form.get('info')?.get('exchangeType')?.value
  }
  //isNewExchange returns true when user select to 'create a new one'
  get isNewExchange() {
    return this.form.get('info')?.get('exchangeType')?.value === "new"
  }
  //isExternalExchange returns true when user select to 'Add an external one'
  get isExternalExchange() {
    return this.form.get('info')?.get('exchangeType')?.value === "external"
  }
  //submitExternalExchangeDisable returns true when the input provided for adding an external exchnage is invaid
  get submitExternalExchangeDisable() {
    return !this.form.controls['external'].valid || (this.isCreatedExternalSubmit && !this.isCreatedExternalFailed)
  }

  isCreatedExternalSubmit = false;
  isCreatedExternalFailed = false;
  //createExternalExchange is to submit the request to 'create an external exchange'
  createExternalExchange() {
    this.isCreatedExternalSubmit = true;
    this.isCreatedExternalFailed = false;
    var externalExchange = {
      name: this.form.controls['info'].get('name')?.value,
      description: this.form.controls['info'].get('description')?.value,
      federation_uuid: this.fed_uuid,
      traffic_server_access_info: {
        host: this.form.get('external')?.get('trafficServerHost')?.value?.trim(),
        port: Number(this.form.get('external')?.get('trafficServerPort')?.value?.trim()),
      },
      nginx_access_info: {
        host: this.form.get('external')?.get('nginxHost')?.value?.trim(),
        port: Number(this.form.get('external')?.get('nginxPort')?.value?.trim()),
      }
    }
    if (this.form.controls['external'].valid && this.form.controls['info'].valid) {
      this.fedservice.createExternalExchange(this.fed_uuid, externalExchange)
        .subscribe(
          data => {
            this.isCreatedExternalFailed = false;
            this.router.navigateByUrl('/federation/fate/' + this.fed_uuid)
          },
          err => {
            this.errorMessage = err.error.message;
            this.isCreatedExternalFailed = true;
          }
        );
    } else {
      this.errorMessage = "invalid input";
      this.isCreatedExternalFailed = true;
    }
    return
  }

  endpointStatus = ENDPOINTSTATUS;
  constantGather = constantGather
  endpointlist: any = [];
  errorMessage: any;
  isShowEndpointFailed: boolean = false;
  isPageLoading: boolean = true;
  noEndpoint = false;
  //showEndpointList is to get the ready endpoint list
  showEndpointList() {
    this.isShowEndpointFailed = false;
    this.isPageLoading = true;
    this.endpointlist = []
    this.endpointService.getEndpointList()
      .subscribe((data: any) => {
        if (data.data) {
          for (var ep of data.data) {
            if (ep?.status === 2) this.endpointlist?.push(ep)
          }
          if (this.endpointlist.length === 0) {
            this.noEndpoint = true;
            this.isPageLoading = false;
            return
          }
        }
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowEndpointFailed = true;
          this.isPageLoading = false;
        });
  }

  endpoint!: EndpointType | null
  //selectedEndpoint is the selected endpoint in the 'Step 2'
  get selectedEndpoint() {
    return this.endpoint
  }
  set selectedEndpoint(value) {
    this.endpoint = value
    if (value) {
      this.onSelectEndpoint()
    }
  }

  use_cert: boolean = false;
  skip_cert: boolean = false;
  //setRadioDisplay is triggered when the selection of certificate's mode is changed
  setRadioDisplay(val: any) {
    if (val == "use") {
      this.use_cert = true;
      this.skip_cert = false;
    }
    if (val == "skip") {
      this.use_cert = false;
      this.skip_cert = true;
      this.form.get('certificate')?.get('fml_manager_client_cert_mode')?.setValue(1);
      this.form.get('certificate')?.get('fml_manager_server_cert_mode')?.setValue(1);
      this.form.get('certificate')?.get('proxy_server_cert_mode')?.setValue(1);
    }
  }

  //cert_disabled is to validate the configuration of certificate section and disabled the 'Next' button if needed
  get cert_disabled() {
    var case1 = this.use_cert && this.isChartContainsPortalservices && (this.form.controls['certificate'].get('proxy_server_cert_mode')?.value === 1 ||
      this.form.controls['certificate'].get('fml_manager_server_cert_mode')?.value === 1
      || this.form.controls['certificate'].get('fml_manager_client_cert_mode')?.value === 1)
    var case2 = this.use_cert && !this.isChartContainsPortalservices && this.form.controls['certificate'].get('proxy_server_cert_mode')?.value === 1
    return (case1 || case2)
  }

  //service_type_disabled is to validate the configuration of Service type section and disabled the 'Next' button if needed
  get service_type_disabled() {
    return !this.form.controls['serviceType'].get('serviceType')?.valid
  }

  //reset form when selection change
  endpointSelectOk: boolean = false;
  infraUUID: any = ""
  onSelectEndpoint() {
    if (this.selectedEndpoint) {
      this.endpointSelectOk = true;
      this.form.get('endpoint')?.get('endpoint_uuid')?.setValue(this.selectedEndpoint.uuid);
      this.infraUUID = this.endpoint?.infra_provider_uuid
      this.showInfraDetail(this.infraUUID)
    }
  }

  infraConfigDetail: any;
  isShowInfraDetailFailed: boolean = false;
  hasRegistry = false;
  hasRegistrySecret = false;
  registrySecretConfig: any;
  //showInfraDetail is to get the registry/registry secret information that saved on the infra of the selected endpoint
  showInfraDetail(uuid: string) {
    this.isShowInfraDetailFailed = false;
    this.infraservice.getInfraDetail(uuid)
      .subscribe((data: any) => {
        this.infraConfigDetail = data.data.kubernetes_provider_info;
        this.hasRegistry = this.infraConfigDetail.registry_config_fate.use_registry
        this.hasRegistrySecret = this.infraConfigDetail.registry_config_fate.use_registry_secret
        this.registrySecretConfig = this.infraConfigDetail.registry_config_fate.registry_secret_config
        this.form.get('registry')?.get('useRegistrySecret')?.setValue(this.hasRegistrySecret);
        this.form.get('registry')?.get('useRegistry')?.setValue(this.hasRegistry);
        this.change_use_registry()
        this.change_registry_secret()
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowInfraDetailFailed = true;
        }
      );
  }

  get useRegistrySecret() {
    return this.form.controls['registry'].get('useRegistrySecret')?.value;
  }
  get useRegistry() {
    return this.form.controls['registry'].get('useRegistry')?.value;
  }
  get registryConfig() {
    return this.form.controls['registry'].get('registryConfig')?.value;
  }

  //onChange_use_registry is triggered when the value of 'useRegistry' is changed
  onChange_use_registry(val: any) {
    this.change_use_registry()
    this.selectChange(val)
  }
  change_use_registry() {
    if (this.useRegistry) {
      this.form.get('registry')?.get('registryConfig')?.setValue(this.infraConfigDetail.registry_config_fate.registry);
    } else {
      this.form.get('registry')?.get('registryConfig')?.setValue("-");
    }
  }

  //onChange_use_registry_secre is triggered when the value of 'useRegistrySecret' is changed
  onChange_use_registry_secret(val: any) {
    this.change_registry_secret()
    this.selectChange(val)
  }
  change_registry_secret() {
    if (this.useRegistrySecret) {
      this.form.get('registry')?.get('server_url')?.setValue(this.registrySecretConfig.server_url);
      this.form.get('registry')?.get('username')?.setValue(this.registrySecretConfig.username);
      this.form.get('registry')?.get('password')?.setValue(this.registrySecretConfig.password);
    }
    else {
      this.form.get('registry')?.get('server_url')?.setValue("https://x");
      this.form.get('registry')?.get('username')?.setValue("-");
      this.form.get('registry')?.get('password')?.setValue("-");
    }
  }

  //registry_disabled is to valid if the registry/registry secret configuration are invalid
  get registry_disabled() {
    var registry_secret_valid: any = true;
    if (this.useRegistrySecret) {
      registry_secret_valid = (this.form.controls['registry'].get('username')?.value?.trim() != '') && (this.form.controls['registry'].get('password')?.value?.trim() != '') && this.form.get('registry.server_url')?.valid && (this.form.controls['registry'].get('server_url')?.value?.trim() != '')
    } else {
      registry_secret_valid = true
    }
    var registry_valid = true;
    if (this.useRegistry) {
      registry_valid = this.form.controls['registry'].get('registryConfig')?.value?.trim() != ''
    } else {
      registry_valid = true
    }
    return !(registry_secret_valid && registry_valid)
  }

  //formResetChild is to reset group form, child is the group name in the form
  formResetChild(child: string) {
    //need to reset twice due to the issues of Clarity Steppers
    this.form.controls[child].reset();
    this.form.controls[child].reset();
    if (child === 'yaml' && this.codeMirror) {
      this.codeMirror.setValue('')
    }
  }

  chartType = CHARTTYPE;
  chartlist: any = [];
  isShowChartFailed: boolean = false;
  //showChartList is to get the chart list of deloying exchange
  showChartList() {
    this.chartlist = [];
    this.isPageLoading = true;
    this.isShowChartFailed = false;
    this.chartservice.getChartList()
      .subscribe((data: any) => {
        if (data.data) {
          for (let chart of data.data) {
            if (chart.type === 1) {
              this.chartlist.push(chart);
            }
          }
        }
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowChartFailed = true;
          this.isPageLoading = false;
        });
  }

  isChartContainsPortalservices = false;
  //onChartChange is triggered when the selection of chart is changed
  onChartChange(val: any) {
    var chart_uuid = this.form.controls['chart'].get('chart_uuid')?.value;
    this.chartservice.getChartDetail(chart_uuid)
      .subscribe((data: any) => {
        if (data.data) {
          //isChartContainsPortalservices to decide the certificate we need when creating the exchange
          this.isChartContainsPortalservices = data.data.contain_portal_services;
        }
        this.isPageLoading = false;
        this.isShowChartFailed = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowChartFailed = true;
          this.isPageLoading = false;
        });
    this.selectChange(val);
    this.resetCert();
  }

  //enablePSP returns if enable the pod secrurity policy
  get enablePSP() {
    return this.form.controls['psp'].get('enablePSP')?.value;
  }

  //resetCert is triggered when the selection of chart is changed
  resetCert() {
    this.formResetChild('certificate');
    this.form.get('certificate')?.get('cert')?.setValue("skip");
    this.setRadioDisplay("skip");
  }

  //selectChange is to reset YAML value when the releted confiuration is changed
  selectChange(val: any) {
    this.formResetChild('yaml');
    this.isGenerateSubmit = false;
    this.isGenerateFailed = false;
  }

  isGenerateSubmit = false;
  isGenerateFailed = false;
  codeMirror: any
  //generateYaml is to get the exchange initial yaml based on the configuration provided by user
  generateYaml() {
    this.isGenerateSubmit = true;
    var chart_id = this.form.controls['chart'].get('chart_uuid')?.value;
    var namespace = this.form.controls['namespace'].get('namespace')?.value;
    var name = this.form.controls['info'].get('name')?.value;
    var service_type = Number(this.form.controls['serviceType'].get('serviceType')?.value);
    if (namespace === '') namespace = 'fate-exchange';
    this.fedservice.getExchangeYaml(chart_id, namespace, name, service_type, this.registryConfig, this.useRegistry, this.useRegistrySecret, this.enablePSP).subscribe(
      data => {
        this.form.get('yaml')?.get('yaml')?.setValue(data.data);
        // if code mirror object and yaml DOM are existing, just set value to the code mirror object
        // else initialize code mirror object and yaml editor window
        if (this.codeMirror && this.hasYAMLTextAreaDOM) {
          this.codeMirror.setValue(data.data)
        } else {
          this.initCodeMirror(data.data)
        }
        this.isGenerateFailed = false;
      },
      err => {
        this.errorMessage = err.error.message;
        this.isGenerateFailed = true;
      }
    )
  }

  // initYAMLEditorByEvent is triggered when clicking the yaml textarea(if there is no highlight) to reinitialize the YAML editor
  initYAMLEditorByEvent(event: any) {
    if (event && event.target && event.target.value) {
      this.isGenerateSubmit = true;
      this.initCodeMirror(event.target.value)
    }
  }

  // initCodeMirror is to initialize Code Mirror YAML editor
  initCodeMirror(yamlContent: any) {
    const yamlHTML = document.getElementById('yaml') as any
    this.codeMirror = window.CodeMirror.fromTextArea(yamlHTML, {
      value: yamlContent,
      mode: 'yaml',
      lineNumbers: true,
      indentUnit: 1,
      lineWrapping: true,
      tabSize: 2,
    })
    this.codeMirror.on('change', (cm: any) => {
      this.codeMirror.save()
    })
    this.hasYAMLTextAreaDOM = true
  }

  hasYAMLTextAreaDOM = false
  // onClickNext is triggered when clicking the 'Next' button before the step 'YAML'
  onClickNext() {
    this.hasYAMLTextAreaDOM = false;
    // when the form is reset by the clarity stepper component and the step 'YAML' is collapsed, the DOM of step 'YAML' will be cleared. 
    // So we need to check if we need to initialized code mirror window in the next step (step 'YAML)
    if (document.getElementById('yaml') !== null) {
      this.hasYAMLTextAreaDOM = true;
    }
  }

  get editorModeHelperMessage() {
    return this.codeMirror && !this.hasYAMLTextAreaDOM && this.form.controls['yaml'].get('yaml')?.value
  }
  //submitDisable returns if disabled the submit button of create a new exchange
  get submitDisable() {
    if (!this.form.controls['info'].valid || this.cert_disabled || this.service_type_disabled || !this.isGenerateSubmit || (this.isCreatedSubmit && !this.isCreatedFailed)) {
      return true
    } else {
      return false
    }
  }
  isCreatedSubmit = false;
  isCreatedFailed = false;
  //createNewExchange is to submmit the request of 'create a new exchange'
  createNewExchange() {
    this.isCreatedFailed = false;
    this.isCreatedSubmit = true;
    const exchangeInfo = {
      chart_uuid: this.form.controls['chart'].get('chart_uuid')?.value,
      deployment_yaml: this.codeMirror.getTextArea().value,
      description: this.form.controls['info'].get('description')?.value,
      endpoint_uuid: this.form.controls['endpoint'].get('endpoint_uuid')?.value,
      federation_uuid: this.fed_uuid,
      registry_config: {
        use_registry: this.useRegistry,
        use_registry_secret: this.useRegistrySecret,
        registry: this.useRegistry ? this.registryConfig : "",
        registry_secret_config: {
          server_url: this.useRegistrySecret ? this.form.controls['registry'].get('server_url')?.value?.trim() : "",
          username: this.useRegistrySecret ? this.form.controls['registry'].get('username')?.value?.trim() : "",
          password: this.useRegistrySecret ? this.form.controls['registry'].get('password')?.value?.trim() : "",
        }
      },
      fml_manager_client_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('fml_manager_client_cert_mode')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('fml_manager_client_cert_uuid')?.value,
      },
      fml_manager_server_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('fml_manager_server_cert_mode')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('fml_manager_server_cert_uuid')?.value
      },
      name: this.form.controls['info'].get('name')?.value,
      namespace: this.form.controls['namespace'].get('namespace')?.value,
      proxy_server_cert_info: {
        binding_mode: Number(this.form.controls['certificate'].get('proxy_server_cert_mode')?.value),
        common_name: "",
        uuid: this.form.controls['certificate'].get('proxy_server_cert_uuid')?.value
      }
    }
    this.fedservice.createExchange(this.fed_uuid, exchangeInfo)
      .subscribe(
        data => {
          this.isCreatedFailed = false;
          this.router.navigateByUrl('/federation/fate/' + this.fed_uuid)
        },
        err => {
          this.errorMessage = err.error.message;
          this.isCreatedFailed = true;
        }
      );
  }

  //validURL is to validate URL is valid
  validURL(str: string) {
    var pattern = new RegExp(
      '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.)+[a-z]{2,}|' + // domain name
      '((\\d{1,3}\\.){3}\\d{1,3}))' + // OR ip (v4) address
      '(\\?[;&a-z\\d%_.~+=-]*)?' + // query string
      '(\\#[-a-z\\d_]*)?$', 'i'); // fragment locator
    return !!pattern.test(str);
  }

  //server_url_suggestion to return the suggested server url based on the provided registry
  get server_url_suggestion() {
    var url_suggestion = "";
    var header = "https://"
    if (this.registryConfig === "" || this.registryConfig === null || this.registryConfig === "-") {
      url_suggestion = header + "index.docker.io/v1/"
    } else {
      var url = this.registryConfig.split('/')[0]
      if (this.validURL(url)) {
        url_suggestion = header + url
      } else {
        url_suggestion = header + "index.docker.io/v1/"
      }
    }
    return url_suggestion
  }

  //valid_server_url return if the the server url is valid or not
  get valid_server_url() {
    return this.form.get('registry.server_url')?.valid && this.form.controls['registry'].get('server_url')?.value?.trim() != ''
  }

}
