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
import { Router } from '@angular/router';
import { EndpointService } from 'src/app/services/common/endpoint.service';
import { ENDPOINTSTATUS, constantGather } from 'src/utils/constant';

@Component({
  selector: 'app-endpoint-mg',
  templateUrl: './endpoint-mg.component.html',
  styleUrls: ['./endpoint-mg.component.scss']
})
export class EndpointMgComponent implements OnInit {

  constructor(private endpointservice: EndpointService, private router: Router) { }

  ngOnInit(): void {
    this.showEndpointList();
  }

  endpointStatus = ENDPOINTSTATUS;
  constantGather = constantGather

  endpointlist: any = []
  errorMessage = "Service Error!"
  isShowEndpointFailed: boolean = false
  isPageLoading: boolean = true
  //showEndpointList is to get showEndpointList
  showEndpointList() {
    this.isShowEndpointFailed = false;
    this.isPageLoading = true;
    this.endpointservice.getEndpointList()
      .subscribe((data: any) => {
        this.endpointlist = data.data;
        this.isPageLoading = false;
      },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isShowEndpointFailed = true;
          this.isPageLoading = false;
        });
  }

  openDeleteModal: boolean = false;
  isDeleteSubmit: boolean = false;
  isDeleteFailed: boolean = false;
  pendingDataList: string = '';
  //openDeleteConfrimModal is to open the confirmaton modal of deletion and initialized the variables
  openDeleteConfrimModal(uuid: string) {
    this.pendingEndpoint = uuid;
    this.isDeleteFailed = false;
    this.isDeleteFailed = false;
    this.openDeleteModal = true;
    this.isDeleteSubmit = false;
    this.uninstall = true;
  }

  pendingEndpoint: string = "";
  uninstall: boolean = true;
  //deleteEndpoint is to delete the Endpoint
  deleteEndpoint() {
    this.isDeleteSubmit = true;
    this.isDeleteFailed = false;
    this.endpointservice.deleteEndpoint(this.pendingEndpoint, this.uninstall)
      .subscribe(() => {
        this.reloadCurrentRoute();
      },
        err => {
          this.isDeleteFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  //refresh is for refresh button
  refresh() {
    this.showEndpointList();
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

}
