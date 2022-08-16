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
import { CustomComparator } from 'src/utils/comparator';
import { ModelService } from '../../../service/model.service';
import { ProjectService } from '../../../service/project.service';
import { constantGather } from '../../../../config/constant'
import { FormBuilder, FormGroup } from '@angular/forms';
import { ValidatorGroup } from '../../../../config/validators'

export interface ProjectModel {
  name: string,
  version: string,
  created_time: string,
  last_call_time: string
}
export interface ParticipantListResponse {
  code: number,
  data: [],
  message: "success"
}
export interface ProjectDetailResponse {
  code: number,
  data: {},
  message: "success"
}
export interface ProjectDetail {
  auto_approval_enabled: boolean,
  creation_time: string,
  description: string,
  manager: string,
  managing_site_name: string,
  managing_site_party_id: number,
  name: string,
  uuid: string
}

@Component({
  selector: 'app-model',
  templateUrl: './model.component.html',
  styleUrls: ['./model.component.css']
})
export class ModelComponent implements OnInit {

  openModal: boolean = false;
  serviceName: any;
  type: string = "";
  parameters_json: any = "";
  constantGather = constantGather

  form: FormGroup;
  constructor(private route: ActivatedRoute, private projectservice: ProjectService, private router: Router, private modelservice: ModelService, private fb: FormBuilder) {
    this.showModelList();
    // form for 'publish model'
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

  ngOnInit(): void {
  }
  routeParams = this.route.parent!.snapshot.paramMap;
  uuid = String(this.routeParams.get('id'));
  errorMessage: string = "";
  modelList: any;
  isShowModelFailed: boolean = false;
  isPageLoading: boolean = true;
  g: boolean = true;
  // showModelList is to get model list in the current project
  showModelList() {
    this.isPageLoading = true;
    this.projectservice.getModelList(this.uuid)
      .subscribe((data: any) => {
        this.modelList = data.data?.sort((n1: ProjectModel, n2: ProjectModel) => { return this.createTimeComparator.compare(n1, n2) });
        this.isPageLoading = false;
        this.isShowModelFailed = false;
      },
        err => {
          this.isShowModelFailed = true;
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
        }
      );
  }

  // refresh button
  refresh() {
    this.showModelList();
  }

  openDeleteModal: boolean = false;
  pendingModelId: string = '';
  //openConfirmModal is triggered to open the modal for user to confirm the deletion
  openConfirmModal(model_id: string) {
    this.pendingModelId = model_id;
    this.openDeleteModal = true;
  }
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  // deleteModel is to submit the request to 'delete' model
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

  pendingModeluuid: string = '';
  // openPublishModal is triggered when 
  openPublishModal(model_uuid: string) {
    this.openModal = true;
    this.pendingModeluuid = model_uuid;
    this.getModelDeployType(model_uuid);
  }

  isGetTypeSubmit: boolean = false;
  isGetTypeFailed: boolean = false;
  modelTypeList: any = [];
  // getModelDeployType is to get the available deployment type of model by model uuid
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
  // publishModel is to request to publish the model
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

  // Comparator for datagrid
  createTimeComparator = new CustomComparator("create_time", "string");

}
