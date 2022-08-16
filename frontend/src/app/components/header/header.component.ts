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
import { ActivatedRoute, Router } from '@angular/router';
import { AuthService } from 'src/app/services/common/auth.service';
import { uncompile } from 'src/utils/compile';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ConfirmedValidator } from 'src/utils/validators';
import { AppService } from 'src/app/app.service'
import { MessageService } from 'src/app/components/message/message.service';
@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.scss']
})
export class HeaderComponent implements OnInit {

  form: FormGroup;
  constructor(private authService: AuthService, private route: ActivatedRoute, public i18: AppService,
    private router: Router, private fb: FormBuilder, private $msg: MessageService) {
    //form is for the form in the modal of 'Change Password'
    this.form = this.fb.group({
      curPassword: [''],
      newPassword: ['', [Validators.required]],
      confirmPassword: ['', [Validators.required]]

    }, {
      validator: ConfirmedValidator('newPassword', 'confirmPassword')
    })
  }


  username: string = '';
  ngOnInit(): void {
    // get stored username
    const username = sessionStorage.getItem('username')
    if (username) {
      this.username = uncompile(username)
    } else {
      this.authService.getCurrentUser().subscribe(
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

  condition: boolean = false;
  open: boolean = false;
  returnUrl: string = "/login";
  isLoggedIn = true;
  isLoginFailed = false;
  errorMessage = 'Service Error!';
  isLogoutFailed = false;
  isLogOutSubmit = false;
  langFlag: boolean = false
  //toggleDropdown is for user dropdown button
  toggleDropdown() {
    this.langFlag = false
    this.condition = !this.condition;
  }

  //logout is for logout button
  logout(): void {
    this.isLogOutSubmit = true;
    this.authService.logout()
      .subscribe(
        data => {
          this.isLoggedIn = false;
          sessionStorage.removeItem('username')
          sessionStorage.removeItem('userId')
          this.router.navigate([this.returnUrl])
        },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isLogoutFailed = true
          this.$msg.error(this.errorMessage, 2000)
        }
      );
  }
  //langDropdown is for language dropdown button
  langDropdown() {
    this.langFlag = !this.langFlag;
    this.condition = false
    if (this.langFlag) this.condition = false;
  }
  curPassword: any = '';
  newPassword: any = '';
  confirmPassword: any = '';
  openModal: boolean = false;
  //resetModal is to reset the modal when close the modal of 'Change Password'
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
  //changePassword is for submitting 'Change Password'
  changePassword() {
    var userId = sessionStorage.getItem('userId');
    this.isChangePwdSubmit = true;
    if (!this.form.valid) {
      this.isChangePwdFailed = true;
      this.errorMessage = "Invaild input."
      return;
    }
    this.authService.changePassword(this.curPassword, this.newPassword, userId)
      .subscribe(data => {
        this.isChangePwdFailed = false;
        this.isChangePwdSuccessed = true;
        setTimeout(() => {
          this.logout();
        }, 3000)
      },
        err => {
          this.isChangePwdFailed = true;
          //validate if the error is because the current password provided by user is incorrect
          if (err.error.message === "crypto/bcrypt: hashedPassword is not the hash of the given password") {
            this.errorMessage = "The input of current password is incorrect. "
          } else {
            this.errorMessage = err.error.message;
          }
        });
  }


}
