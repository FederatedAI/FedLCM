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
import { PermisionResponse, Permission, UserResponse } from '../view/user-mg/user-mg.component';
import { Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class UserMgService {

  constructor(private http: HttpClient) { }
  
  getUser() {
    return this.http.get<UserResponse>('/user');
  }

  updatePermision(permission: Permission, id: number): Observable<PermisionResponse> {
    return this.http.put<PermisionResponse>('/user/' + String(id) + '/permission', permission);
  }

  changePassword(curPassword: string, newPassword: string, userId: any): Observable<any> {
    return this.http.put<any>('/user/' + String(userId) + '/password', {
      "cur_password": curPassword,
      "new_Password": newPassword
    });
  }
}

