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
import { AuthService } from 'src/app/service/auth.service';
import { Router } from '@angular/router';
import { compile } from '../../../utils/compile'
import { MessageService } from 'src/app/components/message/message.service'
import jwt_decode from "jwt-decode";

@Component({
  selector: 'app-log-in',
  templateUrl: './log-in.component.html',
  styleUrls: ['./log-in.component.css']
})

export class LogInComponent implements OnInit {
  // form for login info
  form: any = {
    username: null,
    password: null
  };
  username: string = "";
  password: string = "";
  loading = false;
  submitted = false;
  isLoggedIn = false;
  isLoginFailed = false;
  errorMessage = '';
  decode: any;

  constructor(private authService: AuthService, private router: Router, private $msg: MessageService) {
  }

  ngOnInit(): void {
    const redirect = sessionStorage.getItem('sitePortal-redirect')
    if (redirect) {
      this.$msg.warning('serverMessage.default401')
    }
  }
  // submitLogin is to submit the request to login
  submitLogin(): void {
    this.submitted = true;
    this.loading = true;
    const { username, password } = this.form;
    this.authService.login(username, password)
      .subscribe(
        data => {
          this.isLoginFailed = false;
          this.isLoggedIn = true;
          //decode JWT token
          var token = data.data;
          this.decode = jwt_decode(token);
          // store username, id in seesion storage
          const encryptName: string = compile(this.decode["name"]);
          sessionStorage.setItem('username', encryptName);
          sessionStorage.setItem('userId', this.decode["id"]);
          // redirect to previous page
          const redirect = sessionStorage.getItem('sitePortal-redirect')
          if (redirect) {
            this.router.navigate([redirect])
            sessionStorage.removeItem('sitePortal-redirect')
          } else {
            this.router.navigate(['']);
          }
          try {
            this.$msg.close()
          } catch (error) {

          }
        },
        err => {
          this.errorMessage = err.error.message;
          this.isLoginFailed = true;
        }
      );
  }

}
