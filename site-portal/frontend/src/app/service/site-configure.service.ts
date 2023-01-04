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
import { SiteInfoResponse } from 'src/app/service/site.service'; 

@Injectable({
  providedIn: 'root'
})
export class SiteConfigureService {

  constructor(private http: HttpClient) { }

  putConfigUpdate(siteUpdatedInfo: any) : Observable<SiteInfoResponse> {
    return this.http.put<SiteInfoResponse>('/site', siteUpdatedInfo);
  }
  
  connectFML(connectInfo:any): Observable<any> {
    return this.http.post('/site/fmlmanager/connect', connectInfo);
  }

  unregisterFML() {
    return this.http.post('/site/fmlmanager/unregister', {});
  }

  testFATEFlow(host: string, https: boolean, port: number): Observable<any> {
    return this.http.post('/site/fateflow/connect', { 
      host,
      https,
      port
    });
  }

  testKubeFlow(kubeconfig: string, minio_access_key: string, minio_endpoint: string, minio_region: string, minio_secret_key: string, minio_ssl_enabled:boolean): Observable<any> {
    return this.http.post('/site/kubeflow/connect', { 
      kubeconfig,
      minio_access_key,
      minio_endpoint,
      minio_region,
      minio_secret_key,
      minio_ssl_enabled
    });
  }
}
