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
import { ProjectService } from '../../../service/project.service';

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
export interface AutoApprove {
  enabled: boolean;
}

@Component({
  selector: 'app-info',
  templateUrl: './info.component.html',
  styleUrls: ['./info.component.css']
})
export class InfoComponent implements OnInit {

  constructor(private route: ActivatedRoute, private projectservice: ProjectService, private router: Router) {
    this.showProjectDetail();
  }

  ngOnInit(): void {
  }
  options: boolean = false;
  routeParams = this.route.parent!.snapshot.paramMap;
  //uuid is project uuid
  uuid = String(this.routeParams.get('id'));
  projectDetailResponse: any;
  projectDetail: any = {};
  isPageLoading: boolean = true;
  isShowjobInfoFailed: boolean = false;
  //showProjectDetail is to get the project detail information
  showProjectDetail() {
    this.projectservice.getProjectDetail(this.uuid)
      .subscribe((data: ProjectDetailResponse) => {
        this.projectDetailResponse = data;
        this.projectDetail = this.projectDetailResponse.data;
        this.options = this.projectDetail.auto_approval_enabled;
        this.isPageLoading = false;
      },
        err => {
          this.isShowjobInfoFailed = true;
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
        });
  }

  autoapprove: AutoApprove = {
    enabled: false
  };
  errorMessage: string = "";
  isUpdateFailed: boolean = false;
  isautoapprovesubmit: boolean = false;
  // updateAutoApprove is to update the options to enable auto-approval or not for the jobs in currrent project
  updateAutoApprove() {
    this.isautoapprovesubmit = true;
    this.autoapprove.enabled = this.options;
    this.projectservice.putAutoApprove(this.autoapprove, this.uuid)
      .subscribe(data => {
      },
        err => {
          this.errorMessage = err.error.message;
          this.isUpdateFailed = true;
        });

  }
  
  isOpenLeaveModal: boolean = false;
  leaveProjectUUID: string = "";
  // openLeaveProjectModal is triggered to confirm the request when user click the 'leave' the project' button (only with the joined project)
  openLeaveProjectModal(projectUUID: string) {
    this.isOpenLeaveModal = true;
    this.leaveProjectUUID = projectUUID;
  }

  isLeaveProjectSubmit: boolean = false;
  isLeaveProjectFailed: boolean = false;
  associatedDataExistWhenLeave: boolean = false;
  // leaveProject is to send the leave the project request (only with the joined project)
  leaveProject(projectUUID: string) {
    this.isLeaveProjectSubmit = true;
    this.projectservice.leaveProject(projectUUID)
      .subscribe(
        data => {
          this.router.navigate(['project-management']);
        },
        err => {
          this.isLeaveProjectFailed = true;
          // identity the error reason
          if (err.error.message == '') {
            this.errorMessage = "Request failed."
          } else if (err.error.message === "at least one data association exists, data: site 1 training data") {
            this.errorMessage = ""
            this.associatedDataExistWhenLeave = true;
          } else {
            this.errorMessage = err.error.message;
          }
        }
      );
  }

  isOpenCloseModal: boolean = false;
  closeProjectUUID: string = "";
  // openCloseProjectModal is triggered and open the confirm modal to send 'close project' request (only with the managed project)
  openCloseProjectModal(projectUUID: string) {
    this.isOpenCloseModal = true;
    this.closeProjectUUID = projectUUID;
  }

  isCloseProjectSubmit: boolean = false;
  isCloseProjectFailed: boolean = false;
  // closeProject is to send the close the project request (only with the managed project)
  closeProject(projectUUID: string) {
    this.isCloseProjectSubmit = true;
    this.projectservice.closeProject(projectUUID)
      .subscribe(
        data => {
          this.router.navigate(['project-management'])
        },
        err => {
          this.isCloseProjectFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }

  // redirectToData is to routing to data management in the current project
  redirectToData(projectUUID: string) {
    this.router.navigate(['project-management', 'project-detail', projectUUID, 'data']);
  }
  
  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
}
