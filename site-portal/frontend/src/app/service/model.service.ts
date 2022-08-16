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
import { ModelListResponse } from '../view/model-mg/model-mg.component';
import { ModelDetailResponse } from '../view/model-detail/model-detail.component';

@Injectable({
  providedIn: 'root'
})
export class ModelService {

  constructor(private http: HttpClient) { }

  getModelList() {
    return this.http.get<ModelListResponse>('/model');
  }

  getModelDetail(uuid: string) {
    return this.http.get<ModelDetailResponse>('/model/' + uuid);
  }

  deleteModel(uuid: string): Observable<any> {
    return this.http.delete('/model/' + uuid);
  }

  getModelSupportedDeploymentType(uuid: string) {
    return this.http.get<any>('/model/' + uuid + '/supportedDeploymentTypes');
  }

  publishModel(uuid: string, deployment_type: number, parameters_json: string, service_name: string): Observable<any> {
    return this.http.post('/model/' + uuid + '/publish', {
      "deployment_type": deployment_type,
      "parameters_json": parameters_json,
      "service_name": service_name
    });
  }
}
