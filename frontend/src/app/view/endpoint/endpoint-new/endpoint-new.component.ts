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
import { Router } from '@angular/router';
import { EndpointService } from 'src/app/services/common/endpoint.service'
import { ValidatorGroup } from 'src/utils/validators';
import { InfraService } from 'src/app/services/common/infra.service';
import * as CodeMirror from 'codemirror'
import { EndpointType } from 'src/app/view/endpoint/endpoint-model'
import { InfraType } from 'src/app/view/infra/infra-model'


@Component({
  selector: 'app-endpoint-new',
  templateUrl: './endpoint-new.component.html',
  styleUrls: ['./endpoint-new.component.scss']
})
export class EndpointNewComponent implements OnInit {
  form: FormGroup;
  constructor(private formBuilder: FormBuilder, private endpointService: EndpointService, private infraservice: InfraService, private router: Router) {
    //form is the form to create a new endpoint
    this.form = this.formBuilder.group({
      infra: this.formBuilder.group({
        infra_object: [],
        infra_name: ['']
      }),
      endpoint: this.formBuilder.group({
        endpoint_type: [],
        endpoint_namespace: ['']
      }),
      install: this.formBuilder.group(
        ValidatorGroup([
          {
            name: 'name',
            value: '',
            type: ['word'],
            max: 20,
            min: 2
          },
          {
            name: 'description',
            type: ['']
          },
          {
            name: 'yaml',
            type: ['']
          },
          {
            name: 'service_username',
            type: [''],
            value: 'admin'
          },
          {
            name: 'service_password',
            type: [''],
            value: 'admin'
          },
          {
            name: 'hostname',
            type: [''],
            value: 'kubefate.net'
          },
          {
            name: 'need_ingress',
            type: ['']
          },
          {
            name: 'ingress_controller_service_mode',
            type: [''],
            value: [0]
          },
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
      )
    });
  }

  ngOnInit(): void {
    this.getInfraList()
  }
  //needAdd is true when there is an endpoint on the selected infra but not recorded in the lifecycle manager
  needAdd: boolean = false;
  scanList: EndpointType[] = []
  message = 'Service Error!'
  type: 'success' | 'info' | 'warning' | 'danger' = 'danger'
  questResultFalg = false
  infra!: InfraType | null
  endpointNamespaceList: string[] = []
  //selectedInfra is the selected infra in the 'Step 1'
  get selectedInfra() {
    return this.infra
  }
  set selectedInfra(value) {
    this.infra = value
    if (value) {
      this.onSelectInfra()
    }
  }

  endpoint!: EndpointType | null
  //selectedEndpoint is selected endpoint in the 'Step 2' when there is scaned endpoint list.
  get selectedEndpoint() {
    return this.endpoint
  }
  set selectedEndpoint(value) {
    this.endpoint = value
    if (!this.endpoint?.is_managed && this.endpoint?.is_compatible) {
      this.needAdd = true;
      this.needInstall = false;
    } else {
      this.needAdd = false
      this.needInstall = false;
    }
  }
  // endpoint config disable function
  get endpointDisabled() {
    if (this.endpointNamespaceList && this.endpointNamespaceList.length < 1) {
      if (this.needAdd || this.needInstall) {
        return false
      } else {
        return true
      }
      
    } else {
      const namespace = this.form.get('endpoint')?.get('endpoint_namespace')?.value
      if (namespace) {
        if (this.needAdd && this.needInstall) {
          return true
        } else {
          return false
        }
      } else {
        return true
      }
    }
  }
  // is show Install an Ingress Controller for me button
  get showIngressControllerService() {
    
    if (this.endpointNamespaceList && this.endpointNamespaceList.length > 0) {      
      return false
    } else {
      return !this.needAdd
    }
  }
  infralist = [];
  noInfra = false;
  //getInfraList is to get the infra list in the 'Step 1'
  getInfraList() {
    this.infraservice.getInfraList().subscribe(
      data => {
        this.infralist = data.data;
        if (this.infralist.length === 0) {
          this.noInfra = true;
        }
      },
      err => {
        this.infralist = []
        this.type = 'danger'
        if (err.error.message) this.message = err.error.message
        this.questResultFalg = true
      }
    )
  }

  needInstall = false;
  scanLoading = false;
  //postEndpointScan is to post the request to scan the endpoint on the selected infra
  postEndpointScan(type: any) {
    if (this.infra) {
      
      // must select namespace options
      if (this.endpointNamespaceList && this.endpointNamespaceList.length >0 ) {
        if (this.form.get('endpoint')?.get('endpoint_type')?.value && this.form.get('endpoint')?.get('endpoint_namespace')?.value) {
          this.scanLoading = true;
          const data = {
            infra_provider_uuid: this.infra.uuid,
            type: this.form.get('endpoint')?.get('endpoint_type')?.value,
            namespace: this.form.get('endpoint')?.get('endpoint_namespace')?.value
          }
          this.endpointService.postEndpointScan(data).subscribe(
            (data: any) => {
              this.scanList = data.data
              if (this.scanList === null || this.scanList.length === 0) {
                this.scanList = []
                this.needInstall = true;
                this.needAdd = false;
              } else {
                this.selectedEndpoint = this.scanList[0];
              }
              this.scanLoading = false
            },
            err => {
              this.type = 'danger'
              this.message = err.error.message;
              this.questResultFalg = true
              this.scanLoading = false
            }
          )
        }
      } else {
        this.scanLoading = true;
        const data = {
          infra_provider_uuid: this.infra.uuid,
          type: this.form.get('endpoint')?.get('endpoint_type')?.value
        }
        this.endpointService.postEndpointScan(data).subscribe(
          (data: any) => {
            this.scanList = data.data
            if (this.scanList === null || this.scanList.length === 0) {
              this.scanList = []
              this.needInstall = true;
              this.needAdd = false;
            } else {
              this.selectedEndpoint = this.scanList[0];
            }
            this.scanLoading = false
          },
          err => {
            this.type = 'danger'
            this.message = err.error.message;
            this.questResultFalg = true
            this.scanLoading = false
          }
        )
      }

    }
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
      this.form.get('install')?.get('yaml')?.setValue(cm.getValue())
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
    return this.codeMirror && !this.hasYAMLTextAreaDOM && this.form.controls['install'].get('yaml')?.value
  }

  //onSelectInfra is to reset form when selection change
  infraSelectOk: boolean = false;
  onSelectInfra() {
    if (this.selectedInfra) {
      this.infraSelectOk = true;
      this.form.get('infra')?.get('infra_name')?.setValue(this.selectedInfra.name);
      this.form.get('infra')?.get('infra_object')?.setValue(this.selectedInfra);
      this.formResetChild('endpoint');
      this.formResetChild('install');
      this.form.get('install')?.get('service_username')?.setValue('admin')
      this.form.get('install')?.get('service_password')?.setValue('admin')
      this.form.get('install')?.get('hostname')?.setValue('kubefate.net')
      this.showInfraDetail(this.selectedInfra.uuid)
      this.selectedEndpoint = null
      this.needAdd = false;
      this.needInstall = false;
      this.scanList = [];
      this.isGenerateSubmit = false;
      this.isGenerateFailed = false;
      this.isCreateEndpointFailed = false;
      this.isCreateEndpointSubmit = false;
      this.scanLoading = false;
      this.noInfra = false;
      this.message = "";
    }
  }

  infraConfigDetail: any;
  isShowInfraDetailFailed: boolean = false;
  hasRegistry = false;
  hasRegistrySecret = false;
  registrySecretConfig: any;
  errorMessage = ""
  //showInfraDetail is to get the infra detail to get the saved registry/registry secret configuration
  showInfraDetail(uuid: string) {
    this.isShowInfraDetailFailed = false;
    this.infraservice.getInfraDetail(uuid)
      .subscribe((data: any) => {        
        this.infraConfigDetail = data.data.kubernetes_provider_info;
        this.endpointNamespaceList = data.data.kubernetes_provider_info?.namespaces_list || []
        this.hasRegistry = this.infraConfigDetail.registry_config_fate.use_registry
        this.hasRegistrySecret = this.infraConfigDetail.registry_config_fate.use_registry_secret
        this.registrySecretConfig = this.infraConfigDetail.registry_config_fate.registry_secret_config
        this.form.get('install')?.get('useRegistrySecret')?.setValue(this.hasRegistrySecret);
        this.form.get('install')?.get('useRegistry')?.setValue(this.hasRegistry);
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
    return this.form.controls['install'].get('useRegistrySecret')?.value;
  }
  get useRegistry() {
    return this.form.controls['install'].get('useRegistry')?.value;
  }
  get registryConfig() {
    return this.form.controls['install'].get('registryConfig')?.value;
  }
  //onChange_use_registry is triggered when the value of 'use_registry' is changed
  onChange_use_registry(val: any) {
    this.change_use_registry()
    this.selectChange(val)
  }
  change_use_registry() {
    if (this.useRegistry) {
      this.form.get('install')?.get('registryConfig')?.setValue(this.infraConfigDetail.registry_config_fate.registry);
    } else {
      this.form.get('install')?.get('registryConfig')?.setValue("-");
    }
  }

  //onChange_use_registry_secret is triggered when the value of 'use_registry_secret' is changed
  onChange_use_registry_secret(val: any) {
    this.change_registry_secret()
    this.selectChange(val)
  }
  change_registry_secret() {
    if (this.useRegistrySecret) {
      this.form.get('install')?.get('server_url')?.setValue(this.registrySecretConfig.server_url);
      this.form.get('install')?.get('username')?.setValue(this.registrySecretConfig.username);
      this.form.get('install')?.get('password')?.setValue(this.registrySecretConfig.password);
    }
    else {
      this.form.get('install')?.get('server_url')?.setValue("https://x");
      this.form.get('install')?.get('username')?.setValue("-");
      this.form.get('install')?.get('password')?.setValue("-");
    }
  }

  //registry_disabled is to valid if the registry/registry secret configuration are invalid
  get registry_disabled() {
    var registry_secret_valid: any = true;
    if (this.useRegistrySecret) {
      registry_secret_valid = (this.form.controls['install'].get('username')?.value?.trim() != '') && (this.form.controls['install'].get('password')?.value?.trim() != '') && this.form.get('install.server_url')?.valid && (this.form.controls['install'].get('server_url')?.value?.trim() != '')
    } else {
      registry_secret_valid = true
    }
    var registry_valid = true;
    if (this.useRegistry) {
      registry_valid = this.form.controls['install'].get('registryConfig')?.value?.trim() != ''
    } else {
      registry_valid = true
    }
    return !(registry_secret_valid && registry_valid)
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
    return this.form.get('install.server_url')?.valid && this.form.controls['install'].get('server_url')?.value?.trim() != ''
  }

  isGenerateSubmit = false;
  isGenerateFailed = false;
  codeMirror: any
  //generateYaml() is to get the endpoint installation yaml based on the configuration provided by user
  generateYaml() {
    this.isGenerateSubmit = true;
    const yamlHTML = document.getElementById('yaml') as any
    if (!this.needAdd) {
      var service_username = this.form.controls['install'].get('service_username')?.value;
      var service_password = this.form.controls['install'].get('service_password')?.value;
      var hostname = this.form.controls['install'].get('hostname')?.value;
      var registry_server_url = this.useRegistrySecret ? this.form.controls['install'].get('server_url')?.value?.trim() : "";
      var registry_username = this.useRegistrySecret ? this.form.controls['install'].get('username')?.value?.trim() : "";
      var registry_password = this.useRegistrySecret ? this.form.controls['install'].get('password')?.value?.trim() : "";
      const namespace = this.form.get('endpoint')?.get('endpoint_namespace')?.value
      this.endpointService.getKubefateYaml(service_username, service_password, hostname, this.useRegistry, this.registryConfig, this.useRegistrySecret, registry_server_url, registry_username, registry_password, namespace).subscribe(
        data => {
          this.form.get('install')?.get('yaml')?.setValue(data.data);
          this.isGenerateFailed = false;
          if (this.codeMirror && this.hasYAMLTextAreaDOM) {
            this.codeMirror.setValue(data.data)
          } else {
            this.initCodeMirror(data.data)
          }
        },
        err => {
          this.message = err.error.message;
          this.isGenerateFailed = true;
        }
      )
    }
  }

  //selectChange is to reset YAML value when the releted confiuration is changed
  selectChange(val: any) {
    this.form.get('install')?.get('yaml')?.setValue(null);
    if (this.codeMirror) {
      this.codeMirror.setValue('')
    }

    this.isGenerateSubmit = false;
    this.isGenerateFailed = false;
  }

  //formResetChild is to reset group form, child is the group name in the form
  formResetChild(child: string) {
    //need to reset twice due to the issues of Clarity Steppers
    this.form.controls[child].reset();
    this.form.controls[child].reset();
  }

  //generateYamlDisabled is to disabled the generate Yaml buttom when requied config provided by user id not vaild
  get generateYamlDisabled() {
    return !(this.form.controls['install'].get('service_username')?.value && this.form.controls['install'].get('service_password')?.value && this.form.controls['install'].get('hostname')?.value)
  }

  //onSeclectIngress is triggered when the checkbox of ingress controller is selected
  onSeclectIngress() {
    if (!this.form.get('install')?.get('need_ingress')?.value) {
      this.form.get('install')?.get('ingress_controller_service_mode')?.setValue(0);
    }
  }
  get needIngress() {
    return this.form.get('install')?.get('need_ingress')?.value
  }

  isCreateEndpointFailed = false;
  isCreateEndpointSubmit = false;
  //createEndpoint is to submit 'create endpoint' request
  createEndpoint() {
    // this.isCreateEndpointFailed = false;
    // this.isCreateEndpointSubmit = true;
    if (this.form.valid) {
      const endpointConfig: any = {
        description: this.form.get('install')?.get('description')?.value,
        infra_provider_uuid: this.infra?.uuid,
        install: this.needInstall,
        kubefate_deployment_yaml: this.form.get('install')?.get('yaml')?.value,
        name: this.form.get('install')?.get('name')?.value,
        type: this.form.get('endpoint')?.get('endpoint_type')?.value,
        ingress_controller_service_mode: Number(this.form.get('install')?.get('ingress_controller_service_mode')?.value),
      }
      
      if (this.form.get('endpoint')?.get('endpoint_namespace')?.value) {
        endpointConfig.namespace = this.form.get('endpoint')?.get('endpoint_namespace')?.value
      }
      // validate
      if (this.needInstall && endpointConfig.kubefate_deployment_yaml === '') {
        this.message = "Kubefate Deployment YAML is required.";
        this.isCreateEndpointFailed = true;
      } else if (endpointConfig.type != 'KubeFATE') {
        this.message = "Invalid endpoint type";
        this.isCreateEndpointFailed = true;
      } else {
        this.endpointService.createEndpoint(endpointConfig).subscribe(
          (data: any) => {
            this.router.navigateByUrl('/endpoint')
          },
          err => {
            this.message = err.error.message;
            this.isCreateEndpointFailed = true;
          }
        )
      }
    }
  }
  //createDisabled is to valid all the input and returns if create buttom need to be disabled
  get createDisabled() {
    var mustHave = this.form.controls['install'].get('name')?.value
    var notUseIngress = !this.form.get('install')?.get('need_ingress')?.value && (Number(this.form.get('install')?.get('ingress_controller_service_mode')?.value) === 0)
    var useIngress = this.form.get('install')?.get('need_ingress')?.value &&
      ((Number(this.form.get('install')?.get('ingress_controller_service_mode')?.value) === 1) || (
        Number(this.form.get('install')?.get('ingress_controller_service_mode')?.value) === 2))
    return !(mustHave && (this.needAdd || (!this.generateYamlDisabled && this.form.controls['install'].get('yaml')?.value && (notUseIngress || useIngress)))) || (this.isCreateEndpointSubmit && !this.isCreateEndpointFailed)
  }
}
