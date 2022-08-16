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

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { InfraResponse } from 'src/app/view/infra/infra-model'; 

@Injectable({
  providedIn: 'root'
})

export class InfraService {

  constructor(private http: HttpClient) { } 
  
  getInfraList() {
    return this.http.get<InfraResponse>('/infra');
  }

  getInfraDetail(uuid:string) {
    return this.http.get<InfraResponse>('/infra/'+uuid);
  }

  deleteInfra(uuid:string): Observable<any> {
    return this.http.delete('/infra/'+uuid);
  }

  createInfra(infraInfo:any): Observable<any> {
    return this.http.post('/infra', infraInfo);
  }

  updateInfraProvider(infraInfo: any, uuid:string) : Observable<any> {
    return this.http.put<any>('/infra/'+uuid, infraInfo)
  }

  testK8sConnection(kubeconfig_content:any){
    return this.http.post('/infra/kubernetes/connect',{
      kubeconfig_content
    });
  }

}
