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
import { FormBuilder, FormGroup } from '@angular/forms';
import '@cds/core/icon/register.js';
import { ClarityIcons, searchIcon } from '@cds/core/icon';
import { ModelService } from '../../service/model.service';
import { Router } from '@angular/router';
import { ValidatorGroup } from '../../../config/validators'
import { CustomComparator } from 'src/utils/comparator';
import { constantGather } from '../../../config/constant'

ClarityIcons.addIcons(searchIcon);

export interface ModelElement {
  component_name: string,
  create_time: string,
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
export interface ModelListResponse {
  code: number,
  data: [],
  message: "success"
}

@Component({
  selector: 'app-model-mg',
  templateUrl: './model-mg.component.html',
  styleUrls: ['./model-mg.component.css']
})

export class ModelMgComponent implements OnInit {
  timeComparator = new CustomComparator("create_time", "string");
  filterSearchValue: string = "";
  openModal: boolean = false;
  serviceName: any;
  type: string = "";
  parameters_json: any = "";
  openDeleteModal: boolean = false;
  pendingModelId: string = '';
  storageDataList: any[] = []
  constantGather = constantGather
  form: FormGroup;
  constructor(private fb: FormBuilder, private modelservice: ModelService, private router: Router) {
    this.showModelList();
    //form for publishing model
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
  ngOnInit(): void { }

  //openConfirmModal is triggered to open the modal for user to confirm the deletion
  openConfirmModal(model_id: string) {
    this.pendingModelId = model_id;
    this.openDeleteModal = true;
  }

  errorMessage: any;
  isShowModelFailed: boolean = false;
  modelList: any;
  isPageLoading: boolean = true;
  //showModelList is to get the model list
  showModelList() {
    this.isPageLoading = true;
    this.modelservice.getModelList()
      .subscribe((data: ModelListResponse) => {
        this.modelList = data.data?.sort((n1: ModelElement, n2: ModelElement) => { return this.timeComparator.compare(n1, n2) });
        this.storageDataList = this.modelList;
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowModelFailed = true;
          this.isPageLoading = false;
        });
  }
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  //deleteModel is to submit the request for deleting model
  deleteModel(uuid: string) {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.modelservice.deleteModel(uuid)
      .subscribe(() => {
        this.reloadCurrentRoute();
      },
        err => {
          this.isDeleteFailed = true;
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

  //filterModelHandle is to process model list with filtered result 
  filterModelHandle(data: any) {
    this.modelList = []
    if (data.searchValue.trim() === '' && data.eligibleList.length < 1) {
      this.modelList = this.storageDataList
    } else {
      data.eligibleList.forEach((el: any) => {
        this.modelList.push(el)
      })
    }
  }

  pendingModeluuid: string = '';
  //openPublishModal is to open the modal for publish model
  openPublishModal(model_uuid: string) {
    this.openModal = true;
    this.pendingModeluuid = model_uuid;
    this.getModelDeployType(model_uuid);
  }

  isGetTypeSubmit: boolean = false;
  isGetTypeFailed: boolean = false;
  modelTypeList: any = [];
  //getModelDeployType is to get the deployment type of model
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
  //publish model is to send request to publich model
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

  //refresh button
  refresh() {
    this.showModelList();
    this.isShowfilter = false;
  }
  isShowfilter: boolean = false;
  showFilter() {
    this.isShowfilter = !this.isShowfilter;
  }
}
