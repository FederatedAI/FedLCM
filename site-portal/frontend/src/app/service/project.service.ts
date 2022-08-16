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

import { HttpClient, HttpParams } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { DataListResponse } from '../view/data-mg/data-mg.component';
import { AutoApprove, ParticipantListResponse, ProjectDetailResponse } from '../view/project-details/project-details.component';
import { ProjectListResponse } from '../view/project-mg/project-mg.component';

@Injectable({
  providedIn: 'root'
})
export class ProjectService {

  constructor(private http: HttpClient) { }

  getProjectList() {
    return this.http.get<ProjectListResponse>('/project');
  }

  getProjectDetail(uuid: string) {
    return this.http.get<ProjectDetailResponse>('/project/' + uuid);
  }

  putAutoApprove(autoapprove: AutoApprove, uuid: string): Observable<AutoApprove> {
    return this.http.put<AutoApprove>('/project/' + uuid + '/autoapprovalstatus', autoapprove)
  }

  getParticipantList(uuid: string, all: boolean) {
    let params = new HttpParams().set('all', all);
    return this.http.get<ParticipantListResponse>('/project/' + uuid + '/participant', { params: params });
  }

  createProject(auto_approval_enabled: boolean, description: string, name: string): Observable<any> {
    return this.http.post('/project', {
      auto_approval_enabled,
      description,
      name
    });
  }

  postInvitation(proj_uuid: string, description: string, name: string, party_id: number, uuid: string): Observable<any> {
    return this.http.post('/project/' + proj_uuid + '/invitation', {
      description,
      name,
      party_id,
      uuid
    });
  }

  acceptInvitation(proj_uuid: string): Observable<any> {
    return this.http.post('/project/' + proj_uuid + '/join', {
    });
  }

  rejectInvitation(proj_uuid: string): Observable<any> {
    return this.http.post('/project/' + proj_uuid + '/reject', {
    });
  }

  deleteParticipant(proj_uuid: string, party_uuid: string): Observable<any> {
    return this.http.delete('/project/' + proj_uuid + '/participant/' + party_uuid);
  }

  getAssociatedDataList(uuid: string) {
    return this.http.get<DataListResponse>('/project/' + uuid + '/data');
  }

  getParticipantAssociatedDataList(uuid: string, participant: string) {
    let params = new HttpParams().set('participant', participant);
    return this.http.get<DataListResponse>('/project/' + uuid + '/data', { params: params });
  }

  getLocalDataList(uuid: string) {
    return this.http.get<DataListResponse>('/project/' + uuid + '/data/local');
  }

  associateData(proj_uuid: string, data_id: string, name: string): Observable<any> {
    return this.http.post('/project/' + proj_uuid + '/data', {data_id, name});
  }

  deleteAssociatedData(proj_uuid: string, data_uuid: string): Observable<any> {
    return this.http.delete('/project/' + proj_uuid + '/data/' + data_uuid);
  }

  getJobList(uuid: string) {
    return this.http.get<any>('/project/' + uuid + '/job');
  }

  createJob(uuid: string, newJobDetail: any): Observable<any> {
    return this.http.post('/project/' + uuid + '/job', newJobDetail);
  }

  generateJobConfig(newJobDetail: any): Observable<any> {
    return this.http.post('/job/conf/create', newJobDetail);
  }

  getJobDetail(job_uuid: string) {
    return this.http.get<any>('/job/' + job_uuid);
  }

  approveJob(job_uuid: string): Observable<any> {
    return this.http.post('/job/' + job_uuid + '/approve', {});
  }

  refreshJob(job_uuid: string): Observable<any> {
    return this.http.post('/job/' + job_uuid + '/refresh', {});
  }

  rejectJob(job_uuid: string): Observable<any> {
    return this.http.post('/job/' + job_uuid + '/reject', {});
  }

  getModelList(uuid: string) {
    return this.http.get<any>('/project/' + uuid + '/model');
  }

  getPredictParticipant(predict_model: string) {
    let params = new HttpParams().set('modelUUID', predict_model);
    return this.http.get<any>('/job/predict/participant', { params: params });
  }

  deleteJob(job_uuid: string,): Observable<any> {
    return this.http.delete('/job/' + job_uuid);
  }

  downloadPredictJobResult(job_id: string) {
    return this.http.get('/job/' + job_id + '/data-result/download', { responseType: 'text' });
  }

  leaveProject(proj_uuid: string): Observable<any> {
    return this.http.post('/project/' + proj_uuid + '/leave', {});
  }

  closeProject(proj_uuid: string): Observable<any> {
    return this.http.post('/project/' + proj_uuid + '/close', {});
  }

  getAlgorithmData() {
    return this.http.get('/job/components');
  }

  getDslAndConf(data: any, type: 'generateDslFromDag' | 'generateConfFromDag') {
    if (type === 'generateDslFromDag') {
      return this.http.post('/job/' + type, { 'raw_json': JSON.stringify(data.reqData) });
    } else {
      return this.http.post('/job/' + type, { 'dag_json': { 'raw_json': JSON.stringify(data.reqData) }, 'job_conf': data.jobDetail });
    }
  }
}