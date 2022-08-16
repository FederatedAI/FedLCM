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
import { FormBuilder } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ValidatorGroup } from 'src/utils/validators';
import { InfraService } from 'src/app/services/common/infra.service';

@Component({
  selector: 'app-infra-detail',
  templateUrl: './infra-detail.component.html',
  styleUrls: ['./infra-detail.component.scss']
})
export class InfraDetailComponent implements OnInit {

  //updateInfraForm is the form to update the infra information
  updateInfraForm = this.fb.group(
    ValidatorGroup([
      {
        name: 'infraname',
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
        name: 'type',
        type: ['require']
      },
      {
        name: 'kubeconfig',
        value: '',
        type: ['require']
      },
      {
        name: 'use_registry',
        value: false,
        type: ['']
      },
      {
        name: 'registry',
        value: '',
        type: ['']
      },
      {
        name: 'use_registry_secret',
        value: false,
        type: ['']
      },
      {
        name: 'server_url',
        value: '',
        type: ['internet']
      },
      {
        name: 'username',
        value: '',
        type: ['']
      },
      {
        name: 'password',
        value: '',
        type: ['']
      }
    ])

  )

  constructor(private infraservice: InfraService, private router: Router, private route: ActivatedRoute, private fb: FormBuilder) {
  }

  ngOnInit(): void {
    this.showInfraDetail(this.uuid)
  }

  uuid = String(this.route.snapshot.paramMap.get('id'));
  infraDetail: any;
  errorMessage = "Service Error!"
  code: any
  isShowInfraDetailFailed: boolean = false;
  isPageLoading: boolean = true;
  //showInfraDetail is to get the infra detailed information by UUID
  async showInfraDetail(uuid: string) {
    //first, get infra detail
    await new Promise((re, rj) => {
      this.isPageLoading = true;
      this.infraservice.getInfraDetail(uuid)
        .subscribe((data: any) => {
          const yamlHTML = document.getElementById('yaml') as any
          const value = data.data.kubernetes_provider_info.kubeconfig_content
          if (!this.code) {
            this.code = window.CodeMirror.fromTextArea(yamlHTML, {
              value: '',
              mode: 'yaml',
              lineNumbers: true,
              indentUnit: 1,
              lineWrapping: true,
              tabSize: 2,
              readOnly: true
            })
          }
          this.code.setValue(value)
          this.infraDetail = data.data
          this.isPageLoading = false
          this.isShowInfraDetailFailed = false
          re(this.isPageLoading)
          re(this.isShowInfraDetailFailed)
        },
          err => {
            if (err.error.message) this.errorMessage = err.error.message
            this.isPageLoading = false
            this.isShowInfraDetailFailed = true
            re(this.isPageLoading)
            re(this.isShowInfraDetailFailed)
          }
        );
    })
    //second, test infra status
    if (!this.isShowInfraDetailFailed && !this.isPageLoading) {
      this.testConnection(true)
    }

  }

  openDeleteModal: boolean = false;
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  pendingDataList: string = '';
  //openDeleteConfrimModal is to open the confirmaton modal of deletion and initialized the variables
  openDeleteConfrimModal() {
    this.isDeleteFailed = false;
    this.isDeleteFailed = false;
    this.openDeleteModal = true;
    this.isDeleteSubmit = false;
  }

