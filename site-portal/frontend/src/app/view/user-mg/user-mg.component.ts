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
import { Router } from '@angular/router';
import { UserMgService } from '../../service/user-mg.service';
import { MessageService } from '../../components/message/message.service'

export interface UserResponse {
  Code: number,
  Message: string,
  Data: PublicUser[],
}
export interface PermisionResponse {
  Code: number,
  Message: string,
  Data: {},
}
export interface PublicUser {
  fateboard_access: boolean,
  id: number,
  name: string,
  notebook_access: boolean,
  site_portal_access: boolean,
  uuid: string,
}
export interface Permission {
  fateboard_access: boolean,
  notebook_access: boolean,
  site_portal_access: boolean,
}

@Component({
  selector: 'app-user-mg',
  templateUrl: './user-mg.component.html',
  styleUrls: ['./user-mg.component.css']
})

export class UserMgComponent implements OnInit, OnDestroy {
  userResponse: any;
  userList: any;
  // preUserList is Deep Copy of user list for later compare
  preUserList: any;

  constructor(private userService: UserMgService, private router: Router, private msg: MessageService) {
    this.getUserList();
  }

  isPageLoading: boolean = true;
  isGetUserListFailed: boolean = false;
  // getUserList is to get current user list
  getUserList() {
    this.isPageLoading = true;
    this.isGetUserListFailed = false;
    this.userService.getUser()
      .subscribe((data: UserResponse) => {
        this.userResponse = data;
        this.userList = this.userResponse.data;
        //Deep Copy
        this.preUserList = JSON.parse(JSON.stringify(this.userList));
        this.isPageLoading = false;
      },
        err => {
          this.errorMessage = err.error.message;
          this.isGetUserListFailed = true;
          this.isPageLoading = false;

        }
      );
  }

  ngOnInit(): void {
  }
  ngOnDestroy(): void {
    this.msg.close()
  }

  isUpdateSubmit: boolean = false;
  isUpdateFailed: boolean = false;
  errorMessage: string = '';
  // updateUserPermission is to update user permission
  updateUserPermission() {
    for (let i = 0; i < this.userList.length; i++) {
      // validate if the user permission options is changed 
      var user = this.userList[i];
      var preUser = this.preUserList[i];
      var prePermission = {
        fateboard_access: preUser.fateboard_access,
        notebook_access: preUser.notebook_access,
        site_portal_access: preUser.site_portal_access
      }
      var curPermission = {
        fateboard_access: user.fateboard_access,
        notebook_access: user.notebook_access,
        site_portal_access: user.site_portal_access
      }
      //if is not changed, skip
      if (prePermission.fateboard_access === curPermission.fateboard_access && prePermission.site_portal_access === curPermission.site_portal_access
        && prePermission.notebook_access === curPermission.notebook_access) continue;
      this.isUpdateSubmit = true;
      this.isUpdateFailed = false;
      this.userService.updatePermision(curPermission, user.id)
        .subscribe(data => {
          this.msg.success('serverMessage.default200', 1000)
          this.reloadCurrentRoute();
        },
          err => {
            this.isUpdateFailed = true;
            this.errorMessage = err.error.message;
          }
        );
    }
  }

  //reloadCurrentRoute is to reload current page
  reloadCurrentRoute() {
    let currentUrl = this.router.url;
    this.router.navigateByUrl('/', { skipLocationChange: true }).then(() => {
      this.router.navigate([currentUrl]);
    });
  }

  // refresh button
  refresh() {
    this.getUserList();
  }
}
