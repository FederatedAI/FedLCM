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
import '@cds/core/icon/register.js';
import { checkCircleIcon, ClarityIcons, exclamationCircleIcon, userIcon, vmBugIcon, worldIcon } from '@cds/core/icon';
import { AuthService } from 'src/app/service/auth.service';
import { Router } from '@angular/router';
import { SiteService } from 'src/app/service/site.service';
import { uncompile } from '../../../utils/compile'
import { AppService } from 'src/app/app.service'
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ConfirmedValidator } from '../../../config/validators'
import { UserMgService } from 'src/app/service/user-mg.service';
import { MessageService } from 'src/app/components/message/message.service';

ClarityIcons.addIcons(userIcon, checkCircleIcon, checkCircleIcon, exclamationCircleIcon, worldIcon, vmBugIcon);

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})

export class HeaderComponent implements OnInit {
  open: boolean = false;
  returnUrl: string = "/login";
  isLoggedIn = true;
  isLoginFailed = false;
  errorMessage = '';
  isLogoutFailed = false;
  isLogOutSubmit = false;
  form: FormGroup;
  constructor(private authService: AuthService, private router: Router, private siteService: SiteService, public i18: AppService, private fb: FormBuilder, private userService: UserMgService, private msg: MessageService) {
    //change password form
    this.form = this.fb.group({
      curPassword: [''],
      newPassword: ['', [Validators.required]],
      confirmPassword: ['', [Validators.required]]
    }, {
      validator: ConfirmedValidator('newPassword', 'confirmPassword')
    })
  }

  ngOnInit(): void {
    // get stored username
    const username = sessionStorage.getItem('username')
    if (username) {
      this.username = uncompile(username)
    } else {
      // get current user
      this.siteService.getCurrentUser().subscribe(
        data => {
          if (data.data) {
            this.username = data.data
          }
        },
        err => {
          this.router.navigate(['/login'])
        }
      )
    }
  }

  username: string = '';
  condition: boolean = false;
  langFlag: boolean = false
  //Trigger user dropdown menu
  toggleDropdown() {
    this.condition = !this.condition;
    if (this.condition) this.langFlag = false;
  }

  //Trigger language dropdown menu
  languageDropdown() {
    this.langFlag = !this.langFlag;
    if (this.langFlag) this.condition = false;
  }

  //logout button
  logout(): void {
    this.isLogOutSubmit = true;
    this.authService.logout()
      .subscribe(
        data => {
          this.isLoggedIn = false;
          sessionStorage.removeItem('username');
          sessionStorage.removeItem('userId')
          this.router.navigate([this.returnUrl]);
        },
        err => {
          this.errorMessage = err.error.message;
          this.isLogoutFailed = true;
        }
      );
  }

  curPassword: any = '';
  newPassword: any = '';
  confirmPassword: any = '';
  openModal: boolean = false;
  //reset 'change password' modal when close
  resetModal() {
    this.form.reset();
    this.openModal = false;
    this.isChangePwdSubmit = false;
    this.isChangePwdFailed = false;
    this.isChangePwdSuccessed = false;
  }

  isChangePwdSubmit: boolean = false;
  isChangePwdFailed: boolean = false;
  isChangePwdSuccessed: boolean = false;
  //submit 'change password' request
  changePassword() {
    var userId = sessionStorage.getItem('userId');
    this.isChangePwdSubmit = true;
    if (!this.form.valid) {
      this.isChangePwdFailed = true;
      this.errorMessage = "Invaild input."
      return;
    }
    this.userService.changePassword(this.curPassword, this.newPassword, userId)
      .subscribe(data => {
        this.isChangePwdFailed = false;
        this.isChangePwdSuccessed = true;
        setTimeout(() => {
          this.logout();
        }, 3000)
      },
        err => {
          this.isChangePwdFailed = true;
          if (err.error.message === "crypto/bcrypt: hashedPassword is not the hash of the given password") {
            this.errorMessage = "The input of current password is incorrect. "
          } else {
            this.errorMessage = err.error.message;
          }
        });
  }
}
