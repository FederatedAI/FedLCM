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
import { ActivatedRoute, Router } from '@angular/router';
import { ModelService } from '../../service/model.service';
import { FormBuilder, FormGroup } from '@angular/forms';
import { ValidatorGroup } from '../../../config/validators'
import { ClarityIcons, cloudNetworkIcon } from '@cds/core/icon';
import { constantGather } from '../../../config/constant'

ClarityIcons.addIcons(cloudNetworkIcon);
export interface ModelDetailResponse {
  code: number,
  data: {},
  message: "success"
}
export interface ModelDetail {
  component_name: string,
  create_time: string,
  evaluation: any,
  job_name: string,
  job_uuid: string,
  model_id: string,
  model_version: string,
  name: string,
  party_id: number,
  project_name: string,
  project_uuid: string,
  role: string,
  uuid: string
}

@Component({
  selector: 'app-model-detail',
  templateUrl: './model-detail.component.html',
  styleUrls: ['./model-detail.component.css']
})

export class ModelDetailComponent implements OnInit {
  form: FormGroup;
  constructor(private route: ActivatedRoute, private modelservice: ModelService, private router: Router, private fb: FormBuilder) {
    this.showModelDetail(this.uuid);
    this.form = this.fb.group(
      ValidatorGroup([
        {
          name: 'serviceName',
          value: '',
          type: ['noSpace'],
          max: 20,
          min: 2
        },
        {
          name: 'type',
          type: ['']
        },
        {
          name: 'parameters_json',
          type: ['notRequired', 'json']
        }
      ])
    );
  }

  openModal: boolean = false;
  openDeleteModal: boolean = false;
  serviceName: any;
  type: string = "";
  parameters_json: any = "";
  constantGather = constantGather;
  ngOnInit(): void {

  }
  //uuid is model's uuid
  uuid = String(this.route.snapshot.paramMap.get('id'));
  modeldata: ModelDetail = {
    component_name: "",
    create_time: "",
    evaluation: {},
    job_name: "",
    job_uuid: "",
    model_id: "",
    model_version: "",
    name: "",
    party_id: 0,
    project_name: "",
    project_uuid: "",
    role: "",
    uuid: ""
  };
  errorMessage: any;
  isShowModelDetailFailed: boolean = false;
  key: any;
  isPageLoading: boolean = true;
  //showModelDetail is to get the model detail
  showModelDetail(uuid: string) {
    this.modelservice.getModelDetail(uuid)
      .subscribe((data: any) => {
        this.modeldata = data.data;
        this.key = [];
        for (let keyvalue in this.modeldata.evaluation) {
          this.key.push(keyvalue);
        }
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
          this.isShowModelDetailFailed = true;
        }
      );
  }

  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  //deleteModel is to submit request to delete Model
  deleteModel(uuid: string) {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.modelservice.deleteModel(uuid)
      .subscribe(() => {
        this.router.navigate(['model-management']);
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  //back button
  back() {
    this.router.navigate(['model-management']);
  }

  pendingModeluuid: string = '';
  //openPublishModal is triggerd by 'Publish' button, then open the publich Model modal
  openPublishModal(model_uuid: string) {
    this.openModal = true;
    this.pendingModeluuid = model_uuid;
    this.getModelDeployType(model_uuid);
  }

  isGetTypeSubmit: boolean = false;
  isGetTypeFailed: boolean = false;
  modelTypeList: any = [];
  //getModelDeployType is to get the supported model deployment type
  getModelDeployType(uuid: string) {
    this.isGetTypeSubmit = true;
    this.isGetTypeFailed = false;
    this.modelservice.getModelSupportedDeploymentType(uuid)
      .subscribe((data: any) => {
        this.modelTypeList = data.data;
      },
        err => {
          this.isGetTypeFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  isPublishSubmit: boolean = false;
  isPublishFailed: boolean = false;
  //publishModel is to send request to publish model
  publishModel() {
    this.isPublishSubmit = true;
    this.isPublishFailed = false;
    if (!this.form.valid) {
      this.isPublishFailed = true;
      this.errorMessage = "Invaild input."
      return;
    }
    if (this.parameters_json === "") this.parameters_json = "{}";
    this.modelservice.publishModel(this.pendingModeluuid, Number(this.type), this.parameters_json, this.serviceName)
      .subscribe(() => {
        this.reloadCurrentRoute();
      },
        err => {
          this.isPublishFailed = true;
          this.errorMessage = err.error.message;
        });
  }
  
  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
}

