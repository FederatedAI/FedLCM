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
import { JOBSTATUS, JOBTYPE, constantGather } from '../../../../config/constant'
import { CustomComparator } from 'src/utils/comparator';
@Component({
  selector: 'app-job',
  templateUrl: './job.component.html',
  styleUrls: ['./job.component.css']
})
export class JobComponent implements OnInit {

  constructor(private route: ActivatedRoute, private projectservice: ProjectService, private router: Router) {
    this.showJobList();
  }

  jobStatus = JOBSTATUS
  jobType = JOBTYPE
  constantGather = constantGather
  options: boolean = false;
  project: any;
  openModal: boolean = false;
  inviteoption = false;

  ngOnInit(): void {
  }

  routeParams = this.route.parent!.snapshot.paramMap;
  // uuid is project uuid
  uuid = String(this.routeParams.get('id'));
  errorMessage: string = "";
  jobList: any;
  isShowjobFailed: boolean = false;
  isPageLoading: boolean = true;
  // showJobList is to get job list in current project
  showJobList() {
    this.projectservice.getJobList(this.uuid)
      .subscribe((data: any) => {
        this.jobList = data.data?.sort((n1: any, n2: any) => { return this.jobComparator.compare(n1, n2) });
        this.isPageLoading = false;
      },
        err => {
          this.isShowjobFailed = true;
          this.errorMessage = err.error.message;
          this.isPageLoading = false;
        }
      );
  }
  
  approveJobFailed: boolean = false;
  approveJobsSubmit: boolean = false;
  // approve is to approve the job invitation
  approve(job_uuid: string) {
    this.approveJobsSubmit = true;
    this.approveJobFailed = false;
    this.projectservice.approveJob(job_uuid)
      .subscribe(data => {
        this.reloadCurrentRoute();
      },
        err => {
          this.approveJobFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }

  refreshJobFailed: boolean = false;
  refreshJobsSubmit: boolean = false;
  //refresh button
  refresh() {
    this.refreshJobsSubmit = true;
    for (let job of this.jobList) {
      this.projectservice.refreshJob(job.uuid)
        .subscribe(data => {
        },
          err => {
            this.refreshJobFailed = true;
            this.errorMessage = err.error.message;
          }
        );
    }
    this.reloadCurrentRoute();
  }

  isOpenRejectModal: boolean = false;
  reject_job_uuid: string = "";
  rejectErrorMessage: string = "";
  openRejectModal(job_uuid: string) {
    this.isOpenRejectModal = true;
    this.reject_job_uuid = job_uuid;
  }
  rejectJobFailed: boolean = false;
  rejectJobsSubmit: boolean = false;
  // reject is to reject job invitation
  reject(job_uuid: string) {
    this.rejectJobsSubmit = true;
    this.projectservice.rejectJob(job_uuid)
      .subscribe(data => {
        this.reloadCurrentRoute();
      },
        err => {
          this.rejectJobFailed = true;
          this.rejectErrorMessage = err.error.message;
          if (this.rejectErrorMessage === '') {
            this.rejectErrorMessage = "Request failed."
          }
        }
      );
  }

  submitDeleteFailed: boolean = false;
  deleteJobSubmit: boolean = false;
  deleteerrorMessage: any;
  // deleteJob is to send 'delete job' request
  deleteJob(job_uuid: string) {
    this.deleteJobSubmit = true;
    this.projectservice.deleteJob(job_uuid)
      .subscribe(() => {
        this.reloadCurrentRoute();
      },
        err => {
          this.submitDeleteFailed = true;
          this.deleteerrorMessage = err.error.message;
        });
  }

  pendingJobId: any;
  openDeleteModal: boolean = false;
  //openConfirmModal is triggered to open the modal for user to confirm the deletion
  openConfirmModal(job_id: string) {
    this.pendingJobId = job_id;
    this.openDeleteModal = true;
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }
  
  // Comparator for datagrid
  createTimeComparator = new CustomComparator("creation_time", "string");
  statusComparator = new CustomComparator("status", "number");
  typeComparator = new CustomComparator("type", "number");
  partyComparator = new CustomComparator("initiating_site_name", "string");
  jobComparator = new CustomComparator("creation_time", "job");
}
