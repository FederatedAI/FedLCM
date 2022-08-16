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

import { Component, OnInit} from '@angular/core';
import { FormBuilder } from '@angular/forms';
import { Router } from '@angular/router';
import { InfraService } from 'src/app/services/common/infra.service';
import { ValidatorGroup } from 'src/utils/validators'
import { InfraResponse } from './infra-model';

@Component({
  selector: 'app-infra',
  templateUrl: './infra.component.html',
  styleUrls: ['./infra.component.scss']
})
export class InfraComponent implements OnInit {

  selectedInfraList: any = [];
  openModal: boolean = false;
  //newInfraForm is the form to create an new infra
  newInfraForm = this.fb.group(
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
        type: ['require']
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
  constructor(private fb: FormBuilder, private infraservice: InfraService, private router: Router) {
  }

  ngOnInit(): void {
    this.showInfraList();
    this.newInfraForm.controls['use_registry'].setValue(false);
    this.newInfraForm.controls['use_registry_secret'].setValue(false);
  }

  errorMessage = "Service Error!"
  isShowInfraFailed: boolean = false;
  infralist: any;
  isPageLoading: boolean = true;
  //showInfraList is to get the infra list
  showInfraList() {
    this.isShowInfraFailed = false;
    this.isPageLoading = true;
    this.infraservice.getInfraList()
      .subscribe((data: InfraResponse) => {
        this.infralist = data.data;
        this.isPageLoading = false;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isShowInfraFailed = true
          this.isPageLoading = false
        });
  }

  //onOpenModal is to open the modal of 'Create an new infra'
  onOpenModal() {
    this.openModal = true;
  }
  //resetModal() is to close the modal of 'Create an new infra' and reset the modal of 'Create an new infra' 
  resetModal() {
    this.newInfraForm.reset();
    this.openModal = false;
    this.testPassed = false;
    this.isTestFailed = false;
    this.isCreatedFailed = false;
    this.errorMessage = "";
    this.newInfraForm.controls['use_registry'].setValue(false);
    this.newInfraForm.controls['use_registry_secret'].setValue(false);
  }

  get use_registry_secret() {
    return this.newInfraForm.get('use_registry_secret')?.value
  }
  get use_registry() {
    return this.newInfraForm.get('use_registry')?.value
  }
  get registry() {
    return this.newInfraForm.get('registry')?.value
  }
  //onChange_use_registry is triggered when the value of 'use_registry' is changed
  onChange_use_registry() {
    if (!this.use_registry) this.newInfraForm.controls['registry'].setValue("");
  }
  //onChange_use_registry_secret is triggered when the value of 'use_registry_secret' is changed
  onChange_use_registry_secret() {
    if (!this.use_registry_secret) {
      this.newInfraForm.controls['server_url'].setValue("");
      this.newInfraForm.controls['username'].setValue("");
      this.newInfraForm.controls['password'].setValue("");
    }
  }

  openDeleteModal: boolean = false;
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  pendingDataList: string = '';
  //openConfirmModal() is to initialize variables when open the modal of "Delete Infra"
  openConfirmModal() {
    this.isDeleteFailed = false;
    this.isDeleteFailed = false;
    this.openDeleteModal = true;
    this.isDeleteSubmit = false;
  }

  //deleteInfra is to delete the selected infra(s)
  deleteInfra() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    for (let infra of this.selectedInfraList) {
      this.infraservice.deleteInfra(infra.uuid)
        .subscribe(() => {
          this.reloadCurrentRoute();
        },
          err => {
            this.isDeleteFailed = true;
            this.errorMessage = err.error.message;
          });
    }
  }

  isTestFailed: boolean = false;
  testPassed: boolean = false;
  //testConnection is to test the k8s connection by using kubeconfig
  testConnection() {
    this.testPassed = false;
    this.isTestFailed = false;
    if (this.newInfraForm.get('kubeconfig')?.valid) {
      this.infraservice.testK8sConnection(this.newInfraForm.get('kubeconfig')?.value).subscribe(() => {
        this.testPassed = true;
      },
        err => {
          this.isTestFailed = true;
          this.testPassed = false;
          this.errorMessage = err.error.message;
        });
    } else {
      this.errorMessage = "invalid kubeconfig input";
      this.testPassed = false;
    }
  }

  isCreatedFailed: boolean = false;
  //createNewInfra is to create an new infra
  createNewInfra() {
    this.isCreatedFailed = false;
    var infraInfo = {
      description: this.newInfraForm.get('description')?.value,
      kubernetes_provider_info: {
        kubeconfig_content: this.newInfraForm.get('kubeconfig')?.value,
        registry_config_fate: {
          use_registry: this.newInfraForm.get('use_registry')?.value,
          registry: this.use_registry ? this.newInfraForm.get('registry')?.value?.trim() : "",
          use_registry_secret: this.newInfraForm.get('use_registry_secret')?.value,
          registry_secret_config: {
            server_url: this.use_registry_secret ? this.newInfraForm.get('server_url')?.value?.trim() : "",
            username: this.use_registry_secret ? this.newInfraForm.get('username')?.value?.trim() : "",
            password: this.use_registry_secret ? this.newInfraForm.get('password')?.value?.trim() : "",
          },
        }
      },
      name: this.newInfraForm.get('infraname')?.value,
      type: this.newInfraForm.get('type')?.value,
    }

    this.infraservice.createInfra(infraInfo)
      .subscribe(
        data => {
          this.isCreatedFailed = false;
          this.reloadCurrentRoute();
        },
        err => {
          this.errorMessage = err.error.message;
          this.isCreatedFailed = true;
        }
      );

  }

  //submitDisable returns if disabled the submit button of create a new infra
  get submitDisable() {
    var registry_secret_valid = false;
    if (this.use_registry_secret) {
      registry_secret_valid = (this.newInfraForm.get('username')?.value?.trim() !== '') && (this.newInfraForm.get('password')?.value?.trim() !== '') && (!this.newInfraForm.controls['server_url'].errors)
    } else {
      registry_secret_valid = true
    }
    var registry_valid = false;
    if (this.use_registry) {
      registry_valid = (this.registry && this.registry != '')
    } else {
      registry_valid = true
    }
    var basics_valid = !this.newInfraForm.controls['infraname'].errors && !this.newInfraForm.controls['type'].errors
    return !(this.testPassed && registry_secret_valid && registry_valid && basics_valid)
  }

  //refresh is for refresh button
  refresh() {
    this.showInfraList();
  }

  //onKubeconfigChange is triggered when the input of kubeconfig is changed
  onKubeconfigChange(val: any) {
    this.testPassed = false;
    this.isTestFailed = false;
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
    return !this.newInfraForm.controls['server_url'].errors
  }

}
