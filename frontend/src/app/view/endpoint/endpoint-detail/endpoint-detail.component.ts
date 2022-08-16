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
import { EndpointService } from 'src/app/services/common/endpoint.service';
import { ENDPOINTSTATUS, constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-endpoint-detail',
  templateUrl: './endpoint-detail.component.html',
  styleUrls: ['./endpoint-detail.component.scss']
})
export class EndpointDetailComponent implements OnInit {

  constructor(private endpointservice: EndpointService, private router: Router, private route: ActivatedRoute) { }

  ngOnInit(): void {
    this.showEndpointDetail();
  }
  endpointStatus = ENDPOINTSTATUS;
  constantGather = constantGather
  //uuid of current endpoint
  uuid = String(this.route.snapshot.paramMap.get('id'));
  endpointDetail: any
  errorMessage = "Service Error!"
  code: any
  overview = true
  get isOverview() {
    return this.overview
  }
  set isOverview(value) {
    if (value) {
      if (this.code) {
        this.code.setValue(this.endpointDetail.kubefate_deployment_yaml)
        this.overview = value
      } else {
        setTimeout(() => {
          const yamlHTML = document.getElementById('yaml') as any
          this.code = window.CodeMirror.fromTextArea(yamlHTML, {
            value: '',
            mode: 'yaml',
            lineNumbers: true,
            indentUnit: 1,
            lineWrapping: true,
            tabSize: 2,
            readOnly: true
          })
          this.code.setValue(this.endpointDetail ? this.endpointDetail.kubefate_deployment_yaml : '')
          this.overview = value
        });
      }
    } else {
      this.code = null
    }
  }

  isShowEndpointDetailFailed: boolean = false;
  isPageLoading: boolean = true;
  //showEndpointDetail is to get the endpoint detail
  showEndpointDetail() {
    this.isShowEndpointDetailFailed = false;
    this.endpointservice.getEndpointDetail(this.uuid)
      .subscribe((data: any) => {
        const value = data.data.kubefate_deployment_yaml
        if (this.code) {
          this.code.setValue(value)
        } else {
          const yamlHTML = document.getElementById('yaml') as any
          this.code = window.CodeMirror.fromTextArea(yamlHTML, {
            value: '',
            mode: 'yaml',
            lineNumbers: true,
            indentUnit: 1,
            lineWrapping: true,
            tabSize: 2,
            readOnly: true
          })
          this.code.setValue(value)
        }
        this.endpointDetail = data.data;
        this.isPageLoading = false;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isPageLoading = false
          this.isShowEndpointDetailFailed = true
        }
      );
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
    this.uninstall = true;
  }
  uninstall: boolean = true;
  //deleteEndpoint is to delete the endpoint
  deleteEndpoint() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.endpointservice.deleteEndpoint(this.uuid, this.uninstall)
      .subscribe(() => {
        this.router.navigate(['/endpoint']);
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  ischeckFailed = false
  isLoading = false;
  ischeckSuccess = false;
  //check is to check the connection of endpoint
  check() {
    this.isLoading = true;
    this.endpointservice.checkEndpoint(this.uuid)
      .subscribe(
        (data: any) => {
          this.isLoading = false;
          this.ischeckFailed = false;
          this.ischeckSuccess = true;
        },
        err => {
          this.ischeckFailed = true
          this.errorMessage = err.error.message;
          this.isLoading = false;
          this.ischeckSuccess = false;
        })
  }
}
