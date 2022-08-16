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
import { HttpClient, HttpRequest } from '@angular/common/http';
import { Observable } from 'rxjs';
import { DataListResponse, DataElement, DataColumnResponse } from '../view/data-mg/data-mg.component';
import { DataDetailResponse, Id_Meta } from '../view/data-details/data-details.component';

@Injectable({
  providedIn: 'root'
})
export class DataService {
  constructor(private http: HttpClient) { }

  getDataList() {
    return this.http.get<DataListResponse>('/data');
  }

  uploadData(formData: FormData): Observable<any> {
    const req = new HttpRequest('POST', '/data', formData, {
      reportProgress: true
    })
    return this.http.request(req)
  }

  getDataDetail(data_id: string) {
    return this.http.get<DataDetailResponse>('/data/' + data_id);
  }

  deleteData(data_id: string): Observable<any> {
    return this.http.delete('/data/' + data_id);
  }

  downloadDataDetail(data_id: string) {
    return this.http.get('/data/' + data_id + '/file', { responseType: 'text' });
  }

  getDataColumn(data_id: string) {
    return this.http.get<DataColumnResponse>('/data/' + data_id + '/columns');
  }

  putMetaUpdate(putidmeta: any, data_id: string): Observable<Id_Meta> {
    return this.http.put<any>('/data/' + data_id + '/idmetainfo', putidmeta);
  }
}
