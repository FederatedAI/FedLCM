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

import { Component, OnInit, OnDestroy } from '@angular/core';
import '@cds/core/icon/register.js';
import { ClarityIcons, newIcon, refreshIcon, tasksIcon, windowCloseIcon } from '@cds/core/icon';
import { ProjectService } from '../../service/project.service';
import { Router, ActivatedRoute } from '@angular/router';
import { FormBuilder } from '@angular/forms';
import { ValidatorGroup } from '../../../config/validators'
import { MessageService } from '../../components/message/message.service'
import { CustomComparator } from 'src/utils/comparator';
import { SiteService } from 'src/app/service/site.service';

ClarityIcons.addIcons(windowCloseIcon, newIcon, tasksIcon, refreshIcon);

export interface JoinedProject {
  creation_time: string,
  local_data_num: number,
  managed_by_this_site: boolean,
  manager: string,
  managing_site_name: string,
  managing_site_party_id: number,
  name: string,
  participants_num: number,
  pending_job_exist: boolean,
  remote_data_num: number,
  running_job_num: number,
  success_job_num: number,
  uuid: string
}
export interface InvitedProject {
  creation_time: string,
  manager: string,
  managing_site_name: string,
  managing_site_party_id: number,
  name: string,
  uuid: string
}
export interface ProjectListResponse {
  code: number,
  data: {
    invited_projects: InvitedProject[],
    joined_projects: JoinedProject[]
  },
  message: "success"
}
export interface NewProject {
  auto_approval_enabled: boolean,
  description: string,
  name: string
}

@Component({
  selector: 'app-project-mg',
  templateUrl: './project-mg.component.html',
  styleUrls: ['./project-mg.component.css']
})

export class ProjectMgComponent implements OnInit, OnDestroy {

  constructor(private projectservice: ProjectService, private route: ActivatedRoute,
    private router: Router, private fb: FormBuilder, private msg: MessageService, private siteService: SiteService) {
    this.showProjectList();
  }

  ngOnInit(): void {
  }
  ngOnDestroy(): void {
    this.msg.close()
  }

  openModal: boolean = false;
  projName: string = '';
  desc: string = '';
  options: any;
  selectOptions: string = "all";
  projectListResponse: any;
  projectlist: any = [];
  joinedProjectList: any = [];
  invitedProjectList: any = [];
  myProjectList: JoinedProject[] = [];
  othersProjectList: JoinedProject[] = [];
  closedProjectList: JoinedProject[] = [];
  myPendingJobExist: boolean = false;
  othersPendingJobExist: boolean = false;

  // form is for 'create form' project
  form = this.fb.group(
    ValidatorGroup([
      {
        name: 'projName',
        value: '',
        type: ['word'],
        max: 20,
        min: 2
      },
      {
        name: 'desc',
        type: ['']
      },
      {
        name: 'options',
        type: ['']
      },
    ])
  );

  isShowProjectFailed: boolean = false;
  isPageLoading: boolean = true;
  // showProjectList is to get all project list and filter with type
  showProjectList() {
    this.isPageLoading = true;
    this.myProjectList = [];
    this.othersProjectList = [];
    this.closedProjectList = [];
    this.projectservice.getProjectList()
      .subscribe((data: ProjectListResponse) => {
        this.projectListResponse = data;
        this.projectlist = this.projectListResponse.data;
        this.joinedProjectList = this.projectlist.joined_projects;
        this.invitedProjectList = this.projectlist.invited_projects;
        this.closedProjectList = this.projectlist.closed_projects? this.projectlist.closed_projects : [];
        for (let project of this.joinedProjectList) {
          if (project.managed_by_this_site) {
            this.myProjectList.push(project);
            if (project.pending_job_exist && !this.myPendingJobExist) this.myPendingJobExist = true;
          } else {
            this.othersProjectList.push(project);
            if (project.pending_job_exist && !this.othersPendingJobExist) this.othersPendingJobExist = true;
          }
        }
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isShowProjectFailed = true;
          this.isPageLoading = false;
        });
  }

  isCreateSubmitted: boolean = false;
  isCreatedFailed: boolean = false;
  errorMessage: string = "";
  // createNewProject is to request for creating new project
  createNewProject() {
    this.isCreateSubmitted = true;
    this.isCreatedFailed = false;
    // validation before submit the request
    if (this.projName === '') {
      this.isCreatedFailed = true;
      this.errorMessage = 'Name can not be empty';
      return;
    }
    if (!this.form.valid) {
      this.errorMessage = 'Invalid information.';
      this.isCreatedFailed = true;
      return;
    }
    this.projectservice.createProject(this.options, this.desc, this.projName)
      .subscribe(
        data => {
          this.msg.success('serverMessage.create200', 1000)
          this.isCreatedFailed = false;
          this.reloadCurrentRoute();
        },
        err => {
          this.errorMessage = err.error.message;
          this.isCreatedFailed = true;
        }
      );

  }

