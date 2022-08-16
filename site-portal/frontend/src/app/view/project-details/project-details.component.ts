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
import { Location } from '@angular/common'
import { ProjectService } from '../../service/project.service';

export interface Job {
  name: string,
  time: string,
  id: string,
  type: string,
  intiator: string,
  status: string,
  job_invited: boolean,
}
export interface AssociateData {
  updated_time: string,
  created_time: string,
  data_provider: string,
  name: string,
  associate_status: boolean
}
export interface Participant {
  name: string,
  party_id: string,
  desc: string,
  created_time: string,
  status: string
}
export interface PartyUser {
  creation_time: string,
  description: string,
  name: string,
  party_id: number,
  status: number,
  uuid: string,
  selected: boolean
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
export interface DataListResponse {
  code: number,
  data: DataDetail[],
  message: "success"
}
export interface DataDetail {
  creation_time: string,
  data_id: string,
  is_local: boolean,
  name: string,
  providing_site_name: string,
  providing_site_party_id: number,
  providing_site_uuid: string,
  update_time: string
}
export interface LocalData {
  creation_time: string,
  data_id: string,
  name: string,
  selected: boolean
}
export interface AutoApprove {
  enabled: boolean;
}

@Component({
  selector: 'app-project-details',
  templateUrl: './project-details.component.html',
  styleUrls: ['./project-details.component.css']
})

export class ProjectDetailsComponent implements OnInit {

  constructor(public route: ActivatedRoute, public localtion: Location, public router: Router, private projectservice: ProjectService) {
    this.showProjectDetail();
    this.showJobList();
  }

  options: boolean = false;
  project: any;
  openModal: boolean = false;
  inviteoption = false;
  back() {
    this.router.navigate(['project-management']);
  }
  ngOnInit(): void {
  }

  routeParams = this.route.parent!.snapshot.paramMap;
  // uuid is project uuid
  uuid = String(this.routeParams.get('id'));
  errorMessage: string = "";
  isShowjobInfoFailed: boolean = false;
  //showProjectDetail is to get the project detail information
  showProjectDetail() {
    this.isShowjobInfoFailed = false;
    this.projectservice.getProjectDetail(this.uuid)
      .subscribe((data: any) => {
      },
        err => {
          this.isShowjobInfoFailed = true;
          this.errorMessage = err.error.message;
        });
  }

  pendingJobList: any;
  isShowjobFailed: boolean = false;
  // showJobList is to get the current job list in the current project, then filted the job invitation for alerting user
  showJobList() {
    this.pendingJobList = [];
    this.projectservice.getJobList(this.uuid)
      .subscribe((data: any) => {
        for (let pendingjob of data.data) {
          if (pendingjob.pending_on_this_site) {
            this.pendingJobList.push(pendingjob);
          }
        }
      },
        err => {
          this.isShowjobFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }
  // refresh button
  refresh() {
    this.reloadCurrentRoute();
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
}
