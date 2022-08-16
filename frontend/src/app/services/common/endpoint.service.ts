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
import { HttpUrlEncodingCodec } from '@angular/common/http';

@Injectable({
  providedIn: 'root'
})
export class EndpointService {

  constructor(private http: HttpClient) { } 
  codec = new HttpUrlEncodingCodec;

  getEndpointList() {
    return this.http.get<any>('/endpoint');
  }

  getEndpointDetail(uuid:string) {
    return this.http.get<any>('/endpoint/'+uuid);
  }

  deleteEndpoint(uuid:string, uninstall:boolean): Observable<any> {
    let params = new HttpParams().set('uninstall', uninstall);
    return this.http.delete('/endpoint/'+uuid,{params: params});
  }

  checkEndpoint (uuid:string) {
    return this.http.post('/endpoint/'+uuid+'/kubefate/check', {});
  }
  postEndpointScan (uuid:string, type: string) {
    return this.http.post('/endpoint/scan', {
      infra_provider_uuid: uuid,
      type: type
    });
  }
  getKubefateYaml(service_username: string, service_password: string, hostname: string, use_registry: boolean, registry: string, use_registry_secret: boolean, registry_server_url: string, registry_username: string, registry_password: string) {
    let params = new HttpParams()
    .set('service_username', service_username)
    .set('service_password', service_password)
    .set('hostname', hostname)
    .set('use_registry', use_registry)
    .set('registry', registry)
    .set('use_registry_secret', use_registry_secret)
    .set('registry_server_url', registry_server_url)
    .set('registry_username', registry_username)
    .set('registry_password', registry_password);
    return this.http.get<any>('/endpoint/kubefate/yaml',{params: params});
  }

  createEndpoint(endpointConfig:any) {
    return this.http.post('/endpoint', endpointConfig);
  }

}
