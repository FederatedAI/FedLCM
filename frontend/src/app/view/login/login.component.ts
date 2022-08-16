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
import { FormBuilder } from '@angular/forms';
import { Router, ActivatedRoute } from '@angular/router';
import { compile } from 'src/utils/compile';
import jwt_decode from 'jwt-decode';
import { AuthService } from '../../services/common/auth.service';
import { MessageService } from 'src/app/components/message/message.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent implements OnInit {
  //form is login information form
  form: any = {
    username: null,
    password: null
  };
  constructor(private authService: AuthService, private router: Router, private $msg: MessageService) { }

  ngOnInit(): void {
    const redirect = sessionStorage.getItem('lifecycleManager-redirect')
    if (redirect) {
      this.$msg.warning('serverMessage.default401')
    }
  }

  username: string = "";
  password: string = "";
  loading = false;
  submitted = false;
  isLoggedIn = false;
  isLoginFailed = false;
  errorMessage = "Service Error!";
  decode: any;
  //onSubmit is to submit the login information
  onSubmit(): void {
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
          // redirect pre page
          const redirect = sessionStorage.getItem('lifecycleManager-redirect')
          if (redirect) {
            this.router.navigate([redirect])
            sessionStorage.removeItem('lifecycleManager-redirect')
            try {
              this.$msg.close()
            } catch (error) {

            }
          } else {
            this.router.navigate(['']);
            try {
              this.$msg.close()
            } catch (error) {

            }
          }
        },
        err => {
          if (err.error.message) this.errorMessage = err.error.message
          this.isLoginFailed = true;
        }
      );
  }
}
