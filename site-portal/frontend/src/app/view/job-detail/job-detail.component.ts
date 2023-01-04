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

import { Component, OnInit, OnDestroy, ChangeDetectorRef } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { HttpClient } from '@angular/common/http'
import { ProjectService } from '../../service/project.service';
import { constantGather, JOBSTATUS, JOBTYPE } from '../../../config/constant'
import { MessageService } from '../../components/message/message.service'
import * as fileSaver from 'file-saver';
import Dag from '../../../config/dag'
import '@cds/core/icon/register.js';
import { checkIcon, ClarityIcons, refreshIcon, timesIcon } from '@cds/core/icon';

ClarityIcons.addIcons(refreshIcon, checkIcon, timesIcon);

@Component({
  selector: 'app-job-detail',
  templateUrl: './job-detail.component.html',
  styleUrls: ['./job-detail.component.css']
})
export class JobDetailComponent implements OnInit, OnDestroy {

  constructor(private route: ActivatedRoute, public http: HttpClient, private projectservice: ProjectService, private router: Router, private msg: MessageService, private cdRef: ChangeDetectorRef) {
  }
  ngOnDestroy(): void {
    this.msg.close()
  }

  //default expaned setting of clr-accordion-panel
  panelOpen1: boolean = true;
  panelOpen2: boolean = false;
  _panelOpen3: boolean = false;
  panelOpen4: boolean = true;
  get panelOpen3() {
    return this._panelOpen3
  }
  //draw the flow chart
  set panelOpen3(value) {
    this._panelOpen3 = value
    setTimeout(() => {
      let dag: any = new Dag(this.job.dsl_json, this.job.conf_json, "#svg-canvas", "#component_info")
      dag.tooltip_css = "dagTips";
      dag.Generate();
      dag.Draw();
    })
  }
  options: boolean = false;
  project: any;
  job: any = {};
  dsl: string = "";
  algoconfig: string = "";
  constantGather = constantGather
  jobStatus = JOBSTATUS
  jobType = JOBTYPE
  metrics_key: any;
  projidFromRoute = ''
  ngOnInit(): void {
    const routeParams = this.route.snapshot.paramMap;    
    const jobIdFromRoute = String(routeParams.get('jobid'));
    this.projidFromRoute = String(routeParams.get('projid'));
    
    this.showJobDetail(jobIdFromRoute);
  }
  ngAfterViewChecked() {
    this.cdRef.detectChanges()
  }

  getJobDetailFailed: boolean = false;
  errorMessage: string = "";
  predicting_result_list: any = []
  pageLoading: boolean = true;
  loadingCompleted: boolean = false;
  //showJobDetail is to getthe job detail by job uuid
  showJobDetail(job_id: string) {
    this.metrics_key = [];
    this.predicting_result_list = [];
    this.projectservice.getJobDetail(job_id)
      .subscribe(data => {
        this.job = data.data
        this.pageLoading = false;
        if (this.job.type === this.jobType.Modeling) {
          for (let key in this.job.result_info.training_result) {
            this.metrics_key.push(key);
          }
        }
        if (this.job.type === this.jobType.Predict) {
          if (this.job.result_info.predicting_result.data) {
            for (let sample of this.job.result_info.predicting_result.data) {
              const data_list = [];
              for (let data of sample) {
                if (data === null) {
                  data_list.push('N/A');
                } else if (data === Object(data)) {
                  data_list.push(JSON.stringify(data));
                } else {
                  data_list.push(data);
                }
              }
              this.predicting_result_list.push(data_list);
            }
          }
        }
      },
        err => {
          this.getJobDetailFailed = true;
          this.pageLoading = false;
          this.errorMessage = err.error.message;
        }
      );
  }

  //back button
  back() {
    const routeParams = this.route.snapshot.paramMap;
    const projIdFromRoute = String(routeParams.get('projid'));
    this.router.navigate(['project-management', 'project-detail', projIdFromRoute, 'job']);
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

  //actions
  approveJobFailed: boolean = false;
  approveJobsSubmit: boolean = false;
  //approve is to approve the job that pending on current site
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
  //refresh job button
  refresh(job_uuid: string) {
    this.refreshJobsSubmit = true;
    this.projectservice.refreshJob(job_uuid)
      .subscribe(data => {
        this.reloadCurrentRoute();
      },
        err => {
          this.refreshJobFailed = true;
          this.errorMessage = err.error.message;
        }
      );
  }

  openRejectModal: boolean = false;
  rejectJobFailed: boolean = false;
  rejectJobsSubmit: boolean = false;
  rejectErrorMessage: any;
  //reject is to reject the job that pending on current site
  reject(job_uuid: string) {
    this.rejectJobsSubmit = true;
    this.projectservice.rejectJob(job_uuid)
      .subscribe(data => {
        this.reloadCurrentRoute();
      },
        err => {
          this.rejectJobFailed = true;
          this.rejectErrorMessage = err.error.message;
        }
      );
  }

  openDeleteModal: boolean = false;
  submitDeleteFailed: boolean = false;
  deleteJobSubmit: boolean = false;
  deleteerrorMessage: any;
  //deleteJob is to submit 'delete job' request
  deleteJob(job_uuid: string) {
    this.deleteJobSubmit = true;
    this.projectservice.deleteJob(job_uuid)
      .subscribe(() => {
        this.back();
      },
        err => {
          this.submitDeleteFailed = true;
          this.deleteerrorMessage = err.error.message;
        });
  }

  isDownloadSubmit: boolean = false;
  isDownloadFailed: boolean = false;
  //downloadPredictResult is to download the result of prediction job 
  downloadPredictResult(job_id: string, job_name: string) {
    window.open(window.location.origin + '/api/v1/job/' + job_id + '/data-result/download')
  //   this.isDownloadSubmit = true;
  //   this.isDownloadFailed = false;
    // this.projectservice.downloadPredictJobResult(job_id)
    //   .subscribe((data: any) => {
    //     this.isDownloadSubmit = true;
    //     let blob: any = new Blob([data], { type: 'text/plain' });
    //     const url = window.URL.createObjectURL(blob);
    //     fileSaver.saveAs(blob, job_name);
    //   }), (error: any) => {
    //     this.isDownloadFailed = true;
    //     this.errorMessage = 'Error downloading the file';
    //   }
  };
}