  //deleteInfra is to submit 'delete infra' request
  deleteInfra() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.infraservice.deleteInfra(this.uuid)
      .subscribe(() => {
        this.router.navigate(['/infra']);
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  openModal: boolean = false;
  //onOpenModal is to open the modal of 'Update the infra info' and initialize the form with current value
  onOpenModal() {
    this.updateInfraForm.controls['infraname'].setValue(this.infraDetail.name);
    this.updateInfraForm.controls['type'].setValue(this.infraDetail.type);
    this.updateInfraForm.controls['description'].setValue(this.infraDetail.description);
    this.updateInfraForm.controls['kubeconfig'].setValue(this.infraDetail.kubernetes_provider_info.kubeconfig_content);
    this.updateInfraForm.controls['use_registry_secret'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.use_registry_secret);
    this.updateInfraForm.controls['server_url'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry_secret_config.server_url);
    this.updateInfraForm.controls['username'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry_secret_config.username);
    this.updateInfraForm.controls['password'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry_secret_config.password);
    this.updateInfraForm.controls['registry'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry);
    this.updateInfraForm.controls['use_registry'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.use_registry);
    this.openModal = true;
    this.isTestFailedModal = this.isTestFailed
    this.testPassedModal = this.testPassed
  }
  //resetModal is to reset the modal when close the modal
  resetModal() {
    this.updateInfraForm.reset();
    this.openModal = false
    this.errorMessageModal = ""
    this.isUpdateFailed = false
    this.isTestFailedModal = false
    this.testPassedModal = false
  }

  get use_registry_secret() {
    return this.updateInfraForm.get('use_registry_secret')?.value
  }
  get use_registry() {
    return this.updateInfraForm.get('use_registry')?.value
  }

  //onChange_use_registry is triggered when the value of 'use_registry' is changed
  onChange_use_registry() {
    if (this.use_registry) this.updateInfraForm.controls['registry'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry);
    if (!this.use_registry) this.updateInfraForm.controls['registry'].setValue("");
  }

  //onChange_use_registry_secret is triggered when the value of 'use_registry_secret' is changed
  onChange_use_registry_secret() {
    if (this.use_registry_secret) {
      this.updateInfraForm.controls['server_url'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry_secret_config.server_url);
      this.updateInfraForm.controls['username'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry_secret_config.username);
      this.updateInfraForm.controls['password'].setValue(this.infraDetail.kubernetes_provider_info.registry_config_fate.registry_secret_config.password);
    }
    if (!this.use_registry_secret) {
      this.updateInfraForm.controls['server_url'].setValue("");
      this.updateInfraForm.controls['username'].setValue("");
      this.updateInfraForm.controls['password'].setValue("");
    }
  }
  get registry() {
    return this.updateInfraForm.get('registry')?.value
  }

  // variable in infra detail page
  isTestFailed = false
  testPassed = false
  isTestLoading = false
  // variable in modal
  isTestFailedModal = false
  testPassedModal = true
  isTestLoadingModal = false
  errorMessageModal = ""
  //testConnection is to test the k8s connection by using kubeconfig
  testConnection(notUpdate: boolean) {
    var kubeconfigContent = ""
    if (notUpdate) {
      kubeconfigContent = this.infraDetail.kubernetes_provider_info.kubeconfig_content
      this.isTestLoading = true;
      this.isTestFailed = false;
    } else {
      kubeconfigContent = this.updateInfraForm.get('kubeconfig')?.value
      this.isTestLoadingModal = true;
      this.isTestFailedModal = false;
    }
    this.infraservice.testK8sConnection(kubeconfigContent).subscribe(() => {
      if (notUpdate) {
        this.testPassed = true
        this.isTestLoading = false
      } else {
        this.testPassedModal = true
        this.isTestLoadingModal = false
      }
    },
      err => {
        if (notUpdate) {
          this.isTestFailed = true
          this.testPassed = false
          this.errorMessage = err.error.message
          this.isTestLoading = false
        } else {
          this.isTestFailedModal = true
          this.testPassedModal = false
          this.errorMessageModal = err.error.message
          this.isTestLoadingModal = false
        }
      });
  }

  //onKubeconfigChange is triggered when the input of kubeconfig is changed
  onKubeconfigChange(val: any) {
    this.testPassedModal = false;
    this.isTestFailedModal = false;
  }

  isUpdateFailed: boolean = false;
  //updateInfra is to update the info of infra
  updateInfra() {
    this.isUpdateFailed = false;
    var infraInfo = {
      description: this.updateInfraForm.get('description')?.value,
      kubernetes_provider_info: {
        kubeconfig_content: this.updateInfraForm.get('kubeconfig')?.value,
        registry_config_fate: {
          use_registry: this.updateInfraForm.get('use_registry')?.value,
          registry: this.use_registry ? this.updateInfraForm.get('registry')?.value?.trim() : "",
          use_registry_secret: this.updateInfraForm.get('use_registry_secret')?.value,
          registry_secret_config: {
            server_url: this.use_registry_secret ? this.updateInfraForm.get('server_url')?.value?.trim() : "",
            username: this.use_registry_secret ? this.updateInfraForm.get('username')?.value?.trim() : "",
            password: this.use_registry_secret ? this.updateInfraForm.get('password')?.value?.trim() : "",
          }
        }
      },
      name: this.updateInfraForm.get('infraname')?.value,
      type: this.updateInfraForm.get('type')?.value,
    }
    this.infraservice.updateInfraProvider(infraInfo, this.uuid)
      .subscribe(
        data => {
          this.isUpdateFailed = false;
          this.reloadCurrentRoute();
        },
        err => {
          this.errorMessage = err.error.message;
          this.isUpdateFailed = true;
        }
      );
  }

  //submitDisable returns if disabled the submit button of create a new infra
  get submitDisable() {
    var registry_secret_valid = true;
    if (this.use_registry_secret) {
      registry_secret_valid = (this.updateInfraForm.get('username')?.value?.trim() != '') && (this.updateInfraForm.get('password')?.value?.trim() != '') && !(this.updateInfraForm.controls['server_url'].errors)
    } else {
      registry_secret_valid = true
    }
    var registry_valid = true;
    if (this.use_registry) {
      registry_valid = this.updateInfraForm.get('registry')?.value?.trim() != ''
    } else {
      registry_valid = true
    }
    var basics_valid = !this.updateInfraForm.controls['infraname'].errors && !this.updateInfraForm.controls['type'].errors
    return !(this.testPassedModal && registry_secret_valid && registry_valid && basics_valid)
  }

  //refresh is for refresh button
  refresh() {
    this.showInfraDetail(this.uuid);
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
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
    if (this.registry === '' || this.registry === null) {
      url_suggestion = header + "index.docker.io/v1/"
    } else {
      var url = this.registry.split('/')[0]
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
    return !this.updateInfraForm.controls['server_url'].errors
  }

}