  // acceptInvitation is to accept the project invitation from other party
  acceptProjectInvitationFailed: boolean = false;
  acceptProjectInvitationSubmit: boolean = false;
  acceptInvitation(projectUUID: string) {
    this.acceptProjectInvitationFailed = false
    this.acceptProjectInvitationSubmit = true
    this.projectservice.acceptInvitation(projectUUID)
      .subscribe(
        data => {
          this.reloadCurrentRoute()
        },
        err => {
          this.acceptProjectInvitationFailed = true
          this.errorMessage = err.error.message
        }
      );
  }

  isOpenRejectModal: boolean = false;
  rejectProjectUUID: string = "";
  rejectErrorMessage: string = "";
  // openRejectModal is trigger to open the confirmation modal when user try to 'reject project invitation'
  openRejectModal(projectUUID: string) {
    this.isOpenRejectModal = true;
    this.rejectProjectUUID = projectUUID;
  }

  rejectProjectInvitationFailed: boolean = false;
  rejectProjectInvitationSubmit: boolean = false;
  // rejectInvitation is to request for 'reject project invitation'
  rejectInvitation(projectUUID: string) {
    this.rejectProjectInvitationSubmit = true;
    this.rejectProjectInvitationFailed = false;
    this.projectservice.rejectInvitation(projectUUID)
      .subscribe(
        data => {
          this.reloadCurrentRoute();
        },
        err => {
          this.rejectProjectInvitationFailed = true;
          this.errorMessage = err.error.message;
          if (this.rejectErrorMessage === '') {
            this.rejectErrorMessage = "Request failed."
          }
        }
      );
  }

  // closeModal is to close the modal of 'creat project'
  closeModal() {
    this.openModal = false;
    this.isCreateSubmitted = false;
    this.isCreatedFailed = false;
  }

  //newProjectAuthentication is to send request to get current user for authentication in case the session is expired
  newProjectAuthentication() {
    this.openModal = true
    this.siteService.getCurrentUser().subscribe(
      data => { },
      err => { }
    )
  }

  isOpenLeaveModal: boolean = false;
  leaveProjectUUID: string = "";
  // openLeaveProjectModal is trigger to open the confirmation modal when user try to 'leave project' (only with joined project)
  openLeaveProjectModal(projectUUID: string) {
    this.isOpenLeaveModal = true;
    this.leaveProjectUUID = projectUUID;
  }

  isLeaveProjectSubmit: boolean = false;
  isLeaveProjectFailed: boolean = false;
  associatedDataExistWhenLeave: boolean = false;
  // leaveProject is to send 'leave project' request (only with joined project)
  leaveProject(projectUUID: string) {
    this.isLeaveProjectSubmit = true;
    this.projectservice.leaveProject(projectUUID)
      .subscribe(
        data => {
          this.reloadCurrentRoute();
        },
        err => {
          this.isLeaveProjectFailed = true;
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
  
  //redirectToData is routing to Data management page in current project
  redirectToData(projectUUID: string) {
    this.router.navigate(['project-management', 'project-detail', projectUUID, 'data']);
  }

  isOpenCloseModal: boolean = false;
  closeProjectUUID: string = "";
  // openCloseProjectModal is trigger to open the confirmation modal when user try to 'close project' (only with managed project)
  openCloseProjectModal(projectUUID: string) {
    this.isOpenCloseModal = true;
    this.closeProjectUUID = projectUUID;
  }
  
  isCloseProjectSubmit: boolean = false;
  isCloseProjectFailed: boolean = false;
  // closeProject is to request for closing project (only with managed project)
  closeProject(projectUUID: string) {
    this.isCloseProjectSubmit = true;
    this.projectservice.closeProject(projectUUID)
      .subscribe(
        data => {
          this.reloadCurrentRoute();
        },
        err => {
          this.isCloseProjectFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/user-management', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
  
  // refresh button
  refresh() {
    this.showProjectList();
  }

  // comparator for datagrid
  timeComparator = new CustomComparator("creation_time", "string");
  participantComparator = new CustomComparator("participants_num", "number");
  managingSiteComparator = new CustomComparator("managing_site_name", "string");
}

